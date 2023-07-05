import React from 'react'
import ReactDOM from 'react-dom'
import { Auth0Provider } from '@auth0/auth0-react'

import App from './App'

const auth0 = {
  domain: process.env.REACT_APP_AUTH0_DOMAIN || '',
  clientId: process.env.REACT_APP_AUTH0_CLIENT_ID || '',
  audience: `https://${process.env.REACT_APP_AUTH0_DOMAIN}/api/v2/`,
  scope: 'read:current_user',
}

const graphqlEndpoint = process.env.REACT_APP_GRAPHQL_ENDPOINT || ''

ReactDOM.render(
  <React.StrictMode>
    <Auth0Provider
      domain={auth0.domain}
      clientId={auth0.clientId}
      audience={auth0.audience}
      redirectUri={window.location.origin}
      useRefreshTokens={true}
      connection="twitter"
      scope={auth0.scope}
    >
      <App graphqlEndpoint={graphqlEndpoint} />
    </Auth0Provider>
  </React.StrictMode>,
  document.getElementById('root')
)
