export type ServerRole = 'main' | 'node1' | 'node2' | 'main_node1'

export interface Server {
  id: number
  role: ServerRole
  name: string
  displayName: string
  publicIp: string
  panelUrl: string
  hy2Port: number
  hy2Version: string
  hy2Running: boolean
  panelVersion: string
  createdAt: string
  lastSeenAt: string | null
}

export interface User {
  id: number
  name: string
  uuid: string
  password: string
  trafficLimitGb: number
  trafficUsedGb: number
  expireAt: string | null
  isActive: boolean
  serverId: number
  createdAt: string
}

export interface Subscription {
  id: number
  userId: number
  userName: string
  token: string
  name: string
  lastAccessedAt: string | null
  createdAt: string
}

export interface CascadeLink {
  id: number
  node1Id: number
  node1Ip: string
  node2Id: number
  node2Ip: string
  node2Password: string
  isActive: boolean
}

export interface HysteriaStatus {
  installed: boolean
  running: boolean
  version: string
  port: number
}

export interface DashboardStats {
  totalUsers: number
  activeUsers: number
  totalTrafficGb: number
  uptime: string
  hysteria: HysteriaStatus
}

export interface CreateUserRequest {
  name: string
  trafficLimitGb: number
  expireDays: number | null
  serverId: number
}

export interface CreateSubscriptionRequest {
  userId: number
  name: string
}

export interface NodeConfig {
  bandwidthUp: string
  bandwidthDown: string
  masqueradeUrl: string
}

export interface HysteriaConfig {
  port: number
  obfsPassword: string
  masqueradeUrl: string
  certSha256: string
  bandwidthUp: string
  bandwidthDown: string
  sni: string
}
