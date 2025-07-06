'use client';

import { GET_ARTICLES } from '@/app/queries';
import { useQuery } from '@apollo/client';
import React, { useState, useRef } from 'react';
import Link from 'next/link';
import { MessageCircle, ExternalLink } from 'lucide-react';

interface ArticleData {
  uid: string;
  title: string;
  abstract: string;
  url: string;
  uri: string;
}

export default function Articles() {
  const { loading, error, data } = useQuery(GET_ARTICLES);
  const [, setSearchResults] = useState<string | null>(null);
  const [searchedArticles, setSearchedArticles] = useState<ArticleData[] | null>(null);
  const articlesRef = useRef<HTMLDivElement>(null);

  const resetSearch = () => {
    setSearchResults(null);
    setSearchedArticles(null);
  };

  if (loading) {
    return (
      <div className="flex justify-center items-center min-h-[50vh]">
        <div className="flex flex-col items-center space-y-4">
          <div className="animate-spin rounded-full h-12 w-12 border-t-2 border-b-2 border-pink-400"></div>
          <div className="text-xl font-medium text-gray-300">Loading articles...</div>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="flex justify-center items-center min-h-[50vh]">
        <div className="bg-red-900/20 border border-red-500/50 text-red-300 px-6 py-4 rounded-lg max-w-lg">
          <h3 className="text-lg font-semibold mb-2">Error Loading Articles</h3>
          <p>{error.message}</p>
        </div>
      </div>
    );
  }

  const articles = data?.articles || [];

  const renderArticleList = (articleList: ArticleData[]) => (
    <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
      {articleList.map((article) => (
        <div
          key={article.uid}
          className="bg-hypermode-card border border-hypermode-border rounded-lg p-6 hover:bg-hypermode-hover transition-all duration-200 group article-card"
        >
          <div className="flex flex-col h-full">
            {/* Article Header */}
            <div className="flex-1">
              <h2 className="text-xl font-semibold mb-3 text-white group-hover:text-pink-300 transition-colors line-clamp-2">
                {article.title}
              </h2>

              <p className="text-gray-300 mb-4 line-clamp-3 leading-relaxed">{article.abstract}</p>
            </div>

            {/* Article Footer */}
            <div className="flex items-center justify-between pt-4 border-t border-hypermode-border">
              <div className="flex items-center space-x-4 text-sm text-gray-400">
                <span className="truncate">ID: {article.uid}</span>
              </div>

              <div className="flex items-center space-x-2">
                <Link href="/chat">
                  <button className="flex items-center space-x-2 px-3 py-2 bg-pink-600 hover:bg-pink-700 text-white rounded-lg transition-colors text-sm font-medium">
                    <MessageCircle size={16} />
                    <span>Discuss</span>
                  </button>
                </Link>

                <a
                  href={article.url}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="flex items-center space-x-2 px-3 py-2 bg-hypermode-border hover:bg-hypermode-hover text-gray-300 hover:text-white rounded-lg transition-colors text-sm font-medium"
                >
                  <ExternalLink size={16} />
                  <span>Read</span>
                </a>
              </div>
            </div>
          </div>
        </div>
      ))}
    </div>
  );

  return (
    <div className="min-h-screen bg-hypermode-bg">
      <div className="container mx-auto px-4 py-8">
        <div ref={articlesRef}>
          {searchedArticles ? (
            <div className="max-w-7xl mx-auto">
              <div className="flex justify-between items-center mb-8">
                <h2 className="text-3xl font-bold text-white">Search Results</h2>
                <button
                  onClick={resetSearch}
                  className="px-4 py-2 bg-hypermode-card border border-hypermode-border text-gray-300 rounded-lg hover:bg-hypermode-hover hover:text-white transition-colors flex items-center space-x-2"
                >
                  <svg
                    className="w-4 h-4"
                    fill="none"
                    stroke="currentColor"
                    viewBox="0 0 24 24"
                    xmlns="http://www.w3.org/2000/svg"
                  >
                    <path
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      strokeWidth="2"
                      d="M10 19l-7-7m0 0l7-7m-7 7h18"
                    ></path>
                  </svg>
                  <span>Back to Latest Articles</span>
                </button>
              </div>

              {searchedArticles.length === 0 ? (
                <div className="bg-hypermode-card border border-hypermode-border p-12 rounded-lg text-center">
                  <svg
                    className="w-16 h-16 text-gray-500 mx-auto mb-4"
                    fill="none"
                    stroke="currentColor"
                    viewBox="0 0 24 24"
                    xmlns="http://www.w3.org/2000/svg"
                  >
                    <path
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      strokeWidth="2"
                      d="M9.172 16.172a4 4 0 015.656 0M9 10h.01M15 10h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
                    ></path>
                  </svg>
                  <p className="text-gray-300 text-lg mb-2">No articles found matching your search.</p>
                  <p className="text-gray-500">Try different keywords or phrases.</p>
                </div>
              ) : (
                renderArticleList(searchedArticles)
              )}
            </div>
          ) : (
            <div className="max-w-7xl mx-auto">
              {/* Header with Chat Button */}
              <div className="flex justify-between items-center mb-8 pb-4 border-b border-hypermode-border">
                <div>
                  <h2 className="text-3xl font-bold text-white mb-2">Latest Articles</h2>
                  <p className="text-gray-400">Discover and explore news with AI assistance</p>
                </div>

                <Link href="/chat">
                  <button className="flex items-center space-x-3 px-6 py-3 bg-pink-600 hover:bg-pink-700 text-white rounded-lg transition-colors font-medium shadow-lg hover:shadow-xl">
                    <MessageCircle size={20} />
                    <span>Start Chat</span>
                  </button>
                </Link>
              </div>

              {articles.length === 0 ? (
                <div className="bg-hypermode-card border border-hypermode-border p-12 rounded-lg text-center">
                  <div className="text-gray-400 mb-4">
                    <svg
                      className="w-16 h-16 mx-auto mb-4"
                      fill="none"
                      stroke="currentColor"
                      viewBox="0 0 24 24"
                      xmlns="http://www.w3.org/2000/svg"
                    >
                      <path
                        strokeLinecap="round"
                        strokeLinejoin="round"
                        strokeWidth="2"
                        d="M19 20H5a2 2 0 01-2-2V6a2 2 0 012-2h10a2 2 0 012 2v1m2 13a2 2 0 01-2-2V7m2 13a2 2 0 002-2V9a2 2 0 00-2-2h-2m-4-3H9M7 16h6M7 8h6v4H7V8z"
                      ></path>
                    </svg>
                  </div>
                  <p className="text-gray-300 text-lg">No articles found.</p>
                  <p className="text-gray-500 mt-2">Check back later for new content.</p>
                </div>
              ) : (
                renderArticleList(articles)
              )}
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
