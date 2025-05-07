"use client";

import { useState, useEffect } from "react";
import Link from "next/link";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Skeleton } from "@/components/ui/skeleton";
import { ArrowLeftIcon, PlusIcon, SearchIcon, FolderIcon, Trash2Icon } from "lucide-react";
import { toast } from "@/hooks/use-toast";
import { categoryAPI } from "@/lib/api-client";
import { ArticleCard, ArticleCardSkeleton } from "@/components/articles/article-card";
import type { Article } from "@/types/article";

export default function CategoriesPage() {
  const [searchQuery, setSearchQuery] = useState("");
  const [categories, setCategories] = useState<{ name: string; slug?: string; id?: string }[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [selectedCategory, setSelectedCategory] = useState<string | null>(null);
  const [articles, setArticles] = useState<Article[]>([]);
  const [isArticlesLoading, setIsArticlesLoading] = useState(false);
  const [isDeleting, setIsDeleting] = useState<string | null>(null);

  useEffect(() => {
    fetchCategories();
  }, []);
  
  const fetchCategories = async () => {
    try {
      setIsLoading(true);
      const response = await categoryAPI.listCategories(1, 100);
      let categoryArray: any[] = [];
      if (response?.data?.data && Array.isArray(response.data.data)) {
        categoryArray = response.data.data;
      } else if (response?.data?.items && Array.isArray(response.data.items)) {
        categoryArray = response.data.items;
      } else if (response?.data && Array.isArray(response.data)) {
        categoryArray = response.data;
      }
      setCategories(categoryArray.map((category: any) => ({ name: category.name, slug: category.slug, id: category.id })));
    } catch (error) {
      console.error("Failed to fetch categories:", error);
      toast({
        title: "Error",
        description: "Failed to load categories. Please try again later.",
        variant: "destructive",
      });
    } finally {
      setIsLoading(false);
    }
  };

  const handleCategoryClick = async (identifier: string) => {
    setSelectedCategory(identifier);
    setIsArticlesLoading(true);
    setArticles([]);
    try {
      // Fetch articles for the selected category
      const response = await categoryAPI.getArticlesByCategory(identifier, 1, 20);
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
      console.error("Failed to fetch articles by category:", error);
      toast({
        title: "Error",
        description: "Failed to load articles for this category.",
        variant: "destructive",
      });
    } finally {
      setIsArticlesLoading(false);
    }
  };

  const handleDeleteCategory = async (e: React.MouseEvent, categoryId: string, categoryName: string) => {
    e.stopPropagation(); // Prevent triggering the category click event
    
    if (!categoryId) {
      toast({
        title: "Error",
        description: "Category ID is missing, unable to delete.",
        variant: "destructive",
      });
      return;
    }
    
    try {
      setIsDeleting(categoryId);
      await categoryAPI.deleteCategory(categoryId);
      
      // If currently displaying articles for the deleted category, clear the article area
      if (selectedCategory === categoryId || selectedCategory === categoryName) {
        setSelectedCategory(null);
        setArticles([]);
      }
      
      // Remove the deleted category from the list
      setCategories(categories.filter(category => category.id !== categoryId));
      
      toast({
        title: "Success",
        description: `Category "${categoryName}" has been deleted.`,
      });
    } catch (error) {
      console.error("Failed to delete category:", error);
      toast({
        title: "Error",
        description: "Failed to delete category. Please try again later.",
        variant: "destructive",
      });
    } finally {
      setIsDeleting(null);
    }
  };

  const handleCreateCategory = async () => {
    const name = prompt("Enter new category name");
    if (!name) return;
    
    try {
      const response = await categoryAPI.createCategory(name);
      
      // Add the new category to the list
      if (response?.data?.id) {
        setCategories([...categories, { name, id: response.data.id, slug: response.data.slug }]);
        toast({
          title: "Success",
          description: `Category "${name}" has been created.`,
        });
      }
    } catch (error) {
      console.error("Failed to create category:", error);
      toast({
        title: "Error",
        description: "Failed to create category. Please try again later.",
        variant: "destructive",
      });
    }
  };

  const filteredCategories = categories.filter((category: { name: string; slug?: string }) =>
    category.name && category.name.toLowerCase().includes(searchQuery.toLowerCase())
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
        <h1 className="mb-2 text-3xl font-bold">Categories</h1>
        <p className="text-muted-foreground">
          Browse and manage your article categories.
        </p>
      </div>
      <div className="mb-6 flex items-center gap-4">
        <div className="relative flex-grow">
          <SearchIcon className="absolute left-3 top-2.5 h-4 w-4 text-muted-foreground" />
          <Input
            placeholder="Search categories..."
            className="pl-9"
            value={searchQuery}
            onChange={(e: React.ChangeEvent<HTMLInputElement>) => setSearchQuery(e.target.value)}
          />
        </div>
        <Button onClick={handleCreateCategory}>
          <PlusIcon className="mr-2 h-4 w-4" />
          New Category
        </Button>
      </div>
      {isLoading ? (
        <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4">
          {Array.from({ length: 12 }).map((_, index) => (
            <Skeleton key={index} className="h-16 w-full" />
          ))}
        </div>
      ) : filteredCategories.length === 0 ? (
        <div className="flex flex-col items-center justify-center py-12 text-center">
          <FolderIcon className="h-16 w-16 text-muted-foreground/50" />
          <h3 className="mt-4 text-xl font-semibold">No categories found</h3>
          <p className="mt-2 text-muted-foreground">
            {searchQuery
              ? `No categories matching "${searchQuery}"`
              : "You haven't created any categories yet"}
          </p>
        </div>
      ) : (
        <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4">
          {filteredCategories.map((category: { name: string; slug?: string; id?: string }) => (
            <button
              key={category.id || category.name}
              className={`group flex items-center justify-between rounded-lg border border-border bg-card p-4 transition-colors hover:bg-accent w-full text-left ${selectedCategory === category.id || selectedCategory === category.slug ? 'ring-2 ring-primary' : ''}`}
              onClick={() => handleCategoryClick(category.id || category.slug || category.name)}
            >
              <div className="flex items-center gap-3">
                <FolderIcon className="h-5 w-5 text-primary" />
                <div>
                  <div className="font-medium">{category.name}</div>
                </div>
              </div>
              <Button
                variant="ghost"
                size="icon"
                className="opacity-0 group-hover:opacity-100 transition-opacity"
                onClick={(e) => handleDeleteCategory(e, category.id || "", category.name)}
                disabled={isDeleting === category.id}
              >
                {isDeleting === category.id ? (
                  <span className="h-4 w-4 animate-spin rounded-full border-2 border-current border-t-transparent" />
                ) : (
                  <Trash2Icon className="h-4 w-4 text-destructive" />
                )}
                <span className="sr-only">Delete category</span>
              </Button>
            </button>
          ))}
        </div>
      )}
      {selectedCategory && (
        <div className="mt-8">
          <h2 className="mb-4 text-2xl font-semibold">Articles in category: {categories.find(c => c.id === selectedCategory || c.slug === selectedCategory)?.name || selectedCategory}</h2>
          {isArticlesLoading ? (
            <div className="grid grid-cols-1 gap-6 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
              {Array.from({ length: 8 }).map((_, idx) => (
                <ArticleCardSkeleton key={idx} />
              ))}
            </div>
          ) : articles.length === 0 ? (
            <div className="flex flex-col items-center justify-center py-12">
              <h3 className="mt-4 text-xl font-semibold">No articles found in this category</h3>
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