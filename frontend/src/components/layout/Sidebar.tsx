import { NavLink, useNavigate } from 'react-router-dom'
import {
  LayoutDashboard, Server, Users, Link2,
  GitBranch, Settings, LogOut, Zap,
} from 'lucide-react'
import { cn } from '@/lib/utils'

const NAV = [
  { to: '/',              icon: LayoutDashboard, label: 'Главная'       },
  { to: '/servers',       icon: Server,          label: 'Серверы'       },
  { to: '/users',         icon: Users,           label: 'Пользователи'  },
  { to: '/subscriptions', icon: Link2,           label: 'Подписки'      },
  { to: '/cascade',       icon: GitBranch,       label: 'Каскад'        },
  { to: '/settings',      icon: Settings,        label: 'Настройки'     },
]

export function Sidebar() {
  const navigate = useNavigate()

  return (
    <aside
      className="w-[250px] h-screen flex flex-col shrink-0"
      style={{ background: '#111115', borderRight: '1px solid rgba(255,255,255,0.06)' }}
    >
      {/* Логотип */}
      <div
        className="h-[68px] flex items-center gap-3 px-6"
        style={{ borderBottom: '1px solid rgba(255,255,255,0.06)' }}
      >
        <div
          className="w-7 h-7 rounded-lg flex items-center justify-center shrink-0"
          style={{
            background: 'linear-gradient(135deg, #8a7ef8, #5b52d6)',
            boxShadow: '0 0 14px rgba(124,111,247,0.4)',
          }}
        >
          <Zap className="w-4 h-4 text-white" strokeWidth={2.5} />
        </div>
        <span className="text-[16px] font-semibold text-text" style={{ letterSpacing: '-0.01em' }}>
          H2 Panel
        </span>
      </div>

      {/* Навигация */}
      <nav className="flex-1 px-4 py-4 space-y-1 overflow-y-auto">
        {NAV.map(({ to, icon: Icon, label }) => (
          <NavLink
            key={to}
            to={to}
            end={to === '/'}
            className={({ isActive }) =>
              cn(
                'flex items-center gap-3 px-4 py-3 rounded-xl text-[15px] transition-all duration-100 group',
                isActive ? 'text-text font-medium' : 'text-sub hover:text-text',
              )
            }
            style={({ isActive }) =>
              isActive ? { background: 'rgba(124,111,247,0.13)' } : undefined
            }
          >
            {({ isActive }) => (
              <>
                <Icon
                  className={cn(
                    'w-[17px] h-[17px] shrink-0 transition-colors duration-100',
                    isActive ? 'text-accent' : 'text-dim group-hover:text-sub',
                  )}
                  strokeWidth={isActive ? 2 : 1.75}
                />
                {label}
              </>
            )}
          </NavLink>
        ))}
      </nav>

      {/* Выход */}
      <div className="px-4 py-4" style={{ borderTop: '1px solid rgba(255,255,255,0.06)' }}>
        <button
          onClick={() => { localStorage.removeItem('auth_token'); navigate('/login') }}
          className="w-full flex items-center gap-3 px-4 py-3 rounded-xl text-[15px] text-dim hover:text-sub transition-all duration-100 group"
        >
          <LogOut className="w-[17px] h-[17px] shrink-0" strokeWidth={1.75} />
          Выйти
        </button>
      </div>
    </aside>
  )
}
