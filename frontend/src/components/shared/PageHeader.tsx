import type { ReactNode } from 'react'

interface PageHeaderProps {
  title: string
  description?: string
  action?: ReactNode
}

export function PageHeader({ title, description, action }: PageHeaderProps) {
  return (
    <div className="flex items-start justify-between mb-5">
      <div>
        <h1 className="text-[20px] font-semibold text-text tracking-tight">{title}</h1>
        {description && <p className="mt-1 text-[13px] text-dim">{description}</p>}
      </div>
      {action}
    </div>
  )
}
