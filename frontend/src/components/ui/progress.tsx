import { cn } from '@/lib/utils'

export function Progress({ value, className }: { value: number; className?: string }) {
  const pct = Math.min(100, Math.max(0, value))
  const color = pct > 85 ? '#f06376' : pct > 60 ? '#f5a623' : '#7c6ff7'
  return (
    <div
      className={cn('h-[4px] w-full rounded-full overflow-hidden', className)}
      style={{ background: 'rgba(255,255,255,0.07)' }}
    >
      <div
        className="h-full rounded-full transition-all duration-500"
        style={{
          width: `${pct}%`,
          background: `linear-gradient(90deg, ${color}99, ${color})`,
          boxShadow: pct > 0 ? `0 0 6px ${color}60` : 'none',
        }}
      />
    </div>
  )
}
