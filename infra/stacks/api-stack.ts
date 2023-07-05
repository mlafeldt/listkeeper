import * as cdk from 'aws-cdk-lib'
import * as appsync from 'aws-cdk-lib/aws-appsync'
import * as ddb from 'aws-cdk-lib/aws-dynamodb'
import { Construct } from 'constructs'
import { PolicyStatement } from 'aws-cdk-lib/aws-iam'
import { StringParameter } from 'aws-cdk-lib/aws-ssm'
import { GoFunction } from '../constructs/go-function'
import { JsResolver } from '../constructs/js-resolver'

interface ApiStackProps extends cdk.StackProps {
  appName: string
  graphqlSchema: string
  table: ddb.ITable
}

export class ApiStack extends cdk.Stack {
  readonly graphqlEndpoint: string

  constructor(scope: Construct, id: string, props: ApiStackProps) {
    super(scope, id, props)

    const auth0 = {
      domain: StringParameter.valueForStringParameter(this, `/${props.appName}/auth0-domain`),
      clientId: StringParameter.valueForStringParameter(this, `/${props.appName}/auth0-m2m-client-id`),
      clientSecret: StringParameter.valueForStringParameter(this, `/${props.appName}/auth0-m2m-client-secret`),
    }

    const api = new appsync.GraphqlApi(this, 'GraphqlApi', {
      name: id,
      schema: appsync.SchemaFile.fromAsset(props.graphqlSchema),
      logConfig: {
        fieldLogLevel: appsync.FieldLogLevel.ERROR,
      },
      authorizationConfig: {
        defaultAuthorization: {
          authorizationType: appsync.AuthorizationType.API_KEY,
          apiKeyConfig: {
            expires: cdk.Expiration.after(cdk.Duration.days(365)),
          },
        },
        additionalAuthorizationModes: [
          {
            authorizationType: appsync.AuthorizationType.OIDC,
            openIdConnectConfig: {
              oidcProvider: `https://${auth0.domain}`,
            },
          },
        ],
      },
    })

    const resolveGraphql = new GoFunction(this, 'ResolveGraphqlFunc', {
      handlerDir: 'resolve-graphql',
      memorySize: 256,
      environment: {
        TABLE_NAME: props.table.tableName,
        EVENT_BUS_NAME: 'default',
        EVENT_SOURCE_NAME: props.appName,
        AUTH0_DOMAIN: auth0.domain,
        AUTH0_CLIENT_ID: auth0.clientId,
        AUTH0_CLIENT_SECRET: auth0.clientSecret,
      },
    })
    props.table.grantReadWriteData(resolveGraphql.function)
    resolveGraphql.function.addToRolePolicy(
      new PolicyStatement({
        actions: ['events:PutEvents'],
        resources: ['*'],
      })
    )

    const lambdaDS = api.addLambdaDataSource('LambdaDatasource', resolveGraphql.function)
    lambdaDS.createResolver('RegisterUserResolver', { typeName: 'Mutation', fieldName: 'registerUser' })
    lambdaDS.createResolver('UpdateUserResolver', { typeName: 'Mutation', fieldName: 'updateUser' })
    lambdaDS.createResolver('DeleteUserResolver', { typeName: 'Mutation', fieldName: 'deleteUser' })

    const tableDS = api.addDynamoDbDataSource('DynamoDatasource', props.table)
    new JsResolver(this, 'GetUserResolver', {
      dataSource: tableDS,
      typeName: 'Query',
      fieldName: 'getUser',
      source: 'resolvers/getUser.ts',
    })
    new JsResolver(this, 'GetEventsResolver', {
      dataSource: tableDS,
      typeName: 'Query',
      fieldName: 'getLatestFollowerEvents',
      source: 'resolvers/getLatestFollowerEvents.ts',
    })

    new JsResolver(this, 'PingResolver', {
      dataSource: api.addNoneDataSource('NoneDatasource'),
      typeName: 'Query',
      fieldName: 'ping',
      source: 'resolvers/ping.ts',
    })

    this.graphqlEndpoint = api.graphqlUrl

    new cdk.CfnOutput(this, 'GraphqlEndpoint', { value: api.graphqlUrl })
    new cdk.CfnOutput(this, 'GraphqlApiKey', { value: api.apiKey! })
  }
}
