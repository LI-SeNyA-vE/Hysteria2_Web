import { apiFetch } from './client'

export interface PanelSettings {
  nodeToken: string
}

export const getSettings = () => apiFetch<PanelSettings>('/settings')
