import { apiFetch } from './client'
import type { Server, DashboardStats } from '@/types'

export const getServers = () => apiFetch<Server[]>('/servers')
export const getStats = () => apiFetch<DashboardStats>('/stats')
export const deleteServer = (id: number) =>
  apiFetch<void>(`/servers/${id}`, { method: 'DELETE' })
