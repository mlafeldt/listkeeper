import { Construct } from 'constructs'
import * as appsync from 'aws-cdk-lib/aws-appsync'
import { buildSync } from 'esbuild'

export interface JsResolverProps {
  dataSource: appsync.BaseDataSource
  typeName: 'Query' | 'Mutation'
  fieldName: string
  source: string
}

export class JsResolver extends Construct {
  public readonly resolver: appsync.Resolver

  constructor(scope: Construct, id: string, props: JsResolverProps) {
    super(scope, id)

    // Import API based on data source so there's one less prop to pass in
    const api = appsync.GraphqlApi.fromGraphqlApiAttributes(this, 'GraphqlApi', {
      graphqlApiId: props.dataSource.ds.apiId,
    })

    // Use esbuild to transpile/bundle the resolver code
    const buildResult = buildSync({
      entryPoints: [props.source],
      bundle: true,
      write: false,
      external: ['@aws-appsync/utils'],
      format: 'esm',
      target: 'es2020',
      sourcemap: 'inline',
      sourcesContent: false,
    })

    if (buildResult.errors.length > 0) {
      throw new Error(`Failed to build ${props.source}: ${buildResult.errors[0].text}`)
    }
    if (buildResult.outputFiles.length === 0) {
      throw new Error(`Failed to build ${props.source}: no output files`)
    }

    // Create AppSync function from bundled code
    const runtime = appsync.FunctionRuntime.JS_1_0_0
    const func = new appsync.AppsyncFunction(this, 'Func', {
      api,
      dataSource: props.dataSource,
      name: props.fieldName + props.typeName,
      code: appsync.Code.fromInline(buildResult.outputFiles[0].text),
      runtime,
    })

    // Create dummy pipeline resolver for lack of JS unit resolvers
    this.resolver = new appsync.Resolver(this, 'Resolver', {
      api,
      typeName: props.typeName,
      fieldName: props.fieldName,
      pipelineConfig: [func],
      // prettier-ignore
      code: appsync.Code.fromInline(
        [
          'export function request(ctx)  { return {} }',
          'export function response(ctx) { return ctx.prev.result }'
        ].join('\n')
      ),
      runtime,
    })
  }
}
