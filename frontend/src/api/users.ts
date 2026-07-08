import { apiFetch } from './client'
import type { User, CreateUserRequest } from '@/types'

export const getUsers = () => apiFetch<User[]>('/users')

export const createUser = (data: CreateUserRequest) =>
  apiFetch<User>('/users', { method: 'POST', body: JSON.stringify(data) })

export const updateUser = (id: number, data: Partial<User>) =>
  apiFetch<User>(`/users/${id}`, { method: 'PUT', body: JSON.stringify(data) })

export const deleteUser = (id: number) =>
  apiFetch<void>(`/users/${id}`, { method: 'DELETE' })

export const toggleUser = (id: number, active: boolean) =>
  apiFetch<User>(`/users/${id}`, { method: 'PUT', body: JSON.stringify({ isActive: active }) })
