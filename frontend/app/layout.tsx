import type { Metadata } from 'next';
import './globals.css';
import { ApolloWrapper } from './apollo-wrapper';

export const metadata: Metadata = {
  title: 'Hypernews',
  description: 'Hypermode Article Reader',
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en">
      <body>
        <ApolloWrapper>{children}</ApolloWrapper>
      </body>
    </html>
  );
}
