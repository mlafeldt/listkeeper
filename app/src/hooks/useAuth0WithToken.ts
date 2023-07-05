import { useState, useEffect } from 'react'
import { useAuth0 } from '@auth0/auth0-react'

export const useAuth0WithToken = () => /* eslint-disable-line @typescript-eslint/explicit-module-boundary-types */ {
  const auth0 = useAuth0()
  const [token, setToken] = useState('')
  const [error, setError] = useState()

  useEffect(() => {
    if (!auth0.isAuthenticated) return
    auth0
      .getAccessTokenSilently()
      .then((token: string) => setToken(token))
      .catch((err) => setError(err))
  }, [auth0])

  return {
    ...auth0,
    isLoading: auth0.isLoading || (auth0.isAuthenticated && !token),
    error: auth0.error || error,
    token,
  }
}
