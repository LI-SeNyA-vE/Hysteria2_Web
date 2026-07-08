import { useQuery } from '@tanstack/react-query'
import { ArrowRight, GitBranch, CheckCircle2, AlertCircle } from 'lucide-react'
import { getServers } from '@/api/servers'
import { Card, CardHeader, CardTitle, CardContent } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'

const NODE_ACTIVE = {
  boxShadow: '0 0 0 1px rgba(124,111,247,0.3)',
  background: 'rgba(124,111,247,0.06)',
}
const NODE_INACTIVE = {
  boxShadow: '0 0 0 1px rgba(255,255,255,0.08)',
  background: 'rgba(255,255,255,0.03)',
}

export function Cascade() {
  const { data: servers = [] } = useQuery({
    queryKey: ['servers'],
    queryFn: getServers,
    placeholderData: [],
  })

  const node1 = servers.find(s => s.role === 'node1' || s.role === 'main_node1')
  const node2 = servers.find(s => s.role === 'node2')
  const configured = !!node1 && !!node2

  return (
    <div className="min-h-screen" style={{ padding: '40px 48px' }}>
      <div style={{ marginBottom: 28 }}>
        <h1 className="text-[22px] font-semibold text-text" style={{ letterSpacing: '-0.02em' }}>Каскад</h1>
        <p className="text-[13px] text-dim" style={{ marginTop: 6 }}>Перенаправление трафика через цепочку узлов</p>
      </div>

      <div style={{ maxWidth: 640 }}>
        {/* Схема */}
        <Card style={{ marginBottom: 20 }}>
          <CardContent className="pt-7">
            <div className="flex items-center justify-center gap-4 py-6">
              <NodeBox label="Узел 1" ip={node1?.publicIp} active={!!node1} />

              <div className="flex flex-col items-center gap-1 text-dim">
                <ArrowRight className="w-4 h-4" />
                <span className="text-[11px]">hy2-client</span>
                <span className="text-[11px]">SOCKS5</span>
              </div>

              <NodeBox label="Узел 2" ip={node2?.publicIp} active={!!node2} />

              <div className="flex flex-col items-center gap-1 text-dim">
                <ArrowRight className="w-4 h-4" />
                <span className="text-[12px]">интернет</span>
              </div>

              <div className="px-4 py-3 rounded-xl text-center" style={NODE_INACTIVE}>
                <div className="text-[12px] text-dim" style={{ marginBottom: 4 }}>Интернет</div>
                <div className="text-xl">🌐</div>
              </div>
            </div>

            <div className="flex justify-center">
              {configured ? (
                <Badge variant="success" dot>
                  <CheckCircle2 className="w-3 h-3 mr-1" />
                  Каскад настроен
                </Badge>
              ) : (
                <Badge variant="muted" dot>
                  <AlertCircle className="w-3 h-3 mr-1" />
                  Не настроен — добавьте Узел 1 и Узел 2
                </Badge>
              )}
            </div>
          </CardContent>
        </Card>

        {/* Как работает */}
        <Card>
          <CardHeader>
            <CardTitle>Как работает каскад</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-3.5">
              {[
                'Узел 1 запускает hy2-server и принимает подключения от клиентов',
                'Узел 1 также запускает hy2-client → подключается к Узлу 2 → создаёт локальный SOCKS5 на 127.0.0.1:1080',
                'hy2-server на Узле 1 направляет весь исходящий трафик через SOCKS5 (то есть через Узел 2)',
                'Узел 2 имеет одного системного пользователя, выходит в интернет и возвращает трафик обратно',
              ].map((text, i) => (
                <div key={i} className="flex gap-3">
                  <span
                    className="w-6 h-6 rounded-lg flex items-center justify-center text-[12px] font-semibold text-dim shrink-0 mt-0.5"
                    style={{ background: 'rgba(255,255,255,0.05)', boxShadow: '0 0 0 1px rgba(255,255,255,0.08)' }}
                  >
                    {i + 1}
                  </span>
                  <span className="text-[14px] text-sub leading-relaxed">{text}</span>
                </div>
              ))}
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  )
}

function NodeBox({ label, ip, active }: { label: string; ip?: string; active: boolean }) {
  return (
    <div className="px-5 py-4 rounded-xl text-center transition-all duration-200" style={{ minWidth: 130, ...(active ? NODE_ACTIVE : NODE_INACTIVE) }}>
      <GitBranch className={`w-4 h-4 mx-auto mb-2.5 ${active ? 'text-accent' : 'text-dim'}`} strokeWidth={1.75} />
      <div className={`text-[13px] font-semibold ${active ? 'text-text' : 'text-dim'}`}>{label}</div>
      <div className="text-[12px] text-dim" style={{ marginTop: 4 }}>{ip ?? 'не настроен'}</div>
    </div>
  )
}
