import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { ordersApi } from '@/api/endpoints'
import { Card } from '@/components/ui/Card'
import { Badge } from '@/components/ui/Badge'
import { Button } from '@/components/ui/Button'
import { Modal } from '@/components/ui/Modal'
import { Input } from '@/components/ui/Input'
import { EmptyState } from '@/components/ui/EmptyState'
import { formatCents, formatDate } from '@/lib/utils'
import { ClipboardList, Search, ChevronRight } from 'lucide-react'
import type { Order, OrderStatus } from '@/types'

const COLUMNS: { status: OrderStatus; label: string; color: string }[] = [
  { status: 'pendente', label: 'Pendente', color: 'bg-amber-500' },
  { status: 'pago', label: 'Pago', color: 'bg-blue-500' },
  { status: 'preparando', label: 'Preparando', color: 'bg-indigo-500' },
  { status: 'enviado', label: 'Enviado', color: 'bg-purple-500' },
  { status: 'entregue', label: 'Entregue', color: 'bg-green-500' },
]

const statusVariant: Record<string, 'warning' | 'info' | 'success' | 'danger' | 'default'> = {
  pendente: 'warning',
  pago: 'info',
  preparando: 'info',
  enviado: 'info',
  entregue: 'success',
  cancelado: 'danger',
}

export function OrdersPage() {
  const qc = useQueryClient()
  const [search, setSearch] = useState('')
  const [selectedOrder, setSelectedOrder] = useState<Order | null>(null)
  const [trackingCode, setTrackingCode] = useState('')

  const { data, isLoading } = useQuery({
    queryKey: ['orders', search],
    queryFn: () => ordersApi.list({ per_page: 200, search: search || undefined }),
  })

  const updateMut = useMutation({
    mutationFn: ({ id, status, tracking }: { id: string; status: OrderStatus; tracking?: string }) =>
      ordersApi.updateStatus(id, { status, tracking_code: tracking }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['orders'] })
      setSelectedOrder(null)
    },
  })

  const orders = data?.data || []

  function moveNext(order: Order) {
    const flow: OrderStatus[] = ['pendente', 'pago', 'preparando', 'enviado', 'entregue']
    const idx = flow.indexOf(order.status)
    if (idx < 0 || idx >= flow.length - 1) return
    const nextStatus = flow[idx + 1]
    if (nextStatus === 'enviado') {
      setSelectedOrder(order)
      return
    }
    updateMut.mutate({ id: order.id, status: nextStatus })
  }

  function handleSendWithTracking() {
    if (!selectedOrder) return
    updateMut.mutate({
      id: selectedOrder.id,
      status: 'enviado',
      tracking: trackingCode || undefined,
    })
    setTrackingCode('')
  }

  return (
    <div className="space-y-4">
      {/* Search */}
      <div className="relative w-full sm:w-72">
        <Search size={16} className="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400" />
        <input
          type="text"
          placeholder="Buscar pedidos..."
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          className="w-full pl-9 pr-3 py-2 text-sm rounded-lg border border-border focus:outline-none focus:ring-2 focus:ring-primary/20"
        />
      </div>

      {isLoading ? (
        <div className="flex items-center justify-center py-12">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary" />
        </div>
      ) : orders.length === 0 ? (
        <EmptyState
          icon={<ClipboardList size={48} />}
          title="Nenhum pedido"
          description="Os pedidos aparecerão aqui conforme forem realizados."
        />
      ) : (
        /* Kanban Board */
        <div className="flex gap-4 overflow-x-auto pb-4">
          {COLUMNS.map((col) => {
            const colOrders = orders.filter((o: Order) => o.status === col.status)
            return (
              <div key={col.status} className="flex-shrink-0 w-72">
                <div className="flex items-center gap-2 mb-3">
                  <div className={`w-3 h-3 rounded-full ${col.color}`} />
                  <h3 className="text-sm font-medium text-gray-900">{col.label}</h3>
                  <Badge variant="outline">{colOrders.length}</Badge>
                </div>
                <div className="space-y-3">
                  {colOrders.map((order: Order) => (
                    <Card key={order.id} className="p-4">
                      <div className="space-y-2">
                        <div className="flex items-start justify-between">
                          <div>
                            <p className="text-sm font-medium text-gray-900">
                              {order.customer_name}
                            </p>
                            <p className="text-xs text-muted">{order.customer_email}</p>
                          </div>
                          <Badge variant={statusVariant[order.status]}>
                            {order.status}
                          </Badge>
                        </div>
                        <div className="flex items-center justify-between">
                          <p className="text-sm font-bold text-gray-900">
                            {formatCents(order.total_cents)}
                          </p>
                          <p className="text-xs text-muted">{formatDate(order.created_at)}</p>
                        </div>
                        {order.items && order.items.length > 0 && (
                          <p className="text-xs text-muted">
                            {order.items.length} item(s)
                          </p>
                        )}
                        {order.tracking_code && (
                          <p className="text-xs text-info">Rastreio: {order.tracking_code}</p>
                        )}
                        {col.status !== 'entregue' && (
                          <Button
                            size="sm"
                            variant="ghost"
                            className="w-full mt-1"
                            onClick={() => moveNext(order)}
                            loading={updateMut.isPending}
                          >
                            Avancar <ChevronRight size={14} />
                          </Button>
                        )}
                      </div>
                    </Card>
                  ))}
                </div>
              </div>
            )
          })}

          {/* Cancelled column */}
          {orders.some((o: Order) => o.status === 'cancelado') && (
            <div className="flex-shrink-0 w-72">
              <div className="flex items-center gap-2 mb-3">
                <div className="w-3 h-3 rounded-full bg-red-500" />
                <h3 className="text-sm font-medium text-gray-900">Cancelado</h3>
                <Badge variant="outline">
                  {orders.filter((o: Order) => o.status === 'cancelado').length}
                </Badge>
              </div>
              <div className="space-y-3">
                {orders
                  .filter((o: Order) => o.status === 'cancelado')
                  .map((order: Order) => (
                    <Card key={order.id} className="p-4 opacity-60">
                      <p className="text-sm font-medium text-gray-900">{order.customer_name}</p>
                      <p className="text-sm text-gray-700">{formatCents(order.total_cents)}</p>
                      <p className="text-xs text-muted">{formatDate(order.created_at)}</p>
                    </Card>
                  ))}
              </div>
            </div>
          )}
        </div>
      )}

      {/* Tracking code modal */}
      <Modal
        open={!!selectedOrder}
        onClose={() => { setSelectedOrder(null); setTrackingCode('') }}
        title="Enviar Pedido"
        size="sm"
      >
        <div className="space-y-4">
          <p className="text-sm text-gray-700">
            Informe o codigo de rastreio (opcional) para marcar o pedido como enviado.
          </p>
          <Input
            label="Codigo de Rastreio"
            value={trackingCode}
            onChange={(e) => setTrackingCode(e.target.value)}
            placeholder="BR123456789XX"
          />
          <div className="flex justify-end gap-3">
            <Button variant="ghost" onClick={() => { setSelectedOrder(null); setTrackingCode('') }}>
              Cancelar
            </Button>
            <Button onClick={handleSendWithTracking} loading={updateMut.isPending}>
              Marcar como Enviado
            </Button>
          </div>
        </div>
      </Modal>
    </div>
  )
}
