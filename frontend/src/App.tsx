import { createBrowserRouter, RouterProvider } from 'react-router-dom'
import { AppLayout } from '@/components/layout/AppLayout'
import { Login } from '@/pages/Login'
import { Dashboard } from '@/pages/Dashboard'
import { Servers } from '@/pages/Servers'
import { Users } from '@/pages/Users'
import { Subscriptions } from '@/pages/Subscriptions'
import { Cascade } from '@/pages/Cascade'
import { Settings } from '@/pages/Settings'

const router = createBrowserRouter([
  { path: '/login', element: <Login /> },
  {
    path: '/',
    element: <AppLayout />,
    children: [
      { index: true,                  element: <Dashboard />    },
      { path: 'servers',              element: <Servers />      },
      { path: 'users',                element: <Users />        },
      { path: 'subscriptions',        element: <Subscriptions /> },
      { path: 'cascade',              element: <Cascade />      },
      { path: 'settings',             element: <Settings />     },
    ],
  },
])

export default function App() {
  return <RouterProvider router={router} />
}
