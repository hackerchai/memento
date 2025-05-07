"use client";

import { useState, useEffect } from "react";
import Link from "next/link";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Skeleton } from "@/components/ui/skeleton";
import { ArrowLeftIcon, PlusIcon, SearchIcon, TagIcon } from "lucide-react";
import { toast } from "@/hooks/use-toast";
import { tagAPI } from "@/lib/api-client";
import { ArticleCard, ArticleCardSkeleton } from "@/components/articles/article-card";
import type { Article } from "@/types/article";

export default function TagsPage() {
  const [searchQuery, setSearchQuery] = useState("");
  const [tags, setTags] = useState<{ name: string; slug?: string }[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [selectedTag, setSelectedTag] = useState<string | null>(null);
  const [articles, setArticles] = useState<Article[]>([]);
  const [isArticlesLoading, setIsArticlesLoading] = useState(false);

  useEffect(() => {
    const fetchTags = async () => {
      try {
        const response = await tagAPI.listTags(1, 100);
        let tagArray: any[] = [];
        if (response?.data?.data && Array.isArray(response.data.data)) {
          tagArray = response.data.data;
        } else if (response?.data?.items && Array.isArray(response.data.items)) {
          tagArray = response.data.items;
        } else if (response?.data && Array.isArray(response.data)) {
          tagArray = response.data;
        }
        setTags(tagArray.map((tag: any) => ({ name: tag.name, slug: tag.slug })));
      } catch (error) {
        console.error("Failed to fetch tags:", error);
        toast({
          title: "Error",
          description: "Failed to load tags. Please try again later.",
          variant: "destructive",
        });
      } finally {
        setIsLoading(false);
      }
    };
    fetchTags();
  }, []);

  const handleTagClick = async (tagName: string) => {
    setSelectedTag(tagName);
    setIsArticlesLoading(true);
    setArticles([]);
    try {
      // Fetch articles for the selected tag
      const response = await tagAPI.getArticlesByTag(tagName, 1, 20);
      let articlesArray: any[] = [];
      if (response?.data?.data && Array.isArray(response.data.data)) {
        articlesArray = response.data.data;
      } else if (response?.data?.items && Array.isArray(response.data.items)) {
        articlesArray = response.data.items;
      } else if (response?.data && Array.isArray(response.data)) {
        articlesArray = response.data;
      }
      // Map API fields to frontend fields for correct display in ArticleCard
      const mapped = articlesArray.map((item) => ({
        ...item,
        ogImage: item.ogImage || item.og_image_url,
        isRead: typeof item.isRead === 'boolean' ? item.isRead : item.is_read,
        isStarred: typeof item.isStarred === 'boolean' ? item.isStarred : item.is_starred,
      }));
      setArticles(mapped);
    } catch (error) {
      console.error("Failed to fetch articles by tag:", error);
      toast({
        title: "Error",
        description: "Failed to load articles for this tag.",
        variant: "destructive",
      });
    } finally {
      setIsArticlesLoading(false);
    }
  };

  const filteredTags = tags.filter((tag: { name: string; slug?: string }) =>
    tag.name && tag.name.toLowerCase().includes(searchQuery.toLowerCase())
  );

  return (
    <div className="container mx-auto px-4 py-8">
      <div className="mb-6">
        <Button variant="ghost" asChild>
          <Link href="/articles">
            <ArrowLeftIcon className="mr-2 h-4 w-4" />
            Back to Articles
          </Link>
        </Button>
      </div>
      <div className="mb-8">
        <h1 className="mb-2 text-3xl font-bold">Tags</h1>
        <p className="text-muted-foreground">
          Browse and manage your article tags.
        </p>
      </div>
      <div className="mb-6 flex items-center gap-4">
        <div className="relative flex-grow">
          <SearchIcon className="absolute left-3 top-2.5 h-4 w-4 text-muted-foreground" />
          <Input
            placeholder="Search tags..."
            className="pl-9"
            value={searchQuery}
            onChange={(e: React.ChangeEvent<HTMLInputElement>) => setSearchQuery(e.target.value)}
          />
        </div>
        <Button>
          <PlusIcon className="mr-2 h-4 w-4" />
          New Tag
        </Button>
      </div>
      {isLoading ? (
        <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4">
          {Array.from({ length: 12 }).map((_, index) => (
            <Skeleton key={index} className="h-16 w-full" />
          ))}
        </div>
      ) : filteredTags.length === 0 ? (
        <div className="flex flex-col items-center justify-center py-12 text-center">
          <TagIcon className="h-16 w-16 text-muted-foreground/50" />
          <h3 className="mt-4 text-xl font-semibold">No tags found</h3>
          <p className="mt-2 text-muted-foreground">
            {searchQuery
              ? `No tags matching "${searchQuery}"`
              : "You haven't created any tags yet"}
          </p>
        </div>
      ) : (
        <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4">
          {filteredTags.map((tag: { name: string; slug?: string }) => (
            <button
              key={tag.name}
              className={`group flex items-center justify-between rounded-lg border border-border bg-card p-4 transition-colors hover:bg-accent w-full text-left ${selectedTag === tag.name ? 'ring-2 ring-primary' : ''}`}
              onClick={() => handleTagClick(tag.name)}
            >
              <div className="flex items-center gap-3">
                <TagIcon className="h-5 w-5 text-primary" />
                <div>
                  <div className="font-medium">{tag.name}</div>
                </div>
              </div>
            </button>
          ))}
        </div>
      )}
      {selectedTag && (
        <div className="mt-8">
          <h2 className="mb-4 text-2xl font-semibold">Articles for tag: {selectedTag}</h2>
          {isArticlesLoading ? (
            <div className="grid grid-cols-1 gap-6 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
              {Array.from({ length: 8 }).map((_, idx) => (
                <ArticleCardSkeleton key={idx} />
              ))}
            </div>
          ) : articles.length === 0 ? (
            <div className="flex flex-col items-center justify-center py-12">
              <h3 className="mt-4 text-xl font-semibold">No articles found for this tag</h3>
            </div>
          ) : (
            <div className="grid grid-cols-1 gap-6 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
              {articles.map((article: Article) => (
                <ArticleCard key={article.id} article={article} />
              ))}
            </div>
          )}
        </div>
      )}
    </div>
  );
}