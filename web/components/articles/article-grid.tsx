"use client";

import { useState, useEffect } from "react";
import { Article } from "@/types/article";
import { ArticleCard, ArticleCardSkeleton } from "@/components/articles/article-card";
import { toast } from "@/hooks/use-toast";
import { BookmarkIcon } from "lucide-react";
import { articleAPI } from "@/lib/api-client";
import { Button } from "@/components/ui/button";

interface ArticleGridProps {
  initialFilters?: {
    isRead?: boolean;
    isStarred?: boolean;
  };
}

export function ArticleGrid({ initialFilters = {} }: ArticleGridProps) {
  const [articles, setArticles] = useState<Article[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [isLoadingMore, setIsLoadingMore] = useState(false);
  const [page, setPage] = useState(1);
  const [hasMore, setHasMore] = useState(true);
  const perPage = 12;

  // Fetch articles with the given page and filters
  const fetchArticles = async (currentPage: number, append = false) => {
    try {
      // Set loading state based on whether we're appending or replacing
      if (append) {
        setIsLoadingMore(true);
      } else {
        setIsLoading(true);
      }
      
      // Create API filter parameters
      const filterOptions: any = {};
      
      // Map React prop names to API parameter names
      if (initialFilters.isRead !== undefined) {
        filterOptions.is_read = initialFilters.isRead;
      }
      
      if (initialFilters.isStarred !== undefined) {
        filterOptions.is_starred = initialFilters.isStarred;
      }
      
      
      // Use API client to fetch articles
      const response = await articleAPI.listArticles(currentPage, perPage, filterOptions);
      
      // Process the response
      let articlesArray = [];
      
      if (response?.data?.data && Array.isArray(response.data.data)) {
        articlesArray = response.data.data;
      } else if (response?.data?.items && Array.isArray(response.data.items)) {
        articlesArray = response.data.items;
      } else if (response?.data && Array.isArray(response.data)) {
        articlesArray = response.data;
      }
      
      if (articlesArray.length > 0) {
        // Normalize article data structure
        const fetchedArticles = articlesArray.map((item: any) => ({
          id: item.id,
          url: item.url,
          title: item.title || "Untitled Article",
          description: item.description || item.author || "",
          content: item.content,
          ogImage: item.og_image_url || item.og_image || item.ogImage,
          createdAt: item.created_at || item.createdAt,
          tags: item.tags || [],
          category: item.category || item.category_name,
          isRead: item.is_read || false,
          isStarred: item.is_starred || false
        }));
        
        // Add to existing articles or replace them
        if (append) {
          setArticles(prev => [...prev, ...fetchedArticles]);
        } else {
          setArticles(fetchedArticles);
        }
        
        // Determine if there are more articles to load
        const total = response?.data?.total || 0;
        const currentCount = fetchedArticles.length;
        
        if (total) {
          setHasMore(currentPage * perPage < total);
        } else {
          setHasMore(currentCount >= perPage);
        }
      } else {
        if (!append) {
          setArticles([]);
        }
        setHasMore(false);
      }
    } catch (error) {
      console.error("Failed to fetch articles:", error);
      toast({
        title: "Error",
        description: "Failed to load articles. Please try again later.",
        variant: "destructive",
      });
    } finally {
      // Reset loading states
      if (append) {
        setIsLoadingMore(false);
      } else {
        setIsLoading(false);
      }
    }
  };

  // On mount, fetch articles with initial filters
  useEffect(() => {
    fetchArticles(1, false);
    
    // Clean up function
    return () => {
    };
  }, []); // Empty dependency array, we only fetch on mount

  // Handle user requesting more articles
  const handleLoadMore = () => {
    const nextPage = page + 1;
    setPage(nextPage);
    fetchArticles(nextPage, true);
  };

  // Handle article deletion
  const handleDelete = async (id: string) => {
    try {
      await articleAPI.deleteArticle(id);
      setArticles(articles.filter(article => article.id !== id));
      
      toast({
        title: "Article deleted",
        description: "The article has been successfully removed.",
      });
    } catch (error) {
      console.error("Failed to delete article:", error);
      toast({
        title: "Error",
        description: "Failed to delete article. Please try again later.",
        variant: "destructive",
      });
    }
  };

  // Show loading state
  if (isLoading) {
    return (
      <div className="grid grid-cols-1 gap-6 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
        {Array.from({ length: 8 }).map((_, index) => (
          <ArticleCardSkeleton key={index} />
        ))}
      </div>
    );
  }

  // Show empty state if no articles
  if (articles.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center py-12">
        <BookmarkIcon className="h-16 w-16 text-muted-foreground/50" />
        <h3 className="mt-4 text-xl font-semibold">No articles found</h3>
        <p className="mt-2 text-center text-muted-foreground">
          Try changing your filter settings or add new articles
        </p>
      </div>
    );
  }

  // Show articles
  return (
    <div className="space-y-6">
      <div className="grid grid-cols-1 gap-6 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
        {articles.map((article) => (
          <ArticleCard 
            key={article.id} 
            article={article} 
            onDelete={handleDelete} 
          />
        ))}
      </div>
      
      {hasMore && (
        <div className="flex justify-center pt-4">
          <Button 
            variant="outline" 
            onClick={handleLoadMore} 
            disabled={isLoadingMore}
          >
            {isLoadingMore ? "Loading..." : "Load more articles"}
          </Button>
        </div>
      )}
    </div>
  );
}