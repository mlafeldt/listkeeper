import { type CodegenConfig } from '@graphql-codegen/cli'

const config: CodegenConfig = {
  schema: ['../*.graphql'],
  documents: ['src/**/*.tsx', '!src/gql/**/*'],
  generates: {
    './src/gql/': {
      preset: 'client',
      presetConfig: {
        fragmentMasking: false,
      },
    },
  },
  config: {
    strictScalars: true,
    scalars: {
      AWSDate: 'string',
      AWSDateTime: 'string',
      AWSEmail: 'string',
      AWSIPAddress: 'string',
      AWSJSON: 'string',
      AWSPhone: 'string',
      AWSTime: 'string',
      AWSTimestamp: 'number',
      AWSURL: 'string',
    },
  },
  hooks: { afterOneFileWrite: ['prettier --write'] },
}

export default config
