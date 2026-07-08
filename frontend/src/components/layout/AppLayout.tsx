import { Outlet, Navigate } from 'react-router-dom'
import { Sidebar } from './Sidebar'

export function AppLayout() {
  const token = localStorage.getItem('auth_token')
  if (!token) return <Navigate to="/login" replace />

  return (
    <div className="flex h-screen bg-bg overflow-hidden">
      <Sidebar />
      <main className="flex-1 overflow-y-auto bg-bg">
        <Outlet />
      </main>
    </div>
  )
}
