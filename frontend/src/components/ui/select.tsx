import { forwardRef, type SelectHTMLAttributes } from 'react'
import { cn } from '@/lib/utils'

interface SelectProps extends SelectHTMLAttributes<HTMLSelectElement> {
  label?: string
}

export const Select = forwardRef<HTMLSelectElement, SelectProps>(
  ({ className, label, id, children, ...props }, ref) => (
    <div className="w-full">
      {label && (
        <label htmlFor={id} className="block text-[12px] font-medium text-dim uppercase tracking-[0.06em] mb-2">
          {label}
        </label>
      )}
      <select
        ref={ref}
        id={id}
        className={cn(
          'w-full h-10 px-3.5 rounded-lg text-[14px] text-text outline-none',
          'transition-all duration-150 cursor-pointer appearance-none',
          className,
        )}
        style={{
          background: 'rgba(255,255,255,0.05)',
          boxShadow: '0 0 0 1px rgba(255,255,255,0.09)',
        }}
        onFocus={e => {
          e.currentTarget.style.boxShadow = '0 0 0 1px rgba(124,111,247,0.5), 0 0 0 3px rgba(124,111,247,0.15)'
        }}
        onBlur={e => {
          e.currentTarget.style.boxShadow = '0 0 0 1px rgba(255,255,255,0.09)'
        }}
        {...props}
      >
        {children}
      </select>
    </div>
  )
)
Select.displayName = 'Select'
