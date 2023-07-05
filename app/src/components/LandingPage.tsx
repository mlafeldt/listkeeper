import React from 'react'
import { useAuth0 } from '@auth0/auth0-react'

export const LandingPage = (): JSX.Element => {
  const { loginWithRedirect } = useAuth0()

  return (
    <div className="px-6 pt-6 pb-16 md:pt-24 md:pb-24">
      <h1 className="tracking-snug pb-6 text-center text-4xl font-black leading-snug text-gray-800 sm:text-5xl sm:leading-snug md:text-7xl md:leading-tight">
        Keep track of your&nbsp;
        <br className="hidden sm:block" />
        followers <u>and</u> unfollowers
      </h1>
      <p className="text-center text-lg leading-normal text-gray-700 md:text-2xl md:leading-9">
        Listkeeper watches your Twitter account and&nbsp;
        <br className="hidden sm:block" />
        tells you about new or lost followers.&nbsp;
        <br className="hidden sm:block" />
        Never miss an update again.
      </p>
      <div className="mx-auto flex justify-center pt-8">
        <button className="btn btn-lg" onClick={() => loginWithRedirect()}>
          Sign in with Twitter
        </button>
      </div>
      <div className="pt-3 text-center text-base text-gray-500">Currently in beta - whitelisted users only</div>
    </div>
  )
}
