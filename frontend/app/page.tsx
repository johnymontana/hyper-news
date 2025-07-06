import React from 'react';
import Articles from '../components/articles';
import { MessageCircle, Zap } from 'lucide-react';

export default function Page() {
  return (
    <main className="min-h-screen bg-hypermode-bg">
      {/* Hero Section */}
      <div className="bg-gradient-to-br from-hypermode-bg via-hypermode-card to-hypermode-bg border-b border-hypermode-border">
        <div className="max-w-7xl mx-auto px-4 py-16">
          <div className="text-center space-y-6">
            <div className="flex items-center justify-center space-x-3 mb-4">
              <div className="p-3 bg-blue-600/20 rounded-lg border border-blue-500/30">
                <Zap className="w-8 h-8 text-blue-400" />
              </div>
              <h1 className="text-6xl font-bold bg-gradient-to-r from-blue-400 via-blue-300 to-blue-500 bg-clip-text text-transparent">
                HyperNews
              </h1>
            </div>
            <p className="text-xl text-gray-300 max-w-2xl mx-auto leading-relaxed">
              Discover and explore news with AI assistance. Ask questions, get insights, and dive deeper into the
              stories that matter.
            </p>

            <div className="flex items-center justify-center space-x-4 pt-4">
              <div className="flex items-center space-x-2 text-sm text-gray-400">
                <div className="w-2 h-2 bg-green-400 rounded-full animate-pulse"></div>
                <span>AI Assistant Ready</span>
              </div>
              <div className="w-px h-4 bg-hypermode-border"></div>
              <div className="flex items-center space-x-2 text-sm text-gray-400">
                <MessageCircle size={14} />
                <span>Interactive Chat Available</span>
              </div>
            </div>
          </div>
        </div>
      </div>

      {/* Articles Section */}
      <Articles />
    </main>
  );
}
