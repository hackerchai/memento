package storage

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/hackerchai/memento/internal/config"
	"github.com/hackerchai/memento/pkg/xlog"
)

// LocalImageStorage implements ImageStorage using the local filesystem.
type LocalImageStorage struct {
	basePath string // Base directory for storing images
	baseURL  string // Base URL for constructing public links
	logger   *xlog.Logger
}

// NewLocalImageStorage creates a new LocalImageStorage.
func NewLocalImageStorage(cfg *config.Config, logger *xlog.Logger) (*LocalImageStorage, error) {
	// Load BasePath from config
	basePath := "assets/images"
	if cfg != nil && cfg.Storage.Local.BasePath != "" {
		basePath = cfg.Storage.Local.BasePath
	}
	if err := os.MkdirAll(basePath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create base image storage directory '%s': %w", basePath, err)
	}

	// Get BaseURL from config
	baseURL := cfg.Server.BaseURL // Assuming BaseURL is directly under Server config
	if baseURL == "" {
		logger.WarnX(context.Background()).Msg("Server BaseURL is not configured, generated image URLs might be incorrect.")
		// Handle error or default? Let's proceed but log.
	}
	// Ensure baseURL doesn't end with a slash
	baseURL = strings.TrimSuffix(baseURL, "/")

	return &LocalImageStorage{
		basePath: basePath,
		baseURL:  baseURL,
		logger:   logger.With().Str("component", "LocalImageStorage").Logger(),
	}, nil
}

// getStoragePath calculates the full local file path for an image based on userID and imageName (hash.ext).
// It extracts the first two characters of the hash for subdirectory structure.
func (s *LocalImageStorage) getStoragePath(userID uuid.UUID, imageName string) (string, error) {
	if len(imageName) < 3 { // Need at least 2 hash chars + extension dot
		return "", errors.New("invalid image name format (expected hash.ext)")
	}
	hashPart := imageName[:len(imageName)-len(filepath.Ext(imageName))] // Extract hash without extension
	if len(hashPart) < 2 {
		return "", errors.New("invalid image name format (hash part too short)")
	}

	h1 := hashPart[0:1]
	h2 := hashPart[1:2]

	// Path: /basePath/userID/h1/h2/imageName
	fullPath := filepath.Join(s.basePath, userID.String(), h1, h2, imageName)
	return fullPath, nil
}

// getPublicRelativePath calculates the relative URL path for an image.
func (s *LocalImageStorage) getPublicRelativePath(userID uuid.UUID, imageName string) (string, error) {
	if len(imageName) < 3 {
		return "", errors.New("invalid image name format (expected hash.ext)")
	}
	hashPart := imageName[:len(imageName)-len(filepath.Ext(imageName))]
	if len(hashPart) < 2 {
		return "", errors.New("invalid image name format (hash part too short)")
	}
	h1 := hashPart[0:1]
	h2 := hashPart[1:2]
	// Relative Path: /assets/images/userID/h1/h2/imageName
	// We need the prefix that the static handler will use.
	// Let's assume the static handler serves `s.basePath` at `/assets/images` prefix.
	// A better approach might be to inject the public prefix itself.
	publicPrefix := "/assets/images" // Hardcoded assumption for now!
	relativePath := filepath.ToSlash(filepath.Join(publicPrefix, userID.String(), h1, h2, imageName))
	return relativePath, nil
}

// GetPublicURL provides the public interface method to calculate the URL.
func (s *LocalImageStorage) GetPublicURL(ctx context.Context, userID uuid.UUID, imageName string) (string, error) {
	relativePath, err := s.getPublicRelativePath(userID, imageName)
	if err != nil {
		return "", err
	}
	// Construct full URL: baseURL + relativePath
	fullURL := s.baseURL + relativePath
	return fullURL, nil
}

// Store saves the image data to the local filesystem.
func (s *LocalImageStorage) Store(ctx context.Context, userID uuid.UUID, imageName string, imageData []byte) (string, error) {
	log := s.logger.With().Str("imageName", imageName).Stringer("userID", userID).Logger()

	storagePath, err := s.getStoragePath(userID, imageName)
	if err != nil {
		log.ErrorX(ctx).Err(err).Msg("Failed to calculate storage path")
		return "", err
	}

	// Ensure the specific subdirectory exists
	dir := filepath.Dir(storagePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		log.ErrorX(ctx).Err(err).Str("dir", dir).Msg("Failed creating image subdirectory")
		return "", fmt.Errorf("failed creating storage directory: %w", err)
	}

	// Write the file
	if err := os.WriteFile(storagePath, imageData, 0644); err != nil {
		log.ErrorX(ctx).Err(err).Str("path", storagePath).Msg("Failed writing image to disk")
		return "", fmt.Errorf("failed writing image file: %w", err)
	}

	log.InfoX(ctx).Str("path", storagePath).Msg("Image stored successfully")

	// Calculate public URL using the public method
	publicURL, err := s.GetPublicURL(ctx, userID, imageName)
	if err != nil {
		log.ErrorX(ctx).Err(err).Msg("Failed to calculate public URL after storing image")
		return "", fmt.Errorf("failed calculating public URL: %w", err)
	}

	return publicURL, nil
}

// Exists checks if an image file exists at the calculated path.
func (s *LocalImageStorage) Exists(ctx context.Context, userID uuid.UUID, imageName string) (bool, error) {
	storagePath, err := s.getStoragePath(userID, imageName)
	if err != nil {
		s.logger.ErrorX(ctx).Err(err).Str("imageName", imageName).Stringer("userID", userID).Msg("Failed to calculate storage path for Exists check")
		return false, err // Return error if path calculation fails
	}

	if _, err := os.Stat(storagePath); err == nil {
		// File exists
		return true, nil
	} else if errors.Is(err, os.ErrNotExist) {
		// File does not exist
		return false, nil
	} else {
		// Other error (e.g., permission issues)
		s.logger.ErrorX(ctx).Err(err).Str("path", storagePath).Msg("Error checking image existence")
		return false, err
	}
}
