import { useState, type FormEvent } from 'react'
import { useNavigate, Navigate } from 'react-router-dom'
import { Zap } from 'lucide-react'
import { login } from '@/api/auth'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'

const CARD_STYLE = {
  background: 'linear-gradient(180deg, #18181e 0%, #141418 100%)',
  boxShadow: '0 0 0 1px rgba(255,255,255,0.08), 0 8px 32px rgba(0,0,0,0.5), 0 1px 2px rgba(0,0,0,0.4)',
}

export function Login() {
  const navigate = useNavigate()
  const [password, setPassword] = useState('')
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)

  if (localStorage.getItem('auth_token')) return <Navigate to="/" replace />

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault()
    setError('')
    setLoading(true)

    if (import.meta.env.DEV) {
      localStorage.setItem('auth_token', 'dev-token')
      navigate('/')
      return
    }

    try {
      const { token } = await login(password)
      localStorage.setItem('auth_token', token)
      navigate('/')
    } catch {
      setError('Неверный пароль')
      setLoading(false)
    }
  }

  return (
    <div className="min-h-screen bg-bg flex items-center justify-center p-4">
      {/* Фоновое свечение */}
      <div className="absolute inset-0 pointer-events-none" style={{ overflow: 'hidden' }}>
        <div
          className="absolute top-1/3 left-1/2 -translate-x-1/2 -translate-y-1/2 w-[600px] h-[400px] rounded-full"
          style={{ background: 'rgba(124,111,247,0.06)', filter: 'blur(120px)' }}
        />
      </div>

      <div className="relative w-full" style={{ maxWidth: 360 }}>
        {/* Логотип */}
        <div className="flex flex-col items-center" style={{ marginBottom: 28 }}>
          <div
            className="w-12 h-12 rounded-2xl flex items-center justify-center mb-4"
            style={{
              background: 'rgba(124,111,247,0.12)',
              boxShadow: '0 0 0 1px rgba(124,111,247,0.2), 0 0 30px rgba(124,111,247,0.15)',
            }}
          >
            <Zap className="w-5 h-5 text-accent" strokeWidth={2.5} />
          </div>
          <h1 className="text-[22px] font-semibold text-text" style={{ letterSpacing: '-0.02em' }}>H2 Panel</h1>
          <p className="text-[14px] text-dim" style={{ marginTop: 6 }}>Управление Hysteria2</p>
        </div>

        {/* Карточка входа */}
        <div className="relative rounded-2xl p-7" style={CARD_STYLE}>
          {/* Верхняя полоска-свечение */}
          <div
            className="absolute top-0 inset-x-0 h-px rounded-t-2xl"
            style={{ background: 'linear-gradient(90deg, transparent, rgba(124,111,247,0.4), transparent)' }}
          />

          <form onSubmit={handleSubmit} className="space-y-4">
            <Input
              id="password"
              type="password"
              label="Пароль"
              placeholder="Введите пароль панели"
              value={password}
              onChange={e => setPassword(e.target.value)}
              error={error}
              autoFocus
            />
            <Button type="submit" size="lg" className="w-full" loading={loading}>
              Войти
            </Button>
          </form>
        </div>

        <p className="text-center text-[12px] text-dim" style={{ marginTop: 16 }}>
          Hysteria2 Panel · VPN Management System
        </p>
      </div>
    </div>
  )
}
