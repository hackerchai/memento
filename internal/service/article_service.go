package service

import (
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/go-shiori/go-readability"
	"github.com/gocolly/colly/v2"
	"github.com/google/uuid"
	"github.com/hackerchai/memento/internal/config"
	"github.com/hackerchai/memento/internal/entity"
	"github.com/hackerchai/memento/internal/repository"
	"github.com/hackerchai/memento/internal/sse"
	"github.com/hackerchai/memento/internal/storage"
	"github.com/hackerchai/memento/pkg/xlog"
	"github.com/uptrace/bun"
)

const (
	defaultCategoryName  = "Default"
	imageCacheDir        = "assets/images"
	defaultUserAgent     = "Memento Bot/1.0 (+https://github.com/hackerchai/memento)" // Be a good citizen
	fetchTimeout         = 30 * time.Second
	maxImageDownloadSize = 10 * 1024 * 1024 // 10 MB limit per image
)

// ArticleService provides article saving and processing logic.
type ArticleService struct {
	db            *bun.DB
	articleRepo   *repository.ArticleRepository
	categoryRepo  *repository.CategoryRepository
	tagRepo       *repository.TagRepository
	appConfigRepo *repository.AppConfigRepository
	imageStorage  storage.ImageStorage // Injected image storage
	sseBroker     *sse.Broker          // Add SSE Broker field
	logger        *xlog.Logger
	config        *config.Config // Keep for potential non-user specific static config
	// httpClient is removed, handled by imageStorage or within methods needing direct fetch
}

// NewArticleService creates a new ArticleService.
func NewArticleService(
	db *bun.DB,
	articleRepo *repository.ArticleRepository,
	categoryRepo *repository.CategoryRepository,
	tagRepo *repository.TagRepository,
	appConfigRepo *repository.AppConfigRepository,
	imageStorage storage.ImageStorage, // Added imageStorage parameter
	sseBroker *sse.Broker, // Add sseBroker parameter
	logger *xlog.Logger,
	config *config.Config,
) *ArticleService {
	return &ArticleService{
		db:            db,
		articleRepo:   articleRepo,
		categoryRepo:  categoryRepo,
		tagRepo:       tagRepo,
		appConfigRepo: appConfigRepo,
		imageStorage:  imageStorage, // Store the injected storage
		sseBroker:     sseBroker,    // Store the injected broker
		logger:        logger.With().Str("service", "ArticleService").Logger(),
		config:        config,
		// httpClient removed
	}
}

// SaveArticleFromURLInput defines the input for saving an article.
type SaveArticleFromURLInput struct {
	UserID       uuid.UUID
	URL          string
	CategoryName string   // Optional, defaults to "Default"
	TagNames     []string // Optional
}

// SaveArticleFromURL initiates the process of saving an article from a URL.
// It first checks for duplicates and creates a placeholder article in the DB.
// Then, it launches a background goroutine to fetch, parse, and update the article.
// It returns the initially created article (with ID and pending status) immediately.
func (s *ArticleService) SaveArticleFromURL(ctx context.Context, input *SaveArticleFromURLInput) (*entity.Article, error) {
	s.logger.InfoX(ctx).Str("url", input.URL).Stringer("userID", input.UserID).Msg("Received request to save article")

	// 1. Validate URL
	parsedURL, err := url.ParseRequestURI(input.URL)
	if err != nil || (parsedURL.Scheme != "http" && parsedURL.Scheme != "https") {
		s.logger.WarnX(ctx).Err(err).Str("url", input.URL).Msg("Invalid URL provided")
		return nil, fmt.Errorf("invalid URL: %w", err)
	}
	normalizedURL := parsedURL.String() // Use normalized URL

	// 2. Check if article already exists for this user
	_, err = s.articleRepo.FindByURL(ctx, normalizedURL, input.UserID)
	if err == nil {
		s.logger.WarnX(ctx).Str("url", normalizedURL).Stringer("userID", input.UserID).Msg("Article with this URL already exists for the user")
		return nil, errors.New("article already exists")
	}
	if !errors.Is(err, sql.ErrNoRows) {
		s.logger.ErrorX(ctx).Err(err).Str("url", normalizedURL).Stringer("userID", input.UserID).Msg("Failed to check for existing article")
		return nil, fmt.Errorf("failed checking existing article: %w", err)
	}

	// 3. Find or Create Category
	categoryName := input.CategoryName
	if categoryName == "" {
		categoryName = defaultCategoryName
	}
	category, err := s.categoryRepo.FindOrCreateByName(ctx, categoryName, input.UserID)
	if err != nil {
		s.logger.ErrorX(ctx).Err(err).Str("categoryName", categoryName).Stringer("userID", input.UserID).Msg("Failed to find or create category")
		return nil, fmt.Errorf("failed to process category: %w", err)
	}

	// 4. Prepare initial article entity (placeholder)
	article := &entity.Article{
		UserID:     input.UserID,
		CategoryID: &category.ID,
		URL:        normalizedURL,
		Title:      "Processing...", // Placeholder title
		Status:     entity.StatusPending,
		// Other fields will be populated by the background task
	}

	// 5. Create the placeholder article in the database (without tags for now)
	// Note: We pass nil for tags here. Tags will be associated in the background update.
	if err := s.articleRepo.Create(ctx, article, nil); err != nil {
		// Handle potential race condition where another request created it *just* now
		if strings.Contains(err.Error(), "article already exists") { // Check based on repo error
			s.logger.WarnX(ctx).Str("url", normalizedURL).Stringer("userID", input.UserID).Msg("Article created by concurrent request after check")
			return nil, errors.New("article already exists")
		}
		s.logger.ErrorX(ctx).Err(err).Str("url", normalizedURL).Stringer("userID", input.UserID).Msg("Failed to create placeholder article")
		return nil, fmt.Errorf("failed to create article placeholder: %w", err)
	}

	s.logger.InfoX(ctx).Stringer("articleID", article.ID).Msg("Placeholder article created, launching background processing")

	// 6. Launch background goroutine for processing
	go s.processArticleInBackground(article.ID, input.UserID, normalizedURL, input.TagNames)

	// 7. Return the initial article object (ID is now populated)
	return article, nil
}

// processArticleInBackground fetches, parses, and updates the article content.
func (s *ArticleService) processArticleInBackground(articleID, userID uuid.UUID, articleURL string, tagNames []string) {
	// Create a new background context that won't be cancelled by the original request ending
	bgCtx := context.Background()
	log := s.logger.With().Stringer("articleID", articleID).Stringer("userID", userID).Logger()
	log.InfoX(bgCtx).Msg("Starting background processing")

	var fetchedArticle *readability.Article
	var processingErr error
	var finalStatus entity.ArticleStatus = entity.StatusFailed // Default to failed
	var errorMessage string
	var finalTitle string      // Store title for notification
	var finalOgImageURL string // Store final OG image URL

	defer func() {
		// Update status in DB regardless of success/failure before notifying
		s.updateArticleStatus(bgCtx, articleID, userID, finalStatus)

		// Send SSE Notification after DB update
		var eventData interface{}
		var eventType sse.EventType

		if finalStatus == entity.StatusCompleted {
			eventType = sse.EventTypeArticleProcessingComplete
			eventData = sse.ArticleProcessedData{
				EventData: sse.EventData{ArticleID: articleID},
				Status:    finalStatus,
				Title:     finalTitle, // Send the fetched title
			}
			log.InfoX(bgCtx).Msg("Article processing completed successfully, preparing success notification")
		} else {
			eventType = sse.EventTypeArticleProcessingFailed
			eventData = sse.ArticleProcessedData{
				EventData: sse.EventData{ArticleID: articleID},
				Status:    finalStatus,
				Error:     errorMessage,
			}
			log.WarnX(bgCtx).Str("error", errorMessage).Msg("Article processing failed, preparing failure notification")
		}

		if s.sseBroker != nil {
			messageBytes, err := sse.FormatSSEMessage(eventType, eventData)
			if err != nil {
				log.ErrorX(bgCtx).Err(err).Msg("Failed to format SSE message")
			} else {
				s.sseBroker.NotifyUser(userID, messageBytes)
				log.InfoX(bgCtx).Str("event_type", string(eventType)).Msg("Sent SSE notification to user")
			}
		} else {
			log.WarnX(bgCtx).Msg("SSE Broker is nil, cannot send notification")
		}

		// Trigger LLM tasks after main processing is complete (success or fail? Only success for now)
		if finalStatus == entity.StatusCompleted {
			// TODO: Implement actual LLM calls, potentially checking user preferences first
			// go s.SummarizeArticle(bgCtx, articleID, userID)
			// go s.AutoTagArticle(bgCtx, articleID, userID)
		}
	}()

	// 1. Fetch and Extract Content
	fetchedArticle, _, processingErr = s.fetchAndExtractContent(bgCtx, articleURL)
	if processingErr != nil {
		errorMessage = fmt.Sprintf("Failed to fetch/extract content: %v", processingErr)
		log.ErrorX(bgCtx).Err(processingErr).Msg(errorMessage)
		// finalStatus remains StatusFailed
		return // Exit early
	}
	finalTitle = fetchedArticle.Title // Store title for success notification
	log.InfoX(bgCtx).Str("extracted_title", finalTitle).Msg("Content extracted successfully")

	// Fetch user-specific app config
	appConfig, err := s.appConfigRepo.GetByUserID(bgCtx, userID)
	if err != nil {
		// Log error but proceed with default behavior (no image caching)
		log.ErrorX(bgCtx).Err(err).Msg("Failed to retrieve app config for user, disabling image caching for this run")
		appConfig = &entity.AppConfig{UserID: userID} // Use default (false)
	}

	// Initialize final OG image URL with the original one
	finalOgImageURL = fetchedArticle.Image

	// 2. Process and Cache Images (if enabled in user config)
	processedHTML := fetchedArticle.Content
	if appConfig.ScrapeImgOffline {
		var imgErr error
		processedHTML, imgErr = s.processAndCacheImages(bgCtx, fetchedArticle.Content, articleID, userID, articleURL)
		if imgErr != nil {
			log.WarnX(bgCtx).Err(imgErr).Msg("Error processing images, continuing with original HTML content")
			// Don't set finalStatus to Failed just for image errors
			processedHTML = fetchedArticle.Content
		}

		// Also process OG Image if offline mode is enabled and OG image exists
		if fetchedArticle.Image != "" {
			log.InfoX(bgCtx).Str("og_image_url", fetchedArticle.Image).Msg("Processing OG image for offline caching")
			baseArticleURLParsed, _ := url.Parse(articleURL)
			ogImageURL, err := baseArticleURLParsed.Parse(fetchedArticle.Image)
			if err != nil {
				log.WarnX(bgCtx).Str("original_og_image", fetchedArticle.Image).Err(err).Msg("Failed to resolve OG image URL, using original")
			} else {
				absOgImageURL := ogImageURL.String()
				if !strings.HasPrefix(absOgImageURL, "data:") {
					// Use a temporary client for downloading
					tempHttpClient := &http.Client{Timeout: fetchTimeout}
					ogImageData, ogContentType, downloadErr := s.downloadImageData(bgCtx, tempHttpClient, absOgImageURL)
					if downloadErr != nil {
						log.WarnX(bgCtx).Str("og_image_url", absOgImageURL).Err(downloadErr).Msg("Failed to download OG image data, using original")
					} else {
						ogImageName, nameErr := s.generateImageName(absOgImageURL, ogContentType, ogImageData)
						if nameErr != nil {
							log.WarnX(bgCtx).Str("og_image_url", absOgImageURL).Err(nameErr).Msg("Failed to generate OG image name, using original")
						} else {
							publicURL, storeErr := s.imageStorage.Store(bgCtx, userID, ogImageName, ogImageData)
							if storeErr != nil {
								log.WarnX(bgCtx).Str("og_image_name", ogImageName).Err(storeErr).Msg("Failed to store OG image, using original")
							} else {
								log.InfoX(bgCtx).Str("original_url", absOgImageURL).Str("cached_url", publicURL).Msg("OG image cached successfully")
								finalOgImageURL = publicURL // Update to the cached URL
							}
						}
					}
				}
			}
		}
	} else {
		log.InfoX(bgCtx).Msg("Image caching is disabled for this user")
	}

	// 3. Find or Create Tags
	tagsMap, err := s.tagRepo.FindOrCreateByName(bgCtx, tagNames, userID)
	if err != nil {
		// Log error but don't fail the process
		log.ErrorX(bgCtx).Err(err).Strs("tagNames", tagNames).Msg("Failed to find or create tags, proceeding without them")
		tagsMap = make(map[string]*entity.Tag) // Ensure it's not nil
	}
	finalTags := make([]*entity.Tag, 0, len(tagsMap))
	for _, tag := range tagsMap {
		finalTags = append(finalTags, tag)
	}

	// 4. Prepare final article data for update
	articleToUpdate := &entity.Article{
		ID:             articleID,
		UserID:         userID,
		Title:          finalTitle, // Use fetched title
		Html:           &processedHTML,
		Author:         &fetchedArticle.Byline,
		Description:    &fetchedArticle.Excerpt,
		PlainText:      &fetchedArticle.TextContent,
		LLMDescription: nil,
		OgImageURL:     &finalOgImageURL,
		IsOffline:      appConfig.ScrapeImgOffline,
		Status:         entity.StatusCompleted, // Tentative status for update
	}

	// 5. Update Article in DB with content and tags
	if err := s.articleRepo.Update(bgCtx, articleToUpdate, finalTags); err != nil {
		errorMessage = fmt.Sprintf("Failed to update article in DB: %v", err)
		log.ErrorX(bgCtx).Err(err).Msg(errorMessage)
		// finalStatus remains StatusFailed
		return // Exit early
	}

	// If we reach here, processing was successful
	finalStatus = entity.StatusCompleted
	// Defer function will handle logging, status update, and notification
}

// fetchAndExtractContent uses colly and go-readability to get article data.
func (s *ArticleService) fetchAndExtractContent(ctx context.Context, articleURL string) (*readability.Article, string, error) {
	log := s.logger.With().Str("url", articleURL).Logger()
	var rawHTML string
	var extractedArticle *readability.Article
	var firstError error
	var errOnce sync.Once

	c := colly.NewCollector(
		colly.UserAgent(defaultUserAgent),
		// Add other options like AllowedDomains, MaxDepth(0), Async?, etc. if needed
		// colly.MaxDepth(0), // Don't follow links from the article page itself
		// colly.Async(true), // Consider if needed, adds complexity
	)

	c.SetRequestTimeout(fetchTimeout)

	// Handle HTML response
	c.OnResponse(func(r *colly.Response) {
		rawHTML = string(r.Body)
		// Use go-readability
		parsed, err := readability.FromReader(bytes.NewReader(r.Body), r.Request.URL)
		if err != nil {
			log.ErrorX(ctx).Err(err).Msg("Failed to extract readable content")
			errOnce.Do(func() { firstError = fmt.Errorf("readability failed: %w", err) })
			return
		}
		extractedArticle = &parsed
	})

	// Handle errors during request
	c.OnError(func(r *colly.Response, err error) {
		log.ErrorX(ctx).Err(err).Int("status_code", r.StatusCode).Msg("Colly request failed")
		errOnce.Do(func() { firstError = fmt.Errorf("fetch failed (status %d): %w", r.StatusCode, err) })
	})

	// Start the scrape
	if err := c.Visit(articleURL); err != nil {
		// This initial visit error might be redundant if OnError captures it, but good as fallback
		errOnce.Do(func() { firstError = fmt.Errorf("colly visit initiation failed: %w", err) })
	}

	c.Wait() // Wait for async operations if any

	if firstError != nil {
		return nil, rawHTML, firstError
	}
	if extractedArticle == nil {
		return nil, rawHTML, errors.New("failed to extract article content (unknown reason)")
	}

	return extractedArticle, rawHTML, nil
}

// processAndCacheImages finds images in HTML, downloads/caches them if enabled, and updates src attributes.
func (s *ArticleService) processAndCacheImages(ctx context.Context, htmlContent string, articleID, userID uuid.UUID, baseArticleURL string) (string, error) {
	log := s.logger.With().Stringer("articleID", articleID).Stringer("userID", userID).Logger()
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	if err != nil {
		return htmlContent, fmt.Errorf("failed to parse HTML for image processing: %w", err)
	}

	baseURL, _ := url.Parse(baseArticleURL)
	var processingErrors []string
	var wg sync.WaitGroup
	// Create a temporary HTTP client *here* specifically for downloading images in this function.
	// Or, the ImageStorage implementation could handle the download itself.
	// For now, keeping download logic here temporarily before calling Store.
	tempHttpClient := &http.Client{Timeout: fetchTimeout}
	imgChan := make(chan struct{}, 5) // Limit concurrent image downloads

	doc.Find("img").Each(func(i int, img *goquery.Selection) {
		src, exists := img.Attr("src")
		if !exists || src == "" {
			return
		}
		imgURL, err := baseURL.Parse(src)
		if err != nil {
			log.WarnX(ctx).Str("original_src", src).Err(err).Msg("Failed to resolve image URL")
			processingErrors = append(processingErrors, fmt.Sprintf("resolve %s: %v", src, err))
			return
		}
		absImgURL := imgURL.String()
		if strings.HasPrefix(absImgURL, "data:") {
			return
		}

		wg.Add(1)
		go func(imgNode *goquery.Selection, targetURL string) {
			defer wg.Done()
			imgChan <- struct{}{}
			defer func() { <-imgChan }()

			// 1. Download the image data
			imageData, contentType, err := s.downloadImageData(ctx, tempHttpClient, targetURL)
			if err != nil {
				log.WarnX(ctx).Str("image_url", targetURL).Err(err).Msg("Failed to download image data")
				processingErrors = append(processingErrors, fmt.Sprintf("download %s: %v", targetURL, err))
				return // Skip storing if download fails
			}

			// 2. Determine filename (hash + extension)
			imageName, err := s.generateImageName(targetURL, contentType, imageData)
			if err != nil {
				log.WarnX(ctx).Str("image_url", targetURL).Err(err).Msg("Failed to generate image name")
				processingErrors = append(processingErrors, fmt.Sprintf("name %s: %v", targetURL, err))
				return
			}

			// 3. Check if image already exists (optional optimization)
			exists, err := s.imageStorage.Exists(ctx, userID, imageName)
			if err != nil {
				log.WarnX(ctx).Str("image_name", imageName).Err(err).Msg("Failed to check image existence")
				// Proceed to store anyway, Store should handle potential overwrites if needed
			}

			var publicURL string
			if exists {
				// If it exists, we still need the public URL
				publicURL, err = s.imageStorage.GetPublicURL(ctx, userID, imageName)
				if err != nil {
					log.WarnX(ctx).Str("image_name", imageName).Err(err).Msg("Failed to get public URL for existing image")
					processingErrors = append(processingErrors, fmt.Sprintf("get_url %s: %v", imageName, err))
					return
				}
				log.DebugX(ctx).Str("image_name", imageName).Msg("Image already cached")
			} else {
				// 4. Store the image using the injected storage service
				publicURL, err = s.imageStorage.Store(ctx, userID, imageName, imageData)
				if err != nil {
					log.WarnX(ctx).Str("image_name", imageName).Err(err).Msg("Failed to store image")
					processingErrors = append(processingErrors, fmt.Sprintf("store %s: %v", imageName, err))
					return // Skip updating src if store fails
				}
				log.DebugX(ctx).Str("original_url", targetURL).Str("local_path", publicURL).Msg("Image cached and stored")
			}

			// 5. Update the src attribute in the goquery document
			imgNode.SetAttr("src", publicURL)

		}(img, absImgURL)
	})

	wg.Wait()

	newHTML, err := doc.Html()
	if err != nil {
		return htmlContent, fmt.Errorf("failed to serialize HTML after image processing: %w", err)
	}
	if len(processingErrors) > 0 {
		return newHTML, fmt.Errorf("encountered errors processing images: %s", strings.Join(processingErrors, "; "))
	}
	return newHTML, nil
}

// downloadImageData downloads image data from a URL.
// This helper method is kept temporarily within the service.
func (s *ArticleService) downloadImageData(ctx context.Context, client *http.Client, imgURL string) ([]byte, string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", imgURL, nil)
	if err != nil {
		return nil, "", fmt.Errorf("failed creating image request: %w", err)
	}
	req.Header.Set("User-Agent", defaultUserAgent)

	resp, err := client.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("failed fetching image: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("failed fetching image: status %d", resp.StatusCode)
	}

	contentType := resp.Header.Get("Content-Type")

	limitedReader := &io.LimitedReader{R: resp.Body, N: maxImageDownloadSize}
	imgData, err := io.ReadAll(limitedReader)
	if err != nil {
		return nil, contentType, fmt.Errorf("failed reading image data: %w", err)
	}
	if limitedReader.N == 0 {
		return nil, contentType, fmt.Errorf("image exceeded max size limit (%d bytes)", maxImageDownloadSize)
	}

	return imgData, contentType, nil
}

// generateImageName creates a unique filename based on content hash and extension.
// This helper method is kept temporarily within the service.
func (s *ArticleService) generateImageName(imgURL string, contentType string, imgData []byte) (string, error) {
	ext := ".jpg" // Default extension
	mimeExts, _ := mime.ExtensionsByType(contentType)
	if len(mimeExts) > 0 {
		ext = mimeExts[0]
	} else {
		urlExt := filepath.Ext(imgURL)
		if urlExt != "" {
			ext = urlExt
		}
	}

	hash := sha256.Sum256(imgData)
	hashString := hex.EncodeToString(hash[:])
	imageName := hashString + ext
	return imageName, nil
}

// updateArticleStatus updates only the status of an article.
func (s *ArticleService) updateArticleStatus(ctx context.Context, articleID, userID uuid.UUID, status entity.ArticleStatus) {
	log := s.logger.With().Stringer("articleID", articleID).Logger()
	// Only log failure updates here, success is logged in the main process
	if status == entity.StatusFailed {
		log.WarnX(ctx).Int("final_status", int(status)).Msg("Updating article status to Failed in DB")
	} else {
		log.InfoX(ctx).Int("final_status", int(status)).Msg("Updating article status to Completed in DB")
	}

	_, err := s.db.NewUpdate().
		Model((*entity.Article)(nil)).
		Set("status = ?, updated_at = NOW()", status). // Also update updated_at
		Where("id = ? AND user_id = ?", articleID, userID).
		Exec(ctx)
	if err != nil {
		log.ErrorX(ctx).Err(err).Int("target_status", int(status)).Msg("Failed to update article status in DB")
	}
}

// --- Placeholder Methods for Future LLM Features ---

// SummarizeArticle (Placeholder) - Intended to be called asynchronously.
func (s *ArticleService) SummarizeArticle(ctx context.Context, articleID, userID uuid.UUID) error {
	log := s.logger.With().Stringer("articleID", articleID).Stringer("userID", userID).Logger()
	log.InfoX(ctx).Msg("LLM Summarization triggered (not implemented)")
	// 1. Fetch article plain text from DB
	// 2. Call LLM service (potentially streaming chunks)
	// 3. If streaming, send sse.EventTypeLLMChunk events via s.sseBroker.NotifyUser
	// 4. Update LLMDescription field in the article
	// 5. Send sse.EventTypeArticleLLMSummaryComplete event via s.sseBroker.NotifyUser
	var summaryData = sse.LLMSummaryData{
		EventData: sse.EventData{ArticleID: articleID},
		Summary:   "This is a placeholder summary.",
	}
	messageBytes, err := sse.FormatSSEMessage(sse.EventTypeArticleLLMSummaryComplete, summaryData)
	if err == nil && s.sseBroker != nil {
		s.sseBroker.NotifyUser(userID, messageBytes)
	} else if err != nil {
		log.ErrorX(ctx).Err(err).Msg("Failed to format LLM summary SSE message")
	}
	return errors.New("SummarizeArticle not implemented")
}

// AutoTagArticle (Placeholder) - Intended to be called asynchronously.
func (s *ArticleService) AutoTagArticle(ctx context.Context, articleID, userID uuid.UUID) error {
	log := s.logger.With().Stringer("articleID", articleID).Stringer("userID", userID).Logger()
	log.InfoX(ctx).Msg("LLM Auto-tagging triggered (not implemented)")
	// 1. Fetch article plain text/title/description from DB
	// 2. Call LLM service to get suggested tags
	// 3. FindOrCreateByName the suggested tags
	// 4. Associate the new tags with the article (replace or append?)
	// 5. Send sse.EventTypeArticleLLMTagsComplete event via s.sseBroker.NotifyUser
	var tagsData = sse.LLMTagsData{
		EventData: sse.EventData{ArticleID: articleID},
		Tags:      []string{"llm", "auto-tag", "placeholder"},
	}
	messageBytes, err := sse.FormatSSEMessage(sse.EventTypeArticleLLMTagsComplete, tagsData)
	if err == nil && s.sseBroker != nil {
		s.sseBroker.NotifyUser(userID, messageBytes)
	} else if err != nil {
		log.ErrorX(ctx).Err(err).Msg("Failed to format LLM tags SSE message")
	}
	return errors.New("AutoTagArticle not implemented")
}
