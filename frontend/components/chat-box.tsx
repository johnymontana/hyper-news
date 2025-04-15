'use client';

import React, { useState } from 'react';
import { useLazyQuery } from '@apollo/client';
import { GENERATE_TEXT_WITH_TOOLS } from '@/app/queries';

interface ChatBoxProps {
  onResultsReceived: (results: string) => void;
}

interface Message {
  text: string;
  isUser: boolean;
  time: string;
}

export default function ChatBox({ onResultsReceived }: ChatBoxProps) {
  const [inputValue, setInputValue] = useState('');
  const [messages, setMessages] = useState<Message[]>([
    {
      text: "I'm your news assistant. Ask me about any news topics you're interested in, and I'll search for relevant articles.",
      isUser: false,
      time: new Date().toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' }),
    },
  ]);

  const [generateText, { loading }] = useLazyQuery(GENERATE_TEXT_WITH_TOOLS, {
    onCompleted: (data) => {
      if (data?.generateTextWithTools) {
        setMessages((currentMessages) => [
          ...currentMessages,
          {
            text: "Here are some articles related to your query. I've displayed them below.",
            isUser: false,
            time: new Date().toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' }),
          },
        ]);

        onResultsReceived(data.generateTextWithTools);
      }
    },
    onError: (error) => {
      setMessages((currentMessages) => [
        ...currentMessages,
        {
          text: `Error fetching results: ${error.message}`,
          isUser: false,
          time: new Date().toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' }),
        },
      ]);
    },
  });

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setInputValue(e.target.value);
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (inputValue.trim() === '' || loading) return;

    setMessages((currentMessages) => [
      ...currentMessages,
      {
        text: inputValue,
        isUser: true,
        time: new Date().toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' }),
      },
    ]);

    setMessages((currentMessages) => [
      ...currentMessages,
      {
        text: 'Searching for articles...',
        isUser: false,
        time: new Date().toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' }),
        isLoading: true,
      } as Message & { isLoading?: boolean },
    ]);

    generateText({
      variables: { prompt: inputValue },
    });

    setInputValue('');
  };

  return (
    <div className="bg-white rounded-lg shadow-sm p-4 flex flex-col h-80">
      <div className="flex-1 overflow-y-auto space-y-3 mb-4 pr-2">
        {messages.map((message, index) => {
          if ((message as any).isLoading && loading) {
            return (
              <div key={index} className="flex justify-start">
                <div className="max-w-xs md:max-w-md bg-gray-100 rounded-lg px-4 py-3">
                  <div className="flex items-center space-x-2">
                    <div className="w-2 h-2 bg-gray-400 rounded-full animate-pulse"></div>
                    <div className="w-2 h-2 bg-gray-400 rounded-full animate-pulse delay-150"></div>
                    <div className="w-2 h-2 bg-gray-400 rounded-full animate-pulse delay-300"></div>
                    <span className="text-gray-500 text-sm">Searching...</span>
                  </div>
                </div>
              </div>
            );
          }

          if ((message as any).isLoading && !loading) {
            return null;
          }

          return (
            <div key={index} className={`flex ${message.isUser ? 'justify-end' : 'justify-start'}`}>
              <div
                className={`max-w-xs md:max-w-md ${
                  message.isUser ? 'bg-blue-500 text-white' : 'bg-gray-100 text-gray-800'
                } rounded-lg px-4 py-3`}
              >
                <p className="text-sm">{message.text}</p>
                <p className="text-xs opacity-70 mt-1 text-right">{message.time}</p>
              </div>
            </div>
          );
        })}
      </div>

      <form onSubmit={handleSubmit} className="relative">
        <input
          type="text"
          value={inputValue}
          onChange={handleInputChange}
          placeholder="Search for news articles..."
          className="w-full border border-gray-300 rounded-lg pl-4 pr-12 py-3 focus:outline-none focus:ring-2 focus:ring-blue-500"
          disabled={loading}
        />
        <button
          type="submit"
          className="absolute right-2 top-1/2 transform -translate-y-1/2 bg-blue-500 text-white p-2 rounded-full hover:bg-blue-600 transition-colors disabled:bg-blue-300 focus:outline-none focus:ring-2 focus:ring-blue-500"
          disabled={loading || !inputValue.trim()}
          aria-label="Send"
        >
          {loading ? (
            <svg className="animate-spin h-5 w-5" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
              <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
              <path
                className="opacity-75"
                fill="currentColor"
                d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
              ></path>
            </svg>
          ) : (
            <svg
              className="h-5 w-5"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
              xmlns="http://www.w3.org/2000/svg"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth="2"
                d="M12 19l9 2-9-18-9 18 9-2zm0 0v-8"
              ></path>
            </svg>
          )}
        </button>
      </form>
    </div>
  );
}
