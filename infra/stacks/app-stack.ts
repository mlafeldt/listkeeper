import * as cdk from 'aws-cdk-lib'
import { Construct } from 'constructs'
import { App, GitHubSourceCodeProvider, CustomRule, RedirectStatus } from '@aws-cdk/aws-amplify-alpha'
import { Role, ServicePrincipal, ManagedPolicy } from 'aws-cdk-lib/aws-iam'
import { StringParameter } from 'aws-cdk-lib/aws-ssm'

interface AppStackProps extends cdk.StackProps {
  appName: string
  appEnv: string
  deployBranch: string
  domainName?: string
  graphqlEndpoint: string
}

export class AppStack extends cdk.Stack {
  constructor(scope: Construct, id: string, props: AppStackProps) {
    super(scope, id, props)

    const role = new Role(this, 'ServiceRole', {
      assumedBy: new ServicePrincipal('amplify.amazonaws.com'),
    })
    role.addManagedPolicy(ManagedPolicy.fromAwsManagedPolicyName('AdministratorAccess'))

    const app = new App(this, 'App', {
      appName: props.appName,
      sourceCodeProvider: new GitHubSourceCodeProvider({
        owner: 'mlafeldt',
        repository: 'listkeeper',
        oauthToken: cdk.SecretValue.unsafePlainText(
          // don't want to pay for Secrets Manager...
          StringParameter.valueForStringParameter(this, `/${props.appName}/github-oauth-token`)
        ),
      }),
      // prettier-ignore
      environmentVariables: {
        REACT_APP_AUTH0_DOMAIN: StringParameter.valueForStringParameter(this, `/${props.appName}/auth0-domain`),
        REACT_APP_AUTH0_CLIENT_ID: StringParameter.valueForStringParameter(this, `/${props.appName}/auth0-spa-client-id`),
        REACT_APP_GRAPHQL_ENDPOINT: props.graphqlEndpoint,
        AMPLIFY_DIFF_DEPLOY: 'false',
        AMPLIFY_DIFF_DEPLOY_ROOT: 'app',
        APP_ENV: props.appEnv,
      },
      role,
    })

    const branch = app.addBranch(props.deployBranch, {
      pullRequestPreview: false,
    })

    if (props.domainName) {
      const domain = app.addDomain('Domain', {
        domainName: props.domainName,
        subDomains: [{ branch, prefix: 'www' }],
      })
      domain.mapRoot(branch)

      app.addCustomRule(
        new CustomRule({
          source: `https://www.${props.domainName}`,
          target: `https://${props.domainName}`,
          status: RedirectStatus.TEMPORARY_REDIRECT,
        })
      )
    }

    app.addCustomRule(CustomRule.SINGLE_PAGE_APPLICATION_REDIRECT)

    new cdk.CfnOutput(this, 'AppId', { value: app.appId })
    new cdk.CfnOutput(this, 'AppUrl', { value: `https://${branch.branchName}.${app.defaultDomain}` })
  }
}
