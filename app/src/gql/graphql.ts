/* eslint-disable */
import { TypedDocumentNode as DocumentNode } from '@graphql-typed-document-node/core'
export type Maybe<T> = T | null
export type InputMaybe<T> = Maybe<T>
export type Exact<T extends { [key: string]: unknown }> = { [K in keyof T]: T[K] }
export type MakeOptional<T, K extends keyof T> = Omit<T, K> & { [SubKey in K]?: Maybe<T[SubKey]> }
export type MakeMaybe<T, K extends keyof T> = Omit<T, K> & { [SubKey in K]: Maybe<T[SubKey]> }
/** All built-in and custom scalars, mapped to their actual values */
export type Scalars = {
  ID: string
  String: string
  Boolean: boolean
  Int: number
  Float: number
  AWSDate: string
  AWSDateTime: string
  AWSEmail: string
  AWSIPAddress: string
  AWSJSON: string
  AWSPhone: string
  AWSTime: string
  AWSTimestamp: number
  AWSURL: string
}

export type Follower = {
  __typename?: 'Follower'
  bio?: Maybe<Scalars['String']>
  handle?: Maybe<Scalars['String']>
  id: Scalars['ID']
  location?: Maybe<Scalars['String']>
  name?: Maybe<Scalars['String']>
  profileImageUrl?: Maybe<Scalars['AWSURL']>
  protected: Scalars['Boolean']
  totalFollowers: Scalars['Int']
}

export type FollowerEvent = {
  __typename?: 'FollowerEvent'
  createdAt: Scalars['AWSDateTime']
  follower: Follower
  followerState: FollowerState
  followerStateReason: FollowerStateReason
  id: Scalars['ID']
  totalFollowers: Scalars['Int']
}

export enum FollowerState {
  Lost = 'LOST',
  New = 'NEW',
}

export enum FollowerStateReason {
  Deleted = 'DELETED',
  Followed = 'FOLLOWED',
  Suspended = 'SUSPENDED',
  Unfollowed = 'UNFOLLOWED',
}

export type Mutation = {
  __typename?: 'Mutation'
  deleteUser?: Maybe<Scalars['ID']>
  registerUser?: Maybe<User>
  updateUser?: Maybe<User>
}

export type MutationDeleteUserArgs = {
  id: Scalars['ID']
}

export type MutationRegisterUserArgs = {
  id: Scalars['ID']
}

export type MutationUpdateUserArgs = {
  id: Scalars['ID']
  input: UpdateUserInput
}

export type Query = {
  __typename?: 'Query'
  getLatestFollowerEvents?: Maybe<Array<FollowerEvent>>
  getUser?: Maybe<User>
  ping: Scalars['String']
}

export type QueryGetLatestFollowerEventsArgs = {
  userId: Scalars['ID']
}

export type QueryGetUserArgs = {
  id: Scalars['ID']
}

export type SlackConfig = {
  __typename?: 'SlackConfig'
  channel?: Maybe<Scalars['String']>
  enabled: Scalars['Boolean']
  webhookUrl?: Maybe<Scalars['AWSURL']>
}

export type SlackInput = {
  channel?: InputMaybe<Scalars['String']>
  enabled: Scalars['Boolean']
  webhookUrl?: InputMaybe<Scalars['AWSURL']>
}

export type UpdateUserInput = {
  ignoreFollowers?: InputMaybe<Array<Scalars['String']>>
  slack?: InputMaybe<SlackInput>
}

export type User = {
  __typename?: 'User'
  bio?: Maybe<Scalars['String']>
  createdAt: Scalars['AWSDateTime']
  handle: Scalars['String']
  id: Scalars['ID']
  ignoreFollowers?: Maybe<Array<Scalars['String']>>
  lastLogin: Scalars['AWSDateTime']
  location?: Maybe<Scalars['String']>
  name: Scalars['String']
  profileImageUrl: Scalars['AWSURL']
  slack: SlackConfig
  updatedAt: Scalars['AWSDateTime']
}

export type RegisterUserMutationVariables = Exact<{
  userId: Scalars['ID']
}>

export type RegisterUserMutation = {
  __typename?: 'Mutation'
  registerUser?: { __typename?: 'User'; id: string } | null
}

export type GetLatestFollowerEventsQueryVariables = Exact<{
  userId: Scalars['ID']
}>

export type GetLatestFollowerEventsQuery = {
  __typename?: 'Query'
  getLatestFollowerEvents?: Array<{
    __typename?: 'FollowerEvent'
    id: string
    totalFollowers: number
    followerState: FollowerState
    followerStateReason: FollowerStateReason
    createdAt: string
    follower: {
      __typename: 'Follower'
      id: string
      handle?: string | null
      name?: string | null
      profileImageUrl?: string | null
      protected: boolean
      totalFollowers: number
    }
  }> | null
}

export type UpdateUserMutationVariables = Exact<{
  userId: Scalars['ID']
  input: UpdateUserInput
}>

export type UpdateUserMutation = {
  __typename?: 'Mutation'
  updateUser?: {
    __typename?: 'User'
    slack: { __typename?: 'SlackConfig'; enabled: boolean; webhookUrl?: string | null; channel?: string | null }
  } | null
}

export type GetUserQueryVariables = Exact<{
  userId: Scalars['ID']
}>

export type GetUserQuery = {
  __typename?: 'Query'
  getUser?: {
    __typename?: 'User'
    slack: { __typename?: 'SlackConfig'; enabled: boolean; webhookUrl?: string | null; channel?: string | null }
  } | null
}

export const RegisterUserDocument = {
  kind: 'Document',
  definitions: [
    {
      kind: 'OperationDefinition',
      operation: 'mutation',
      name: { kind: 'Name', value: 'registerUser' },
      variableDefinitions: [
        {
          kind: 'VariableDefinition',
          variable: { kind: 'Variable', name: { kind: 'Name', value: 'userId' } },
          type: { kind: 'NonNullType', type: { kind: 'NamedType', name: { kind: 'Name', value: 'ID' } } },
        },
      ],
      selectionSet: {
        kind: 'SelectionSet',
        selections: [
          {
            kind: 'Field',
            name: { kind: 'Name', value: 'registerUser' },
            arguments: [
              {
                kind: 'Argument',
                name: { kind: 'Name', value: 'id' },
                value: { kind: 'Variable', name: { kind: 'Name', value: 'userId' } },
              },
            ],
            selectionSet: {
              kind: 'SelectionSet',
              selections: [{ kind: 'Field', name: { kind: 'Name', value: 'id' } }],
            },
          },
        ],
      },
    },
  ],
} as unknown as DocumentNode<RegisterUserMutation, RegisterUserMutationVariables>
export const GetLatestFollowerEventsDocument = {
  kind: 'Document',
  definitions: [
    {
      kind: 'OperationDefinition',
      operation: 'query',
      name: { kind: 'Name', value: 'getLatestFollowerEvents' },
      variableDefinitions: [
        {
          kind: 'VariableDefinition',
          variable: { kind: 'Variable', name: { kind: 'Name', value: 'userId' } },
          type: { kind: 'NonNullType', type: { kind: 'NamedType', name: { kind: 'Name', value: 'ID' } } },
        },
      ],
      selectionSet: {
        kind: 'SelectionSet',
        selections: [
          {
            kind: 'Field',
            name: { kind: 'Name', value: 'getLatestFollowerEvents' },
            arguments: [
              {
                kind: 'Argument',
                name: { kind: 'Name', value: 'userId' },
                value: { kind: 'Variable', name: { kind: 'Name', value: 'userId' } },
              },
            ],
            selectionSet: {
              kind: 'SelectionSet',
              selections: [
                { kind: 'Field', name: { kind: 'Name', value: 'id' } },
                { kind: 'Field', name: { kind: 'Name', value: 'totalFollowers' } },
                {
                  kind: 'Field',
                  name: { kind: 'Name', value: 'follower' },
                  selectionSet: {
                    kind: 'SelectionSet',
                    selections: [
                      {
                        kind: 'Field',
                        name: { kind: 'Name', value: '__typename' },
                        directives: [
                          {
                            kind: 'Directive',
                            name: { kind: 'Name', value: 'skip' },
                            arguments: [
                              {
                                kind: 'Argument',
                                name: { kind: 'Name', value: 'if' },
                                value: { kind: 'BooleanValue', value: true },
                              },
                            ],
                          },
                        ],
                      },
                      { kind: 'Field', name: { kind: 'Name', value: 'id' } },
                      { kind: 'Field', name: { kind: 'Name', value: 'handle' } },
                      { kind: 'Field', name: { kind: 'Name', value: 'name' } },
                      { kind: 'Field', name: { kind: 'Name', value: 'profileImageUrl' } },
                      { kind: 'Field', name: { kind: 'Name', value: 'protected' } },
                      { kind: 'Field', name: { kind: 'Name', value: 'totalFollowers' } },
                    ],
                  },
                },
                { kind: 'Field', name: { kind: 'Name', value: 'followerState' } },
                { kind: 'Field', name: { kind: 'Name', value: 'followerStateReason' } },
                { kind: 'Field', name: { kind: 'Name', value: 'createdAt' } },
              ],
            },
          },
        ],
      },
    },
  ],
} as unknown as DocumentNode<GetLatestFollowerEventsQuery, GetLatestFollowerEventsQueryVariables>
export const UpdateUserDocument = {
  kind: 'Document',
  definitions: [
    {
      kind: 'OperationDefinition',
      operation: 'mutation',
      name: { kind: 'Name', value: 'updateUser' },
      variableDefinitions: [
        {
          kind: 'VariableDefinition',
          variable: { kind: 'Variable', name: { kind: 'Name', value: 'userId' } },
          type: { kind: 'NonNullType', type: { kind: 'NamedType', name: { kind: 'Name', value: 'ID' } } },
        },
        {
          kind: 'VariableDefinition',
          variable: { kind: 'Variable', name: { kind: 'Name', value: 'input' } },
          type: { kind: 'NonNullType', type: { kind: 'NamedType', name: { kind: 'Name', value: 'UpdateUserInput' } } },
        },
      ],
      selectionSet: {
        kind: 'SelectionSet',
        selections: [
          {
            kind: 'Field',
            name: { kind: 'Name', value: 'updateUser' },
            arguments: [
              {
                kind: 'Argument',
                name: { kind: 'Name', value: 'id' },
                value: { kind: 'Variable', name: { kind: 'Name', value: 'userId' } },
              },
              {
                kind: 'Argument',
                name: { kind: 'Name', value: 'input' },
                value: { kind: 'Variable', name: { kind: 'Name', value: 'input' } },
              },
            ],
            selectionSet: {
              kind: 'SelectionSet',
              selections: [
                {
                  kind: 'Field',
                  name: { kind: 'Name', value: 'slack' },
                  selectionSet: {
                    kind: 'SelectionSet',
                    selections: [
                      { kind: 'Field', name: { kind: 'Name', value: 'enabled' } },
                      { kind: 'Field', name: { kind: 'Name', value: 'webhookUrl' } },
                      { kind: 'Field', name: { kind: 'Name', value: 'channel' } },
                    ],
                  },
                },
              ],
            },
          },
        ],
      },
    },
  ],
} as unknown as DocumentNode<UpdateUserMutation, UpdateUserMutationVariables>
export const GetUserDocument = {
  kind: 'Document',
  definitions: [
    {
      kind: 'OperationDefinition',
      operation: 'query',
      name: { kind: 'Name', value: 'getUser' },
      variableDefinitions: [
        {
          kind: 'VariableDefinition',
          variable: { kind: 'Variable', name: { kind: 'Name', value: 'userId' } },
          type: { kind: 'NonNullType', type: { kind: 'NamedType', name: { kind: 'Name', value: 'ID' } } },
        },
      ],
      selectionSet: {
        kind: 'SelectionSet',
        selections: [
          {
            kind: 'Field',
            name: { kind: 'Name', value: 'getUser' },
            arguments: [
              {
                kind: 'Argument',
                name: { kind: 'Name', value: 'id' },
                value: { kind: 'Variable', name: { kind: 'Name', value: 'userId' } },
              },
            ],
            selectionSet: {
              kind: 'SelectionSet',
              selections: [
                {
                  kind: 'Field',
                  name: { kind: 'Name', value: 'slack' },
                  selectionSet: {
                    kind: 'SelectionSet',
                    selections: [
                      { kind: 'Field', name: { kind: 'Name', value: 'enabled' } },
                      { kind: 'Field', name: { kind: 'Name', value: 'webhookUrl' } },
                      { kind: 'Field', name: { kind: 'Name', value: 'channel' } },
                    ],
                  },
                },
              ],
            },
          },
        ],
      },
    },
  ],
} as unknown as DocumentNode<GetUserQuery, GetUserQueryVariables>
