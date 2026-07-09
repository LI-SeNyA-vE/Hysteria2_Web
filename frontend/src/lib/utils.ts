import { clsx, type ClassValue } from 'clsx'
import { twMerge } from 'tailwind-merge'

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
}

export function formatBytes(gb: number): string {
  if (gb === 0) return '0 B'
  if (gb < 0.001) return `${(gb * 1024 * 1024).toFixed(0)} KB`
  if (gb < 1) return `${(gb * 1024).toFixed(1)} MB`
  if (gb < 1024) return `${gb.toFixed(1)} GB`
  return `${(gb / 1024).toFixed(2)} TB`
}

export function formatExpiry(dateStr: string | null): { label: string; expired: boolean } {
  if (!dateStr) return { label: 'Never', expired: false }
  const diff = new Date(dateStr).getTime() - Date.now()
  if (diff < 0) return { label: 'Expired', expired: true }
  const days = Math.floor(diff / 86_400_000)
  if (days === 0) return { label: 'Today', expired: false }
  if (days < 30) return { label: `${days}d`, expired: false }
  const months = Math.floor(days / 30)
  return { label: `${months}mo`, expired: false }
}

export function copyToClipboard(text: string): Promise<void> {
  if (navigator.clipboard && window.isSecureContext) {
    return navigator.clipboard.writeText(text)
  }
  // fallback for HTTP (non-secure context)
  return new Promise((resolve, reject) => {
    const el = document.createElement('textarea')
    el.value = text
    el.style.position = 'fixed'
    el.style.left = '-9999px'
    el.style.top = '-9999px'
    document.body.appendChild(el)
    el.focus()
    el.select()
    const ok = document.execCommand('copy')
    document.body.removeChild(el)
    ok ? resolve() : reject(new Error('execCommand failed'))
  })
}
