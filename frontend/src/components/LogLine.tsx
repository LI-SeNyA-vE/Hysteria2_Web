// eslint-disable-next-line no-control-regex
const ANSI_RE = /\x1b\[[0-9;]*m/g

const LEVEL_COLORS: Record<string, string> = {
  DEBUG:   '#6b7280',
  INFO:    '#60a5fa',
  WARN:    '#fbbf24',
  WARNING: '#fbbf24',
  ERROR:   '#f87171',
  FATAL:   '#f87171',
  PANIC:   '#f87171',
}

export function LogLine({ raw }: { raw: string }) {
  const line = raw.replace(ANSI_RE, '')

  // Формат hysteria2: "2026-07-13T09:43:03Z  INFO  message  {fields}"
  const hy2 = line.match(/^(\d{4}-\d{2}-\d{2}T[\d:]+Z)\s+(DEBUG|INFO|WARN(?:ING)?|ERROR|FATAL|PANIC)\s+(.+)$/)
  if (hy2) {
    const [, ts, level, rest] = hy2
    const lvlColor = LEVEL_COLORS[level] ?? '#94a3b8'
    const tabIdx = rest.lastIndexOf('\t')
    const msg    = tabIdx > 0 ? rest.slice(0, tabIdx).replace(/\t/g, '  ') : rest
    const fields = tabIdx > 0 ? rest.slice(tabIdx + 1) : ''
    return (
      <div className="text-[12px] font-mono leading-relaxed select-text" style={{ whiteSpace: 'pre-wrap', wordBreak: 'break-all' }}>
        <span style={{ color: '#4b5563' }}>{ts}{'  '}</span>
        <span style={{ color: lvlColor, fontWeight: 600 }}>{level}</span>
        <span style={{ color: '#94a3b8' }}>{'  '}{msg}</span>
        {fields && <span style={{ color: '#4b5563' }}>{'  '}{fields}</span>}
      </div>
    )
  }

  // Формат Go log: "2026/07/13 09:43:03 message"
  const goLog = line.match(/^(\d{4}\/\d{2}\/\d{2} [\d:]+) (.+)$/)
  if (goLog) {
    const [, ts, msg] = goLog
    const l = msg.toLowerCase()
    const color = l.includes('error') || l.includes('fatal') ? '#f87171'
      : l.includes('warn') ? '#fbbf24'
      : '#94a3b8'
    return (
      <div className="text-[12px] font-mono leading-relaxed select-text" style={{ whiteSpace: 'pre-wrap', wordBreak: 'break-all' }}>
        <span style={{ color: '#4b5563' }}>{ts}{'  '}</span>
        <span style={{ color }}>{msg}</span>
      </div>
    )
  }

  // Fallback
  const l = line.toLowerCase()
  const color = l.includes('error') || l.includes('fatal') ? '#f87171'
    : l.includes('warn') ? '#fbbf24'
    : '#6b7280'
  return (
    <div className="text-[12px] font-mono leading-relaxed select-text" style={{ color, whiteSpace: 'pre-wrap', wordBreak: 'break-all' }}>
      {line}
    </div>
  )
}
