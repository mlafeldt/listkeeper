import 'source-map-support/register'
import * as cdk from 'aws-cdk-lib'
import * as path from 'path'
import { Schedule } from 'aws-cdk-lib/aws-events'

import { ApiStack } from './stacks/api-stack'
import { AppStack } from './stacks/app-stack'
import { CoreStack } from './stacks/core-stack'
import { DataStack } from './stacks/data-stack'

class ListkeeperApp extends cdk.App {
  constructor(props?: cdk.AppProps) {
    super(props)

    /*** Development environment ***/
    {
      const appName = 'listkeeper-dev'
      const appEnv = 'dev'
      const tags = { APP_NAME: appName, APP_ENV: appEnv }
      const ttlInDays = 7

      const dataStack = new DataStack(this, `${appName}-data`, {
        ttlInDays,
        tags,
      })

      new CoreStack(this, `${appName}-core`, {
        appName,
        schedule: Schedule.cron({ minute: '*/15' }),
        ttlInDays,
        slackUsername: 'Listkeeper (dev)',
        slackIconUrl: 'https://listkeeper.io/slack-icon.png',
        bucket: dataStack.bucket,
        table: dataStack.table,
        tags,
      })

      const apiStack = new ApiStack(this, `${appName}-api`, {
        appName,
        graphqlSchema: path.join(__dirname, '..', 'schema.graphql'),
        table: dataStack.table,
        tags,
      })

      new AppStack(this, `${appName}-app`, {
        appName,
        appEnv,
        deployBranch: 'main',
        graphqlEndpoint: apiStack.graphqlEndpoint,
        tags,
      })
    }

    /*** Production environment ***/
    {
      const appName = 'listkeeper-prod'
      const appEnv = 'prod'
      const tags = { APP_NAME: appName, APP_ENV: appEnv }
      const ttlInDays = 30

      const dataStack = new DataStack(this, `${appName}-data`, {
        ttlInDays,
        tags,
      })

      new CoreStack(this, `${appName}-core`, {
        appName,
        schedule: Schedule.cron({ minute: '0' }),
        ttlInDays,
        slackUsername: 'Listkeeper',
        slackIconUrl: 'https://listkeeper.io/slack-icon.png',
        bucket: dataStack.bucket,
        table: dataStack.table,
        tags,
      })

      const apiStack = new ApiStack(this, `${appName}-api`, {
        appName,
        graphqlSchema: path.join(__dirname, '..', 'schema.graphql'),
        table: dataStack.table,
        tags,
      })

      new AppStack(this, `${appName}-app`, {
        appName,
        appEnv,
        deployBranch: 'prod',
        domainName: 'listkeeper.io',
        graphqlEndpoint: apiStack.graphqlEndpoint,
        tags,
      })
    }
  }
}

new ListkeeperApp().synth()
