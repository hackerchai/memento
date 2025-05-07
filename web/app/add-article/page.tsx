'use client';

import React from 'react';
import AddArticleForm from '../../components/AddArticleForm';
import Link from 'next/link';

export default function AddArticlePage() {
  return (
    <div className="min-h-screen bg-gray-50 py-12">
      <div className="container mx-auto px-4">
        <div className="mb-6 flex justify-between items-center">
          <h1 className="text-3xl font-bold text-gray-900">Add New Article</h1>
          <Link 
            href="/articles" 
            className="px-4 py-2 bg-gray-200 hover:bg-gray-300 text-gray-800 rounded-md transition-colors"
          >
            Back to Articles
          </Link>
        </div>
        
        <AddArticleForm />
      </div>
    </div>
  );
} 