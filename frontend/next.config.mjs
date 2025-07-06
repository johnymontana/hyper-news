/** @type {import('next').NextConfig} */
const nextConfig = {
  eslint: {
    ignoreDuringBuilds: true,
  },
  typescript: {
    ignoreBuildErrors: true,
  },
  images: {
    unoptimized: true,
  },
  experimental: {
    webpackBuildWorker: true,
    parallelServerBuildTraces: true,
    parallelServerCompiles: true,
  },
  // Environment variables configuration
  env: {
    // These will be available at build time and runtime
    NEXT_PUBLIC_GRAPHQL_API_URL: process.env.NEXT_PUBLIC_GRAPHQL_API_URL || 'http://localhost:8686/graphql',
    NEXT_PUBLIC_API_TOKEN: process.env.NEXT_PUBLIC_API_TOKEN || '',
  },
  // Public runtime config for dynamic configuration
  publicRuntimeConfig: {
    graphqlApiUrl: process.env.NEXT_PUBLIC_GRAPHQL_API_URL || 'http://localhost:8686/graphql',
    apiToken: process.env.NEXT_PUBLIC_API_TOKEN || '',
  },
  // Server runtime config (server-side only)
  serverRuntimeConfig: {
    // Add any server-side only configuration here
  },
};

export default nextConfig;
