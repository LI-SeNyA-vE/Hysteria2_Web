import { cn } from '@/lib/utils'
import type { ReactNode } from 'react'

const CARD_STYLE = {
  background: 'linear-gradient(180deg, #18181e 0%, #141418 100%)',
  boxShadow: '0 0 0 1px rgba(255,255,255,0.07), 0 1px 2px rgba(0,0,0,0.4), 0 4px 16px rgba(0,0,0,0.25)',
}

export function Card({ className, style, ...props }: React.HTMLAttributes<HTMLDivElement>) {
  return (
    <div
      className={cn('rounded-2xl', className)}
      style={{ ...CARD_STYLE, ...style }}
      {...props}
    />
  )
}

export function CardHeader({ className, ...props }: React.HTMLAttributes<HTMLDivElement>) {
  return <div className={cn('px-7 pt-6 pb-4', className)} {...props} />
}

export function CardTitle({ className, ...props }: React.HTMLAttributes<HTMLParagraphElement>) {
  return (
    <p className={cn('text-[12px] font-medium text-dim uppercase tracking-[0.08em]', className)} {...props} />
  )
}

export function CardContent({ className, ...props }: React.HTMLAttributes<HTMLDivElement>) {
  return <div className={cn('px-7 pb-7', className)} {...props} />
}

/* Stat card */
export function StatCard({
  label, value, sub, icon,
}: {
  label: string
  value: string | number
  sub?: string
  icon: ReactNode
}) {
  return (
    <div
      className="rounded-2xl p-7 flex flex-col gap-6"
      style={CARD_STYLE}
    >
      <div className="flex items-center justify-between">
        <span className="text-[12px] font-semibold text-dim uppercase tracking-[0.08em]">{label}</span>
        <span className="text-sub">{icon}</span>
      </div>
      <div>
        <div
          className="text-[34px] font-semibold text-text leading-none"
          style={{ letterSpacing: '-0.03em', fontVariantNumeric: 'tabular-nums' }}
        >
          {value}
        </div>
        {sub && <div className="mt-2 text-[13px] text-dim">{sub}</div>}
      </div>
    </div>
  )
}
