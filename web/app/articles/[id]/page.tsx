"use client";

import { useEffect, useState, useCallback } from "react";
import { useParams } from "next/navigation";
import { toast } from "@/hooks/use-toast";
import { articleAPI } from "@/lib/api-client";
import ArticleDetail from "@/components/articles/article-detail";
import { DetailedArticle } from "@/types/article";

export default function ArticleDetailPage() {
  // Get article id from route params
  const params = useParams();
  const id = Array.isArray(params.id) ? params.id[0] : params.id;

  // State for article data and loading
  const [article, setArticle] = useState<DetailedArticle | null>(null);
  const [isLoading, setIsLoading] = useState(true);

  // Fetch article detail
  const fetchArticle = useCallback(async () => {
    setIsLoading(true);
    try {
      const response = await articleAPI.getArticle(id);
      
      if (response?.data) {
        // Map API response to DetailedArticle
        const apiData = response.data;
        
        // Transform tags if needed
        let processedTags = apiData.tags || [];
        
        setArticle({
          id: apiData.id,
          url: apiData.url || '',
          title: apiData.title || 'Untitled Article',
          description: apiData.description || '',
          content: apiData.content || '',
          html: apiData.html || '',
          plain_text: apiData.plain_text || '',
          ogImage: apiData.og_image_url || apiData.ogImage || '',
          createdAt: apiData.created_at || apiData.createdAt || new Date().toISOString(),
          updatedAt: apiData.updated_at || apiData.updatedAt || '',
          category: typeof apiData.category === 'object' ? apiData.category.name : apiData.category || '',
          category_id: apiData.category_id || '',
          tags: processedTags,
          isRead: apiData.is_read || apiData.isRead || false,
          isStarred: apiData.is_starred || apiData.isStarred || false,
          author: apiData.author || '',
          is_offline: apiData.is_offline || false,
          status: apiData.status || 0
        });
      } else {
        throw new Error('Article not found or invalid response format');
      }
    } catch (error) {
      console.error("Failed to fetch article:", error);
      toast({
        title: "Error",
        description: "Failed to load article. Please try again later.",
        variant: "destructive",
      });
      setArticle(null);
    } finally {
      setIsLoading(false);
    }
  }, [id]);

  // Refresh article data handler (modified to fetch, not rescrape)
  const handleRefresh = useCallback(async () => {
    setIsLoading(true); // Indicate loading during refresh
    try {
      // Removed: await articleAPI.rescrapeArticle(id);
      toast({
        title: "Refreshing...",
        description: "Fetching latest article data.",
        variant: "default",
      });

      // Directly fetch the article data
      await fetchArticle(); 

      // Removed: setTimeout(() => fetchArticle(), 2000);

      // Update toast to success after fetch completes (optional, fetchArticle shows its own toasts)
      // toast({ title: "Success", description: "Article data refreshed." }); 

    } catch (error) {
      // fetchArticle already handles its own errors and toasts
      // We might not need duplicate error handling here unless handleRefresh itself fails
      console.error("Error occurred during handleRefresh process:", error);
      // toast({
      //   title: "Error",
      //   description: "Failed to refresh article data.",
      //   variant: "destructive",
      // });
    } finally {
       // setIsLoading(false); // fetchArticle handles its own loading state
    }
  }, [id, fetchArticle]); // Keep id in dependency if fetchArticle potentially uses it differently than the outer scope `id`

  // Initial fetch on mount
  useEffect(() => {
    fetchArticle();
  }, [fetchArticle]);

  return <ArticleDetail article={article} isLoading={isLoading} onRefetch={handleRefresh} />;
}