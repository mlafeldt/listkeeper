import { util, AppSyncIdentityOIDC } from '@aws-appsync/utils'

// With String.prototype.replace(), the pattern can be a
// string or a regex. However, APPSYNC_JS only accepts a
// string that is treated as a regex!
const AUTH0_PROVIDER_PREFIX = '^twitter\\|'

export function authorize(userId: string, identity?: AppSyncIdentityOIDC): string {
  if (identity?.sub && identity.sub !== userId) {
    util.unauthorized()
  }
  return userId.replace(AUTH0_PROVIDER_PREFIX, '')
}
