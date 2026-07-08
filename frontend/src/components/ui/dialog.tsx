import { useEffect, type ReactNode } from 'react'
import { X } from 'lucide-react'
import { cn } from '@/lib/utils'
import { Button } from './button'

interface DialogProps {
  open: boolean
  onClose: () => void
  title: string
  description?: string
  children: ReactNode
  className?: string
}

export function Dialog({ open, onClose, title, description, children, className }: DialogProps) {
  useEffect(() => {
    if (!open) return
    const h = (e: KeyboardEvent) => { if (e.key === 'Escape') onClose() }
    document.addEventListener('keydown', h)
    return () => document.removeEventListener('keydown', h)
  }, [open, onClose])

  if (!open) return null

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4">
      {/* Overlay */}
      <div
        className="absolute inset-0 backdrop-blur-sm"
        style={{ background: 'rgba(0,0,0,0.7)' }}
        onClick={onClose}
      />

      {/* Панель */}
      <div
        className={cn('relative w-full max-w-md rounded-2xl', className)}
        style={{
          background: 'linear-gradient(180deg, #1c1c24 0%, #161619 100%)',
          boxShadow: '0 0 0 1px rgba(255,255,255,0.09), 0 25px 60px rgba(0,0,0,0.7), 0 8px 20px rgba(0,0,0,0.4)',
        }}
      >
        {/* Шапка */}
        <div className="flex items-start justify-between px-7 pt-6 pb-5"
          style={{ borderBottom: '1px solid rgba(255,255,255,0.06)' }}>
          <div>
            <h2 className="text-[17px] font-semibold text-text" style={{ letterSpacing: '-0.01em' }}>{title}</h2>
            {description && <p className="text-[13px] text-dim" style={{ marginTop: 6 }}>{description}</p>}
          </div>
          <Button variant="ghost" size="icon" onClick={onClose}>
            <X className="w-4 h-4" />
          </Button>
        </div>

        {/* Контент */}
        <div className="px-7 py-7">{children}</div>
      </div>
    </div>
  )
}
