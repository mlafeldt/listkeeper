import React, { ReactNode } from 'react'
import { Link } from 'react-router-dom'
import { useAuth0 } from '@auth0/auth0-react'

export const Layout = (props: { children: ReactNode }): JSX.Element => {
  return (
    <div className="flex min-h-screen flex-col">
      <div className="flex-1">
        <Header />
        <main className="flex-grow">
          <div className="mx-auto max-w-screen-xl">{props.children}</div>
        </main>
      </div>
      <Footer />
    </div>
  )
}

const Header = () => {
  const { isLoading, isAuthenticated, loginWithRedirect, logout } = useAuth0()

  return (
    <header className="container mx-auto max-w-screen-xl">
      <div className="items-center space-y-6 px-6 py-5 md:flex md:space-y-0">
        <div className="flex flex-grow justify-center text-gray-900 md:justify-start">
          <Link to="/" className="flex items-center font-semibold tracking-widest">
            <img
              className="mr-px h-8 w-8 p-1"
              src={process.env.PUBLIC_URL + '/apple-touch-icon.png'}
              alt="Listkeeper's logo"
            />
            <span>Listkeeper</span>
          </Link>
        </div>
        <div className="flex items-center justify-center space-x-6">
          {isAuthenticated || isLoading ? (
            <button className="btn btn-md" onClick={() => logout({ returnTo: window.location.origin })}>
              Sign out
            </button>
          ) : (
            <button className="btn btn-md" onClick={() => loginWithRedirect()}>
              Sign in
            </button>
          )}
        </div>
      </div>
    </header>
  )
}

const Footer = () => {
  return (
    <footer className="bg-gray-700 text-gray-400">
      <div className="py-12 text-center">
        Made with 💙 by{' '}
        <a className="underline" target="_blank" rel="noreferrer" href="https://twitter.com/mlafeldt">
          Mathias Lafeldt
        </a>
      </div>
    </footer>
  )
}
