import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { RefreshCw, Copy, Check, Download, Play, Square, RotateCcw, Shield } from 'lucide-react'
import {
  getHysteriaStatus, getHysteriaConfig,
  installHysteria, startHysteria, stopHysteria, reloadConfig,
  saveHysteriaConfig, regenerateCert,
} from '@/api/hysteria'
import { Card, CardHeader, CardTitle, CardContent } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { copyToClipboard } from '@/lib/utils'
import type { HysteriaConfig } from '@/types'

export function Settings() {
  const qc = useQueryClient()

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

  return (
    <div className="min-h-screen" style={{ padding: '40px 48px' }}>
      <div style={{ marginBottom: 28 }}>
        <h1 className="text-[22px] font-semibold text-text" style={{ letterSpacing: '-0.02em' }}>Настройки</h1>
      </div>

      <div className="space-y-5" style={{ maxWidth: 580 }}>
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

        {/* Конфигурация */}
        {config && <ConfigSection config={config} onSaved={() => inv()} />}

        {/* Сертификат */}
        {config && (
          <Card>
            <CardHeader>
              <CardTitle>TLS Сертификат</CardTitle>
            </CardHeader>
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
      </div>
    </div>
  )
}

function ConfigSection({ config, onSaved }: { config: HysteriaConfig; onSaved: () => void }) {
  const [port, setPort] = useState(String(config.port))
  const [obfs, setObfs] = useState(config.obfsPassword)
  const [mask, setMask] = useState(config.masqueradeUrl)
  const [saved, setSaved] = useState(false)

  const saveMut = useMutation({
    mutationFn: () => saveHysteriaConfig({ port: parseInt(port), obfsPassword: obfs, masqueradeUrl: mask }),
    onSuccess: () => { onSaved(); setSaved(true); setTimeout(() => setSaved(false), 2000) },
  })

  return (
    <Card>
      <CardHeader>
        <CardTitle>Конфигурация</CardTitle>
      </CardHeader>
      <CardContent>
        <div className="space-y-4">
          <Input label="Порт" type="number" min="1" max="65535" value={port} onChange={e => setPort(e.target.value)} />
          <Input
            label="OBFS пароль (Salamander, 30 символов)"
            value={obfs} onChange={e => setObfs(e.target.value)}
            hint="Генерируется автоматически при установке"
          />
          <Input
            label="Маскарад URL"
            placeholder="https://news.ycombinator.com"
            value={mask} onChange={e => setMask(e.target.value)}
            hint="Сайт, под который маскируется трафик"
          />
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
