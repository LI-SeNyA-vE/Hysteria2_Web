import { apiFetch } from './client'
import type { Server, DashboardStats, NodeConfig } from '@/types'

export const getServers = () => apiFetch<Server[]>('/servers')
export const getStats = () => apiFetch<DashboardStats>('/stats')
export const deleteServer = (id: number) =>
  apiFetch<void>(`/servers/${id}`, { method: 'DELETE' })
export const getServerLogs = (id: number) =>
  apiFetch<{ lines: string[] }>(`/servers/${id}/logs`)
export const getNodeConfig = (id: number) =>
  apiFetch<NodeConfig>(`/servers/${id}/config`)
export const saveNodeConfig = (id: number, data: NodeConfig) =>
  apiFetch<void>(`/servers/${id}/config`, { method: 'PUT', body: JSON.stringify(data) })
export const pushNodeUpdate = () =>
  apiFetch<{ version: string; message: string }>('/update/push-nodes', { method: 'POST' })
