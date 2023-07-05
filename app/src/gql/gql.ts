/* eslint-disable */
import * as types from './graphql'
import { TypedDocumentNode as DocumentNode } from '@graphql-typed-document-node/core'

/**
 * Map of all GraphQL operations in the project.
 *
 * This map has several performance disadvantages:
 * 1. It is not tree-shakeable, so it will include all operations in the project.
 * 2. It is not minifiable, so the string of a GraphQL query will be multiple times inside the bundle.
 * 3. It does not support dead code elimination, so it will add unused operations.
 *
 * Therefore it is highly recommended to use the babel or swc plugin for production.
 */
const documents = {
  '\n  mutation registerUser($userId: ID!) {\n    registerUser(id: $userId) {\n      id\n    }\n  }\n':
    types.RegisterUserDocument,
  '\n  query getLatestFollowerEvents($userId: ID!) {\n    getLatestFollowerEvents(userId: $userId) {\n      id\n      totalFollowers\n      follower {\n        __typename @skip(if: true) # Apollo must not cache followers by their id\n        id\n        handle\n        name\n        profileImageUrl\n        protected\n        totalFollowers\n      }\n      followerState\n      followerStateReason\n      createdAt\n    }\n  }\n':
    types.GetLatestFollowerEventsDocument,
  '\n  mutation updateUser($userId: ID!, $input: UpdateUserInput!) {\n    updateUser(id: $userId, input: $input) {\n      slack {\n        enabled\n        webhookUrl\n        channel\n      }\n    }\n  }\n':
    types.UpdateUserDocument,
  '\n  query getUser($userId: ID!) {\n    getUser(id: $userId) {\n      slack {\n        enabled\n        webhookUrl\n        channel\n      }\n    }\n  }\n':
    types.GetUserDocument,
}

/**
 * The graphql function is used to parse GraphQL queries into a document that can be used by GraphQL clients.
 *
 *
 * @example
 * ```ts
 * const query = graphql(`query GetUser($id: ID!) { user(id: $id) { name } }`);
 * ```
 *
 * The query argument is unknown!
 * Please regenerate the types.
 */
export function graphql(source: string): unknown

/**
 * The graphql function is used to parse GraphQL queries into a document that can be used by GraphQL clients.
 */
export function graphql(
  source: '\n  mutation registerUser($userId: ID!) {\n    registerUser(id: $userId) {\n      id\n    }\n  }\n'
): (typeof documents)['\n  mutation registerUser($userId: ID!) {\n    registerUser(id: $userId) {\n      id\n    }\n  }\n']
/**
 * The graphql function is used to parse GraphQL queries into a document that can be used by GraphQL clients.
 */
export function graphql(
  source: '\n  query getLatestFollowerEvents($userId: ID!) {\n    getLatestFollowerEvents(userId: $userId) {\n      id\n      totalFollowers\n      follower {\n        __typename @skip(if: true) # Apollo must not cache followers by their id\n        id\n        handle\n        name\n        profileImageUrl\n        protected\n        totalFollowers\n      }\n      followerState\n      followerStateReason\n      createdAt\n    }\n  }\n'
): (typeof documents)['\n  query getLatestFollowerEvents($userId: ID!) {\n    getLatestFollowerEvents(userId: $userId) {\n      id\n      totalFollowers\n      follower {\n        __typename @skip(if: true) # Apollo must not cache followers by their id\n        id\n        handle\n        name\n        profileImageUrl\n        protected\n        totalFollowers\n      }\n      followerState\n      followerStateReason\n      createdAt\n    }\n  }\n']
/**
 * The graphql function is used to parse GraphQL queries into a document that can be used by GraphQL clients.
 */
export function graphql(
  source: '\n  mutation updateUser($userId: ID!, $input: UpdateUserInput!) {\n    updateUser(id: $userId, input: $input) {\n      slack {\n        enabled\n        webhookUrl\n        channel\n      }\n    }\n  }\n'
): (typeof documents)['\n  mutation updateUser($userId: ID!, $input: UpdateUserInput!) {\n    updateUser(id: $userId, input: $input) {\n      slack {\n        enabled\n        webhookUrl\n        channel\n      }\n    }\n  }\n']
/**
 * The graphql function is used to parse GraphQL queries into a document that can be used by GraphQL clients.
 */
export function graphql(
  source: '\n  query getUser($userId: ID!) {\n    getUser(id: $userId) {\n      slack {\n        enabled\n        webhookUrl\n        channel\n      }\n    }\n  }\n'
): (typeof documents)['\n  query getUser($userId: ID!) {\n    getUser(id: $userId) {\n      slack {\n        enabled\n        webhookUrl\n        channel\n      }\n    }\n  }\n']

export function graphql(source: string) {
  return (documents as any)[source] ?? {}
}

export type DocumentType<TDocumentNode extends DocumentNode<any, any>> = TDocumentNode extends DocumentNode<
  infer TType,
  any
>
  ? TType
  : never
