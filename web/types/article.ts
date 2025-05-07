export interface Article {
  id: string;
  url: string;
  title: string;
  description?: string;
  content?: string;
  ogImage?: string;
  createdAt: string;
  updatedAt?: string;
  tags: string[];
  category?: string;
  readingTime?: number;
  isRead?: boolean;
  isStarred?: boolean;
}

// Detailed article interface for article detail page
export interface DetailedArticle extends Article {
  html?: string;          // HTML content from API
  plain_text?: string;    // Plain text version of the content
  category_id?: string;   // Category ID if available
  author?: string;        // Author of the article
  is_offline?: boolean;   // Whether article is available offline
  status?: number;        // Article processing status
}