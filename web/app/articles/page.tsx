"use client";

import Link from "next/link";
import { Button } from "@/components/ui/button";
import { UrlForm } from "@/components/articles/url-form";
import { ArticleGrid } from "@/components/articles/article-grid";
import { ArticleFilters } from "@/components/articles/article-filters";
import { PlusIcon } from "lucide-react";
import { useState, useEffect, useRef, useCallback } from "react";
import { toast } from "@/hooks/use-toast";
import { tagAPI, categoryAPI } from "@/lib/api-client";
import { useSearchParams } from "next/navigation";

export default function ArticlesPage() {
  // State for available tags and categories
  const [availableTags, setAvailableTags] = useState<string[]>([]);
  const [availableCategories, setAvailableCategories] = useState<string[]>([]);
  const [isLoading, setIsLoading] = useState(true);

  // Filters state: undefined means no filter is applied
  const [filters, setFilters] = useState<{
    isRead?: boolean;
    isStarred?: boolean;
  }>({});

  // Ref to track if filters have been changed by user interaction
  const filtersUpdatedByUser = useRef(false);

  // Stable filter change handler
  const handleFilterChange = useCallback((newFilters: {
    isRead?: boolean;
    isStarred?: boolean;
  }) => {
    // Mark that this filter change was triggered by a user interaction
    filtersUpdatedByUser.current = true;
    setFilters(newFilters);
  }, []);

  // Fetch tags and categories metadata on mount
  useEffect(() => {
    const fetchMetadata = async () => {
      try {
        const [tagsResponse, categoriesResponse] = await Promise.all([
          tagAPI.listTags(),
          categoryAPI.listCategories()
        ]);

        // Extract tag and category names from responses
        let tagNames: string[] = [];
        let categoryNames: string[] = [];

        // Extract tags
        let tagsArray = [];
        if (tagsResponse?.data?.data && Array.isArray(tagsResponse.data.data)) {
          tagsArray = tagsResponse.data.data;
        } else if (tagsResponse?.data?.items && Array.isArray(tagsResponse.data.items)) {
          tagsArray = tagsResponse.data.items;
        } else if (tagsResponse?.data && Array.isArray(tagsResponse.data)) {
          tagsArray = tagsResponse.data;
        }
        tagNames = tagsArray.map((tag: any) => tag.name || tag);

        // Extract categories
        let categoriesArray = [];
        if (categoriesResponse?.data?.data && Array.isArray(categoriesResponse.data.data)) {
          categoriesArray = categoriesResponse.data.data;
        } else if (categoriesResponse?.data?.items && Array.isArray(categoriesResponse.data.items)) {
          categoriesArray = categoriesResponse.data.items;
        } else if (categoriesResponse?.data && Array.isArray(categoriesResponse.data)) {
          categoriesArray = categoriesResponse.data;
        }
        categoryNames = categoriesArray.map((category: any) => category.name || category);

        setAvailableTags(tagNames);
        setAvailableCategories(categoryNames);
      } catch (error) {
        // Show error toast if metadata fetch fails
        toast({
          title: "Error",
          description: "Failed to load filters. Some functionality may be limited.",
          variant: "destructive",
        });
      } finally {
        setIsLoading(false);
      }
    };
    fetchMetadata();
    return () => {};
  }, []);

  return (
    <div className="container mx-auto px-4 py-8">
      <div className="mb-8">
        <h1 className="mb-2 text-3xl font-bold">My Articles</h1>
        <p className="text-muted-foreground">
          Browse, search, and manage your saved articles.
        </p>
      </div>
      <div className="mb-8">
        <UrlForm />
      </div>
      <ArticleFilters 
        availableTags={availableTags}
        availableCategories={availableCategories}
        isLoading={isLoading}
        onFilterChange={handleFilterChange}
      />
      <div className="mt-6">
        <ArticleGrid 
          initialFilters={filters} 
          key={`grid-${JSON.stringify(filters)}`} 
        />
      </div>
      <div className="fixed bottom-6 right-6 md:hidden">
        <Button size="lg" className="h-14 w-14 rounded-full shadow-lg" asChild>
          <Link href="/articles/add">
            <PlusIcon className="h-6 w-6" />
            <span className="sr-only">Add URL</span>
          </Link>
        </Button>
      </div>
    </div>
  );
}