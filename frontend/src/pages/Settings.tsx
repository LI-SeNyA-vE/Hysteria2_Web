import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { RefreshCw, Copy, Check, Download, Play, Square, RotateCcw, Shield } from 'lucide-react'
import {
  getHysteriaStatus, getHysteriaConfig,
  installHysteria, startHysteria, stopHysteria, reloadConfig,
  saveHysteriaConfig, regenerateCert,
} from '@/api/hysteria'
import { getServers, getNodeConfig, saveNodeConfig } from '@/api/servers'
import { Card, CardHeader, CardTitle, CardContent } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { copyToClipboard } from '@/lib/utils'
import type { HysteriaConfig, Server } from '@/types'

const ROLE_RU: Record<string, string> = {
  main: 'Главная', main_node1: 'Главная', node1: 'Узел 1', node2: 'Узел 2',
}

export function Settings() {
  const qc = useQueryClient()
  const [selectedId, setSelectedId] = useState<number | 'main'>('main')

  const { data: servers = [] } = useQuery({ queryKey: ['servers'], queryFn: getServers, placeholderData: [] })

  const { data: status } = useQuery({
    queryKey: ['hysteria-status'],
    queryFn: getHysteriaStatus,
    refetchInterval: 5_000,
  })
  const { data: config } = useQuery({
    queryKey: ['hysteria-config'],
    queryFn: getHysteriaConfig,
  })

  const inv = () => {
    qc.invalidateQueries({ queryKey: ['hysteria-status'] })
    qc.invalidateQueries({ queryKey: ['hysteria-config'] })
  }

  const installMut = useMutation({ mutationFn: installHysteria,  onSuccess: inv })
  const startMut   = useMutation({ mutationFn: startHysteria,    onSuccess: inv })
  const stopMut    = useMutation({ mutationFn: stopHysteria,     onSuccess: inv })
  const reloadMut  = useMutation({ mutationFn: reloadConfig,     onSuccess: inv })
  const certMut    = useMutation({
    mutationFn: regenerateCert,
    onSuccess: () => qc.invalidateQueries({ queryKey: ['hysteria-config'] }),
  })

  const isMain = selectedId === 'main'
  const remoteServer = isMain ? null : servers.find(s => s.id === selectedId) ?? null

  // Серверы-ноды (не main/main_node1)
  const nodeServers = servers.filter(s => s.role === 'node1' || s.role === 'node2')

  return (
    <div className="min-h-screen" style={{ padding: '40px 48px' }}>
      <div style={{ marginBottom: 28 }}>
        <h1 className="text-[22px] font-semibold text-text" style={{ letterSpacing: '-0.02em' }}>Настройки</h1>
      </div>

      {/* Селектор сервера */}
      {nodeServers.length > 0 && (
        <div className="flex gap-2" style={{ marginBottom: 24 }}>
          <ServerTab
            label="Главная панель"
            active={isMain}
            onClick={() => setSelectedId('main')}
          />
          {nodeServers.map(s => (
            <ServerTab
              key={s.id}
              label={`${ROLE_RU[s.role]} · ${s.publicIp}`}
              active={selectedId === s.id}
              onClick={() => setSelectedId(s.id)}
            />
          ))}
        </div>
      )}

      <div className="space-y-5" style={{ maxWidth: 580 }}>
        {isMain ? (
          <>
            {/* Сервис Hysteria2 */}
            <Card>
              <CardHeader>
                <CardTitle>Сервис Hysteria2</CardTitle>
              </CardHeader>
              <CardContent>
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-3">
                    <Badge dot variant={status?.running ? 'success' : status?.installed ? 'warning' : 'muted'}>
                      {status?.running ? 'Работает' : status?.installed ? 'Остановлен' : 'Не установлен'}
                    </Badge>
                    {status?.version && <span className="text-[13px] text-dim">{status.version}</span>}
                  </div>
                  <div className="flex gap-1.5">
                    {!status?.installed ? (
                      <Button size="sm" loading={installMut.isPending} onClick={() => installMut.mutate()}>
                        <Download className="w-3.5 h-3.5" /> Установить
                      </Button>
                    ) : (
                      <>
                        {status.running ? (
                          <Button size="sm" variant="outline" loading={stopMut.isPending} onClick={() => stopMut.mutate()}>
                            <Square className="w-3.5 h-3.5" /> Стоп
                          </Button>
                        ) : (
                          <Button size="sm" loading={startMut.isPending} onClick={() => startMut.mutate()}>
                            <Play className="w-3.5 h-3.5" /> Запустить
                          </Button>
                        )}
                        <Button size="sm" variant="outline" loading={reloadMut.isPending} onClick={() => reloadMut.mutate()}>
                          <RefreshCw className="w-3.5 h-3.5" /> Перезагрузить
                        </Button>
                      </>
                    )}
                  </div>
                </div>
              </CardContent>
            </Card>

            {config && <ConfigSection config={config} onSaved={() => inv()} />}

            {config && (
              <Card>
                <CardHeader><CardTitle>TLS Сертификат</CardTitle></CardHeader>
                <CardContent>
                  <div
                    className="flex items-start gap-3 mb-4 p-3.5 rounded-xl"
                    style={{ background: 'rgba(124,111,247,0.07)', boxShadow: '0 0 0 1px rgba(124,111,247,0.15)' }}
                  >
                    <Shield className="w-4 h-4 text-accent mt-0.5 shrink-0" />
                    <p className="text-[13px] text-sub leading-relaxed">
                      Self-signed сертификат. Клиенты используют <code className="text-text font-mono">pinSHA256</code> для проверки — домен не нужен.
                    </p>
                  </div>
                  <div className="flex gap-2 mb-3">
                    <Input value={config.certSha256 || 'Сертификат не сгенерирован'} readOnly className="font-mono text-[12px]" />
                    <CopyBtn text={config.certSha256} />
                  </div>
                  <Button size="sm" variant="outline" loading={certMut.isPending} onClick={() => certMut.mutate()}>
                    <RotateCcw className="w-3.5 h-3.5" /> Перегенерировать
                  </Button>
                </CardContent>
              </Card>
            )}
          </>
        ) : (
          remoteServer && <NodeConfigSection server={remoteServer} />
        )}
      </div>
    </div>
  )
}

function ServerTab({ label, active, onClick }: { label: string; active: boolean; onClick: () => void }) {
  return (
    <button
      onClick={onClick}
      className="px-4 py-2 rounded-xl text-[13px] font-medium transition-colors"
      style={active ? {
        background: 'rgba(255,255,255,0.12)',
        color: 'var(--color-text, #f1f5f9)',
        boxShadow: '0 0 0 1px rgba(255,255,255,0.15)',
      } : {
        background: 'rgba(255,255,255,0.04)',
        color: 'var(--color-dim, #64748b)',
        boxShadow: '0 0 0 1px rgba(255,255,255,0.06)',
      }}
    >
      {label}
    </button>
  )
}

function NodeConfigSection({ server }: { server: Server }) {
  const qc = useQueryClient()
  const { data } = useQuery({
    queryKey: ['node-config', server.id],
    queryFn: () => getNodeConfig(server.id),
  })

  const [bwUp, setBwUp]     = useState('')
  const [bwDown, setBwDown] = useState('')
  const [masq, setMasq]     = useState('')
  const [saved, setSaved]   = useState(false)
  const [loaded, setLoaded] = useState(false)

  // Инициализация полей после загрузки
  if (data && !loaded) {
    setBwUp(data.bandwidthUp)
    setBwDown(data.bandwidthDown)
    setMasq(data.masqueradeUrl)
    setLoaded(true)
  }

  const saveMut = useMutation({
    mutationFn: () => saveNodeConfig(server.id, { bandwidthUp: bwUp, bandwidthDown: bwDown, masqueradeUrl: masq }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['node-config', server.id] })
      setSaved(true)
      setTimeout(() => setSaved(false), 2000)
    },
  })

  return (
    <Card>
      <CardHeader>
        <CardTitle>Конфигурация ноды</CardTitle>
      </CardHeader>
      <CardContent>
        <p className="text-[12px] text-dim" style={{ marginBottom: 16 }}>
          Изменения применяются автоматически при следующем heartbeat (≤10 сек).
        </p>
        <div className="space-y-4">
          <Input
            label="Маскарад URL"
            placeholder="https://news.ycombinator.com (глобальное если пусто)"
            value={masq}
            onChange={e => setMasq(e.target.value)}
            hint="Пусто = использовать глобальную настройку"
          />
          <div className="grid grid-cols-2 gap-3">
            <Input
              label="Лимит Upload"
              placeholder="напр. 100 mbps"
              value={bwUp}
              onChange={e => setBwUp(e.target.value)}
              hint="Пусто = глобальный лимит"
            />
            <Input
              label="Лимит Download"
              placeholder="напр. 1 gbps"
              value={bwDown}
              onChange={e => setBwDown(e.target.value)}
              hint="Пусто = глобальный лимит"
            />
          </div>
          <Button size="sm" loading={saveMut.isPending} onClick={() => saveMut.mutate()}>
            {saved ? <><Check className="w-3.5 h-3.5" /> Сохранено</> : 'Сохранить'}
          </Button>
          {saveMut.isError && (
            <p className="text-[12px] text-danger">{String(saveMut.error)}</p>
          )}
        </div>
      </CardContent>
    </Card>
  )
}

function ConfigSection({ config, onSaved }: { config: HysteriaConfig; onSaved: () => void }) {
  const [port, setPort]     = useState(String(config.port))
  const [obfs, setObfs]     = useState(config.obfsPassword)
  const [mask, setMask]     = useState(config.masqueradeUrl)
  const [bwUp, setBwUp]     = useState(config.bandwidthUp)
  const [bwDown, setBwDown] = useState(config.bandwidthDown)
  const [sni, setSni]       = useState(config.sni)
  const [saved, setSaved]   = useState(false)

  const saveMut = useMutation({
    mutationFn: () => saveHysteriaConfig({
      port: parseInt(port) || config.port,
      obfsPassword: obfs,
      masqueradeUrl: mask,
      bandwidthUp: bwUp,
      bandwidthDown: bwDown,
      sni,
    }),
    onSuccess: () => { onSaved(); setSaved(true); setTimeout(() => setSaved(false), 2000) },
  })

  return (
    <Card>
      <CardHeader>
        <CardTitle>Конфигурация</CardTitle>
      </CardHeader>
      <CardContent>
        <div className="space-y-4">
          <Input
            label="UDP порт"
            type="number"
            min="1"
            max="65535"
            value={port}
            onChange={e => setPort(e.target.value)}
            hint="Требует перезапуска hysteria2"
          />
          <Input
            label="OBFS пароль (Salamander)"
            value={obfs}
            onChange={e => setObfs(e.target.value)}
            hint="Одинаковый на сервере и клиенте"
          />
          <Input
            label="Маскарад URL"
            placeholder="https://news.ycombinator.com"
            value={mask}
            onChange={e => setMask(e.target.value)}
            hint="Сайт, под который маскируется трафик"
          />
          <Input
            label="SNI (подставляется в ссылку клиента)"
            placeholder="напр. yandex.ru"
            value={sni}
            onChange={e => setSni(e.target.value)}
            hint="Пусто = без SNI в URI"
          />
          <div className="grid grid-cols-2 gap-3">
            <Input
              label="Лимит Upload (сервер→клиент)"
              placeholder="напр. 100 mbps"
              value={bwUp}
              onChange={e => setBwUp(e.target.value)}
              hint="Пусто = без лимита"
            />
            <Input
              label="Лимит Download (клиент→сервер)"
              placeholder="напр. 1 gbps"
              value={bwDown}
              onChange={e => setBwDown(e.target.value)}
              hint="Пусто = без лимита"
            />
          </div>
          <Button size="sm" loading={saveMut.isPending} onClick={() => saveMut.mutate()}>
            {saved ? <><Check className="w-3.5 h-3.5" /> Сохранено</> : 'Сохранить'}
          </Button>
        </div>
      </CardContent>
    </Card>
  )
}

function CopyBtn({ text }: { text: string }) {
  const [copied, setCopied] = useState(false)
  return (
    <Button variant="outline" size="icon" className="shrink-0"
      onClick={async () => { await copyToClipboard(text); setCopied(true); setTimeout(() => setCopied(false), 2000) }}>
      {copied ? <Check className="w-4 h-4 text-success" /> : <Copy className="w-4 h-4" />}
    </Button>
  )
}
