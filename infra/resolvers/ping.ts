import { NONERequest } from '@aws-appsync/utils'

export function request(): NONERequest {
  return { payload: {} }
}

export function response(): string {
  return 'pong'
}
