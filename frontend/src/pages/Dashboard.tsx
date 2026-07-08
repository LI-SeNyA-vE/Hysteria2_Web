import { useEffect, useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Activity, Users, HardDrive, Clock, Play, Square, Download, RefreshCw, ArrowUpCircle, ExternalLink } from 'lucide-react'
import { getStats } from '@/api/servers'
import { startHysteria, stopHysteria, installHysteria, reloadConfig } from '@/api/hysteria'
import { checkUpdate, applyUpdate } from '@/api/update'
import { StatCard } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Progress } from '@/components/ui/progress'
import { formatBytes } from '@/lib/utils'
import type { DashboardStats } from '@/types'

const PLACEHOLDER: DashboardStats = {
  totalUsers: 0, activeUsers: 0, totalTrafficGb: 0, uptime: '—',
  hysteria: { installed: false, running: false, version: 'v2.6.1', port: 443 },
}

const CARD_STYLE = {
  background: 'linear-gradient(180deg, #18181e 0%, #141418 100%)',
  boxShadow: '0 0 0 1px rgba(255,255,255,0.07), 0 1px 2px rgba(0,0,0,0.4), 0 4px 16px rgba(0,0,0,0.25)',
}

export function Dashboard() {
  const qc = useQueryClient()
  const { data = PLACEHOLDER, isError, isSuccess } = useQuery({
    queryKey: ['stats'],
    queryFn: getStats,
    placeholderData: (prev: typeof PLACEHOLDER | undefined) => prev ?? PLACEHOLDER,
    refetchInterval: 10_000,
  })

  // Липкий флаг: появился — держится до первого успешного ответа сервера
  const [offline, setOffline] = useState(false)
  useEffect(() => {
    if (isError) setOffline(true)
    else if (isSuccess) setOffline(false)
  }, [isError, isSuccess])

  const { data: updateInfo } = useQuery({
    queryKey: ['update-check'],
    queryFn: checkUpdate,
    refetchInterval: 60 * 60 * 1000, // раз в час
    retry: false,
  })

  const inv = () => qc.invalidateQueries({ queryKey: ['stats'] })
  const invDelayed = () => setTimeout(inv, 600)
  const installMut = useMutation({ mutationFn: installHysteria, onSuccess: inv })
  const startMut   = useMutation({ mutationFn: startHysteria,   onSuccess: invDelayed })
  const stopMut    = useMutation({ mutationFn: stopHysteria,    onSuccess: invDelayed })
  const reloadMut  = useMutation({ mutationFn: reloadConfig,    onSuccess: invDelayed })
  const updateMut  = useMutation({ mutationFn: applyUpdate })

  const hy = data.hysteria

  return (
    <div className="min-h-screen" style={{ padding: '40px 48px' }}>

      {/* Заголовок страницы */}
      <div style={{ marginBottom: 32 }}>
        <h1 className="text-[22px] font-semibold text-text" style={{ letterSpacing: '-0.02em' }}>
          Главная
        </h1>
        {offline && (
          <p className="text-[13px] text-dim" style={{ marginTop: 6 }}>Нет соединения с сервером</p>
        )}
      </div>

      {/* Баннер обновления */}
      {updateInfo?.updateAvailable && (
        <div
          className="rounded-2xl px-6 py-4 flex items-center justify-between"
          style={{
            background: 'linear-gradient(90deg, rgba(234,179,8,0.12) 0%, rgba(234,179,8,0.06) 100%)',
            boxShadow: '0 0 0 1px rgba(234,179,8,0.25)',
            marginBottom: 20,
          }}
        >
          <div className="flex items-center gap-3">
            <ArrowUpCircle className="w-4 h-4 text-yellow-400 flex-shrink-0" />
            <span className="text-[14px] text-sub">
              Доступно обновление{' '}
              <span className="text-text font-semibold">{updateInfo.latestVersion}</span>
              <span className="text-dim"> (текущая {updateInfo.currentVersion})</span>
            </span>
          </div>
          <div className="flex items-center gap-3">
            <a
              href={updateInfo.releaseUrl}
              target="_blank"
              rel="noreferrer"
              className="flex items-center gap-1.5 text-[13px] text-dim hover:text-sub transition-colors"
            >
              <ExternalLink className="w-3.5 h-3.5" />
              Что нового
            </a>
            {updateMut.isSuccess ? (
              <span className="text-[13px] text-green-400">Перезапуск...</span>
            ) : (
              <Button
                size="sm"
                loading={updateMut.isPending}
                onClick={() => updateMut.mutate()}
                style={{ background: 'rgba(234,179,8,0.2)', borderColor: 'rgba(234,179,8,0.3)', color: '#fde047' }}
              >
                <Download className="w-3 h-3" /> Обновить
              </Button>
            )}
          </div>
        </div>
      )}

      {/* Статистика */}
      <div className="grid grid-cols-4 gap-5" style={{ marginBottom: 28 }}>
        <StatCard
          label="Пользователей"
          value={data.totalUsers}
          sub={`${data.activeUsers} активных`}
          icon={<Users className="w-4 h-4" strokeWidth={1.5} />}
        />
        <StatCard
          label="Активных"
          value={data.activeUsers}
          sub={data.totalUsers > 0 ? `из ${data.totalUsers}` : '—'}
          icon={<Activity className="w-4 h-4" strokeWidth={1.5} />}
        />
        <StatCard
          label="Трафик"
          value={formatBytes(data.totalTrafficGb)}
          sub="суммарно"
          icon={<HardDrive className="w-4 h-4" strokeWidth={1.5} />}
        />
        <StatCard
          label="Аптайм"
          value={data.uptime}
          sub="сервиса"
          icon={<Clock className="w-4 h-4" strokeWidth={1.5} />}
        />
      </div>

      {/* Hysteria2 */}
      <div className="rounded-2xl" style={{ ...CARD_STYLE, marginBottom: 20 }}>
        <div className="px-7 py-5 flex items-center justify-between"
          style={{ borderBottom: '1px solid rgba(255,255,255,0.06)' }}>
          <div className="flex items-center gap-4">
            <span className="text-[15px] font-medium text-text">Hysteria2</span>
            <Badge dot variant={hy.running ? 'success' : hy.installed ? 'warning' : 'muted'}>
              {hy.running ? 'Работает' : hy.installed ? 'Остановлен' : 'Не установлен'}
            </Badge>
          </div>
          <div className="flex items-center gap-2">
            {!hy.installed ? (
              <Button size="sm" loading={installMut.isPending} onClick={() => installMut.mutate()}>
                <Download className="w-3 h-3" /> Установить
              </Button>
            ) : (
              <>
                {hy.running ? (
                  <Button size="sm" variant="outline" loading={stopMut.isPending} onClick={() => stopMut.mutate()}>
                    <Square className="w-3 h-3" /> Стоп
                  </Button>
                ) : (
                  <Button size="sm" loading={startMut.isPending} onClick={() => startMut.mutate()}>
                    <Play className="w-3 h-3" /> Запустить
                  </Button>
                )}
                <Button size="sm" variant="outline" loading={reloadMut.isPending} onClick={() => reloadMut.mutate()}>
                  <RefreshCw className="w-3 h-3" /> Перезагрузить
                </Button>
              </>
            )}
          </div>
        </div>

        <div className="px-7 py-6">
          <div className="flex items-center gap-8 text-[14px] text-dim">
            <span>Версия <span className="text-sub font-medium">{hy.version}</span></span>
            <span>Порт <span className="text-sub font-medium">:{hy.port}</span></span>
          </div>

          {data.totalUsers > 0 && (
            <div style={{ marginTop: 20 }}>
              <div className="flex justify-between text-[13px] text-dim" style={{ marginBottom: 8 }}>
                <span>Активность</span>
                <span>{data.activeUsers} / {data.totalUsers}</span>
              </div>
              <Progress value={data.totalUsers > 0 ? (data.activeUsers / data.totalUsers) * 100 : 0} />
            </div>
          )}
        </div>
      </div>

      {/* Подсказка для пустого состояния */}
      {data.totalUsers === 0 && (
        <div className="rounded-2xl" style={CARD_STYLE}>
          <div className="px-7 py-6">
            <p className="text-[12px] font-semibold text-dim uppercase tracking-[0.06em]" style={{ marginBottom: 18 }}>С чего начать</p>
            <ol className="space-y-4">
              {[
                ['Настройки', 'Установите Hysteria2 и настройте порт и obfs'],
                ['Пользователи', 'Создайте первого пользователя — пароль генерируется автоматически'],
                ['Подписки', 'Создайте подписку и передайте base64-ссылку клиенту'],
              ].map(([section, desc], i) => (
                <li key={i} className="flex gap-3">
                  <span className="text-[13px] text-dim font-medium tabular-nums" style={{ marginTop: 2 }}>{i + 1}.</span>
                  <span className="text-[14px] text-sub">
                    <span className="text-text font-medium">{section}</span>
                    {' — '}{desc}
                  </span>
                </li>
              ))}
            </ol>
          </div>
        </div>
      )}
    </div>
  )
}
