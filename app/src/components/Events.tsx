import React, { useEffect } from 'react'
import ReactTooltip from 'react-tooltip'
import { useLazyQuery } from '@apollo/client'
import { formatDistanceToNowStrict } from 'date-fns'
import { LockClosedIcon, StarIcon } from '@heroicons/react/solid'
import clsx from 'clsx'

import { graphql } from '../gql'
import { type FollowerEvent } from '../gql/graphql'

const DEFAULT_PROFILE_IMAGE = '/default_profile.png'

const FOLLOWER_STAR_STEPS = [1_000, 5_000, 10_000]

const GET_LATEST_FOLLOWER_EVENTS = graphql(/* GraphQL */ `
  query getLatestFollowerEvents($userId: ID!) {
    getLatestFollowerEvents(userId: $userId) {
      id
      totalFollowers
      follower {
        __typename @skip(if: true) # Apollo must not cache followers by their id
        id
        handle
        name
        profileImageUrl
        protected
        totalFollowers
      }
      followerState
      followerStateReason
      createdAt
    }
  }
`)

export const Events = (props: { userId: string }): JSX.Element => {
  const [getEvents, { data, loading, error }] = useLazyQuery(GET_LATEST_FOLLOWER_EVENTS, {
    variables: { userId: props.userId },
    fetchPolicy: 'network-only',
  })

  useEffect(() => {
    getEvents()
  }, [getEvents])

  return (
    <div className="px-6 pt-6 pb-16">
      <h3 className="pb-1 text-2xl font-normal leading-10 text-gray-900">Follower Events</h3>
      <div className="mb-7 text-sm text-gray-500">
        <span>See who followed or unfollowed you on Twitter. </span>
        <span className="cursor-pointer underline" onClick={() => getEvents()}>
          Refresh
        </span>
      </div>
      {loading ? (
        <p className="text-sm text-gray-700">Loading...</p>
      ) : error ? (
        <p className="text-sm text-red-500">Error: {error.message}</p>
      ) : (
        <EventList events={data?.getLatestFollowerEvents ?? []} />
      )}
    </div>
  )
}

const EventList = ({ events }: { events: Array<FollowerEvent> }) => {
  if (events.length === 0) {
    return <p className="text-sm text-gray-700">No events yet! Try again in an hour.</p>
  }

  return (
    <ul>
      {events.map((e: FollowerEvent) => (
        <EventItem event={e} key={e.id} />
      ))}
      <ReactTooltip place="right" />
    </ul>
  )
}

const EventItem = ({ event }: { event: FollowerEvent }) => {
  return (
    <li
      className={clsx(
        'grid grid-cols-10 gap-4 border-b px-1 py-3 text-sm transition duration-150 ease-in-out hover:bg-gray-50 md:px-3',
        { NEW: 'bg-green-50', LOST: 'bg-white' }[event.followerState]
      )}
    >
      <div
        className="col-span-3 truncate text-gray-700 sm:col-span-2"
        data-tip={new Date(event.createdAt).toString().replace(/ \(.*\)/, '')}
      >
        {formatDistanceToNowStrict(new Date(event.createdAt), { addSuffix: true })}
      </div>
      <div className="col-span-7 text-sm leading-5">
        <div className="flex items-center pb-1">
          <div className="pr-2">
            <img
              className="hidden h-5 w-5 flex-shrink-0 select-none rounded-full bg-white text-white ring-2 ring-white sm:flex"
              src={getProfileImage(event.follower.profileImageUrl)}
              onError={handleImageError(DEFAULT_PROFILE_IMAGE)}
              alt={event.follower.id}
              loading="lazy"
            />
          </div>
          <div className="items-center justify-center truncate text-gray-900 sm:flex">
            {
              {
                FOLLOWED: (
                  <a target="_blank" rel="noreferrer" href={'https://twitter.com/' + event.follower.handle}>
                    {event.follower.name}
                    <span className="hidden sm:inline"> (@{event.follower.handle})</span> followed you
                  </a>
                ),
                UNFOLLOWED: (
                  <a target="_blank" rel="noreferrer" href={'https://twitter.com/' + event.follower.handle}>
                    {event.follower.name}
                    <span className="hidden sm:inline"> (@{event.follower.handle})</span> unfollowed you
                  </a>
                ),
                DELETED: `Follower #${event.follower.id} was deleted`,
                SUSPENDED: `Follower #${event.follower.id} was suspended`,
              }[event.followerStateReason]
            }
            {event.follower.protected && (
              <span className="ml-1 hidden sm:flex">
                <LockClosedIcon className="h-4 w-4 text-gray-500" data-tip="Protected account" />
              </span>
            )}
            <span
              className="ml-1 hidden sm:flex"
              data-tip={numberWithCommas(event.follower.totalFollowers) + ' followers'}
            >
              {FOLLOWER_STAR_STEPS.map(
                (step, i) =>
                  event.follower.totalFollowers >= step && <StarIcon key={i} className="h-4 w-4 text-gray-500" />
              )}
            </span>
          </div>
        </div>
      </div>
    </li>
  )
}

const numberWithCommas = (n: number) => {
  return n.toString().replace(/\B(?=(\d{3})+(?!\d))/g, ',')
}

const getProfileImage = (url: string | undefined | null) => {
  return url?.replace(/_normal\./, '_bigger.') ?? DEFAULT_PROFILE_IMAGE
}

// eslint-disable-next-line @typescript-eslint/no-explicit-any
const handleImageError = (fallback: string) => (event: any) => (event.target.src = fallback)
