'use client';

import { GET_ARTICLES } from '@/app/queries';
import { useQuery } from '@apollo/client';
import React from 'react';

interface ArticleData {
  uid: string;
  title: string;
  abstract: string;
  url: string;
  uri: string;
}

export default function Articles() {
  const { loading, error, data } = useQuery(GET_ARTICLES);

  // const parseArticles = (articlesJsonString: string): ArticleData[] => {
  //   try {
  //     const parsed = JSON.parse(articlesJsonString);
  //     return parsed.data.articles;
  //   } catch (e) {
  //     console.error('Error parsing articles JSON:', e);
  //     return [];
  //   }
  // };

  if (loading) {
    return (
      <div className="flex justify-center items-center min-h-screen">
        <div className="animate-pulse text-xl font-medium">Loading articles...</div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="flex justify-center items-center min-h-screen">
        <div className="text-xl font-medium text-red-500">Error loading articles: {error.message}</div>
      </div>
    );
  }

  const articles = data?.articles ? (data.articles) : [];

  return (
    <div className="container mx-auto px-4 py-8">
      <h1 className="text-3xl font-bold mb-8 border-b border-gray-200 pb-4">Latest Articles</h1>

      {articles.length === 0 ? (
        <p className="text-center text-gray-600">No articles found.</p>
      ) : (
        <ul className="space-y-6">
          {articles.map((article) => (
            <li
              key={article.uid}
              className="border border-gray-200 p-6 rounded-lg hover:shadow-md transition-shadow duration-200"
            >
              <a href={article.url} target="_blank" rel="noopener noreferrer" className="block">
                <h2 className="text-xl font-semibold mb-2 hover:underline">{article.title}</h2>
                <p className="text-gray-600 mb-3">{article.abstract}</p>
                <div className="text-sm text-gray-500">Article ID: {article.uid}</div>
              </a>
            </li>
          ))}
        </ul>
      )}
    </div>
  );
}
