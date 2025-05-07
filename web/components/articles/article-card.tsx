"use client";

import { useState } from "react";
import Link from "next/link";
import Image from "next/image";
import { Article } from "@/types/article";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardFooter } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { 
  BookmarkIcon, 
  ExternalLinkIcon, 
  EyeIcon, 
  MoreHorizontalIcon, 
  StarIcon, 
  TrashIcon 
} from "lucide-react";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { formatDistanceToNow } from "date-fns";
import { cn } from "@/lib/utils";
import { articleAPI } from "@/lib/api-client";
import { toast } from "@/hooks/use-toast";

interface ArticleCardProps {
  article: Article;
  onDelete?: (id: string) => void;
}

export function ArticleCard({ article, onDelete }: ArticleCardProps) {
  const [isImageLoaded, setIsImageLoaded] = useState(false);
  const [localArticle, setLocalArticle] = useState<Article>(article);
  const [isUpdating, setIsUpdating] = useState(false);
  
  // Ensure we have valid tags array
  const tags = localArticle.tags || [];
  
  // Format date safely
  const getFormattedDate = () => {
    try {
      if (!localArticle.createdAt) return "Recently";
      return formatDistanceToNow(new Date(localArticle.createdAt), { addSuffix: true });
    } catch (error) {
      console.error("Error formatting date:", error);
      return "Invalid date";
    }
  };

  // Handle toggling read status
  const handleToggleRead = async (e: React.MouseEvent) => {
    e.preventDefault();
    e.stopPropagation();
    
    if (isUpdating) return;
    
    try {
      setIsUpdating(true);
      const newReadStatus = !localArticle.isRead;
      
      await articleAPI.updateArticleStatus(localArticle.id, { is_read: newReadStatus });
      
      setLocalArticle({
        ...localArticle,
        isRead: newReadStatus
      });
      
      toast({
        title: newReadStatus ? "Marked as read" : "Marked as unread",
        duration: 1500,
      });
    } catch (error) {
      console.error("Failed to update read status:", error);
      toast({
        title: "Update failed",
        description: "Could not update article read status. Please try again later.",
        variant: "destructive",
      });
    } finally {
      setIsUpdating(false);
    }
  };

  // Handle toggling starred status
  const handleToggleStar = async (e: React.MouseEvent) => {
    e.preventDefault();
    e.stopPropagation();
    
    if (isUpdating) return;
    
    try {
      setIsUpdating(true);
      const newStarredStatus = !localArticle.isStarred;
      
      await articleAPI.updateArticleStatus(localArticle.id, { is_starred: newStarredStatus });
      
      setLocalArticle({
        ...localArticle,
        isStarred: newStarredStatus
      });
      
      toast({
        title: newStarredStatus ? "Added to starred" : "Removed from starred",
        duration: 1500,
      });
    } catch (error) {
      console.error("Failed to update starred status:", error);
      toast({
        title: "Update failed",
        description: "Could not update article starred status. Please try again later.",
        variant: "destructive",
      });
    } finally {
      setIsUpdating(false);
    }
  };

  // Fallback image for articles without an image
  const imageUrl = localArticle.ogImage || "https://source.unsplash.com/random/800x600?article";

  return (
    <Link href={`/articles/${localArticle.id}`}>
      <Card className={cn(
        "overflow-hidden transition-all duration-200 hover:shadow-md cursor-pointer",
        localArticle.isRead ? "bg-muted/40" : ""
      )}>
        <div className="relative aspect-video overflow-hidden">
          {!isImageLoaded && (
            <Skeleton className="absolute inset-0 h-full w-full" />
          )}
          <Image
            src={imageUrl}
            alt={localArticle.title}
            fill
            className={`object-cover transition-opacity duration-300 ${
              isImageLoaded ? "opacity-100" : "opacity-0"
            }`}
            onLoad={() => setIsImageLoaded(true)}
            onError={() => {
              console.error("Failed to load image:", imageUrl);
              setIsImageLoaded(true); // Show something rather than just a skeleton
            }}
          />
          <div className="absolute inset-0 bg-gradient-to-t from-black/50 to-transparent" />
          <div className="absolute bottom-3 left-3 right-3">
            <div className="flex flex-wrap gap-1">
              {tags.slice(0, 3).map((tag) => (
                <Badge key={tag} variant="secondary" className="bg-black/50 hover:bg-black/70">
                  {tag}
                </Badge>
              ))}
              {tags.length > 3 && (
                <Badge variant="secondary" className="bg-black/50">
                  +{tags.length - 3}
                </Badge>
              )}
            </div>
          </div>
          
          {/* Read and starred status indicators */}
          <div className="absolute right-2 top-2 flex items-center gap-1">
            {localArticle.isRead && (
              <Badge variant="secondary" className="bg-black/70 px-2 py-1">
                <EyeIcon className="h-3 w-3 mr-1" />
                <span className="text-xs">Read</span>
              </Badge>
            )}
            {localArticle.isStarred && (
              <Badge variant="secondary" className="bg-amber-500/70 px-2 py-1">
                <StarIcon className="h-3 w-3 mr-1" />
                <span className="text-xs">Starred</span>
              </Badge>
            )}
          </div>
        </div>
        <CardContent className="p-4">
          <div className="flex flex-col space-y-2">
            <h3 className="line-clamp-2 font-semibold text-lg">{localArticle.title}</h3>
            <p className="line-clamp-2 text-sm text-muted-foreground">
              {localArticle.description || "No description available"}
            </p>
          </div>
        </CardContent>
        <CardFooter className="flex items-center justify-between p-4 pt-0">
          <div className="flex items-center text-xs text-muted-foreground">
            {getFormattedDate()}
            {localArticle.category && (
              <Badge variant="outline" className="ml-2 text-xs">
                {localArticle.category}
              </Badge>
            )}
          </div>
          <div className="flex items-center gap-2">
            <Button
              variant="ghost"
              size="icon"
              className={cn("h-8 w-8", localArticle.isRead ? "text-primary" : "text-muted-foreground")}
              onClick={handleToggleRead}
              disabled={isUpdating}
            >
              <EyeIcon className="h-4 w-4" />
              <span className="sr-only">{localArticle.isRead ? "Mark as unread" : "Mark as read"}</span>
            </Button>
            
            <Button
              variant="ghost"
              size="icon"
              className={cn("h-8 w-8", localArticle.isStarred ? "text-amber-500" : "text-muted-foreground")}
              onClick={handleToggleStar}
              disabled={isUpdating}
            >
              <StarIcon className="h-4 w-4" />
              <span className="sr-only">{localArticle.isStarred ? "Remove star" : "Add star"}</span>
            </Button>
            
            <Button
              variant="ghost"
              size="icon"
              className="h-8 w-8"
              onClick={(e) => {
                e.preventDefault();
                e.stopPropagation();
                window.open(localArticle.url, '_blank', 'noopener,noreferrer');
              }}
            >
              <ExternalLinkIcon className="h-4 w-4" />
              <span className="sr-only">Open original</span>
            </Button>
            
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <Button 
                  variant="ghost" 
                  size="icon" 
                  className="h-8 w-8"
                  onClick={(e) => {
                    e.preventDefault();
                    e.stopPropagation();
                  }}
                >
                  <MoreHorizontalIcon className="h-4 w-4" />
                  <span className="sr-only">More options</span>
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent 
                align="end"
                onClick={(e) => {
                  e.preventDefault();
                  e.stopPropagation();
                }}
              >
                <DropdownMenuItem
                  className="text-destructive focus:text-destructive"
                  onClick={() => onDelete && onDelete(localArticle.id)}
                >
                  <TrashIcon className="mr-2 h-4 w-4" />
                  Delete
                </DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>
          </div>
        </CardFooter>
      </Card>
    </Link>
  );
}

export function ArticleCardSkeleton() {
  return (
    <Card className="overflow-hidden">
      <Skeleton className="aspect-video" />
      <CardContent className="p-4">
        <div className="flex flex-col space-y-2">
          <Skeleton className="h-6 w-full" />
          <Skeleton className="h-4 w-full" />
          <Skeleton className="h-4 w-3/4" />
        </div>
      </CardContent>
      <CardFooter className="flex items-center justify-between p-4 pt-0">
        <Skeleton className="h-4 w-24" />
        <div className="flex gap-2">
          <Skeleton className="h-8 w-8 rounded-full" />
          <Skeleton className="h-8 w-8 rounded-full" />
          <Skeleton className="h-8 w-8 rounded-full" />
          <Skeleton className="h-8 w-8 rounded-full" />
        </div>
      </CardFooter>
    </Card>
  );
}