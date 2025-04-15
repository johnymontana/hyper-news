'use client';

import { GET_ARTICLES } from '@/app/queries';
import { useQuery } from '@apollo/client';
import React, { useState, useRef } from 'react';
import ChatBox from './chat-box';

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

  const handleSearchResults = (results: string) => {
    setSearchResults(results);

    try {
      const parsedResults = JSON.parse(results);

      if (Array.isArray(parsedResults)) {
        setSearchedArticles(parsedResults);
      } else {
        console.log('Received non-array result:', parsedResults);
        setSearchedArticles(null);
      }
    } catch (error) {
      console.error('Error parsing search results:', error);
      setSearchedArticles(null);
    }

    setTimeout(() => {
      if (articlesRef.current) {
        articlesRef.current.scrollIntoView({ behavior: 'smooth' });
      }
    }, 100);
  };

  const resetSearch = () => {
    setSearchResults(null);
    setSearchedArticles(null);
  };

  if (loading) {
    return (
      <div className="flex justify-center items-center min-h-[50vh]">
        <div className="flex flex-col items-center">
          <div className="animate-spin rounded-full h-12 w-12 border-t-2 border-b-2 border-blue-500 mb-4"></div>
          <div className="text-xl font-medium text-gray-700">Loading articles...</div>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="flex justify-center items-center min-h-[50vh]">
        <div className="bg-red-50 border border-red-200 text-red-700 px-6 py-4 rounded-lg max-w-lg">
          <h3 className="text-lg font-semibold mb-2">Error Loading Articles</h3>
          <p>{error.message}</p>
        </div>
      </div>
    );
  }

  const articles = data?.articles || [];

  const renderArticleList = (articleList: ArticleData[]) => (
    <div className="grid grid-cols-1 gap-6">
      {articleList.map((article) => (
        <div
          key={article.uid}
          className="bg-white border border-gray-200 p-6 rounded-lg hover:shadow-md transition-shadow duration-200"
        >
          <a href={article.url} target="_blank" rel="noopener noreferrer" className="block">
            <h2 className="text-xl font-semibold mb-2 text-blue-600 hover:underline">{article.title}</h2>
            <p className="text-gray-600 mb-3">{article.abstract}</p>
            <div className="text-sm text-gray-400">ID: {article.uid}</div>
          </a>
        </div>
      ))}
    </div>
  );

  return (
    <div className="container mx-auto px-4 py-8">
      <div className="mb-12 max-w-3xl mx-auto">
        <h2 className="text-2xl font-bold mb-6 text-center text-gray-800">Ask About News</h2>
        <ChatBox onResultsReceived={handleSearchResults} />
      </div>

      <div ref={articlesRef}>
        {searchedArticles ? (
          <div className="max-w-5xl mx-auto">
            <div className="flex justify-between items-center mb-6">
              <h2 className="text-2xl font-bold text-gray-800">Search Results</h2>
              <button
                onClick={resetSearch}
                className="px-4 py-2 bg-gray-100 text-gray-700 rounded-lg hover:bg-gray-200 transition-colors flex items-center"
              >
                <svg
                  className="w-4 h-4 mr-2"
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
                Back to Latest Articles
              </button>
            </div>

            {searchedArticles.length === 0 ? (
              <div className="bg-gray-50 p-8 rounded-lg text-center">
                <svg
                  className="w-16 h-16 text-gray-400 mx-auto mb-4"
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
                <p className="text-gray-600 text-lg">No articles found matching your search.</p>
                <p className="text-gray-500 mt-2">Try different keywords or phrases.</p>
              </div>
            ) : (
              renderArticleList(searchedArticles)
            )}
          </div>
        ) : (
          <div className="max-w-5xl mx-auto">
            <h2 className="text-2xl font-bold mb-6 text-gray-800 border-b border-gray-200 pb-3">Latest Articles</h2>
            {articles.length === 0 ? (
              <div className="bg-gray-50 p-8 rounded-lg text-center">
                <p className="text-gray-600">No articles found.</p>
              </div>
            ) : (
              renderArticleList(articles)
            )}
          </div>
        )}
      </div>
    </div>
  );
}
