import { useEffect, useRef, useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { Terminal } from 'lucide-react'
import { getPanelLogs, getHysteriaLogs } from '@/api/logs'
import { LogLine } from '@/components/LogLine'

type Source = 'panel' | 'hysteria'

const TABS: { id: Source; label: string }[] = [
  { id: 'panel',    label: 'Панель'    },
  { id: 'hysteria', label: 'Hysteria2' },
]

export function Logs() {
  const [source, setSource] = useState<Source>('panel')
  const bottomRef = useRef<HTMLDivElement>(null)

  const { data, isFetching } = useQuery({
    queryKey: ['logs', source],
    queryFn: source === 'panel' ? getPanelLogs : getHysteriaLogs,
    refetchInterval: 3_000,
    placeholderData: (prev) => prev,
  })

  const lines = data?.lines ?? []

  // Авто-скролл вниз при появлении новых строк
  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: 'smooth' })
  }, [lines.length])

  return (
    <div className="min-h-screen flex flex-col" style={{ padding: '40px 48px' }}>
      {/* Заголовок */}
      <div style={{ marginBottom: 24 }}>
        <h1 className="text-[22px] font-semibold text-text" style={{ letterSpacing: '-0.02em' }}>Логи</h1>
        <p className="text-[13px] text-dim" style={{ marginTop: 6 }}>Вывод процессов в реальном времени</p>
      </div>

      {/* Табы */}
      <div className="flex items-center gap-2" style={{ marginBottom: 16 }}>
        {TABS.map(tab => (
          <button
            key={tab.id}
            onClick={() => setSource(tab.id)}
            className="px-4 py-1.5 rounded-lg text-[13px] font-medium transition-colors"
            style={source === tab.id ? {
              background: 'rgba(124,111,247,0.15)',
              color: '#a5b4fc',
              boxShadow: '0 0 0 1px rgba(124,111,247,0.3)',
            } : {
              background: 'rgba(255,255,255,0.04)',
              color: 'var(--color-dim, #64748b)',
              boxShadow: '0 0 0 1px rgba(255,255,255,0.07)',
            }}
          >
            {tab.label}
          </button>
        ))}
        <div className="flex items-center gap-1.5 ml-auto">
          {isFetching && (
            <span className="text-[12px] text-dim">обновление...</span>
          )}
          <span className="text-[12px] text-dim">{lines.length} строк</span>
        </div>
      </div>

      {/* Терминал */}
      <div
        className="flex-1 rounded-2xl overflow-hidden flex flex-col"
        style={{
          background: '#0d0d10',
          boxShadow: '0 0 0 1px rgba(255,255,255,0.07), 0 4px 16px rgba(0,0,0,0.4)',
          minHeight: 480,
        }}
      >
        {/* Шапка терминала */}
        <div
          className="flex items-center gap-2 px-4 py-3"
          style={{ borderBottom: '1px solid rgba(255,255,255,0.06)' }}
        >
          <div className="w-3 h-3 rounded-full" style={{ background: '#ff5f57' }} />
          <div className="w-3 h-3 rounded-full" style={{ background: '#febc2e' }} />
          <div className="w-3 h-3 rounded-full" style={{ background: '#28c840' }} />
          <span className="text-[12px] text-dim ml-2 font-mono">
            {source === 'panel' ? 'hysteria2-panel' : 'hysteria2 server'}
          </span>
          <Terminal className="w-3.5 h-3.5 text-dim ml-auto" />
        </div>

        {/* Строки логов */}
        <div className="flex-1 overflow-y-auto px-4 py-3" style={{ maxHeight: 600 }}>
          {lines.length === 0 ? (
            <p className="text-[13px] font-mono" style={{ color: '#4b5563' }}>
              {source === 'hysteria'
                ? '— Hysteria2 ещё не запущен или нет вывода —'
                : '— Логи панели появятся здесь —'}
            </p>
          ) : (
            lines.map((line, i) => <LogLine key={i} raw={line} />)
          )}
          <div ref={bottomRef} />
        </div>
      </div>
    </div>
  )
}
