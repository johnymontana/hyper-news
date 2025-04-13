import React from 'react';
import Articles from '../components/articles';

export default function Page() {
  return (
    <main className="min-h-screen bg-white">
      <div className="max-w-6xl mx-auto">
        <Articles />
      </div>
    </main>
  );
}
