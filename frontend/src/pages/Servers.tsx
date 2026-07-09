import { useState, useEffect, useRef } from 'react'
import { useQuery, useQueryClient, useMutation } from '@tanstack/react-query'
import { Server, ChevronDown, ChevronUp, Copy, Check, ScrollText, RefreshCw, Upload } from 'lucide-react'
import { getServers, getServerLogs, pushNodeUpdate } from '@/api/servers'
import { getSettings } from '@/api/settings'
import { Card, CardContent } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Dialog } from '@/components/ui/dialog'
import { copyToClipboard } from '@/lib/utils'
import type { Server as ServerType, ServerRole } from '@/types'

const ROLE_RU: Record<ServerRole, string> = {
  main:       'Главный',
  node1:      'Узел 1',
  node2:      'Узел 2',
  main_node1: 'Главный + Узел 1',
}
const ROLE_COLOR: Record<ServerRole, 'default' | 'success' | 'warning' | 'muted'> = {
  main:       'default',
  node1:      'success',
  node2:      'warning',
  main_node1: 'default',
}

const ICON_BOX = {
  background: 'rgba(255,255,255,0.05)',
  boxShadow: '0 0 0 1px rgba(255,255,255,0.08)',
}

const CARD_BOX = {
  background: 'linear-gradient(180deg, #18181e 0%, #141418 100%)',
  boxShadow: '0 0 0 1px rgba(255,255,255,0.07), 0 1px 2px rgba(0,0,0,0.4), 0 4px 16px rgba(0,0,0,0.25)',
}

// Main-нода онлайн всегда (нет heartbeat-loop на себя), ноды — по lastSeenAt
function isOnline(server: ServerType): boolean {
  if (server.role === 'main' || server.role === 'main_node1') return true
  if (!server.lastSeenAt) return false
  return Date.now() - new Date(server.lastSeenAt).getTime() < 30_000
}

export function Servers() {
  const qc = useQueryClient()
  const { data: servers = [], isFetching } = useQuery({
    queryKey: ['servers'],
    queryFn: getServers,
    placeholderData: [],
    refetchInterval: 30_000,
  })

  const isEmpty = servers.length === 0
  const [connectOpen, setConnectOpen] = useState(isEmpty)
  const [logsServer, setLogsServer] = useState<ServerType | null>(null)
  const [updateMsg, setUpdateMsg] = useState('')

  const pushMut = useMutation({
    mutationFn: pushNodeUpdate,
    onSuccess: (res) => {
      setUpdateMsg(`Ноды получат ${res.version} на следующем heartbeat (~10 сек)`)
      setTimeout(() => setUpdateMsg(''), 6000)
      setTimeout(() => qc.invalidateQueries({ queryKey: ['servers'] }), 15_000)
    },
  })

  const nodes = servers.filter(s => s.role !== 'main' && s.role !== 'main_node1')

  return (
    <div className="min-h-screen" style={{ padding: '40px 48px' }}>
      <div className="flex items-center justify-between" style={{ marginBottom: 28 }}>
        <div>
          <h1 className="text-[22px] font-semibold text-text" style={{ letterSpacing: '-0.02em' }}>Серверы</h1>
          <p className="text-[13px] text-dim" style={{ marginTop: 6 }}>Узлы каскадной сети</p>
        </div>
        <div className="flex items-center gap-2">
          {nodes.length > 0 && (
            <div className="flex items-center gap-2">
              {updateMsg && <span className="text-[12px] text-success">{updateMsg}</span>}
              <Button
                variant="outline" size="sm"
                loading={pushMut.isPending}
                onClick={() => pushMut.mutate()}
              >
                <Upload className="w-3.5 h-3.5" /> Обновить ноды
              </Button>
            </div>
          )}
          <Button
            variant="outline" size="sm"
            loading={isFetching}
            onClick={() => qc.invalidateQueries({ queryKey: ['servers'] })}
          >
            <RefreshCw className="w-3.5 h-3.5" /> Обновить
          </Button>
        </div>
      </div>

      <ConnectNodeBlock
        open={connectOpen}
        onToggle={() => setConnectOpen(o => !o)}
        servers={servers}
      />

      <div style={{ marginTop: 24 }}>
        {servers.length === 0 ? (
          <div className="rounded-2xl py-16 text-center" style={CARD_BOX}>
            <div className="w-12 h-12 rounded-2xl flex items-center justify-center mx-auto mb-4" style={ICON_BOX}>
              <Server className="w-6 h-6 text-dim" strokeWidth={1.5} />
            </div>
            <p className="text-[15px] font-medium text-sub">Серверы не настроены</p>
            <p className="text-[13px] text-dim" style={{ marginTop: 6 }}>Используйте блок выше чтобы добавить первый узел</p>
          </div>
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-5">
            {servers.map(server => (
              <Card key={server.id}>
                <CardContent className="pt-6">
                  <div className="flex items-start justify-between mb-5">
                    <div className="flex items-center gap-3">
                      <div className="w-9 h-9 rounded-xl flex items-center justify-center" style={ICON_BOX}>
                        <Server className="w-4 h-4 text-sub" strokeWidth={1.5} />
                      </div>
                      <span className="text-[15px] font-medium text-text">{server.name}</span>
                    </div>
                    <Badge variant={ROLE_COLOR[server.role]}>{ROLE_RU[server.role]}</Badge>
                  </div>

                  <div className="space-y-3">
                    <InfoRow label="IP адрес"   value={server.publicIp}      />
                    <InfoRow label="Порт"        value={`:${server.hy2Port}`} />
                    <InfoRow label="Hysteria2"   value={server.hy2Version}    />
                    {server.panelVersion && (
                      <InfoRow label="Панель" value={server.panelVersion} />
                    )}
                    <InfoRow label="Статус"      value={
                      isOnline(server)
                        ? <Badge dot variant="success">В сети</Badge>
                        : <Badge dot variant="danger">Оффлайн</Badge>
                    } />
                  </div>

                  <div style={{ marginTop: 16, borderTop: '1px solid rgba(255,255,255,0.06)', paddingTop: 14 }}>
                    <Button variant="outline" size="sm" onClick={() => setLogsServer(server)}>
                      <ScrollText className="w-3.5 h-3.5" /> Логи hysteria2
                    </Button>
                  </div>
                </CardContent>
              </Card>
            ))}
          </div>
        )}
      </div>

      <ServerLogsDialog server={logsServer} onClose={() => setLogsServer(null)} />
    </div>
  )
}

// ── ConnectNodeBlock ───────────────────────────────────────────────────────────

function ConnectNodeBlock({
  open, onToggle, servers,
}: {
  open: boolean
  onToggle: () => void
  servers: ServerType[]
}) {
  const { data: settings } = useQuery({
    queryKey: ['settings'],
    queryFn: getSettings,
    enabled: open,
  })

  const [selectedRole, setSelectedRole] = useState<'node2' | 'node1'>('node2')
  const [cascadeTarget, setCascadeTarget] = useState('')

  const panelUrl = window.location.origin
  const nodeToken = settings?.nodeToken ?? ''
  const node2Servers = servers.filter(s => s.role === 'node2')

  const INSTALL_URL = 'https://github.com/LI-SeNyA-vE/Hysteria2_Web/releases/latest/download/install.sh'
  const envParts = [
    `PANEL_ROLE=${selectedRole}`,
    `PANEL_MAIN_URL=${panelUrl}`,
    nodeToken ? `PANEL_NODE_TOKEN=${nodeToken}` : 'PANEL_NODE_TOKEN=<загрузка...>',
    selectedRole === 'node1' && cascadeTarget ? `PANEL_CASCADE_TARGET=${cascadeTarget}` : null,
  ].filter(Boolean).join(' ')
  const installCmd = `curl -fsSL ${INSTALL_URL} | sudo env ${envParts} bash`

  return (
    <div className="rounded-2xl overflow-hidden" style={{
      background: 'linear-gradient(180deg, #18181e 0%, #141418 100%)',
      boxShadow: '0 0 0 1px rgba(255,255,255,0.07), 0 1px 2px rgba(0,0,0,0.4), 0 4px 16px rgba(0,0,0,0.25)',
    }}>
      {/* Заголовок */}
      <button
        onClick={onToggle}
        className="w-full flex items-center justify-between px-6 py-4 text-left hover:bg-white/[0.02] transition-colors"
      >
        <span className="text-[14px] font-medium text-sub">+ Подключить новый сервер</span>
        {open
          ? <ChevronUp className="w-4 h-4 text-dim" />
          : <ChevronDown className="w-4 h-4 text-dim" />
        }
      </button>

      {open && (
        <div className="px-6 pb-6" style={{ borderTop: '1px solid rgba(255,255,255,0.06)' }}>
          <div className="space-y-5 pt-5">

            {/* Выбор роли */}
            <div>
              <p className="text-[12px] text-dim mb-2">Роль нового сервера</p>
              <div className="flex gap-2">
                {(['node2', 'node1'] as const).map(role => (
                  <button
                    key={role}
                    onClick={() => { setSelectedRole(role); setCascadeTarget('') }}
                    className="px-4 py-1.5 rounded-lg text-[13px] font-medium transition-colors"
                    style={selectedRole === role ? {
                      background: 'rgba(255,255,255,0.12)',
                      color: 'var(--color-text, #f1f5f9)',
                      boxShadow: '0 0 0 1px rgba(255,255,255,0.15)',
                    } : {
                      background: 'rgba(255,255,255,0.04)',
                      color: 'var(--color-dim, #64748b)',
                      boxShadow: '0 0 0 1px rgba(255,255,255,0.06)',
                    }}
                  >
                    {role}
                  </button>
                ))}
              </div>
            </div>

            {/* Cascade target (только для node1) */}
            {selectedRole === 'node1' && (
              <div>
                <p className="text-[12px] text-dim mb-2">Каскад через (node2)</p>
                {node2Servers.length === 0 ? (
                  <p className="text-[12px]" style={{ color: '#f59e0b' }}>
                    Сначала установите хотя бы один node2, затем вернитесь сюда
                  </p>
                ) : (
                  <select
                    value={cascadeTarget}
                    onChange={e => setCascadeTarget(e.target.value)}
                    className="w-full rounded-xl px-4 py-2.5 text-[13px] text-sub outline-none"
                    style={{
                      background: 'rgba(255,255,255,0.04)',
                      boxShadow: '0 0 0 1px rgba(255,255,255,0.08)',
                      color: 'inherit',
                    }}
                  >
                    <option value="">Первый доступный node2</option>
                    {node2Servers.map(s => (
                      <option key={s.id} value={s.name}>{s.name} ({s.publicIp})</option>
                    ))}
                  </select>
                )}
              </div>
            )}

            {/* Готовая команда */}
            <div>
              <p className="text-[12px] text-dim mb-2">
                Выполните на новом сервере:
              </p>
              <CopyField value={installCmd} />
            </div>

            <p className="text-[12px] text-dim" style={{ opacity: 0.55 }}>
              Команда устанавливает всё автоматически — вопросов не будет.
              Сервер появится в списке через ~30 секунд.
            </p>
          </div>
        </div>
      )}
    </div>
  )
}

// ── CopyField ──────────────────────────────────────────────────────────────────

function CopyField({ value }: { value: string }) {
  const [copied, setCopied] = useState(false)

  const copy = () => {
    copyToClipboard(value).then(() => {
      setCopied(true)
      setTimeout(() => setCopied(false), 1800)
    })
  }

  return (
    <div
      className="flex items-start gap-3 rounded-xl px-4 py-3 cursor-pointer"
      style={{ background: 'rgba(255,255,255,0.04)', boxShadow: '0 0 0 1px rgba(255,255,255,0.08)' }}
      onClick={copy}
    >
      <span
        className="flex-1 text-[12px] text-sub break-all select-none"
        style={{ fontFamily: 'monospace', lineHeight: '1.7' }}
      >
        {value}
      </span>
      <button
        onClick={e => { e.stopPropagation(); copy() }}
        className="flex-shrink-0 mt-0.5 text-dim hover:text-sub transition-colors"
        title="Копировать"
      >
        {copied
          ? <Check className="w-4 h-4 text-green-400" />
          : <Copy className="w-4 h-4" />
        }
      </button>
    </div>
  )
}

// ── ServerLogsDialog ──────────────────────────────────────────────────────────

function ServerLogsDialog({ server, onClose }: { server: ServerType | null; onClose: () => void }) {
  const bottomRef = useRef<HTMLDivElement>(null)

  const { data, dataUpdatedAt } = useQuery({
    queryKey: ['server-logs', server?.id],
    queryFn: () => getServerLogs(server!.id),
    enabled: !!server,
    refetchInterval: 5_000,
  })

  const lines = data?.lines ?? []

  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: 'smooth' })
  }, [dataUpdatedAt])

  return (
    <Dialog
      open={!!server}
      onClose={onClose}
      title={`Логи — ${server?.name ?? ''}`}
      description={`hysteria2 • обновляется каждые 5 сек`}
    >
      <div
        className="rounded-xl overflow-y-auto"
        style={{
          height: 420,
          background: 'rgba(0,0,0,0.35)',
          boxShadow: '0 0 0 1px rgba(255,255,255,0.07)',
          padding: '12px 14px',
        }}
      >
        {lines.length === 0 ? (
          <p className="text-[12px] text-dim text-center" style={{ marginTop: 80 }}>
            Логи ещё не получены — нода пришлёт их при следующем heartbeat (≤10 сек)
          </p>
        ) : (
          lines.map((line, i) => (
            <div
              key={i}
              className="text-[11px] leading-relaxed select-text"
              style={{ fontFamily: 'monospace', color: logColor(line), wordBreak: 'break-all' }}
            >
              {line}
            </div>
          ))
        )}
        <div ref={bottomRef} />
      </div>
    </Dialog>
  )
}

function logColor(line: string): string {
  const l = line.toLowerCase()
  if (l.includes('error') || l.includes('fatal') || l.includes('ERR')) return '#f87171'
  if (l.includes('warn')) return '#fbbf24'
  return 'rgba(255,255,255,0.55)'
}

// ── InfoRow ────────────────────────────────────────────────────────────────────

function InfoRow({ label, value }: { label: string; value: React.ReactNode }) {
  return (
    <div className="flex items-center justify-between">
      <span className="text-[13px] text-dim">{label}</span>
      <span className="text-[14px] text-sub font-medium">{value}</span>
    </div>
  )
}
