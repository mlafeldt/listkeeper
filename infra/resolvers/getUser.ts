import { util, Context, AppSyncIdentityOIDC, DynamoDBQueryRequest } from '@aws-appsync/utils'
import { User } from '../../app/src/gql/graphql'
import { authorize } from './shared'

export function request(ctx: Context<{ id: string }>): DynamoDBQueryRequest {
  const userId = authorize(ctx.args.id, ctx.identity as AppSyncIdentityOIDC)

  return {
    operation: 'Query',
    query: {
      expression: 'PK = :PK',
      expressionValues: util.dynamodb.toMapValues({
        ':PK': `USER#${userId}`,
      }),
    },
    index: 'UserIndex',
  }
}

export function response(ctx: Context): User {
  const { result, error } = ctx

  if (error) {
    util.error(error.message, error.type)
  }
  if (result.items.length === 0) {
    util.error('user not found')
  }

  const user = result.items[0]

  return {
    id: user.UserID,
    handle: user.Handle,
    name: user.Name,
    location: user.Location,
    bio: user.Bio,
    profileImageUrl: user.ProfileImageURL,
    slack: {
      enabled: user.Slack?.Enabled ?? false,
      webhookUrl: user.Slack?.WebhookURL,
      channel: user.Slack?.Channel,
    },
    ignoreFollowers: user.IgnoreFollowers,
    createdAt: user.CreatedAt,
    updatedAt: user.UpdatedAt,
    lastLogin: user.LastLogin,
  }
}
