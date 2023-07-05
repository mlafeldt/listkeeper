import React, { useEffect } from 'react'
import { BrowserRouter as Router, Switch, Route, Redirect } from 'react-router-dom'
import { ApolloClient, ApolloProvider, InMemoryCache, createHttpLink, useMutation } from '@apollo/client'
import { setContext } from '@apollo/client/link/context'
import * as Fathom from 'fathom-client'

import { useAuth0WithToken } from './hooks/useAuth0WithToken'

import { Events } from './components/Events'
import { LandingPage } from './components/LandingPage'
import { Layout } from './components/Layout'
import { NotFound } from './components/NotFound'
import { Settings } from './components/Settings'

import { graphql } from './gql'

import './App.css'

const REGISTER_USER = graphql(/* GraphQL */ `
  mutation registerUser($userId: ID!) {
    registerUser(id: $userId) {
      id
    }
  }
`)

interface AppProps {
  graphqlEndpoint: string
}

const getApolloClient = (endpoint: string, token: string) => {
  const httpLink = createHttpLink({ uri: endpoint })

  const authLink = setContext((_, { headers }) => {
    return {
      headers: {
        ...headers,
        authorization: token,
      },
    }
  })

  return new ApolloClient({
    link: authLink.concat(httpLink),
    cache: new InMemoryCache(),
  })
}

const App = (props: AppProps): JSX.Element => {
  const { user, token, isLoading, isAuthenticated, error } = useAuth0WithToken()

  useEffect(() => {
    Fathom.load('QLLPJBDU', {
      includedDomains: ['listkeeper.io'],
    })
  }, [])

  return (
    <ApolloProvider client={getApolloClient(props.graphqlEndpoint, token)}>
      <Router>
        <Layout>
          {isLoading ? null : error ? (
            <div>Error: {error.message}</div>
          ) : isAuthenticated && user?.sub ? (
            <Main userId={user.sub} />
          ) : (
            <LandingPage />
          )}
        </Layout>
      </Router>
    </ApolloProvider>
  )
}

const Main = (props: { userId: string }) => {
  const [registerUser, { error }] = useMutation(REGISTER_USER, {
    variables: { userId: props.userId },
    onError: () => null, // don't throw Unhandled Rejection error
    ignoreResults: true, // only ensure the user is created in the backend
  })

  useEffect(() => {
    registerUser()
  }, [registerUser])

  return (
    <>
      {error ? (
        <div>Error: {error.message}</div>
      ) : (
        <Switch>
          <Route exact path="/">
            <Redirect to="/events" />
          </Route>
          <Route path="/events">
            <Events userId={props.userId} />
          </Route>
          <Route path="/settings">
            <Settings userId={props.userId} />
          </Route>
          <Route component={NotFound} />
        </Switch>
      )}
    </>
  )
}

export default App
