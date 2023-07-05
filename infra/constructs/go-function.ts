import * as cdk from 'aws-cdk-lib'
import * as lambda from 'aws-cdk-lib/aws-lambda'
import { Construct } from 'constructs'
import { RetentionDays } from 'aws-cdk-lib/aws-logs'
import * as path from 'path'

const HANDLER_BASE_DIR = path.join(__dirname, '..', '..', 'functions', 'bin')

export interface GoFunctionProps {
  handlerDir: string
  memorySize?: number
  timeout?: cdk.Duration
  environment?: {
    [key: string]: string
  }
  onSuccess?: lambda.IDestination
}

export class GoFunction extends Construct {
  public readonly function: lambda.IFunction

  constructor(scope: Construct, id: string, props: GoFunctionProps) {
    super(scope, id)

    this.function = new lambda.Function(this, id, {
      code: lambda.Code.fromAsset(path.join(HANDLER_BASE_DIR, props.handlerDir)),
      handler: 'bootstrap',
      runtime: lambda.Runtime.PROVIDED_AL2,
      architecture: lambda.Architecture.ARM_64,
      memorySize: props.memorySize || 128,
      timeout: props.timeout || cdk.Duration.seconds(10),
      logRetention: RetentionDays.ONE_MONTH,
      environment: props.environment,
      onSuccess: props.onSuccess,
    })
  }
}
