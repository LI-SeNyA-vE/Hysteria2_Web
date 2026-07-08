import { forwardRef, type InputHTMLAttributes } from 'react'
import { cn } from '@/lib/utils'

interface InputProps extends InputHTMLAttributes<HTMLInputElement> {
  label?: string
  hint?: string
  error?: string
}

export const Input = forwardRef<HTMLInputElement, InputProps>(
  ({ className, label, hint, error, id, ...props }, ref) => (
    <div className="w-full">
      {label && (
        <label htmlFor={id} className="block text-[12px] font-medium text-dim uppercase tracking-[0.06em] mb-2">
          {label}
        </label>
      )}
      <input
        ref={ref}
        id={id}
        className={cn(
          'w-full h-10 px-3.5 rounded-lg text-[14px] text-text outline-none',
          'placeholder:text-dim transition-all duration-150',
          className,
        )}
        style={{
          background: 'rgba(255,255,255,0.05)',
          boxShadow: error
            ? '0 0 0 1px rgba(240,99,118,0.5)'
            : '0 0 0 1px rgba(255,255,255,0.09)',
        }}
        onFocus={e => {
          if (!error) e.currentTarget.style.boxShadow = '0 0 0 1px rgba(124,111,247,0.5), 0 0 0 3px rgba(124,111,247,0.15)'
        }}
        onBlur={e => {
          if (!error) e.currentTarget.style.boxShadow = '0 0 0 1px rgba(255,255,255,0.09)'
        }}
        {...props}
      />
      {hint && !error && <p className="mt-1.5 text-[12px] text-dim">{hint}</p>}
      {error && <p className="mt-1.5 text-[12px] text-danger">{error}</p>}
    </div>
  )
)
Input.displayName = 'Input'
