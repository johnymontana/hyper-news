'use client';
import { Avatar, ChatInterface } from '@aichatkit/ui';
import { useApolloClient } from '@apollo/client';
import { PlusIcon, SendIcon, XIcon, ArrowLeftIcon, Zap } from 'lucide-react';
import { useEffect, useState } from 'react';
import Link from 'next/link';
import { ApolloAdapter } from '@aichatkit/apollo-adapter';
import { LocalStorageAdapter } from '@aichatkit/localstorage-adapter';

// Import the required CSS
import '@aichatkit/ui/dist/base.css';

export default function ChatPage() {
  const apolloClient = useApolloClient();
  const [networkAdapter, setNetworkAdapter] = useState<ApolloAdapter | null>(null);
  const [storageAdapter, setStorageAdapter] = useState<LocalStorageAdapter | null>(null);
  const [ready, setReady] = useState(false);

  useEffect(() => {
    const initAdapters = async () => {
      try {
        // Create adapters
        const apolloAdapter = new ApolloAdapter({
          // eslint-disable-next-line @typescript-eslint/ban-ts-comment
          // @ts-ignore
          apolloClient,
          apiUrl: process.env.NEXT_PUBLIC_GRAPHQL_API_URL || 'http://localhost:8686/graphql',
        });
        const localStorageAdapter = new LocalStorageAdapter({
          conversationPrefix: 'hypernews_chat_',
          conversationIdsKey: 'hypernews_conversations',
          activeConversationKey: 'hypernews_active_chat',
        });

        // Initialize storage adapter
        await localStorageAdapter.initialize();

        // Set up network callbacks for backend synchronization
        localStorageAdapter.setNetworkCallbacks({
          getConversationItems: (agentId: string) => apolloAdapter.getConversationItems(agentId),
          clearConversationHistory: (agentId: string) => apolloAdapter.clearConversationHistory(agentId),
        });

        setNetworkAdapter(apolloAdapter);
        setStorageAdapter(localStorageAdapter);
        setReady(true);
      } catch (error) {
        console.error('Error initializing chat adapters:', error);
      }
    };

    initAdapters();
  }, [apolloClient]);

  if (!ready) {
    return (
      <div className="fixed inset-0 flex items-center justify-center bg-hypermode-bg">
        <div className="flex flex-col items-center space-y-4">
          <div className="animate-spin rounded-full h-12 w-12 border-t-2 border-b-2 border-pink-400"></div>
          <div className="text-lg text-white">Loading HyperNews Chat...</div>
        </div>
      </div>
    );
  }

  return (
    <div className="fixed inset-0 bg-hypermode-bg text-white overflow-hidden">
      <ChatInterface
        networkAdapter={networkAdapter!}
        storageAdapter={storageAdapter!}
        showSidebar={true}
        hypermodeStyle={true}
        className="w-full h-full"
        chatAreaClassName="hypermode-scrollbar h-full"
        sendButtonIcon={<SendIcon size={18} />}
        newConversationIcon={<PlusIcon size={18} />}
        deleteConversationIcon={<XIcon size={16} />}
        userAvatar={<Avatar initial="U" role="user" hypermodeStyle={true} />}
        assistantAvatar={
          <div className="w-8 h-8 rounded-full bg-gradient-to-br from-pink-500 to-pink-600 flex items-center justify-center">
            <Zap size={16} className="text-white" />
          </div>
        }
        inputPlaceholder="Ask about news, search articles, get analysis..."
        headerContent={
          <div className="flex items-center justify-between border-hypermode-border border-b bg-hypermode-card p-4">
            <div className="flex items-center gap-3">
              <Link
                href="/"
                className="flex items-center gap-2 text-pink-400 hover:text-pink-300 transition-colors group"
              >
                <ArrowLeftIcon size={20} className="group-hover:-translate-x-1 transition-transform" />
                <span>Back to News</span>
              </Link>
              <div className="h-6 w-px bg-hypermode-border" />
              <div className="flex items-center gap-3">
                <div className="w-8 h-8 rounded-full bg-gradient-to-br from-pink-500 to-pink-600 flex items-center justify-center">
                  <Zap size={16} className="text-white" />
                </div>
                <div>
                  <span className="font-semibold text-white">HyperNews Assistant</span>
                  <div className="flex items-center space-x-2 text-xs text-gray-400">
                    <div className="w-1.5 h-1.5 bg-pink-400 rounded-full animate-pulse"></div>
                    <span>AI Ready</span>
                  </div>
                </div>
              </div>
            </div>
          </div>
        }
        sidebarHeaderContent={
          <div className="flex items-center justify-center border-hypermode-border border-b bg-hypermode-card p-4">
            <div className="flex items-center space-x-3">
              <div className="w-8 h-8 rounded-full bg-gradient-to-br from-pink-500 to-pink-600 flex items-center justify-center">
                <Zap size={16} className="text-white" />
              </div>
              <h1 className="text-xl font-bold bg-gradient-to-r from-pink-400 to-pink-300 bg-clip-text text-transparent">
                HyperNews
              </h1>
            </div>
          </div>
        }
        onCardAction={(action) => {
          console.log('Card action triggered:', action);
          switch (action.action) {
            case 'view_article':
              // Open article in new tab
              if (action.data?.url) {
                window.open(action.data.url, '_blank');
              }
              break;
            case 'search_articles':
              // Perform article search
              if (action.data?.query) {
                // Could trigger another tool call or navigate
                console.log('Searching for:', action.data.query);
              }
              break;
            case 'analyze_topic':
              // Analyze topic trends
              if (action.data?.topic) {
                console.log('Analyzing topic:', action.data.topic);
              }
              break;
            default:
              if (action.type === 'link') {
                window.open(action.action, '_blank');
              }
          }
        }}
      />
    </div>
  );
}
