'use client';

import React, { useEffect } from 'react';
import Link from 'next/link';
import { useRouter, useSearchParams } from 'next/navigation';
import { CheckIcon } from 'lucide-react';

export default function ArticleAddedPage() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const title = searchParams?.get('title') || '';

  // Redirect to articles page after 3 seconds
  useEffect(() => {
    const timer = setTimeout(() => {
      router.push('/articles');
    }, 3000);

    return () => clearTimeout(timer);
  }, [router]);

  return (
    <div className="min-h-screen bg-gray-50 flex items-center justify-center py-12">
      <div className="max-w-md w-full bg-white shadow-lg rounded-lg p-8 text-center">
        <div className="h-16 w-16 bg-green-100 text-green-500 rounded-full flex items-center justify-center mx-auto mb-4">
          <CheckIcon className="h-10 w-10" />
        </div>
        
        <h1 className="text-2xl font-bold text-gray-900 mb-2">Article Added Successfully!</h1>
        
        <p className="text-gray-600 mb-6">
          {title 
            ? `"${title}" has been added to your collection.` 
            : 'Your article has been added to your collection.'}
        </p>
        
        <div className="flex flex-col space-y-3">
          <Link 
            href="/add-article" 
            className="inline-block w-full py-2 px-4 bg-blue-600 hover:bg-blue-700 text-white font-medium rounded-md transition-colors"
          >
            Add Another Article
          </Link>
          
          <Link 
            href="/articles" 
            className="inline-block w-full py-2 px-4 bg-gray-200 hover:bg-gray-300 text-gray-800 font-medium rounded-md transition-colors"
          >
            View All Articles
          </Link>
        </div>
        
        <p className="text-sm text-gray-500 mt-6">
          Redirecting to articles page in 3 seconds...
        </p>
      </div>
    </div>
  );
} 