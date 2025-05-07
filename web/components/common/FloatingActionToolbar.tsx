"use client";

import { useState, useRef, useEffect, useMemo } from 'react';
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { 
  Dialog, 
  DialogContent, 
  DialogHeader, 
  DialogTitle, 
  DialogFooter,
  DialogDescription,
  DialogTrigger
} from "@/components/ui/dialog";
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover"
import { 
  BookmarkIcon, 
  RefreshCwIcon, 
  Star as StarIcon,
  BookOpen as BookOpenIcon,
  Book as BookClosedIcon,
  Check as CheckIcon,
  Plus as PlusIcon,
  Tag as TagIcon,
  X as XIcon,
  FolderIcon,
  Loader2Icon,
  ExternalLink as ExternalLinkIcon,
} from "lucide-react";
import { Badge } from "@/components/ui/badge";
import { cn } from "@/lib/utils";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Command, CommandEmpty, CommandGroup, CommandInput, CommandItem, CommandList, CommandSeparator } from "@/components/ui/command";
import { categoryAPI } from "@/lib/api-client";
import { useToast } from "@/components/ui/use-toast";

// Define more basic types to avoid dependency issues
interface Tag {
  name: string;
  slug?: string;
}

interface Category {
  name: string;
  slug: string;
}

interface DetailedArticle {
  id: string;
  title: string;
  isRead?: boolean;
  isStarred?: boolean;
  tags: (string | Tag)[];
  category?: {
    name: string;
    slug: string;
  } | string;
  url?: string;
}

// Define props interface for component
interface FloatingActionToolbarProps {
  article: DetailedArticle;
  onUpdateStatus: (status: { is_read?: boolean; is_starred?: boolean }) => Promise<void>;
  onUpdateCategory: (category: string) => Promise<void>;
  onAddTags: (tags: string[]) => Promise<void>;
  onRemoveTags: (tags: string[]) => Promise<void>;
  onRescrape: () => Promise<void>;
  categories: Category[];
  // Add optional props for existing tags list
  existingTags?: string[];
  // Add a callback for refreshing the parent component
  onRefresh?: () => Promise<void>;
}

export default function FloatingActionToolbar({
  article,
  onUpdateStatus,
  onUpdateCategory,
  onAddTags,
  onRemoveTags,
  onRescrape,
  categories = [],
  existingTags = [],
  onRefresh
}: FloatingActionToolbarProps) {
  const [isExpanded, setIsExpanded] = useState(false);
  const [isLoading, setIsLoading] = useState(false);
  const [newTagInput, setNewTagInput] = useState('');
  const [tagsDialogOpen, setTagsDialogOpen] = useState(false);
  const [categoryPopoverOpen, setCategoryPopoverOpen] = useState(false);
  const [suggestedTags, setSuggestedTags] = useState<string[]>([]);
  const [selectedTags, setSelectedTags] = useState<string[]>([]);
  const [newCategoryName, setNewCategoryName] = useState('');
  const [isCreatingCategory, setIsCreatingCategory] = useState(false);
  const [localArticle, setLocalArticle] = useState<DetailedArticle>(article);
  const [updatingField, setUpdatingField] = useState<{is_read?: boolean, is_starred?: boolean} | null>(null);
  
  const toolbarRef = useRef<HTMLDivElement>(null);
  const tagsDialogContentRef = useRef<HTMLDivElement>(null);
  const categoryPopoverContentRef = useRef<HTMLDivElement>(null);
  const tagsTriggerRef = useRef<HTMLButtonElement>(null);
  const categoryTriggerRef = useRef<HTMLButtonElement>(null);
  const { toast } = useToast();

  useEffect(() => {
    if (!updatingField) {
      // console.log('FloatingActionToolbar: Article prop changed, updating localArticle');
      setLocalArticle(article);
    } else {
      // console.log('FloatingActionToolbar: Ignoring article prop change while updating status locally');
    }
  }, [article, updatingField]);

  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      const target = event.target as Node;
      console.log("handleClickOutside triggered. Target:", target);
      
      if (
        (tagsTriggerRef.current && tagsTriggerRef.current.contains(target)) ||
        (categoryTriggerRef.current && categoryTriggerRef.current.contains(target))
      ) {
        console.log("handleClickOutside: Click on a trigger button, doing nothing.");
        return;
      }
      
      const isOutsideToolbar = toolbarRef.current && !toolbarRef.current.contains(target);
      const isOutsideTagsDialog = !tagsDialogContentRef.current || !tagsDialogContentRef.current.contains(target);
      const isOutsideCategoryPopover = !categoryPopoverContentRef.current || !categoryPopoverContentRef.current.contains(target);

      console.log("handleClickOutside: isOutsideToolbar:", isOutsideToolbar);
      console.log("handleClickOutside: isOutsideTagsDialog:", isOutsideTagsDialog);
      console.log("handleClickOutside: isOutsideCategoryPopover:", isOutsideCategoryPopover);

      if (isOutsideToolbar && isOutsideTagsDialog && isOutsideCategoryPopover) {
        console.log("handleClickOutside: Conditions met to set isExpanded to false.");
        setIsExpanded(false);
      }
    };

    document.addEventListener('mousedown', handleClickOutside);
    return () => {
      document.removeEventListener('mousedown', handleClickOutside);
    };
  }, []); // Empty dependency array is correct here as refs' .current property is accessed dynamically

  // Reset tag selection when dialog opens
  useEffect(() => {
    if (tagsDialogOpen) {
      setSelectedTags([]);
      setNewTagInput('');
    }
  }, [tagsDialogOpen]);

  // Reset category creation fields
  useEffect(() => {
    if (!categoryPopoverOpen) {
      setNewCategoryName('');
    }
  }, [categoryPopoverOpen]);

  // Filter tag suggestions based on input
  useEffect(() => {
    if (newTagInput.trim()) {
      const input = newTagInput.trim().toLowerCase();
      
      // Filter matching tags from existing tags
      const articleTags = getArticleTags();
      const filtered = existingTags
        .filter(tag => 
          tag.toLowerCase().includes(input) && 
          !articleTags.includes(tag) &&
          !selectedTags.includes(tag)
        )
        .slice(0, 5); // Limit suggestions to 5
      setSuggestedTags(filtered);
    } else {
      setSuggestedTags([]);
    }
  }, [newTagInput, existingTags, selectedTags, localArticle.tags]);
  
  // Extract tag names from complex tag data structure
  const getArticleTags = () => {
    return Array.isArray(localArticle.tags)
      ? localArticle.tags.map(tag => typeof tag === 'string' ? tag : tag.name || '')
      : [];
  };

  // Get current category name
  const getCurrentCategoryName = () => {
    if (!localArticle.category) return '';
    if (typeof localArticle.category === 'string') return localArticle.category;
    return localArticle.category.name || '';
  };

  // Handle category selection
  const handleSelectCategory = async (categoryName: string) => {
    const currentCategory = getCurrentCategoryName();
    
    if (categoryName === currentCategory) {
      setCategoryPopoverOpen(false);
      return;
    }
    
    if (isLoading) return;
    
    try {
      setIsLoading(true);
      
      // Update local state immediately for instant feedback
      setLocalArticle(prev => ({
        ...prev,
        category: categoryName
      }));
      
      // Call API to update category
      await onUpdateCategory(categoryName);
      
      toast({
        title: "Category updated",
        description: `Moved to "${categoryName}"`,
      });
      
      // Close popover
      setCategoryPopoverOpen(false);
      
      // Refresh article data
      if (onRefresh) {
        await onRefresh();
      }
    } catch (error) {
      console.error('Failed to update category:', error);
      
      // Restore original state on error
      setLocalArticle(article);
      
      toast({
        title: "Failed to update category",
        variant: "destructive",
      });
    } finally {
      setIsLoading(false);
    }
  };

  // Handle creating new category
  const handleCreateCategory = async () => {
    if (!newCategoryName.trim() || isCreatingCategory || isLoading) return;
    
    try {
      setIsCreatingCategory(true);
      
      // Create new category
      await categoryAPI.createCategory(newCategoryName);
      
      // Update article with new category
      await handleSelectCategory(newCategoryName);
      
      // Reset input and close popover
      setNewCategoryName('');
    } catch (error) {
      console.error('Failed to create category:', error);
      
      toast({
        title: "Failed to create category",
        variant: "destructive",
      });
    } finally {
      setIsCreatingCategory(false);
    }
  };

  // Toggle read status
  const handleToggleRead = async () => {
    if (isLoading) return;
    
    try {
      setIsLoading(true);
      const newStatus = !localArticle.isRead;
      
      setUpdatingField({ is_read: newStatus });
      
      setLocalArticle(prev => ({
        ...prev,
        isRead: newStatus
      }));
      
      console.log("[FloatingActionToolbar] Attempting to call onUpdateStatus for isRead:", { articleId: localArticle.id, newStatus });
      await onUpdateStatus({ is_read: newStatus });
      console.log("[FloatingActionToolbar] onUpdateStatus call completed for isRead.");
      
      toast({
        title: newStatus ? "Article marked as read" : "Article marked as unread",
      });
      
      // Refresh article data (but keep our local state if API call was successful)
      if (onRefresh) {
        await onRefresh();
      }
    } catch (error) {
      console.error('Failed to toggle read status:', error);
      
      // Restore original state
      setLocalArticle(prev => ({
        ...prev,
        isRead: article.isRead
      }));
      
      toast({
        title: "Failed to update status",
        variant: "destructive"
      });
    } finally {
      setIsLoading(false);
      // Clear the updating field flag
      setUpdatingField(null);
    }
  };

  // Toggle star status
  const handleToggleStar = async () => {
    if (isLoading) return;
    
    try {
      setIsLoading(true);
      const newStatus = !localArticle.isStarred;
      
      setUpdatingField({ is_starred: newStatus });
      
      setLocalArticle(prev => ({
        ...prev,
        isStarred: newStatus
      }));
      
      console.log("[FloatingActionToolbar] Attempting to call onUpdateStatus for isStarred:", { articleId: localArticle.id, newStatus });
      await onUpdateStatus({ is_starred: newStatus });
      console.log("[FloatingActionToolbar] onUpdateStatus call completed for isStarred.");
      
      toast({
        title: newStatus ? "Article starred" : "Article unstarred",
      });
      
      // Refresh article data (but keep our local state if API call was successful)
      if (onRefresh) {
        await onRefresh();
      }
    } catch (error) {
      console.error('Failed to toggle star status:', error);
      
      // Restore original state only for the star property
      setLocalArticle(prev => ({
        ...prev,
        isStarred: article.isStarred
      }));
      
      toast({
        title: "Failed to update status",
        variant: "destructive"
      });
    } finally {
      setIsLoading(false);
      // Clear the updating field flag
      setUpdatingField(null);
    }
  };

  // Handle rescrape
  const handleRescrape = async () => {
    if (isLoading) return;
    
    try {
      setIsLoading(true);
      
      // Call API to rescrape
      await onRescrape();
      
      toast({
        title: "Article rescraped",
        description: "Article content is being refreshed in background",
      });
      
      // Refresh article data
      if (onRefresh) {
        await onRefresh();
      }
    } catch (error) {
      console.error('Failed to rescrape:', error);
      
      toast({
        title: "Rescrape failed",
        description: "Could not refresh article content",
        variant: "destructive"
      });
    } finally {
      setIsLoading(false);
    }
  };

  // Add tag to selection
  const handleSelectTag = (tag: string) => {
    if (!selectedTags.includes(tag)) {
      setSelectedTags([...selectedTags, tag]);
      setNewTagInput('');
    }
  };

  // Remove tag from selection
  const handleRemoveSelectedTag = (tag: string) => {
    setSelectedTags(selectedTags.filter(t => t !== tag));
  };

  // Handle adding tags
  const handleAddTag = async () => {
    if (isLoading) return;
    
    try {
      setIsLoading(true);
      const tagsToAdd: string[] = [];
      
      // Add from selection
      if (selectedTags.length > 0) {
        tagsToAdd.push(...selectedTags);
      }
      
      // Add from input field
      if (newTagInput.trim()) {
        const articleTags = getArticleTags();
        const inputTags = newTagInput
          .split(',')
          .map(tag => tag.trim())
          .filter(tag => tag && !articleTags.includes(tag) && !selectedTags.includes(tag));
        
        tagsToAdd.push(...inputTags);
      }
      
      if (tagsToAdd.length > 0) {
        // Update local state
        const updatedTags = [...getArticleTags(), ...tagsToAdd];
        setLocalArticle(prev => ({
          ...prev,
          tags: updatedTags
        }));
        
        // Call API to add tags
        await onAddTags(tagsToAdd);
        
        toast({
          title: "Tags added",
          description: `Added ${tagsToAdd.length} tag${tagsToAdd.length > 1 ? 's' : ''} to article`,
        });
        
        // Reset input
        setNewTagInput('');
        setSelectedTags([]);
        
        // Refresh article data
        if (onRefresh) {
          await onRefresh();
        }
      }
    } catch (error) {
      console.error('Failed to add tags:', error);
      
      // Restore original state
      setLocalArticle(article);
      
      toast({
        title: "Failed to add tags",
        description: "Could not update article tags",
        variant: "destructive"
      });
    } finally {
      setIsLoading(false);
    }
  };

  // Handle removing tags
  const handleRemoveTag = async (tagToRemove: string) => {
    if (isLoading) return;
    
    try {
      setIsLoading(true);
      
      // Update local state
      const updatedTags = getArticleTags().filter(tag => tag !== tagToRemove);
      setLocalArticle(prev => ({
        ...prev,
        tags: updatedTags
      }));
      
      // Call API to remove tags
      await onRemoveTags([tagToRemove]);
      
      toast({
        title: "Tag removed",
        description: `Removed tag "${tagToRemove}" from article`,
      });
      
      // Refresh article data
      if (onRefresh) {
        await onRefresh();
      }
    } catch (error) {
      console.error('Failed to remove tag:', error);
      
      // Restore original state
      setLocalArticle(article);
      
      toast({
        title: "Failed to remove tag",
        description: "Could not update article tags",
        variant: "destructive"
      });
    } finally {
      setIsLoading(false);
    }
  };

  // Handle opening the article URL in a new tab
  const handleOpenArticleUrl = () => {
    if (localArticle.url) {
      window.open(localArticle.url, '_blank', 'noopener,noreferrer');
    } else {
      toast({
        title: "URL not available",
        description: "This article doesn't have a source URL",
        variant: "destructive"
      });
    }
  };

  const articleTags = getArticleTags();
  const currentCategoryName = getCurrentCategoryName();

  // Reset tagsDialogOpen and categoryPopoverOpen when component expands/collapses
  useEffect(() => {
    if (!isExpanded) {
      console.log("isExpanded changed to:", isExpanded, "Setting tagsDialogOpen and categoryPopoverOpen to false.");
      setTagsDialogOpen(false);
      setCategoryPopoverOpen(false);
    }
  }, [isExpanded]);

  // Function to toggle tag dialog or category popover open state
  const toggleTagsDialog = () => {
    if (!tagsDialogOpen) {
      // Close category popover if open
      setCategoryPopoverOpen(false);
    }
    setTagsDialogOpen(!tagsDialogOpen);
  };

  return (
    <div 
      ref={toolbarRef}
      className="fixed right-6 bottom-6 z-50 flex flex-col-reverse items-center gap-2"
    >
      {/* Main floating button */}
      <Button
        onClick={() => {
          console.log("Main toolbar button clicked. Current isExpanded:", isExpanded, "New isExpanded:", !isExpanded);
          setIsExpanded(!isExpanded)
        }}
        className="h-14 w-14 rounded-full shadow-lg bg-primary hover:bg-primary/90"
      >
        <BookmarkIcon className="h-6 w-6" />
      </Button>

      {/* Expanded toolbar */}
      {isExpanded && (
        <div className="flex flex-col gap-2 mb-2 items-center">
          {/* Open Original URL button - Added as the topmost button */}
          <Button
            onClick={handleOpenArticleUrl}
            disabled={isLoading || !localArticle.url}
            className="h-10 w-10 rounded-full shadow-md bg-indigo-600 hover:bg-indigo-700"
            title="Open original article"
          >
            <ExternalLinkIcon className="h-5 w-5" />
          </Button>
          
          {/* Read/Unread button */}
          <Button
            onClick={handleToggleRead}
            disabled={isLoading}
            className={cn(
              "h-10 w-10 rounded-full shadow-md", 
              localArticle.isRead 
                ? "bg-blue-600 hover:bg-blue-700" 
                : "bg-slate-600 hover:bg-slate-700"
            )}
            title={localArticle.isRead ? "Mark as unread" : "Mark as read"}
          >
            {localArticle.isRead ? 
              <BookOpenIcon className="h-5 w-5" /> : 
              <BookClosedIcon className="h-5 w-5" />
            }
          </Button>
          
          {/* Star button */}
          <Button
            onClick={handleToggleStar}
            disabled={isLoading}
            className={cn(
              "h-10 w-10 rounded-full shadow-md", 
              localArticle.isStarred 
                ? "bg-amber-500 hover:bg-amber-600" 
                : "bg-slate-600 hover:bg-slate-700"
            )}
            title={localArticle.isStarred ? "Remove star" : "Add star"}
          >
            <StarIcon className={cn("h-5 w-5", localArticle.isStarred && "fill-white")} />
          </Button>

          {/* Tags dialog - modified to handle proper event flow */}
          <Dialog 
            open={tagsDialogOpen} 
            onOpenChange={(open) => {
              // Only allow programmatic closing through our own UI
              if (!open) {
                setTagsDialogOpen(false);
              }
            }}
          >
            <Button
              ref={tagsTriggerRef}
              onClick={toggleTagsDialog}
              disabled={isLoading}
              className="h-10 w-10 rounded-full shadow-md bg-blue-600 hover:bg-blue-700"
              title="Manage tags"
            >
              <TagIcon className="h-5 w-5" />
            </Button>
            <DialogContent 
              ref={tagsDialogContentRef}
              className="sm:max-w-md" 
              onInteractOutside={(e) => {
                // Default behavior is to close on interact outside, which is fine.
                // The main toolbar's handleClickOutside will handle not collapsing if click is here.
              }}
              onEscapeKeyDown={(e) => {
                // Allow ESC key to close
                setTagsDialogOpen(false);
              }}
            >
              <DialogHeader>
                <DialogTitle>Manage Tags</DialogTitle>
                <DialogDescription>
                  Add or remove tags for this article
                </DialogDescription>
              </DialogHeader>
              <div className="space-y-4 py-4">
                {/* Current tags */}
                <div>
                  <h4 className="mb-2 text-sm font-medium">Current Tags</h4>
                  <div className="flex flex-wrap gap-2">
                    {articleTags.length > 0 ? (
                      articleTags.map((tag, index) => (
                        <Badge key={index} variant="secondary" className="group relative">
                          {tag}
                          <button
                            onClick={(e) => {
                              e.stopPropagation();
                              handleRemoveTag(tag);
                            }}
                            className="ml-1 rounded-full hover:bg-destructive/20 p-0.5"
                            aria-label={`Remove tag ${tag}`}
                          >
                            <XIcon className="h-3 w-3" />
                          </button>
                        </Badge>
                      ))
                    ) : (
                      <p className="text-sm text-muted-foreground">No tags added yet</p>
                    )}
                  </div>
                </div>
                
                {/* Pending tags */}
                {selectedTags.length > 0 && (
                  <div>
                    <h4 className="mb-2 text-sm font-medium">Pending Tags</h4>
                    <div className="flex flex-wrap gap-2">
                      {selectedTags.map((tag, index) => (
                        <Badge key={index} variant="outline" className="bg-blue-100 dark:bg-blue-900/30">
                          {tag}
                          <button
                            onClick={(e) => {
                              e.stopPropagation();
                              handleRemoveSelectedTag(tag);
                            }}
                            className="ml-1 rounded-full hover:bg-destructive/20 p-0.5"
                            aria-label={`Remove ${tag} from selection`}
                          >
                            <XIcon className="h-3 w-3" />
                          </button>
                        </Badge>
                      ))}
                    </div>
                  </div>
                )}
                
                {/* Add new tags */}
                <div>
                  <h4 className="mb-2 text-sm font-medium">Add New Tags</h4>
                  <div className="flex gap-2">
                    <Input
                      value={newTagInput}
                      onChange={(e) => setNewTagInput(e.target.value)}
                      onKeyDown={(e) => {
                        if (e.key === 'Enter') {
                          e.preventDefault();
                          if (suggestedTags.length > 0) {
                            handleSelectTag(suggestedTags[0]); // Select first suggestion
                          } else if (newTagInput.trim()) {
                            handleSelectTag(newTagInput.trim());
                          }
                        }
                      }}
                      placeholder="Add tags (comma separated)"
                      className="flex-grow"
                    />
                    <Button 
                      onClick={(e) => {
                        e.stopPropagation();
                        handleAddTag();
                      }}
                      disabled={isLoading || (!newTagInput.trim() && selectedTags.length === 0)}
                      size="sm" 
                      type="button"
                    >
                      <PlusIcon className="h-4 w-4 mr-1" />
                      Add
                    </Button>
                  </div>
                  
                  {/* Tag suggestions */}
                  {suggestedTags.length > 0 && (
                    <div className="mt-2">
                      <p className="text-xs text-muted-foreground mb-1">Suggested tags:</p>
                      <div className="flex flex-wrap gap-1.5">
                        {suggestedTags.map((tag, index) => (
                          <Badge 
                            key={index} 
                            variant="outline" 
                            className="bg-slate-100 dark:bg-slate-800 cursor-pointer hover:bg-slate-200 dark:hover:bg-slate-700"
                            onClick={(e) => {
                              e.stopPropagation();
                              handleSelectTag(tag);
                            }}
                          >
                            {tag}
                          </Badge>
                        ))}
                      </div>
                    </div>
                  )}
                  
                  <p className="mt-1 text-xs text-muted-foreground">
                    Separate multiple tags with commas, press Enter to quickly add
                  </p>
                </div>
              </div>
              <DialogFooter className="gap-2">
                <Button 
                  type="button" 
                  variant="secondary"
                  onClick={(e) => {
                    e.stopPropagation();
                    setTagsDialogOpen(false);
                  }}
                >
                  Close
                </Button>
                <Button 
                  type="button" 
                  onClick={(e) => {
                    e.stopPropagation();
                    handleAddTag();
                    if (newTagInput.trim() || selectedTags.length > 0) {
                      setTagsDialogOpen(false);
                    }
                  }}
                  disabled={isLoading || (!newTagInput.trim() && selectedTags.length === 0)}
                >
                  Confirm
                </Button>
              </DialogFooter>
            </DialogContent>
          </Dialog>

          {/* Category selection - modified for better event handling */}
          <Popover 
            open={categoryPopoverOpen} 
            onOpenChange={(isOpen) => {
              console.log("Category Popover onOpenChange called. isOpen:", isOpen, "Current categoryPopoverOpen before set:", categoryPopoverOpen);
              setCategoryPopoverOpen(isOpen);
              console.log("Category Popover onOpenChange: categoryPopoverOpen set to:", isOpen);
              if (isOpen) { 
                setTagsDialogOpen(false);
                console.log("Category Popover onOpenChange: tagsDialogOpen set to false because Popover is opening.");
              }
            }}
          >
            <PopoverTrigger asChild>
              <Button
                ref={categoryTriggerRef}
                onClick={() => {
                  console.log("Category <Button> (inside PopoverTrigger) clicked. Current isExpanded:", isExpanded, "Current categoryPopoverOpen:", categoryPopoverOpen);
                  // Let Popover's onOpenChange handle the state.
                }}
                disabled={isLoading}
                className="h-10 w-10 rounded-full shadow-md bg-green-600 hover:bg-green-700"
                title="Change category"
              >
                <FolderIcon className="h-5 w-5" />
              </Button>
            </PopoverTrigger>
            <PopoverContent 
              ref={categoryPopoverContentRef}
              className="p-0 w-[250px]" 
              align="end" 
              side="top"
              sideOffset={5}
              onInteractOutside={(e) => {
                // Default behavior is to close on interact outside.
              }}
            >
              {/* --- END TEMPORARY TEST --- */}
              {/* <div className="p-4">Test: Category Popover Should Be Visible</div> */}
              
              <Command>
                <CommandInput 
                  id="category-search-input"
                  placeholder="Search categories..." 
                  className="h-9" 
                />
                <CommandList className="max-h-[300px]">
                  <CommandEmpty>No categories found</CommandEmpty>
                  <CommandGroup heading="Categories">
                    <ScrollArea className="h-[200px]">
                      {categories.map((category) => (
                        <CommandItem
                          key={category.slug}
                          value={category.name}
                          onSelect={() => {
                            handleSelectCategory(category.name);
                          }}
                          onMouseDown={(e) => {
                            // Prevent default to avoid focus issues
                            e.preventDefault();
                          }}
                          className="cursor-pointer"
                        >
                          <span className="flex-1">{category.name}</span>
                          {currentCategoryName === category.name && (
                            <CheckIcon className="ml-auto h-4 w-4 text-green-600" />
                          )}
                        </CommandItem>
                      ))}
                    </ScrollArea>
                  </CommandGroup>
                  
                  <CommandSeparator />
                  
                  <CommandGroup heading="Create New Category">
                    <div className="flex items-center gap-2 px-2 py-1.5">
                      <Input
                        id="new-category-name-input"
                        value={newCategoryName}
                        onChange={(e) => setNewCategoryName(e.target.value)}
                        placeholder="Enter new category name"
                        className="h-8"
                        onKeyDown={(e) => {
                          if (e.key === 'Enter' && newCategoryName.trim()) {
                            e.preventDefault();
                            e.stopPropagation();
                            handleCreateCategory();
                          }
                        }}
                        onClick={(e) => {
                          // Prevent propagation to keep popover open
                          e.stopPropagation();
                        }}
                      />
                      <Button
                        size="sm"
                        className="h-8 px-2"
                        type="button"
                        onClick={(e) => {
                          e.stopPropagation();
                          handleCreateCategory();
                        }}
                        disabled={!newCategoryName.trim() || isCreatingCategory}
                      >
                        {isCreatingCategory ? (
                          <Loader2Icon className="h-4 w-4 animate-spin" />
                        ) : (
                          <PlusIcon className="h-4 w-4" />
                        )}
                      </Button>
                    </div>
                  </CommandGroup>
                </CommandList>
              </Command>
              
            </PopoverContent>
          </Popover>
          
          {/* Rescrape button */}
          <Button
            onClick={handleRescrape}
            disabled={isLoading}
            className="h-10 w-10 rounded-full shadow-md bg-purple-600 hover:bg-purple-700"
            title="Rescrape article"
          >
            <RefreshCwIcon className="h-5 w-5" />
          </Button>
        </div>
      )}
    </div>
  );
} 