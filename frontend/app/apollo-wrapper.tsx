'use client';

import { ApolloClient, ApolloProvider, InMemoryCache, HttpLink, from } from '@apollo/client';
import { setContext } from '@apollo/client/link/context';
import { RetryLink } from '@apollo/client/link/retry';

export function ApolloWrapper({ children }: React.PropsWithChildren<{}>) {
  // Use environment variable with fallback to localhost for development
  const graphqlUrl = process.env.NEXT_PUBLIC_GRAPHQL_API_URL || 'http://localhost:8686/graphql';

  const httpLink = new HttpLink({
    uri: graphqlUrl,
  });

  const authLink = setContext((_, { headers }) => {
    // Get token from environment variable (empty string for local development)
    const token = process.env.NEXT_PUBLIC_API_TOKEN || '';

    return {
      headers: {
        ...headers,
        // Only add authorization header if token exists
        ...(token && { authorization: `Bearer ${token}` }),
      },
    };
  });

  const retryLink = new RetryLink({
    attempts: {
      max: 3,
    },
  });

  const client = new ApolloClient({
    cache: new InMemoryCache(),
    link: from([retryLink, authLink, httpLink]),
  });

  return <ApolloProvider client={client}>{children}</ApolloProvider>;
}
