import { apiFetch } from './client'
import type { HysteriaStatus, HysteriaConfig } from '@/types'

export const getHysteriaStatus = () => apiFetch<HysteriaStatus>('/hysteria/status')
export const getHysteriaConfig = () => apiFetch<HysteriaConfig>('/hysteria/config')

export const installHysteria = () =>
  apiFetch<void>('/hysteria/install', { method: 'POST' })

export const startHysteria = () =>
  apiFetch<void>('/hysteria/start', { method: 'POST' })

export const stopHysteria = () =>
  apiFetch<void>('/hysteria/stop', { method: 'POST' })

export const reloadConfig = () =>
  apiFetch<void>('/hysteria/reload-config', { method: 'POST' })

export const saveHysteriaConfig = (data: Partial<HysteriaConfig>) =>
  apiFetch<void>('/hysteria/config', { method: 'PUT', body: JSON.stringify(data) })

export const regenerateCert = () =>
  apiFetch<{ sha256: string }>('/hysteria/cert/regenerate', { method: 'POST' })
