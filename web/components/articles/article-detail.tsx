"use client";

import Link from "next/link";
import Image from "next/image";
import { Button } from "@/components/ui/button";
import { CalendarIcon, TagIcon, BookmarkIcon, ArrowLeftIcon, StarIcon, EyeIcon, RefreshCwIcon, UserIcon } from "lucide-react";
import { format } from "date-fns";
import { DetailedArticle } from "@/types/article";
import React, { useState, useEffect, useRef, useCallback } from "react";
import parse, { 
  domToReact, 
  Element as ParserElement, 
  HTMLReactParserOptions,
  attributesToProps, 
  DOMNode
} from "html-react-parser";
import { Prism as SyntaxHighlighter } from "react-syntax-highlighter";
import { oneDark, oneLight } from "react-syntax-highlighter/dist/cjs/styles/prism";
import { useTheme } from "next-themes";
import FloatingActionToolbar from "@/components/common/FloatingActionToolbar";
import { articleAPI, categoryAPI, tagAPI } from "@/lib/api-client";
import { useToast } from "@/components/ui/use-toast";

interface TagObject {
  name: string;
  slug?: string;
}

// Pre-process HTML content to remove conflicting tags
const sanitizeHtml = (html: string): string => {
  if (!html) return '';
  
  // Remove DOCTYPE declarations
  let sanitized = html.replace(/<!DOCTYPE[^>]*>/i, '');
  
  // Remove <html> tags but keep content
  sanitized = sanitized.replace(/<html[^>]*>([\s\S]*?)<\/html>/i, '$1');
  
  // Remove <head> tags and their content
  sanitized = sanitized.replace(/<head[\s\S]*?<\/head>/i, '');
  
  // Remove <body> tags but keep content
  sanitized = sanitized.replace(/<body[^>]*>([\s\S]*?)<\/body>/i, '$1');
  
  return sanitized;
};

// Recursively extract content from nested code blocks
const extractCodeContent = (nodes: any[]): string => {
  // Extract all text and normalize line breaks
  const extractText = (node: any): string => {
    // Text node case
    if (node.type === 'text') {
      return node.data || '';
    }
    
    // Element with children
    if (node.children && node.children.length > 0) {
      return node.children.map(extractText).join('');
    }
    
    return '';
  };
  
  // Get raw text content
  let content = '';
  for (const node of nodes) {
    content += extractText(node);
  }
  
  // Normalize line breaks and remove excess empty lines
  return content
    .replace(/\r\n/g, '\n')      // Normalize Windows line endings
    .replace(/\r/g, '\n')        // Normalize Mac line endings
    .replace(/\n{3,}/g, '\n\n')  // Replace 3+ consecutive newlines with 2
    .trim();                     // Remove leading/trailing whitespace
};

interface ArticleDetailProps {
  article: DetailedArticle | null;
  isLoading?: boolean;
  onRefetch?: () => Promise<void>;
}

export default function ArticleDetail({ article, isLoading = false, onRefetch }: ArticleDetailProps) {
  const [isRefetching, setIsRefetching] = useState(false);
  const [currentArticle, setCurrentArticle] = useState<DetailedArticle | null>(article);
  const [categories, setCategories] = useState<Array<{name: string, slug: string}>>([]);
  const [availableTags, setAvailableTags] = useState<string[]>([]);
  const [isLoadingCategories, setIsLoadingCategories] = useState(false);
  // Track fields being updated to prevent flickering
  const [updatingFields, setUpdatingFields] = useState<{
    is_read?: boolean;
    is_starred?: boolean;
    category?: string;
    tags?: boolean;
    rescrape?: boolean;
  } | null>(null);
  const contentRef = useRef<HTMLDivElement>(null);
  const { theme, systemTheme } = useTheme();
  const { toast } = useToast();
  
  // Determine if dark mode is active
  const isDarkMode = 
    theme === 'dark' || 
    (theme === 'system' && systemTheme === 'dark') || 
    (typeof document !== 'undefined' && document.documentElement.classList.contains('dark'));
  
  useEffect(() => {
    // Update current article when prop changes, but intelligently
    if (article) {
      console.log('ArticleDetail: Article prop changed, updating currentArticle:', {
        id: article.id,
        isRead: article.isRead, 
        isStarred: article.isStarred
      });
      
      // Only update if we're not in the middle of our own update
      if (!updatingFields) {
        setCurrentArticle(article);
      } else {
        console.log('ArticleDetail: Ignoring article update while making local changes', updatingFields);
      }
    }
    
    // Fetch categories for the dropdown
    const fetchCategories = async () => {
      if (!article) return;
      
      try {
        setIsLoadingCategories(true);
        const response = await categoryAPI.listCategories(1, 50);
        if (response?.data?.data) {
          setCategories(response.data.data.map((cat: any) => ({
            name: cat.name,
            slug: cat.slug
          })));
        }
      } catch (error) {
        console.error('Failed to fetch categories:', error);
        toast({
          title: "Error fetching categories",
          description: "Please try again later",
          variant: "destructive"
        });
      } finally {
        setIsLoadingCategories(false);
      }
    };

    // Fetch all available tags
    const fetchTags = async () => {
      if (!article) return;
      
      try {
        const response = await tagAPI.listTags(1, 100);
        if (response?.data?.data) {
          // Extract tag names array
          const tagNames = response.data.data.map((tag: any) => 
            typeof tag === 'string' ? tag : (tag.name || '')
          ).filter(Boolean);
          setAvailableTags(tagNames);
        }
      } catch (error) {
        console.error('Failed to fetch tags:', error);
        // No error display as this is not a critical feature
      }
    };
    
    fetchCategories();
    fetchTags();
  }, [article, toast, updatingFields]);
  
  // Handle article status updates with proper sequence
  const handleUpdateStatus = useCallback(async (status: { is_read?: boolean; is_starred?: boolean }) => {
    console.log("[ArticleDetail] handleUpdateStatus called with status:", status, "and currentArticle ID:", currentArticle?.id);

    if (!currentArticle) {
      console.error('ArticleDetail: Cannot update status, currentArticle is null');
      console.log("[ArticleDetail] currentArticle is null, returning early from handleUpdateStatus.");
      return;
    }
    
    try {
      setIsRefetching(true);
      
      setUpdatingFields(status);
      // console.log('ArticleDetail: Updating status locally:', status);
      
      setCurrentArticle(prev => {
        if (!prev) return null;
        return {
          ...prev,
          isRead: status.is_read !== undefined ? status.is_read : prev.isRead,
          isStarred: status.is_starred !== undefined ? status.is_starred : prev.isStarred
        };
      });
      
      console.log("[ArticleDetail] Attempting to call articleAPI.updateArticleStatus with ID:", currentArticle.id, "and status:", status);
      await articleAPI.updateArticleStatus(currentArticle.id, status);
      console.log("[ArticleDetail] articleAPI.updateArticleStatus call completed.");
      
      toast({
        title: status.is_read !== undefined 
          ? `Article marked as ${status.is_read ? "read" : "unread"}`
          : `Article ${status.is_starred ? "starred" : "unstarred"}`,
      });
      
      const updatedState = {
        isRead: status.is_read !== undefined ? status.is_read : currentArticle.isRead,
        isStarred: status.is_starred !== undefined ? status.is_starred : currentArticle.isStarred
      };
      
      // Refresh article data - TEMPORARILY COMMENT OUT FOR DEBUGGING STATUS UPDATE
      // if (onRefetch) {
      //   await onRefetch();
      // }
      
      setCurrentArticle(prev => {
        if (!prev) return null;
        return {
          ...prev,
          ...updatedState
        };
      });
      
    } catch (error) {
      console.error("[ArticleDetail] Failed to update article status (inside catch block):", error, "for article ID:", currentArticle?.id, "with status:", status);
      
      if (article) {
        setCurrentArticle(prev => {
          if (!prev) return article;
          return {
            ...prev,
            isRead: status.is_read !== undefined ? article.isRead : prev.isRead,
            isStarred: status.is_starred !== undefined ? article.isStarred : prev.isStarred
          };
        });
      }
      
      toast({
        title: "Update failed",
        description: "Could not update article status",
        variant: "destructive"
      });
    } finally {
      setIsRefetching(false);
      setUpdatingFields(null);
      console.log("[ArticleDetail] handleUpdateStatus finally block executed.");
    }
  }, [currentArticle, toast, onRefetch, article]);
  
  // Handle updating article category with improved state management
  const handleUpdateCategory = useCallback(async (categoryName: string) => {
    if (!currentArticle) return;
    
    try {
      setIsRefetching(true);
      
      // Mark that we're updating category
      setUpdatingFields({ category: categoryName });
      console.log('ArticleDetail: Updating category locally:', categoryName);
      
      // Update local state immediately
      setCurrentArticle(prev => {
        if (!prev) return null;
        return {
          ...prev,
          category: categoryName
        };
      });
      
      // Call API to update category
      await articleAPI.updateArticleCategory(currentArticle.id, categoryName);
      
      // Show success notification
      toast({
        title: "Category updated",
        description: `Article moved to "${categoryName}"`,
      });
      
      // Refresh article data
      if (onRefetch) {
        await onRefetch();
      }
      
      // Ensure our local state reflects our update
      setCurrentArticle(prev => {
        if (!prev) return null;
        return {
          ...prev,
          category: categoryName
        };
      });
      
    } catch (error) {
      console.error('ArticleDetail: Failed to update article category:', error);
      
      // Restore original state
      if (article) {
        setCurrentArticle(prev => {
          if (!prev) return article;
          // Only restore the category field
          return {
            ...prev,
            category: article.category
          };
        });
      }
      
      toast({
        title: "Update failed",
        description: "Could not update article category",
        variant: "destructive"
      });
    } finally {
      setIsRefetching(false);
      setUpdatingFields(null);
    }
  }, [currentArticle, article, toast, onRefetch]);
  
  // Handle adding tags to article
  const handleAddTags = useCallback(async (tags: string[]) => {
    if (!currentArticle || tags.length === 0) return;
    
    try {
      setIsRefetching(true);
      
      // Mark that we're updating tags
      setUpdatingFields({ tags: true });
      console.log('ArticleDetail: Adding tags locally:', tags);
      
      // Extract existing tags
      const existingTags = Array.isArray(currentArticle.tags) 
        ? currentArticle.tags.map(tag => typeof tag === 'string' ? tag : (tag as any)?.name || '')
        : [];
      
      // Update local state immediately
      const updatedTags = Array.from(new Set([...existingTags, ...tags]));
      setCurrentArticle(prev => {
        if (!prev) return null;
        
        return {
          ...prev,
          tags: updatedTags
        };
      });
      
      // Call API to add tags
      await articleAPI.addTags(currentArticle.id, tags);
      
      // Show success notification
      toast({
        title: "Tags added",
        description: `Added ${tags.length} tag${tags.length > 1 ? 's' : ''}`,
      });
      
      // Refresh article data
      if (onRefetch) {
        await onRefetch();
      }
      
      // Ensure our local state reflects our update
      setCurrentArticle(prev => {
        if (!prev) return null;
        return {
          ...prev,
          tags: updatedTags
        };
      });
      
    } catch (error) {
      console.error('ArticleDetail: Failed to add tags:', error);
      
      // Restore original state
      if (article) {
        setCurrentArticle(prev => {
          if (!prev) return article;
          return {
            ...prev,
            tags: article.tags
          };
        });
      }
      
      toast({
        title: "Update failed",
        description: "Could not add tags",
        variant: "destructive"
      });
    } finally {
      setIsRefetching(false);
      setUpdatingFields(null);
    }
  }, [currentArticle, article, toast, onRefetch]);
  
  // Handle removing tags from article
  const handleRemoveTags = useCallback(async (tags: string[]) => {
    if (!currentArticle || tags.length === 0) return;
    
    try {
      setIsRefetching(true);
      
      // Mark that we're updating tags
      setUpdatingFields({ tags: true });
      console.log('ArticleDetail: Removing tags locally:', tags);
      
      // Extract existing tags
      const existingTags = Array.isArray(currentArticle.tags) 
        ? currentArticle.tags.map(tag => typeof tag === 'string' ? tag : (tag as any)?.name || '')
        : [];
      
      // Update local state immediately
      const updatedTags = existingTags.filter(tag => !tags.includes(tag));
      setCurrentArticle(prev => {
        if (!prev) return null;
        
        return {
          ...prev,
          tags: updatedTags
        };
      });
      
      // Call API to remove tags
      await articleAPI.removeTags(currentArticle.id, tags);
      
      // Show success notification
      toast({
        title: "Tags removed",
        description: `Removed ${tags.length} tag${tags.length > 1 ? 's' : ''}`,
      });
      
      // Refresh article data
      if (onRefetch) {
        await onRefetch();
      }
      
      // Ensure our local state reflects our update
      setCurrentArticle(prev => {
        if (!prev) return null;
        return {
          ...prev,
          tags: updatedTags
        };
      });
      
    } catch (error) {
      console.error('ArticleDetail: Failed to remove tags:', error);
      
      // Restore original state
      if (article) {
        setCurrentArticle(prev => {
          if (!prev) return article;
          return {
            ...prev,
            tags: article.tags
          };
        });
      }
      
      toast({
        title: "Update failed",
        description: "Could not remove tags",
        variant: "destructive"
      });
    } finally {
      setIsRefetching(false);
      setUpdatingFields(null);
    }
  }, [currentArticle, article, toast, onRefetch]);
  
  // Handle article rescrape with improved state management
  const handleRescrape = useCallback(async () => {
    // Add error log to detect unexpected calls
    console.error("[ArticleDetail] handleRescrape was called! Check if this was intended (e.g., via the rescrape button) or unexpected (e.g., during status update).");

    if (!currentArticle) return;
    
    try {
      setIsRefetching(true);
      
      // We don't need to update any fields locally for rescrape
      // but we'll mark it to prevent race conditions
      setUpdatingFields({ rescrape: true });
      console.log('ArticleDetail: Initiating rescrape for article:', currentArticle.id);
      
      // Call API to rescrape article
      await articleAPI.rescrapeArticle(currentArticle.id);
      
      // Show success notification
      toast({
        title: "Rescrape initiated",
        description: "Article content is being refreshed in background",
      });
      
      // Refresh article data
      if (onRefetch) {
        await onRefetch();
      }
      
    } catch (error) {
      console.error('ArticleDetail: Failed to rescrape article:', error);
      
      toast({
        title: "Rescrape failed",
        description: "Could not refresh article content",
        variant: "destructive"
      });
    } finally {
      setIsRefetching(false);
      setUpdatingFields(null);
    }
  }, [currentArticle, toast, onRefetch]);

  // Wrap onRefetch with logging
  const handleRefetch = useCallback(async () => {
    console.log("ArticleDetail: Starting manual refetch");
    setIsRefetching(true);
    
    try {
      if (onRefetch) {
        console.log("ArticleDetail: Calling parent onRefetch function");
        await onRefetch();
        console.log("ArticleDetail: Parent onRefetch completed successfully");
        
        // Check if we need to update current article state again
        if (article && currentArticle && (
          article.isRead !== currentArticle.isRead || 
          article.isStarred !== currentArticle.isStarred
        )) {
          console.log("ArticleDetail: Detected state difference after refetch, updating currentArticle");
          console.log("ArticleDetail: New article state:", {
            isRead: article.isRead,
            isStarred: article.isStarred
          });
          console.log("ArticleDetail: Current article state:", {
            isRead: currentArticle.isRead,
            isStarred: currentArticle.isStarred
          });
          setCurrentArticle(article);
        }
      } else {
        console.log("ArticleDetail: No onRefetch function provided");
      }
    } catch (error) {
      console.error("ArticleDetail: Error during refetch:", error);
    } finally {
      setIsRefetching(false);
      console.log("ArticleDetail: Manual refetch completed");
    }
  }, [onRefetch, article, currentArticle]);

  // Parse HTML content with custom handlers
  const parseOptions: HTMLReactParserOptions = {
    replace: (domNode) => {
      if (!(domNode as any).type) return undefined;
      
      if ((domNode as ParserElement).type === 'tag') {
        const element = domNode as ParserElement;
        const { name, attribs, children } = element;
        
        // Filter out html and body tags to prevent rendering errors
        if (name === 'html' || name === 'body') {
          return <>{domToReact(children as DOMNode[], parseOptions)}</>;
        }
        
        const props = attributesToProps(attribs);
        
        // Handle code blocks inside pre tags
        if (name === 'pre') {
          // Look for code tags
          const codeElement = children.find(child => 
            (child as ParserElement).name === 'code'
          ) as ParserElement | undefined;
          
          if (codeElement) {
            // Get language information
            // 1. From pre tag's data-language attribute
            let language = attribs['data-language'] || '';
            
            // 2. From code tag's class
            if (!language && codeElement.attribs && codeElement.attribs.class) {
              const langMatch = codeElement.attribs.class.match(/language-(\w+)/);
              if (langMatch) language = langMatch[1];
            }
            
            // 3. From pre tag's class
            if (!language && attribs.class) {
              // Check if there's a language indicator
              const classLangMatch = attribs.class.match(/language-(\w+)/);
              if (classLangMatch) language = classLangMatch[1];
            }
            
            // Also try to detect language from class like "astro-code one-dark-pro" with data-language attribute
            if (!language && attribs.class && attribs['data-language']) {
              language = attribs['data-language'];
            }
            
            // If still no language, try to find other common language indicators
            if (!language && attribs.class) {
              const knownLanguages = ['js', 'ts', 'jsx', 'tsx', 'html', 'css', 'go', 'rust', 'python', 'java', 'c', 'cpp'];
              for (const lang of knownLanguages) {
                if (attribs.class.includes(lang)) {
                  language = lang;
                  break;
                }
              }
            }
            
            // If no language found, set to plaintext
            if (!language) language = 'plaintext';
            
            // Extract code content based on structure
            let codeContent = '';
            
            // Special handling for span.line structured code blocks
            const hasLineSpans = codeElement.children.some(child => 
              (child as ParserElement).name === 'span' && 
              (child as ParserElement).attribs?.class?.includes('line')
            );
            
            if (hasLineSpans) {
              // Process line by line for span.line structure
              codeContent = codeElement.children
                .filter(child => (child as ParserElement).name === 'span')
                .map((lineSpan, index, arr) => {
                  const lineContent = extractTextContent(lineSpan);
                  return lineContent + (index < arr.length - 1 ? '\n' : '');
                })
                .join('');
            } else {
              // Process simple text or other structures
              codeContent = extractCodeContent(codeElement.children);
            }
            
            return (
              <div className="code-block my-6 rounded-lg overflow-hidden shadow-sm">
                <div className="code-lang-label bg-slate-800 text-gray-300 text-xs py-1 px-4 font-mono flex justify-between items-center border-b border-slate-700">
                  <span>{language}</span>
                  <button 
                    className="copy-code-button text-xs py-0.5 px-2 bg-slate-700 border-none rounded cursor-pointer text-gray-300 hover:bg-slate-600"
                    onClick={() => {
                      navigator.clipboard.writeText(codeContent);
                      // Toast or feedback could be added here
                    }}
                  >
                    Copy
                  </button>
                </div>
                <SyntaxHighlighter
                  language={language}
                  style={isDarkMode ? oneDark : oneLight}
                  customStyle={{ margin: 0, borderRadius: 0, maxHeight: '500px' }}
                  showLineNumbers
                >
                  {codeContent}
                </SyntaxHighlighter>
              </div>
            );
          }
        }
        
        // Handle inline code
        if (name === 'code') {
          // Make sure it's not a code block inside a pre tag
          const parentElement = element.parent as ParserElement | null;
          if (!parentElement || parentElement.name !== 'pre') {
            // Solve type issue through explicit type conversion
            const codeContent = domToReact(children as DOMNode[], parseOptions);
            
            return (
              <code className="font-mono text-sm">
                <span
                  className={`px-1.5 py-0.5 rounded ${
                    isDarkMode 
                      ? "bg-pink-900/30 text-pink-300" 
                      : "bg-pink-50 text-pink-600"
                  }`}
                >
                  {codeContent}
                </span>
              </code>
            );
          }
        }
        
        // Add anchor links to headings
        if (['h1', 'h2', 'h3', 'h4', 'h5', 'h6'].includes(name)) {
          // Solve type issue through explicit type conversion
          const content = domToReact(children as DOMNode[], parseOptions);
          // Use String wrapper to ensure toString works properly
          const headingText = String(content);
          const id = headingText
            .toLowerCase()
            .replace(/\s+/g, '-')
            .replace(/[^a-z0-9-]/g, '') || `heading-${Math.random().toString(36).substring(2, 9)}`;
          
          // Apply appropriate heading styles based on heading level
          let headingClasses = '';
          
          switch (name) {
            case 'h1':
              headingClasses = 'text-3xl font-bold mt-8 mb-4';
              break;
            case 'h2':
              headingClasses = 'text-2xl font-semibold mt-6 mb-3 pb-1 border-b';
              break;
            case 'h3':
              headingClasses = 'text-xl font-semibold mt-5 mb-2';
              break;
            default:
              headingClasses = 'text-lg font-medium mt-4 mb-2';
          }
          
          const HeadingTag = name as keyof JSX.IntrinsicElements;
          
          return (
            <HeadingTag
              id={id}
              className={headingClasses}
              onMouseEnter={(e) => {
                const anchor = e.currentTarget.querySelector('.heading-anchor');
                if (anchor) (anchor as HTMLElement).style.opacity = '0.5';
              }}
              onMouseLeave={(e) => {
                const anchor = e.currentTarget.querySelector('.heading-anchor');
                if (anchor) (anchor as HTMLElement).style.opacity = '0';
              }}
            >
              {content}
              <a 
                href={`#${id}`} 
                className="heading-anchor ml-2 opacity-0 transition-opacity duration-200 text-inherit no-underline"
                aria-hidden="true"
              >
                🔗
              </a>
            </HeadingTag>
          );
        }
        
        // Handle links
        if (name === 'a') {
          const href = typeof props.href === 'string' ? props.href : '';
          // Solve type issue through explicit type conversion
          const content = domToReact(children as DOMNode[], parseOptions);
          
          if (href && (href.startsWith('http') || href.startsWith('www'))) {
            return (
              <a 
                href={href} 
                target="_blank" 
                rel="noopener noreferrer"
                className="text-blue-600 dark:text-blue-400 hover:text-blue-800 dark:hover:text-blue-300 underline"
              >
                {content}
                <span className="external-icon text-[0.8em]"> ↗</span>
              </a>
            );
          }
          
          return (
            <a 
              href={href} 
              className={`${href?.startsWith('#') 
                ? 'text-primary hover:text-primary/80 underline no-underline-offset' 
                : 'text-primary hover:text-primary/80 underline'}`
              }
            >
              {content}
            </a>
          );
        }
        
        // Handle images
        if (name === 'img') {
          return (
            <img 
              {...props}
              className="max-w-full h-auto my-6 rounded-md shadow-sm border border-muted" 
              loading="lazy"
            />
          );
        }
        
        // Handle blockquotes
        if (name === 'blockquote') {
          return (
            <blockquote className="border-l-4 border-pink-500 dark:border-pink-600 pl-4 pr-4 py-3 my-6 italic text-pink-700/90 dark:text-pink-300/90 rounded-r-md bg-pink-50 dark:bg-pink-900/30">
              {domToReact(children as DOMNode[], parseOptions)}
            </blockquote>
          );
        }
        
        // Wrap tables for better mobile support
        if (name === 'table') {
          return (
            <div className="overflow-x-auto my-6 rounded-lg border border-muted">
              <table className="min-w-full border-collapse">
                {domToReact(children as DOMNode[], parseOptions)}
              </table>
            </div>
          );
        }
        
        // Style table headers
        if (name === 'th') {
          return (
            <th {...props} className="border border-muted p-3 bg-muted/50 font-semibold text-left text-sm">
              {domToReact(children as DOMNode[], parseOptions)}
            </th>
          );
        }
        
        // Style table cells
        if (name === 'td') {
          return (
            <td {...props} className="border border-muted p-3 text-sm">
              {domToReact(children as DOMNode[], parseOptions)}
            </td>
          );
        }
        
        // Add spacing to paragraphs
        if (name === 'p') {
          return (
            <p className="my-4 leading-relaxed text-base break-words">
              {domToReact(children as DOMNode[], parseOptions)}
            </p>
          );
        }
        
        // Style lists
        if (name === 'ul' || name === 'ol') {
          const ListTag = name as 'ul' | 'ol';
          return (
            <ListTag
              {...props} 
              className={`my-4 pl-6 space-y-1.5 text-base ${name === 'ul' ? 'list-disc' : 'list-decimal'}`}
            >
              {domToReact(children as DOMNode[], parseOptions)}
            </ListTag>
          );
        }
        
        // Add line height to list items
        if (name === 'li') {
          return (
            <li className="leading-relaxed">
              {domToReact(children as DOMNode[], parseOptions)}
            </li>
          );
        }
      }
      return undefined;
    }
  };

  if (isLoading) {
    return (
      <div className="container mx-auto max-w-4xl px-4 py-8">
        <div className="mb-6 h-10 w-3/4 bg-gray-200 dark:bg-gray-700 animate-pulse rounded" />
        <div className="h-6 w-full bg-gray-200 dark:bg-gray-700 animate-pulse rounded mb-4" />
        <div className="h-6 w-5/6 bg-gray-200 dark:bg-gray-700 animate-pulse rounded mb-8" />
        <div className="space-y-3">
          {[...Array(5)].map((_, i) => <div key={i} className="h-4 bg-gray-200 dark:bg-gray-700 animate-pulse rounded" />)}
        </div>
      </div>
    );
  }

  if (!currentArticle && !isLoading) {
    return (
      <div className="container mx-auto max-w-4xl px-4 py-8 text-center">
        <BookmarkIcon className="h-16 w-16 text-muted-foreground/50 mx-auto" />
        <h2 className="mt-4 text-2xl font-semibold">Article Not Found</h2>
        <p className="mt-2 text-muted-foreground">The article you are looking for does not exist or has been removed.</p>
        <Button asChild className="mt-6"><Link href="/articles"><ArrowLeftIcon className="mr-2 h-4 w-4" />Back to Articles</Link></Button>
      </div>
    );
  }

  if (!currentArticle) {
    return (
      <div className="container mx-auto max-w-4xl px-4 py-8">
        <p className="text-center text-muted-foreground">Loading article...</p>
      </div>
    );
  }

  const formattedDate = currentArticle.createdAt ? format(new Date(currentArticle.createdAt), "yyyy-MM-dd HH:mm") : "Unknown date";
  const articleTags = Array.isArray(currentArticle.tags)
    ? currentArticle.tags.map(tag => typeof tag === 'string' ? tag : (tag as any)?.name || '').filter(Boolean)
    : [];

  return (
    <div className="container mx-auto max-w-4xl px-4 py-8">
      <div className="mb-6">
        <Button variant="ghost" asChild><Link href="/articles"><ArrowLeftIcon className="mr-2 h-4 w-4" />Back to Articles</Link></Button>
      </div>

      <article className="space-y-6">
        <div className="mb-8">
          <h1 className="text-3xl font-bold md:text-4xl lg:text-5xl break-words mb-6">{currentArticle.title || 'Untitled Article'}</h1>
          
          <div className="flex flex-wrap items-center gap-4 text-sm text-muted-foreground mt-4 border-t border-b py-4">
            {currentArticle.author && (
              <span className="font-medium flex items-center">
                <UserIcon className="mr-1.5 h-4 w-4" />
                {currentArticle.author}
              </span>
            )}
            {formattedDate !== "Unknown date" && (
              <span className="flex items-center">
                <CalendarIcon className="mr-1.5 h-4 w-4" />
                <time dateTime={currentArticle.createdAt?.toString()}>{formattedDate}</time>
              </span>
            )}
            
            {currentArticle.category && (
              <div className="flex items-center text-blue-600 dark:text-blue-400">
                <TagIcon className="mr-1.5 h-4 w-4" />
                <span className="font-medium">{typeof currentArticle.category === 'string' ? currentArticle.category : (currentArticle.category as any)?.name || ''}</span>
              </div>
            )}
            
            {currentArticle.isRead && (
              <div className="flex items-center text-green-600">
                <EyeIcon className="mr-1.5 h-4 w-4" />
                <span>Read</span>
              </div>
            )}
            
            {currentArticle.isStarred && (
              <div className="flex items-center text-yellow-500">
                <StarIcon className="mr-1.5 h-4 w-4" />
                <span>Starred</span>
              </div>
            )}
          </div>
          
          {articleTags.length > 0 && (
            <div className="flex flex-wrap items-center gap-2 mt-4">
              <span className="font-medium text-muted-foreground mr-2">Tags:</span>
              <div className="flex flex-wrap gap-2">
                {articleTags.map((tag, index) => (
                  <span key={index} className="inline-flex items-center rounded-full bg-secondary px-2.5 py-0.5 text-xs font-medium text-secondary-foreground">
                    {tag}
                  </span>
                ))}
              </div>
            </div>
          )}
        </div>
        
        {onRefetch && (
          <Button variant="outline" size="sm" className="mb-6" disabled={isRefetching}
            onClick={handleRefetch}
          >
            <RefreshCwIcon className={`mr-2 h-4 w-4 ${isRefetching ? 'animate-spin' : ''}`} />
            {isRefetching ? "Refreshing..." : "Refresh Content"}
          </Button>
        )}

        {currentArticle.ogImage && (
          <div className="relative aspect-[16/9] w-full overflow-hidden rounded-lg mb-6 shadow-lg">
            <Image 
              src={currentArticle.ogImage} 
              alt={currentArticle.title || 'Article Image'} 
              fill 
              sizes="(max-width: 768px) 100vw, (max-width: 1200px) 50vw, 33vw"
              className="object-cover"
              priority={true}
              onError={(e: React.SyntheticEvent<HTMLImageElement, Event>) => { (e.currentTarget as HTMLImageElement).style.display = 'none'; }}
             />
          </div>
        )}

        <div 
          ref={contentRef} 
          className="prose prose-quoteless prose-neutral dark:prose-invert max-w-none article-html-content selection:bg-primary/20"
        >
          {currentArticle.html ? parse(sanitizeHtml(currentArticle.html), parseOptions) : (
            <p className="text-muted-foreground">No content available for this article.</p>
          )}
        </div>
      </article>
      
      {/* Add Floating Action Toolbar */}
      {currentArticle && (
        <FloatingActionToolbar
          article={currentArticle}
          categories={categories}
          existingTags={availableTags}
          onUpdateStatus={handleUpdateStatus}
          onUpdateCategory={handleUpdateCategory}
          onAddTags={handleAddTags}
          onRemoveTags={handleRemoveTags}
          onRescrape={handleRescrape}
          onRefresh={handleRefetch}
        />
      )}
    </div>
  );
}

// Simple helper to extract only text from an element
const extractTextContent = (node: any): string => {
  if (!node) return '';
  
  if (node.type === 'text') {
    return node.data || '';
  }
  
  if (node.children && node.children.length > 0) {
    return node.children.map(extractTextContent).join('');
  }
  
  return '';
};