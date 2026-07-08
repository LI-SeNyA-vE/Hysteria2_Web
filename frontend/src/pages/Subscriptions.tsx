import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Plus, Copy, Check, Trash2, ExternalLink, Link2 } from 'lucide-react'
import { getSubscriptions, createSubscription, deleteSubscription } from '@/api/subscriptions'
import { getUsers } from '@/api/users'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Select } from '@/components/ui/select'
import { Dialog } from '@/components/ui/dialog'
import { copyToClipboard } from '@/lib/utils'
import type { CreateSubscriptionRequest } from '@/types'

const TABLE_BOX = {
  borderRadius: 16,
  boxShadow: '0 0 0 1px rgba(255,255,255,0.07), 0 1px 2px rgba(0,0,0,0.4), 0 4px 16px rgba(0,0,0,0.25)',
  background: 'linear-gradient(180deg, #18181e 0%, #141418 100%)',
}

export function Subscriptions() {
  const qc = useQueryClient()
  const [createOpen, setCreateOpen] = useState(false)
  const [copiedId, setCopiedId]     = useState<number | null>(null)

  const { data: subs = [] }  = useQuery({ queryKey: ['subscriptions'], queryFn: getSubscriptions, placeholderData: [] })
  const { data: users = [] } = useQuery({ queryKey: ['users'],         queryFn: getUsers,          placeholderData: [] })

  const createMut = useMutation({
    mutationFn: createSubscription,
    onSuccess: () => { qc.invalidateQueries({ queryKey: ['subscriptions'] }); setCreateOpen(false) },
  })
  const deleteMut = useMutation({
    mutationFn: deleteSubscription,
    onSuccess: () => qc.invalidateQueries({ queryKey: ['subscriptions'] }),
  })

  const subUrl = (token: string) => `${window.location.origin}/sub/${token}`

  const copy = async (id: number, token: string) => {
    await copyToClipboard(subUrl(token))
    setCopiedId(id)
    setTimeout(() => setCopiedId(null), 2000)
  }

  return (
    <div className="min-h-screen" style={{ padding: '40px 48px' }}>

      <div className="flex items-center justify-between" style={{ marginBottom: 28 }}>
        <div>
          <h1 className="text-[22px] font-semibold text-text" style={{ letterSpacing: '-0.02em' }}>Подписки</h1>
          <p className="text-[13px] text-dim" style={{ marginTop: 6 }}>Base64-ссылки для клиентов</p>
        </div>
        <Button onClick={() => setCreateOpen(true)}>
          <Plus className="w-4 h-4" /> Создать
        </Button>
      </div>

      <div style={TABLE_BOX}>
        <table className="w-full border-collapse">
          <thead>
            <tr style={{ borderBottom: '1px solid rgba(255,255,255,0.07)' }}>
              <Th first>Название</Th>
              <Th>Пользователь</Th>
              <Th>Токен</Th>
              <Th>Использовалась</Th>
              <Th last><span /></Th>
            </tr>
          </thead>
          <tbody>
            {subs.length === 0 ? (
              <tr>
                <td colSpan={5} style={{ borderRadius: '0 0 16px 16px' }}>
                  <div className="flex flex-col items-center py-20 gap-4">
                    <div
                      className="w-12 h-12 rounded-2xl flex items-center justify-center"
                      style={{ background: 'rgba(255,255,255,0.04)', boxShadow: '0 0 0 1px rgba(255,255,255,0.07)' }}
                    >
                      <Link2 className="w-6 h-6 text-dim" strokeWidth={1.5} />
                    </div>
                    <div className="text-center">
                      <p className="text-[15px] font-medium text-sub">Нет подписок</p>
                      <p className="text-[13px] text-dim" style={{ marginTop: 6 }}>Создайте ссылку для клиента</p>
                    </div>
                  </div>
                </td>
              </tr>
            ) : (
              subs.map((sub, idx) => {
                const isLast = idx === subs.length - 1
                return (
                  <tr
                    key={sub.id}
                    className="transition-colors duration-75"
                    style={{ borderBottom: isLast ? 'none' : '1px solid rgba(255,255,255,0.05)' }}
                    onMouseEnter={e => (e.currentTarget.style.background = 'rgba(255,255,255,0.025)')}
                    onMouseLeave={e => (e.currentTarget.style.background = '')}
                  >
                    <td
                      className="py-5 text-[14px] font-medium text-text"
                      style={{ paddingLeft: 28, paddingRight: 20, borderRadius: isLast ? '0 0 0 16px' : undefined }}
                    >
                      {sub.name}
                    </td>
                    <td className="px-5 py-5 text-[14px] text-sub">{sub.userName}</td>
                    <td className="px-5 py-5">
                      <code className="text-[13px] text-dim font-mono">/sub/{sub.token.slice(0, 10)}…</code>
                    </td>
                    <td className="px-5 py-5 text-[14px] text-dim">
                      {sub.lastAccessedAt
                        ? new Date(sub.lastAccessedAt).toLocaleDateString('ru-RU')
                        : 'Никогда'}
                    </td>
                    <td
                      className="py-5"
                      style={{ paddingLeft: 8, paddingRight: 24, borderRadius: isLast ? '0 0 16px 0' : undefined }}
                    >
                      <div className="flex items-center gap-0.5 justify-end">
                        <Button variant="ghost" size="icon" onClick={() => copy(sub.id, sub.token)}>
                          {copiedId === sub.id
                            ? <Check className="w-4 h-4 text-success" />
                            : <Copy className="w-4 h-4" />}
                        </Button>
                        <Button variant="ghost" size="icon" onClick={() => window.open(subUrl(sub.token), '_blank')}>
                          <ExternalLink className="w-4 h-4" />
                        </Button>
                        <Button
                          variant="ghost" size="icon"
                          loading={deleteMut.isPending}
                          onClick={() => deleteMut.mutate(sub.id)}
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

      <CreateDialog
        open={createOpen} onClose={() => setCreateOpen(false)}
        onSubmit={d => createMut.mutate(d)} loading={createMut.isPending}
        error={createMut.error?.message}
        users={users.map(u => ({ id: u.id, name: u.name }))}
      />
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

function CreateDialog({ open, onClose, onSubmit, loading, error, users }: {
  open: boolean; onClose: () => void
  onSubmit: (d: CreateSubscriptionRequest) => void
  loading: boolean; error?: string
  users: { id: number; name: string }[]
}) {
  const [name, setName]     = useState('')
  const [userId, setUserId] = useState<number | ''>('')

  return (
    <Dialog open={open} onClose={onClose} title="Новая подписка" description="Генерирует base64 URL для клиента">
      <div className="space-y-4">
        <Input label="Название" placeholder="Мой телефон" value={name} onChange={e => setName(e.target.value)} />
        <Select
          label="Пользователь"
          value={userId === '' ? '' : String(userId)}
          onChange={e => setUserId(Number(e.target.value))}
        >
          <option value="">Выберите пользователя…</option>
          {users.map(u => <option key={u.id} value={u.id}>{u.name}</option>)}
        </Select>
        {error && <p className="text-[12px] text-danger">{error}</p>}
        <div className="flex gap-2 pt-1">
          <Button variant="outline" className="flex-1" onClick={onClose}>Отмена</Button>
          <Button className="flex-1" loading={loading} disabled={!name.trim() || userId === ''}
            onClick={() => onSubmit({ name: name.trim(), userId: userId as number })}>
            Создать
          </Button>
        </div>
      </div>
    </Dialog>
  )
}
