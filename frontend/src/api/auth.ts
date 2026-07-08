import { apiFetch } from './client'

export const login = (password: string) =>
  apiFetch<{ token: string }>('/login', {
    method: 'POST',
    body: JSON.stringify({ password }),
  })
