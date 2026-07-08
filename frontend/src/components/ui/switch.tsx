import { cn } from '@/lib/utils'

interface SwitchProps {
  checked: boolean
  onCheckedChange: (v: boolean) => void
  disabled?: boolean
  className?: string
}

export function Switch({ checked, onCheckedChange, disabled, className }: SwitchProps) {
  return (
    <button
      role="switch"
      aria-checked={checked}
      disabled={disabled}
      onClick={() => onCheckedChange(!checked)}
      className={cn('relative w-10 h-[22px] rounded-full cursor-pointer transition-all duration-200',
        'disabled:opacity-40 disabled:cursor-not-allowed', className)}
      style={{
        background: checked ? 'linear-gradient(135deg, #8a7ef8, #6a5fe6)' : 'rgba(255,255,255,0.08)',
        boxShadow: checked ? '0 0 10px rgba(124,111,247,0.4), 0 0 0 1px rgba(138,126,248,0.3)' : '0 0 0 1px rgba(255,255,255,0.1)',
      }}
    >
      <span
        className={cn(
          'absolute top-[2px] left-[2px] w-[18px] h-[18px] rounded-full bg-white',
          'transition-transform duration-200',
          checked && 'translate-x-[18px]',
        )}
        style={{ boxShadow: '0 1px 3px rgba(0,0,0,0.4)' }}
      />
    </button>
  )
}
