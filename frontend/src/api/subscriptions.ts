import { apiFetch } from './client'
import type { Subscription, CreateSubscriptionRequest } from '@/types'

export const getSubscriptions = () => apiFetch<Subscription[]>('/subscriptions')

export const createSubscription = (data: CreateSubscriptionRequest) =>
  apiFetch<Subscription>('/subscriptions', { method: 'POST', body: JSON.stringify(data) })

export const deleteSubscription = (id: number) =>
  apiFetch<void>(`/subscriptions/${id}`, { method: 'DELETE' })
