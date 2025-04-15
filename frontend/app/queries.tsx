import { gql } from '@apollo/client';

export const GET_ARTICLES = gql`
  query {
    articles: queryArticles(num: 10) {
      uid
      title
      abstract
      url
      uri: url
    }
  }
`;

export const GENERATE_TEXT_WITH_TOOLS = gql`
  query GenerateTextWithTools($prompt: String!) {
    generateTextWithTools(prompt: $prompt)
  }
`;
