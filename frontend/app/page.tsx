import React from 'react';
import Image from 'next/image';
import Articles from '../components/articles';
import { MessageCircle, Zap } from 'lucide-react';

export default function Page() {
  return (
    <main className="min-h-screen bg-hypermode-bg">
      {/* Hero Section */}
      <div className="bg-gradient-to-br from-hypermode-bg via-hypermode-card to-hypermode-bg border-b border-hypermode-border">
        <div className="max-w-7xl mx-auto px-4 py-20">
          <div className="text-center space-y-8">
            <div className="flex items-center justify-center space-x-3 mb-6">
              <div className="p-3 bg-pink-600/20 rounded-lg border border-pink-500/30">
                <Image
                  src="/pink-hypermode.svg"
                  alt="Hypermode Logo"
                  width={32}
                  height={29}
                  className="w-8 h-8"
                  priority
                />
              </div>
              <h1 className="text-5xl md:text-6xl font-bold bg-gradient-to-r from-pink-400 via-pink-300 to-pink-500 bg-clip-text text-transparent leading-tight">
                HyperNews
              </h1>
            </div>
            <p className="text-xl text-gray-300 max-w-3xl mx-auto leading-relaxed">
              Discover and explore news with AI assistance. Ask questions, get insights, and dive deeper into the
              stories that matter.
            </p>

            <div className="flex items-center justify-center space-x-6 pt-6">
              <div className="flex items-center space-x-2 text-sm text-gray-400">
                <div className="w-2 h-2 bg-pink-400 rounded-full animate-pulse"></div>
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
