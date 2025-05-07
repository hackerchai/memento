"use client";

import { useState, useEffect } from "react";
import { Button } from "@/components/ui/button";
import { Label } from "@/components/ui/label";
import { Separator } from "@/components/ui/separator";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuLabel,
  DropdownMenuRadioGroup,
  DropdownMenuRadioItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { CheckIcon, EyeIcon, FilterIcon, StarIcon, XIcon } from "lucide-react";
import { Skeleton } from "@/components/ui/skeleton";

type SortOption = "newest" | "oldest" | "title-asc" | "title-desc";

interface ArticleFiltersProps {
  availableTags: string[];
  availableCategories: string[];
  isLoading?: boolean;
  initialSelectedTags?: string[];
  onFilterChange?: (filters: {
    isRead?: boolean;
    isStarred?: boolean;
  }) => void;
}

export function ArticleFilters({ 
  availableTags, 
  availableCategories, 
  isLoading = false,
  initialSelectedTags = [],
  onFilterChange 
}: ArticleFiltersProps) {
  const [isReadFilter, setIsReadFilter] = useState<boolean>(false);
  const [isStarredFilter, setIsStarredFilter] = useState<boolean>(false);
  const [sortBy, setSortBy] = useState<SortOption>("newest");

  useEffect(() => {
    if (onFilterChange) {
      const filtersToApi: { isRead?: boolean; isStarred?: boolean } = {};
      
      if (isReadFilter) {
        filtersToApi.isRead = false;
      }

      if (isStarredFilter) {
        filtersToApi.isStarred = true;
      }
      

      onFilterChange(filtersToApi);
    }
  }, [isReadFilter, isStarredFilter, onFilterChange]);

  const toggleReadFilter = () => {
    setIsReadFilter((prev: boolean) => !prev);
  };

  const toggleStarredFilter = () => {
    setIsStarredFilter((prev: boolean) => !prev);
  };

  const resetFilters = () => {
    setIsReadFilter(false);
    setIsStarredFilter(false);
    setSortBy("newest");
  };

  if (isLoading) {
    return (
      <div className="space-y-4">
        <div className="flex flex-wrap items-center gap-2">
          <Skeleton className="h-8 w-20" />
          <Skeleton className="h-8 w-28" />
        </div>
        <Separator />
      </div>
    );
  }

  return (
    <div className="space-y-4">
      <div className="flex flex-wrap items-center justify-between gap-2">
        <div className="flex items-center gap-2">
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button variant="outline" size="sm" className="h-8">
                <FilterIcon className="mr-2 h-3.5 w-3.5" />
                Filters
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent className="w-56">
              <DropdownMenuLabel>Status</DropdownMenuLabel>
              <DropdownMenuSeparator />
              <div className="py-1">
                <div
                  className="flex items-center px-2 py-1.5 hover:bg-muted cursor-pointer"
                  onClick={toggleReadFilter}
                >
                  <div className="flex h-4 w-4 items-center justify-center rounded-sm border mr-2">
                    {isReadFilter && <CheckIcon className="h-3 w-3" />}
                  </div>
                  <EyeIcon className={`mr-2 h-3.5 w-3.5 ${isReadFilter ? "text-blue-500" : "text-muted-foreground"}`} />
                  <span className="text-sm">
                    Unread articles
                  </span>
                </div>
                
                <div
                  className="flex items-center px-2 py-1.5 hover:bg-muted cursor-pointer"
                  onClick={toggleStarredFilter}
                >
                  <div className="flex h-4 w-4 items-center justify-center rounded-sm border mr-2">
                    {isStarredFilter && <CheckIcon className="h-3 w-3" />}
                  </div>
                  <StarIcon className={`mr-2 h-3.5 w-3.5 ${isStarredFilter ? "text-amber-500" : "text-muted-foreground"}`} />
                  <span className="text-sm">
                    Starred articles
                  </span>
                </div>
              </div>
            </DropdownMenuContent>
          </DropdownMenu>

          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button variant="outline" size="sm" className="h-8">
                Sort: {getSortLabel(sortBy)}
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent>
              <DropdownMenuRadioGroup value={sortBy} onValueChange={(value: string) => setSortBy(value as SortOption)}>
                <DropdownMenuRadioItem value="newest">Newest</DropdownMenuRadioItem>
                <DropdownMenuRadioItem value="oldest">Oldest</DropdownMenuRadioItem>
                <DropdownMenuRadioItem value="title-asc">Title (A-Z)</DropdownMenuRadioItem>
                <DropdownMenuRadioItem value="title-desc">Title (Z-A)</DropdownMenuRadioItem>
              </DropdownMenuRadioGroup>
            </DropdownMenuContent>
          </DropdownMenu>

          <Button
            variant="ghost"
            size="sm"
            className="h-8"
            onClick={resetFilters}
          >
            <XIcon className="mr-2 h-3.5 w-3.5" />
            Reset filters
          </Button>
        </div>
      </div>

      <div className="flex flex-wrap gap-1 pt-2">
        <div className="flex items-center gap-1 pl-2 border rounded-md pr-1 py-0.5">
          <EyeIcon className="mr-1 h-3 w-3" />
          {isReadFilter ? "Unread only" : "All (Read)"}
          <button
            className="ml-1 rounded-full outline-none ring-offset-background focus:ring-2 focus:ring-ring focus:ring-offset-2"
            onClick={toggleReadFilter}
          >
            <XIcon className="h-3 w-3" />
            <span className="sr-only">Toggle read status</span>
          </button>
        </div>

        <div className="flex items-center gap-1 pl-2 border rounded-md pr-1 py-0.5">
          <StarIcon className="mr-1 h-3 w-3" />
          {isStarredFilter ? "Starred only" : "All (Starred)"}
          <button
            className="ml-1 rounded-full outline-none ring-offset-background focus:ring-2 focus:ring-ring focus:ring-offset-2"
            onClick={toggleStarredFilter}
          >
            <XIcon className="h-3 w-3" />
            <span className="sr-only">Toggle starred status</span>
          </button>
        </div>
      </div>

      <Separator />
    </div>
  );
}

function getSortLabel(sortBy: SortOption): string {
  switch (sortBy) {
    case "newest":
      return "Newest";
    case "oldest":
      return "Oldest";
    case "title-asc":
      return "Title (A-Z)";
    case "title-desc":
      return "Title (Z-A)";
    default:
      return "Newest";
  }
}