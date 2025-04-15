import React from 'react';
import Articles from '../components/articles';

export default function Page() {
  return (
    <main className="min-h-screen bg-gray-50">
      <div className="max-w-7xl mx-auto pt-8 pb-16">
        <header className="mb-12 text-center">
          <h1 className="text-5xl font-bold text-blue-600 mb-2">HyperNews</h1>
          <p className="text-gray-600 text-lg">Discover and explore news with AI assistance</p>
        </header>
        <Articles />
      </div>
    </main>
  );
}
