'use client';

import React, { useState, useEffect } from 'react';
import { articleAPI, categoryAPI, tagAPI } from '../lib/api-client';
import { PlusIcon } from 'lucide-react';
import { useRouter } from 'next/navigation';

// Interface for category and tag data
interface Category {
  id: string;
  name: string;
}

interface Tag {
  id: string;
  name: string;
}

const AddArticleForm: React.FC = () => {
  const router = useRouter();
  
  // Form state
  const [url, setUrl] = useState<string>('');
  const [selectedCategory, setSelectedCategory] = useState<string>('');
  const [selectedTags, setSelectedTags] = useState<string[]>([]);
  const [customTag, setCustomTag] = useState<string>('');
  
  // Data state
  const [categories, setCategories] = useState<Category[]>([]);
  const [tags, setTags] = useState<Tag[]>([]);
  const [isLoading, setIsLoading] = useState<boolean>(false);
  const [error, setError] = useState<string | null>(null);

  // Fetch categories and tags on component mount
  useEffect(() => {
    const fetchData = async () => {
      try {
        // Fetch categories
        const categoriesResponse = await categoryAPI.listCategories(1, 100);
        if (categoriesResponse && categoriesResponse.data) {
          // Handle different response data structures
          let categoryData = [];
          if (categoriesResponse.data.data && Array.isArray(categoriesResponse.data.data)) {
            categoryData = categoriesResponse.data.data;
          } else if (categoriesResponse.data.items && Array.isArray(categoriesResponse.data.items)) {
            categoryData = categoriesResponse.data.items;
          } else if (Array.isArray(categoriesResponse.data)) {
            categoryData = categoriesResponse.data;
          }
          setCategories(categoryData);
        } else {
          setCategories([]);
        }
        
        // Fetch tags
        const tagsResponse = await tagAPI.listTags(1, 100);
        if (tagsResponse && tagsResponse.data) {
          // Handle different response data structures
          let tagData = [];
          if (tagsResponse.data.data && Array.isArray(tagsResponse.data.data)) {
            tagData = tagsResponse.data.data;
          } else if (tagsResponse.data.items && Array.isArray(tagsResponse.data.items)) {
            tagData = tagsResponse.data.items;
          } else if (Array.isArray(tagsResponse.data)) {
            tagData = tagsResponse.data;
          }
          setTags(tagData);
        } else {
          setTags([]);
        }
      } catch (err) {
        console.error('Error fetching categories or tags:', err);
        setError('Failed to load categories or tags. Please try again.');
        setCategories([]);
        setTags([]);
      }
    };
    
    fetchData();
  }, []);

  // Handle form submission
  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    
    // Validate URL
    if (!url.trim()) {
      setError('URL is required');
      return;
    }
    
    try {
      setIsLoading(true);
      setError(null);
      
      // Create article with optional category and tags
      const response = await articleAPI.createArticle(
        url,
        selectedCategory || undefined,
        selectedTags.length > 0 ? selectedTags : undefined
      );
      
      // Reset form states
      setUrl('');
      setSelectedCategory('');
      setSelectedTags([]);
      setCustomTag('');
      
      // Navigate to success page with article title if available
      if (response && response.data && response.data.title) {
        router.push(`/article-added?title=${encodeURIComponent(response.data.title)}`);
      } else {
        router.push('/article-added');
      }
    } catch (err) {
      console.error('Error creating article:', err);
      setError('Failed to create article. Please try again.');
      setIsLoading(false);
    }
  };

  // Handle tag selection
  const handleTagToggle = (tagName: string) => {
    setSelectedTags(prev => 
      prev.includes(tagName)
        ? prev.filter(t => t !== tagName)
        : [...prev, tagName]
    );
  };

  // Handle adding custom tag
  const handleAddCustomTag = () => {
    if (customTag.trim() && !selectedTags.includes(customTag.trim())) {
      setSelectedTags(prev => [...prev, customTag.trim()]);
      setCustomTag('');
    }
  };

  return (
    <div className="max-w-2xl mx-auto p-6 bg-white rounded-lg shadow-md">
      <h2 className="text-2xl font-bold mb-6">Add New Article</h2>
      
      {error && (
        <div className="mb-4 p-3 bg-red-100 text-red-700 rounded-md">
          {error}
        </div>
      )}
      
      <form onSubmit={handleSubmit}>
        {/* URL Input */}
        <div className="mb-4">
          <label htmlFor="url" className="block text-sm font-medium text-gray-700 mb-1">
            Article URL *
          </label>
          <input
            type="url"
            id="url"
            value={url}
            onChange={(e) => setUrl(e.target.value)}
            className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
            placeholder="https://example.com/article"
            required
          />
        </div>
        
        {/* Category Selection */}
        <div className="mb-4">
          <label htmlFor="category" className="block text-sm font-medium text-gray-700 mb-1">
            Category (Optional)
          </label>
          <select
            id="category"
            value={selectedCategory}
            onChange={(e) => setSelectedCategory(e.target.value)}
            className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
          >
            <option value="">Select a category</option>
            {categories && categories.length > 0 && categories.map((category) => (
              <option key={category.id} value={category.name}>
                {category.name}
              </option>
            ))}
          </select>
        </div>
        
        {/* Tags Selection */}
        <div className="mb-6">
          <label className="block text-sm font-medium text-gray-700 mb-1">
            Tags (Optional)
          </label>
          
          <div className="flex flex-wrap gap-2 mb-2">
            {tags && tags.length > 0 && tags.map((tag) => (
              <button
                key={tag.id}
                type="button"
                onClick={() => handleTagToggle(tag.name)}
                className={`px-3 py-1 rounded-full text-sm ${
                  selectedTags.includes(tag.name)
                    ? 'bg-blue-500 text-white'
                    : 'bg-gray-200 text-gray-700'
                }`}
              >
                {tag.name}
              </button>
            ))}
          </div>
          
          {/* Custom Tag Input */}
          <div className="flex mt-2">
            <input
              type="text"
              value={customTag}
              onChange={(e) => setCustomTag(e.target.value)}
              className="flex-grow px-3 py-2 border border-gray-300 rounded-l-md focus:outline-none focus:ring-2 focus:ring-blue-500"
              placeholder="Add custom tag"
            />
            <button
              type="button"
              onClick={handleAddCustomTag}
              className="px-3 py-2 bg-gray-200 text-gray-700 rounded-r-md hover:bg-gray-300 focus:outline-none"
            >
              <PlusIcon className="h-5 w-5" />
            </button>
          </div>
          
          {/* Selected Tags Display */}
          {selectedTags.length > 0 && (
            <div className="mt-2">
              <p className="text-sm text-gray-600 mb-1">Selected tags:</p>
              <div className="flex flex-wrap gap-1">
                {selectedTags.map((tag, index) => (
                  <span
                    key={index}
                    className="px-2 py-1 bg-blue-100 text-blue-800 rounded-md text-xs flex items-center"
                  >
                    {tag}
                    <button
                      type="button"
                      onClick={() => handleTagToggle(tag)}
                      className="ml-1 text-blue-500 hover:text-blue-700"
                    >
                      &times;
                    </button>
                  </span>
                ))}
              </div>
            </div>
          )}
        </div>
        
        {/* Submit Button */}
        <button
          type="submit"
          disabled={isLoading}
          className="w-full py-2 px-4 bg-blue-600 hover:bg-blue-700 text-white font-medium rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 disabled:opacity-50"
        >
          {isLoading ? 'Adding...' : 'Add Article'}
        </button>
      </form>
    </div>
  );
};

export default AddArticleForm; 