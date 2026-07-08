import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Plus, Trash2, QrCode, Copy, Check, Users as UsersIcon, Search } from 'lucide-react'
import { QRCodeSVG } from 'qrcode.react'
import { getUsers, createUser, deleteUser, toggleUser } from '@/api/users'
import { getHysteriaConfig } from '@/api/hysteria'
import { getServers } from '@/api/servers'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Input } from '@/components/ui/input'
import { Switch } from '@/components/ui/switch'
import { Progress } from '@/components/ui/progress'
import { Dialog } from '@/components/ui/dialog'
import { formatBytes, formatExpiry, copyToClipboard } from '@/lib/utils'
import type { CreateUserRequest, User } from '@/types'

const TABLE_BOX = {
  borderRadius: 16,
  boxShadow: '0 0 0 1px rgba(255,255,255,0.07), 0 1px 2px rgba(0,0,0,0.4), 0 4px 16px rgba(0,0,0,0.25)',
  background: 'linear-gradient(180deg, #18181e 0%, #141418 100%)',
}

export function Users() {
  const qc = useQueryClient()
  const [search, setSearch]         = useState('')
  const [createOpen, setCreateOpen] = useState(false)
  const [qrUser, setQrUser]         = useState<User | null>(null)
  const [copiedId, setCopiedId]     = useState<number | null>(null)
  const [uriCopied, setUriCopied]   = useState(false)

  const { data: users = [] }  = useQuery({ queryKey: ['users'],           queryFn: getUsers,           placeholderData: [] })
  const { data: config }      = useQuery({ queryKey: ['hysteria-config'],  queryFn: getHysteriaConfig  })
  const { data: servers = [] } = useQuery({ queryKey: ['servers'],         queryFn: getServers,         placeholderData: [] })

  const mainServer = servers.find(s => s.role === 'main' || s.role === 'main_node1')

  const buildUri = (user: User) => {
    const ip   = mainServer?.publicIp  ?? 'YOUR_IP'
    const port = config?.port          ?? 443
    const obfs = config?.obfsPassword  ?? ''
    const pin  = config?.certSha256    ?? ''
    const sni  = config?.sni           ?? ''

    const parts: string[] = []
    if (obfs) { parts.push('obfs=salamander'); parts.push(`obfs-password=${encodeURIComponent(obfs)}`) }
    if (pin)  parts.push(`pinSHA256=${pin}`)   // колонки не кодируем — hysteria2 ожидает AA:BB:CC
    if (sni)  parts.push(`sni=${encodeURIComponent(sni)}`)
    parts.push('insecure=1')

    return `hysteria2://${encodeURIComponent(user.name)}:${encodeURIComponent(user.password)}@${ip}:${port}?${parts.join('&')}#${encodeURIComponent(user.name)}`
  }

  const createMut = useMutation({
    mutationFn: createUser,
    onSuccess: () => { qc.invalidateQueries({ queryKey: ['users'] }); setCreateOpen(false) },
  })
  const deleteMut = useMutation({
    mutationFn: deleteUser,
    onSuccess: () => qc.invalidateQueries({ queryKey: ['users'] }),
  })
  const toggleMut = useMutation({
    mutationFn: ({ id, active }: { id: number; active: boolean }) => toggleUser(id, active),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['users'] }),
  })

  const filtered = users.filter(u => u.name.toLowerCase().includes(search.toLowerCase()))

  const handleCopy = async (user: User) => {
    await copyToClipboard(buildUri(user))
    setCopiedId(user.id)
    setTimeout(() => setCopiedId(null), 2000)
  }

  const handleUriCopy = async (user: User) => {
    await copyToClipboard(buildUri(user))
    setUriCopied(true)
    setTimeout(() => setUriCopied(false), 2000)
  }

  return (
    <div className="min-h-screen" style={{ padding: '40px 48px' }}>

      {/* Шапка */}
      <div className="flex items-center justify-between" style={{ marginBottom: 28 }}>
        <div>
          <h1 className="text-[22px] font-semibold text-text" style={{ letterSpacing: '-0.02em' }}>
            Пользователи
          </h1>
          <p className="text-[13px] text-dim" style={{ marginTop: 6 }}>{users.length} записей</p>
        </div>
        <Button size="md" onClick={() => setCreateOpen(true)}>
          <Plus className="w-4 h-4" />
          Создать
        </Button>
      </div>

      {/* Поиск */}
      <div className="relative" style={{ marginBottom: 20, maxWidth: 300 }}>
        <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-dim pointer-events-none" />
        <Input placeholder="Поиск…" value={search} onChange={e => setSearch(e.target.value)} className="pl-9" />
      </div>

      <div style={TABLE_BOX}>
        <table className="w-full border-collapse" style={{ borderRadius: 16 }}>
          <thead>
            <tr style={{ borderBottom: '1px solid rgba(255,255,255,0.07)' }}>
              <Th first>Пользователь</Th>
              <Th>Трафик</Th>
              <Th>Срок</Th>
              <Th>Статус</Th>
              <Th last><span /></Th>
            </tr>
          </thead>
          <tbody>
            {filtered.length === 0 ? (
              <tr>
                <td colSpan={5} style={{ borderRadius: '0 0 16px 16px' }}>
                  <div className="flex flex-col items-center py-20 gap-4">
                    <div
                      className="w-12 h-12 rounded-2xl flex items-center justify-center"
                      style={{ background: 'rgba(255,255,255,0.04)', boxShadow: '0 0 0 1px rgba(255,255,255,0.07)' }}
                    >
                      <UsersIcon className="w-6 h-6 text-dim" strokeWidth={1.5} />
                    </div>
                    <div className="text-center">
                      <p className="text-[15px] font-medium text-sub">
                        {search ? 'Ничего не найдено' : 'Пользователей нет'}
                      </p>
                      <p className="text-[13px] text-dim" style={{ marginTop: 6 }}>
                        {search ? 'Попробуйте другой запрос' : 'Нажмите «Создать» чтобы добавить'}
                      </p>
                    </div>
                  </div>
                </td>
              </tr>
            ) : (
              filtered.map((user, idx) => {
                const pct = user.trafficLimitGb > 0
                  ? (user.trafficUsedGb / user.trafficLimitGb) * 100 : 0
                const expiry = formatExpiry(user.expireAt)
                const isLast = idx === filtered.length - 1

                return (
                  <tr
                    key={user.id}
                    className="transition-colors duration-75"
                    style={{ borderBottom: isLast ? 'none' : '1px solid rgba(255,255,255,0.05)' }}
                    onMouseEnter={e => (e.currentTarget.style.background = 'rgba(255,255,255,0.025)')}
                    onMouseLeave={e => (e.currentTarget.style.background = '')}
                  >
                    <td
                      className="py-5 text-[14px] font-medium text-text"
                      style={{ paddingLeft: 28, paddingRight: 20, borderRadius: isLast ? '0 0 0 16px' : undefined }}
                    >
                      {user.name}
                    </td>
                    <td className="px-5 py-5" style={{ minWidth: 190 }}>
                      <div className="text-[13px] text-dim" style={{ marginBottom: 8 }}>
                        {formatBytes(user.trafficUsedGb)}
                        <span style={{ margin: '0 6px', opacity: 0.4 }}>/</span>
                        {user.trafficLimitGb === 0 ? '∞' : formatBytes(user.trafficLimitGb)}
                      </div>
                      <Progress value={pct} />
                    </td>
                    <td className="px-5 py-5">
                      {expiry.expired
                        ? <Badge variant="danger">Истёк</Badge>
                        : <span className="text-[14px] text-dim">{expiry.label}</span>
                      }
                    </td>
                    <td className="px-5 py-5">
                      <Switch
                        checked={user.isActive}
                        onCheckedChange={active => toggleMut.mutate({ id: user.id, active })}
                      />
                    </td>
                    <td
                      className="py-5"
                      style={{ paddingLeft: 8, paddingRight: 24, borderRadius: isLast ? '0 0 16px 0' : undefined }}
                    >
                      <div className="flex items-center gap-0.5 justify-end">
                        <Button variant="ghost" size="icon" onClick={() => handleCopy(user)}>
                          {copiedId === user.id
                            ? <Check className="w-4 h-4 text-success" />
                            : <Copy className="w-4 h-4" />}
                        </Button>
                        <Button variant="ghost" size="icon" onClick={() => { setQrUser(user); setUriCopied(false) }}>
                          <QrCode className="w-4 h-4" />
                        </Button>
                        <Button
                          variant="ghost" size="icon"
                          onClick={() => deleteMut.mutate(user.id)}
                          loading={deleteMut.isPending}
                          className="hover:!text-danger"
                        >
                          <Trash2 className="w-4 h-4" />
                        </Button>
                      </div>
                    </td>
                  </tr>
                )
              })
            )}
          </tbody>
        </table>
      </div>

      <CreateUserDialog
        open={createOpen} onClose={() => setCreateOpen(false)}
        onSubmit={d => createMut.mutate(d)}
        loading={createMut.isPending} error={createMut.error?.message}
      />

      {/* QR-диалог */}
      <Dialog open={!!qrUser} onClose={() => setQrUser(null)} title="Подключение" description={qrUser?.name ?? ''}>
        {qrUser && (() => {
          const uri = buildUri(qrUser)
          return (
            <div className="space-y-4">
              {/* QR */}
              <div className="flex justify-center">
                <div className="p-4 rounded-2xl" style={{ background: '#fff' }}>
                  <QRCodeSVG value={uri} size={220} />
                </div>
              </div>

              {/* URI текст */}
              <div
                className="rounded-xl p-3 font-mono text-[11px] text-dim break-all leading-relaxed select-all"
                style={{ background: 'rgba(255,255,255,0.03)', boxShadow: '0 0 0 1px rgba(255,255,255,0.07)' }}
              >
                {uri}
              </div>

              {/* Кнопка копирования */}
              <Button className="w-full" onClick={() => handleUriCopy(qrUser)}>
                {uriCopied
                  ? <><Check className="w-4 h-4" /> Скопировано</>
                  : <><Copy className="w-4 h-4" /> Скопировать URI</>}
              </Button>
            </div>
          )
        })()}
      </Dialog>
    </div>
  )
}

function Th({ children, first, last }: { children: React.ReactNode; first?: boolean; last?: boolean }) {
  return (
    <th
      className="text-left text-[12px] font-semibold text-dim uppercase tracking-[0.07em]"
      style={{
        paddingTop: 16, paddingBottom: 16,
        paddingLeft: first ? 28 : 20,
        paddingRight: last ? 24 : 20,
        borderRadius: first ? '16px 0 0 0' : last ? '0 16px 0 0' : undefined,
      }}
    >
      {children}
    </th>
  )
}

function CreateUserDialog({ open, onClose, onSubmit, loading, error }: {
  open: boolean; onClose: () => void
  onSubmit: (d: CreateUserRequest) => void
  loading: boolean; error?: string
}) {
  const [name, setName]     = useState('')
  const [traffic, setTraff] = useState('0')
  const [days, setDays]     = useState('')

  return (
    <Dialog open={open} onClose={onClose} title="Новый пользователь" description="Пароль генерируется автоматически">
      <div className="space-y-4">
        <Input label="Имя" placeholder="alice" value={name} onChange={e => setName(e.target.value)} />
        <Input
          label="Лимит трафика (ГБ, 0 = безлимит)"
          type="number" min="0" placeholder="0"
          value={traffic} onChange={e => setTraff(e.target.value)}
        />
        <Input
          label="Срок действия (дней, пусто = бессрочно)"
          type="number" min="1" placeholder="30"
          value={days} onChange={e => setDays(e.target.value)}
        />
        {error && <p className="text-[12px] text-danger">{error}</p>}
        <div className="flex gap-2 pt-1">
          <Button variant="outline" className="flex-1" onClick={onClose}>Отмена</Button>
          <Button className="flex-1" loading={loading} disabled={!name.trim()}
            onClick={() => onSubmit({
              name: name.trim(),
              trafficLimitGb: parseFloat(traffic) || 0,
              expireDays: days ? parseInt(days) : null,
              serverId: 1,
            })}>
            Создать
          </Button>
        </div>
      </div>
    </Dialog>
  )
}
