import { forwardRef, type ButtonHTMLAttributes } from 'react'
import { cn } from '@/lib/utils'

interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: 'default' | 'outline' | 'ghost' | 'danger'
  size?: 'sm' | 'md' | 'lg' | 'icon'
  loading?: boolean
}

const STYLES = {
  default: {
    background: 'linear-gradient(180deg, #8a7ef8 0%, #6a5fe6 100%)',
    boxShadow: '0 1px 2px rgba(0,0,0,0.4), 0 0 0 1px rgba(255,255,255,0.12) inset, 0 -1px 0 rgba(0,0,0,0.3) inset',
    color: '#fff',
  },
  outline: {
    background: 'rgba(255,255,255,0.04)',
    boxShadow: '0 0 0 1px rgba(255,255,255,0.10)',
    color: '#8b8b9e',
  },
  ghost: {
    background: 'transparent',
    boxShadow: 'none',
    color: '#52525f',
  },
  danger: {
    background: 'rgba(240,99,118,0.12)',
    boxShadow: '0 0 0 1px rgba(240,99,118,0.25)',
    color: '#f06376',
  },
}

export const Button = forwardRef<HTMLButtonElement, ButtonProps>(
  ({ className, variant = 'default', size = 'md', loading, children, disabled, style, ...props }, ref) => (
    <button
      ref={ref}
      disabled={disabled ?? loading}
      style={{ ...STYLES[variant], ...style }}
      className={cn(
        'inline-flex items-center justify-center gap-1.5 font-medium rounded-lg',
        'transition-all duration-100 cursor-pointer select-none',
        'hover:brightness-110 active:scale-[0.98]',
        'disabled:opacity-40 disabled:cursor-not-allowed disabled:pointer-events-none',
        size === 'sm'   && 'h-8 px-3 text-[13px] rounded-lg gap-1',
        size === 'md'   && 'h-9 px-4 text-[14px]',
        size === 'lg'   && 'h-10 px-5 text-[14px] rounded-xl',
        size === 'icon' && 'h-8 w-8 p-0 rounded-md',
        className,
      )}
      {...props}
    >
      {loading ? (
        <svg className="animate-spin w-3.5 h-3.5" viewBox="0 0 24 24" fill="none">
          <circle cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="3" opacity=".25" />
          <path fill="currentColor" d="M4 12a8 8 0 018-8V0C5.4 0 0 5.4 0 12h4z" opacity=".75" />
        </svg>
      ) : children}
    </button>
  )
)
Button.displayName = 'Button'
