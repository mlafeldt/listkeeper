import { util, Context, AppSyncIdentityOIDC, DynamoDBQueryRequest } from '@aws-appsync/utils'
import { FollowerEvent } from '../../app/src/gql/graphql'
import { authorize } from './shared'

export function request(ctx: Context<{ userId: string }>): DynamoDBQueryRequest {
  const userId = authorize(ctx.args.userId, ctx.identity as AppSyncIdentityOIDC)

  return {
    operation: 'Query',
    query: {
      expression: 'PK = :PK and begins_with(SK, :SK)',
      expressionValues: util.dynamodb.toMapValues({
        ':PK': `USER#${userId}`,
        ':SK': 'EVENT#',
      }),
    },
    limit: 100,
    scanIndexForward: false,
  }
}

export function response(ctx: Context): FollowerEvent[] {
  const { result, error } = ctx

  if (error) {
    util.error(error.message, error.type)
  }

  return result.items.map((item: any) => ({
    id: item.EventID,
    totalFollowers: item.TotalFollowers,
    follower: {
      id: item.Follower.ID,
      handle: item.Follower.Handle,
      name: item.Follower.Name,
      location: item.Follower.Location,
      bio: item.Follower.Bio,
      profileImageUrl: item.Follower.ProfileImageURL,
      protected: item.Follower.Protected,
      totalFollowers: item.Follower.TotalFollowers,
    },
    followerState: item.FollowerState,
    followerStateReason: item.FollowerStateReason,
    createdAt: item.CreatedAt,
  }))
}
