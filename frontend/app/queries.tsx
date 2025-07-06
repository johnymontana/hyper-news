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

export const CREATE_CONVERSATION = gql`
  mutation CreateConversation {
    createConversation
  }
`;

export const CONTINUE_CHAT = gql`
  query ContinueChat($id: String!, $query: String!) {
    continueChat(id: $id, query: $query) {
      items
      conversationId
    }
  }
`;

export const CHAT_HISTORY = gql`
  query ChatHistory($id: String!) {
    chatHistory(id: $id) {
      items
      count
    }
  }
`;

export const DELETE_CONVERSATION_HISTORY = gql`
  mutation DeleteConversationHistory($id: String!) {
    deleteConversationHistory(id: $id)
  }
`;

export const DELETE_AGENT = gql`
  mutation DeleteAgent($id: String!) {
    deleteAgent(id: $id)
  }
`;
