import * as cdk from 'aws-cdk-lib'
import { Construct } from 'constructs'
import { ITable } from 'aws-cdk-lib/aws-dynamodb'
import { Schedule, Rule, RuleTargetInput } from 'aws-cdk-lib/aws-events'
import { LambdaFunction } from 'aws-cdk-lib/aws-events-targets'
import { LambdaDestination } from 'aws-cdk-lib/aws-lambda-destinations'
import { PolicyStatement } from 'aws-cdk-lib/aws-iam'
import { IBucket } from 'aws-cdk-lib/aws-s3'
import { StringParameter } from 'aws-cdk-lib/aws-ssm'
import { GoFunction } from '../constructs/go-function'

interface CoreStackProps extends cdk.StackProps {
  appName: string
  schedule: Schedule
  ttlInDays: number
  slackUsername: string
  slackIconUrl: string
  bucket: IBucket
  table: ITable
}

export class CoreStack extends cdk.Stack {
  constructor(scope: Construct, id: string, props: CoreStackProps) {
    super(scope, id, props)

    // prettier-ignore
    const twitterVars = {
      TWITTER_CONSUMER_KEY: StringParameter.valueForStringParameter(this, `/${props.appName}/twitter-consumer-key`),
      TWITTER_CONSUMER_SECRET: StringParameter.valueForStringParameter(this, `/${props.appName}/twitter-consumer-secret`),
    }

    const diffFollowers = new GoFunction(this, 'DiffFollowersFunc', {
      handlerDir: 'diff-followers',
      environment: {
        TABLE_NAME: props.table.tableName,
        EVENT_BUS_NAME: 'default',
        EVENT_SOURCE_NAME: props.appName,
        ...twitterVars,
      },
    })
    props.table.grantReadWriteData(diffFollowers.function)
    props.bucket.grantRead(diffFollowers.function)
    diffFollowers.function.addToRolePolicy(
      new PolicyStatement({
        actions: ['events:PutEvents'],
        resources: ['*'],
      })
    )

    const getFollowers = new GoFunction(this, 'GetFollowersFunc', {
      handlerDir: 'get-followers',
      environment: {
        TABLE_NAME: props.table.tableName,
        TABLE_TTL: `${props.ttlInDays * 24}h`,
        BUCKET_NAME: props.bucket.bucketName,
        ...twitterVars,
      },
      onSuccess: new LambdaDestination(diffFollowers.function, { responseOnly: true }),
    })
    props.table.grantReadWriteData(getFollowers.function)
    props.bucket.grantPut(getFollowers.function)

    new Rule(this, 'GetFollowersOnSignup', {
      eventPattern: {
        source: [props.appName], // default bus
        detailType: ['New User Signup'],
      },
      targets: [
        new LambdaFunction(getFollowers.function, {
          event: RuleTargetInput.fromEventPath('$.detail'),
        }),
      ],
    })

    const notifyUser = new GoFunction(this, 'NotifyUserFunc', {
      handlerDir: 'notify-user',
      environment: {
        TABLE_NAME: props.table.tableName,
        SLACK_USERNAME: props.slackUsername,
        SLACK_ICON_URL: props.slackIconUrl,
      },
    })
    props.table.grantReadData(notifyUser.function)

    new Rule(this, 'NotifyUserOnFollowerChange', {
      eventPattern: {
        source: [props.appName], // default bus
        detailType: ['Twitter Follower Change'],
      },
      targets: [
        new LambdaFunction(notifyUser.function, {
          event: RuleTargetInput.fromEventPath('$.detail'),
        }),
      ],
    })

    const enqueueUsers = new GoFunction(this, 'EnqueueUsersFunc', {
      handlerDir: 'enqueue-users',
      environment: {
        TABLE_NAME: props.table.tableName,
        FUNCTION_NAME: getFollowers.function.functionName,
      },
    })
    props.table.grantReadData(enqueueUsers.function)
    getFollowers.function.grantInvoke(enqueueUsers.function)

    new Rule(this, 'ScheduleEnqueueUsers', {
      schedule: props.schedule,
      targets: [new LambdaFunction(enqueueUsers.function)],
    })
  }
}
