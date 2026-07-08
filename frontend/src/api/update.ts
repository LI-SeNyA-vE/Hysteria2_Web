import { apiFetch } from './client'

export interface UpdateInfo {
  currentVersion: string
  latestVersion: string
  updateAvailable: boolean
  releaseUrl: string
}

export const checkUpdate = () => apiFetch<UpdateInfo>('/update/check')
export const applyUpdate = () => apiFetch<{ message: string; version: string }>('/update/apply', { method: 'POST' })
