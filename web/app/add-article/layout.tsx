import React from 'react';
import { Metadata } from 'next';

export const metadata: Metadata = {
  title: 'Add New Article | Memento',
  description: 'Add a new article to your Memento collection',
};

export default function AddArticleLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return <>{children}</>;
} 