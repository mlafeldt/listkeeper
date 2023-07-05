import React from 'react'
import { Link } from 'react-router-dom'

export const NotFound = (): JSX.Element => (
  <div className="px-6 pt-16 pb-16 text-center md:pt-24 md:pb-24">
    <h1 className="pb-4 text-2xl leading-10 text-gray-900">Whoops, couldn&apos;t find that page</h1>
    <Link
      to="/"
      className="text-lg font-medium text-green-500 transition duration-150 ease-in-out hover:text-green-400 active:text-green-600"
    >
      Take me home →
    </Link>
  </div>
)
