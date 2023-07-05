import React from 'react'
import { useState, createRef } from 'react'
import { useQuery, useMutation } from '@apollo/client'
import clsx from 'clsx'
import { useEffect } from 'react'

import { graphql } from '../gql'

const UPDATE_USER = graphql(/* GraphQL */ `
  mutation updateUser($userId: ID!, $input: UpdateUserInput!) {
    updateUser(id: $userId, input: $input) {
      slack {
        enabled
        webhookUrl
        channel
      }
    }
  }
`)

const GET_USER = graphql(/* GraphQL */ `
  query getUser($userId: ID!) {
    getUser(id: $userId) {
      slack {
        enabled
        webhookUrl
        channel
      }
    }
  }
`)

export const Settings = (props: { userId: string }): JSX.Element => {
  const {
    data,
    loading: queryLoading,
    error: queryError,
  } = useQuery(GET_USER, {
    variables: { userId: props.userId },
  })
  const [updateUser, { loading, error }] = useMutation(UPDATE_USER, {
    onError: () => null, // don't throw Unhandled Rejection error
  })
  const [slackEnabled, setSlackEnabled] = useState(false)
  const webhookUrlInput = createRef<HTMLInputElement>()
  const channelInput = createRef<HTMLInputElement>()

  useEffect(() => {
    setSlackEnabled(data?.getUser?.slack.enabled ?? false)
  }, [setSlackEnabled, data?.getUser?.slack])

  if (!data || queryLoading) return <div>Loading...</div>
  if (queryError) return <div>Error: {queryError.message}</div>

  return (
    <div className="flex-grow overflow-y-auto">
      <div className="w-full border-b px-6 py-12">
        <div className="mx-auto max-w-2xl">
          <h3 className="pb-1 text-2xl font-normal leading-10 text-gray-900">Notification Settings</h3>
          <p className="mb-7 text-sm text-gray-500">Configure Slack notifications.</p>
          <div className="flex items-start space-x-3 pb-6">
            <div>
              <span
                className={clsx('toggle', { 'toggle-checked': slackEnabled })}
                role="checkbox"
                tabIndex={0}
                aria-checked={slackEnabled}
                onClick={() => setSlackEnabled(!slackEnabled)}
              >
                <span className={clsx('toggle-switch', { 'toggle-switch-checked': slackEnabled })}></span>
              </span>
            </div>
            <div>
              <div
                className="cursor-pointer select-none text-sm font-medium leading-5 text-gray-700"
                onClick={() => setSlackEnabled(!slackEnabled)}
              >
                Enable Slack notifications
              </div>
            </div>
          </div>

          <form className="md:w-2/3" onSubmit={(event) => event.preventDefault()}>
            <div className="pb-6">
              <label className="block text-sm font-medium leading-5 text-gray-700">Webhook URL</label>
              <input
                className="form-input mt-1 w-full"
                id="slack-webhook-url"
                type="text"
                required={slackEnabled}
                disabled={!slackEnabled}
                ref={webhookUrlInput}
                defaultValue={data?.getUser?.slack.webhookUrl ?? undefined}
              ></input>
            </div>
            <div className="pb-6">
              <label className="block text-sm font-medium leading-5 text-gray-700">Channel name</label>
              <input
                className="form-input mt-1 w-full"
                id="slack-channel"
                type="text"
                required={false}
                disabled={!slackEnabled}
                ref={channelInput}
                defaultValue={data?.getUser?.slack.channel ?? undefined}
              ></input>
            </div>
            <div className="pb-1">
              <button
                className="btn btn-md"
                type="submit"
                onClick={() => {
                  updateUser({
                    variables: {
                      userId: props.userId,
                      input: {
                        slack: {
                          enabled: slackEnabled,
                          webhookUrl: webhookUrlInput.current?.value,
                          channel: channelInput.current?.value,
                        },
                      },
                    },
                  })
                }}
              >
                Save
              </button>
            </div>
          </form>
          {loading && <p>Saving...</p>}
          {error && <p>Error: {error.message}</p>}
        </div>
      </div>
    </div>
  )
}
