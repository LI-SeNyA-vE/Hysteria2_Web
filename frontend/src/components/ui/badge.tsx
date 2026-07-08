import { cn } from '@/lib/utils'

type V = 'default' | 'success' | 'danger' | 'warning' | 'muted'

interface BadgeProps {
  children: React.ReactNode
  variant?: V
  dot?: boolean
  className?: string
}

const BADGE: Record<V, { bg: string; text: string; dot: string }> = {
  default: { bg: 'rgba(124,111,247,0.12)', text: '#9d94f8', dot: '#7c6ff7' },
  success: { bg: 'rgba(62,207,142,0.10)',  text: '#3ecf8e', dot: '#3ecf8e' },
  danger:  { bg: 'rgba(240,99,118,0.10)',  text: '#f06376', dot: '#f06376' },
  warning: { bg: 'rgba(245,166,35,0.10)',  text: '#f5a623', dot: '#f5a623' },
  muted:   { bg: 'rgba(255,255,255,0.05)', text: '#52525f', dot: '#52525f' },
}

export function Badge({ children, variant = 'default', dot = false, className }: BadgeProps) {
  const s = BADGE[variant]
  return (
    <span
      className={cn('inline-flex items-center gap-2 px-3 py-1.5 rounded-full text-[12px] font-medium whitespace-nowrap leading-none', className)}
      style={{ background: s.bg, color: s.text }}
    >
      {dot && (
        <span className="w-[5px] h-[5px] rounded-full shrink-0 animate-pulse"
          style={{ background: s.dot, boxShadow: `0 0 4px ${s.dot}` }} />
      )}
      {children}
    </span>
  )
}
