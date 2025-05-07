import React from 'react';
import { Metadata } from 'next';

export const metadata: Metadata = {
  title: 'Article Added Successfully | Memento',
  description: 'Article has been added to your collection',
};

export default function ArticleAddedLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return <>{children}</>;
} 