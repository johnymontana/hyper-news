import { gql } from '@apollo/client';

export const GET_ARTICLES = gql`
query{
  articles: queryArticles(num: 10) {
    uid
    title
    abstract
    url
    uri: url
}}
`;
