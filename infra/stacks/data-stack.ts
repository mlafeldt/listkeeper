import * as cdk from 'aws-cdk-lib'
import * as ddb from 'aws-cdk-lib/aws-dynamodb'
import * as s3 from 'aws-cdk-lib/aws-s3'
import { Construct } from 'constructs'

interface DataStackProps extends cdk.StackProps {
  ttlInDays: number
}

export class DataStack extends cdk.Stack {
  public readonly bucket: s3.IBucket
  public readonly table: ddb.ITable

  constructor(scope: Construct, id: string, props: DataStackProps) {
    super(scope, id, props)

    const bucket = new s3.Bucket(this, 'Bucket', {
      blockPublicAccess: s3.BlockPublicAccess.BLOCK_ALL,
      encryption: s3.BucketEncryption.S3_MANAGED,
    })
    bucket.addLifecycleRule({
      id: `ExpireObjects`,
      expiration: cdk.Duration.days(props.ttlInDays),
    })
    this.bucket = bucket

    const table = new ddb.Table(this, 'Table', {
      partitionKey: { name: 'PK', type: ddb.AttributeType.STRING },
      sortKey: { name: 'SK', type: ddb.AttributeType.STRING },
      timeToLiveAttribute: 'TTL',
      billingMode: ddb.BillingMode.PAY_PER_REQUEST,
      pointInTimeRecovery: true,
      stream: ddb.StreamViewType.NEW_AND_OLD_IMAGES,
    })
    table.addLocalSecondaryIndex({
      indexName: 'UserIndex',
      sortKey: { name: 'UserIndex', type: ddb.AttributeType.STRING },
      projectionType: ddb.ProjectionType.ALL,
    })
    this.table = table

    new cdk.CfnOutput(this, 'BucketName', { value: bucket.bucketName })
    new cdk.CfnOutput(this, 'TableName', { value: table.tableName })
  }
}
