import { apiFetch } from './client'

export interface LogsResponse {
  lines: string[]
}

export const getPanelLogs    = () => apiFetch<LogsResponse>('/logs?source=panel')
export const getHysteriaLogs = () => apiFetch<LogsResponse>('/logs?source=hysteria')
