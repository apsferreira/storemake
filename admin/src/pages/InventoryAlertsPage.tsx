import { useNavigate } from 'react-router-dom'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { inventoryApi } from '@/api/endpoints'
import { Card } from '@/components/ui/Card'
import { Badge } from '@/components/ui/Badge'
import { Button } from '@/components/ui/Button'
import { EmptyState } from '@/components/ui/EmptyState'
import { formatDate } from '@/lib/utils'
import { ArrowLeft, CheckCircle2, AlertTriangle, AlertOctagon } from 'lucide-react'

export function InventoryAlertsPage() {
  const navigate = useNavigate()
  const qc = useQueryClient()

  const { data, isLoading } = useQuery({
    queryKey: ['inventory-alerts'],
    queryFn: () => inventoryApi.alerts(),
  })

  const ackMut = useMutation({
    mutationFn: (id: string) => inventoryApi.acknowledgeAlert(id),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['inventory-alerts'] })
      qc.invalidateQueries({ queryKey: ['inventory'] })
    },
  })

  const alerts = data?.alerts || []

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center gap-3">
        <button
          onClick={() => navigate('/estoque')}
          className="flex items-center gap-1 text-sm text-gray-500 hover:text-gray-700"
        >
          <ArrowLeft size={16} />
          Estoque
        </button>
        <span className="text-gray-300">/</span>
        <h1 className="text-xl font-bold text-gray-900">Alertas de reposição</h1>
        {alerts.length > 0 && (
          <Badge variant="warning">{alerts.length} pendente{alerts.length > 1 ? 's' : ''}</Badge>
        )}
      </div>

      {/* List */}
      {isLoading ? (
        <Card><div className="p-8 text-center text-gray-400 text-sm">Carregando...</div></Card>
      ) : alerts.length === 0 ? (
        <EmptyState
          icon={<CheckCircle2 size={48} />}
          title="Nenhum alerta pendente"
          description="Todos os itens estão com estoque adequado."
        />
      ) : (
        <div className="space-y-3">
          {alerts.map(alert => (
            <Card key={alert.id}>
              <div className="p-4 flex items-start justify-between gap-4">
                <div className="flex items-start gap-3">
                  <div className={`mt-0.5 shrink-0 ${alert.alert_type === 'out_of_stock' ? 'text-red-500' : 'text-amber-500'}`}>
                    {alert.alert_type === 'out_of_stock' ? (
                      <AlertOctagon size={20} />
                    ) : (
                      <AlertTriangle size={20} />
                    )}
                  </div>
                  <div>
                    <div className="flex items-center gap-2 mb-1">
                      <Badge variant={alert.alert_type === 'out_of_stock' ? 'danger' : 'warning'}>
                        {alert.alert_type === 'out_of_stock' ? 'Sem estoque' : 'Estoque baixo'}
                      </Badge>
                      <span className="text-xs text-gray-400">{formatDate(alert.created_at)}</span>
                    </div>
                    <p className="text-sm text-gray-700">
                      SKU <span className="font-mono font-medium">{alert.master_id.slice(0, 8)}…</span>
                    </p>
                    <p className="text-xs text-gray-500 mt-0.5">
                      Qtd. atual: <strong>{alert.quantity_current}</strong> — Ponto de reposição: <strong>{alert.quantity_reorder}</strong>
                    </p>
                  </div>
                </div>
                <div className="flex gap-2 shrink-0">
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => navigate(`/estoque/${alert.master_id}`)}
                  >
                    Ver SKU
                  </Button>
                  <Button
                    size="sm"
                    onClick={() => ackMut.mutate(alert.id)}
                    disabled={ackMut.isPending}
                  >
                    <CheckCircle2 size={14} />
                    Confirmar
                  </Button>
                </div>
              </div>
            </Card>
          ))}
        </div>
      )}
    </div>
  )
}
