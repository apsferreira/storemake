import { useState } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { inventoryApi } from '@/api/endpoints'
import { Card } from '@/components/ui/Card'
import { Badge } from '@/components/ui/Badge'
import { Button } from '@/components/ui/Button'
import { Modal } from '@/components/ui/Modal'
import { Input } from '@/components/ui/Input'
import { formatDate } from '@/lib/utils'
import { ArrowLeft, Package, TrendingUp, TrendingDown, AlertTriangle, ShoppingCart, Plus } from 'lucide-react'
import type { AdjustQuantityRequest, UpsertAllocationRequest, CreateSupplierOrderRequest } from '@/types'

const movTypeLabel: Record<string, string> = {
  entrada: 'Entrada',
  saida_venda: 'Saída (venda)',
  saida_perda: 'Saída (perda)',
  transferencia: 'Transferência',
  ajuste: 'Ajuste',
  devolucao: 'Devolução',
}

export function InventoryDetailPage() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const qc = useQueryClient()

  const [showAdjust, setShowAdjust] = useState(false)
  const [showAllocate, setShowAllocate] = useState(false)
  const [showOrder, setShowOrder] = useState(false)
  const [adjustForm, setAdjustForm] = useState<AdjustQuantityRequest>({ delta: 0 })
  const [allocForm, setAllocForm] = useState<{ lojaId: string } & UpsertAllocationRequest>({
    lojaId: '',
    quantity_allocated: 0,
    profit_share_pct: 0,
  })
  const [orderForm, setOrderForm] = useState<CreateSupplierOrderRequest>({ quantity_ordered: 0 })

  const { data, isLoading } = useQuery({
    queryKey: ['inventory', id],
    queryFn: () => inventoryApi.get(id!),
    enabled: !!id,
  })

  const adjustMut = useMutation({
    mutationFn: (req: AdjustQuantityRequest) => inventoryApi.adjust(id!, req),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['inventory', id] })
      qc.invalidateQueries({ queryKey: ['inventory'] })
      setShowAdjust(false)
      setAdjustForm({ delta: 0 })
    },
  })

  const allocMut = useMutation({
    mutationFn: ({ lojaId, ...req }: { lojaId: string } & UpsertAllocationRequest) =>
      inventoryApi.allocate(id!, lojaId, req),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['inventory', id] })
      setShowAllocate(false)
      setAllocForm({ lojaId: '', quantity_allocated: 0, profit_share_pct: 0 })
    },
  })

  const orderMut = useMutation({
    mutationFn: (req: CreateSupplierOrderRequest) => inventoryApi.createOrder(id!, req),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['inventory', id] })
      setShowOrder(false)
      setOrderForm({ quantity_ordered: 0 })
    },
  })

  if (isLoading) {
    return <div className="p-8 text-center text-gray-400 text-sm">Carregando...</div>
  }

  if (!data?.master) {
    return <div className="p-8 text-center text-gray-400 text-sm">SKU não encontrado</div>
  }

  const { master, allocations, movements } = data
  const available = master.quantity_total - master.quantity_reserved

  return (
    <div className="space-y-6">
      {/* Breadcrumb */}
      <div className="flex items-center gap-2">
        <button
          onClick={() => navigate('/estoque')}
          className="flex items-center gap-1 text-sm text-gray-500 hover:text-gray-700"
        >
          <ArrowLeft size={16} />
          Estoque
        </button>
        <span className="text-gray-300">/</span>
        <span className="text-sm text-gray-700 font-medium">{master.nome}</span>
      </div>

      {/* Header + actions */}
      <div className="flex items-start justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">{master.nome}</h1>
          {master.sku_global && <p className="text-sm text-gray-400 mt-0.5">SKU: {master.sku_global}</p>}
        </div>
        <div className="flex gap-2 shrink-0">
          <Button variant="outline" onClick={() => setShowAdjust(true)}>Ajustar quantidade</Button>
          <Button variant="outline" onClick={() => setShowAllocate(true)}>Alocar para loja</Button>
          <Button onClick={() => setShowOrder(true)}>
            <ShoppingCart size={16} />
            Pedido de reposição
          </Button>
        </div>
      </div>

      {/* Stats cards */}
      <div className="grid grid-cols-2 sm:grid-cols-4 gap-4">
        {[
          { label: 'Total', value: master.quantity_total, icon: Package, color: 'text-blue-600 bg-blue-50' },
          { label: 'Reservado', value: master.quantity_reserved, icon: TrendingDown, color: 'text-amber-600 bg-amber-50' },
          { label: 'Disponível', value: available, icon: TrendingUp, color: available <= master.reorder_point ? 'text-red-600 bg-red-50' : 'text-green-600 bg-green-50' },
          { label: 'Reorder point', value: master.reorder_point, icon: AlertTriangle, color: 'text-gray-600 bg-gray-100' },
        ].map(stat => (
          <Card key={stat.label}>
            <div className="p-4 flex items-center gap-3">
              <div className={`w-10 h-10 rounded-lg flex items-center justify-center ${stat.color}`}>
                <stat.icon size={20} />
              </div>
              <div>
                <p className="text-xs text-gray-500">{stat.label}</p>
                <p className="text-xl font-bold text-gray-900">{stat.value}</p>
              </div>
            </div>
          </Card>
        ))}
      </div>

      {/* Allocations */}
      <Card>
        <div className="p-4 border-b border-gray-100">
          <h2 className="font-semibold text-gray-800">Alocações por loja</h2>
        </div>
        {!allocations || allocations.length === 0 ? (
          <div className="p-6 text-center text-gray-400 text-sm">
            Nenhuma alocação. Use "Alocar para loja" para distribuir estoque.
          </div>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b border-gray-100 text-left text-gray-500">
                  <th className="px-4 py-3 font-medium">Loja ID</th>
                  <th className="px-4 py-3 font-medium text-center">Alocado</th>
                  <th className="px-4 py-3 font-medium text-center">Vendido</th>
                  <th className="px-4 py-3 font-medium text-center">% Lucro</th>
                  <th className="px-4 py-3 font-medium text-center">Status</th>
                </tr>
              </thead>
              <tbody>
                {allocations.map(alloc => (
                  <tr key={alloc.id} className="border-b border-gray-50 hover:bg-gray-50">
                    <td className="px-4 py-3 font-mono text-xs text-gray-600">{alloc.loja_id.slice(0, 8)}…</td>
                    <td className="px-4 py-3 text-center text-gray-700">{alloc.quantity_allocated}</td>
                    <td className="px-4 py-3 text-center text-gray-500">{alloc.quantity_sold}</td>
                    <td className="px-4 py-3 text-center text-gray-500">{alloc.profit_share_pct}%</td>
                    <td className="px-4 py-3 text-center">
                      {alloc.is_active ? <Badge variant="success">Ativa</Badge> : <Badge variant="default">Inativa</Badge>}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </Card>

      {/* Movements */}
      <Card>
        <div className="p-4 border-b border-gray-100">
          <h2 className="font-semibold text-gray-800">Histórico de movimentações</h2>
        </div>
        {!movements || movements.length === 0 ? (
          <div className="p-6 text-center text-gray-400 text-sm">Nenhuma movimentação registrada.</div>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b border-gray-100 text-left text-gray-500">
                  <th className="px-4 py-3 font-medium">Data</th>
                  <th className="px-4 py-3 font-medium">Tipo</th>
                  <th className="px-4 py-3 font-medium text-center">Qtd.</th>
                  <th className="px-4 py-3 font-medium text-center">Antes</th>
                  <th className="px-4 py-3 font-medium text-center">Depois</th>
                  <th className="px-4 py-3 font-medium">Observação</th>
                </tr>
              </thead>
              <tbody>
                {movements.map(mov => (
                  <tr key={mov.id} className="border-b border-gray-50 hover:bg-gray-50">
                    <td className="px-4 py-3 text-gray-500 whitespace-nowrap">{formatDate(mov.created_at)}</td>
                    <td className="px-4 py-3">
                      <span className={`text-xs font-medium ${mov.quantity > 0 ? 'text-green-600' : 'text-red-500'}`}>
                        {movTypeLabel[mov.movement_type] || mov.movement_type}
                      </span>
                    </td>
                    <td className="px-4 py-3 text-center">
                      <span className={mov.quantity > 0 ? 'text-green-600' : 'text-red-500'}>
                        {mov.quantity > 0 ? '+' : ''}{mov.quantity}
                      </span>
                    </td>
                    <td className="px-4 py-3 text-center text-gray-500">{mov.quantity_before}</td>
                    <td className="px-4 py-3 text-center text-gray-700 font-medium">{mov.quantity_after}</td>
                    <td className="px-4 py-3 text-gray-400 text-xs">{mov.observacao || '—'}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </Card>

      {/* Adjust Modal */}
      <Modal open={showAdjust} onClose={() => setShowAdjust(false)} title="Ajustar quantidade">
        <div className="space-y-4">
          <p className="text-sm text-gray-500">Use valores positivos para entrada e negativos para saída.</p>
          <Input
            label="Delta (quantidade) *"
            type="number"
            value={adjustForm.delta}
            onChange={e => setAdjustForm(f => ({ ...f, delta: Number(e.target.value) }))}
            placeholder="+10 ou -5"
          />
          <Input
            label="Observação"
            value={adjustForm.observacao || ''}
            onChange={e => setAdjustForm(f => ({ ...f, observacao: e.target.value }))}
            placeholder="Motivo do ajuste..."
          />
          <div className="flex justify-end gap-2 pt-2">
            <Button variant="outline" onClick={() => setShowAdjust(false)}>Cancelar</Button>
            <Button
              onClick={() => adjustMut.mutate(adjustForm)}
              disabled={adjustForm.delta === 0 || adjustMut.isPending}
            >
              {adjustMut.isPending ? 'Ajustando...' : 'Confirmar ajuste'}
            </Button>
          </div>
        </div>
      </Modal>

      {/* Allocate Modal */}
      <Modal open={showAllocate} onClose={() => setShowAllocate(false)} title="Alocar para loja">
        <div className="space-y-4">
          <Input
            label="Loja ID *"
            value={allocForm.lojaId}
            onChange={e => setAllocForm(f => ({ ...f, lojaId: e.target.value }))}
            placeholder="UUID da loja"
          />
          <div className="grid grid-cols-2 gap-4">
            <Input
              label="Quantidade alocada *"
              type="number"
              value={allocForm.quantity_allocated}
              onChange={e => setAllocForm(f => ({ ...f, quantity_allocated: Number(e.target.value) }))}
            />
            <Input
              label="% de lucro"
              type="number"
              value={allocForm.profit_share_pct}
              onChange={e => setAllocForm(f => ({ ...f, profit_share_pct: Number(e.target.value) }))}
              placeholder="0–100"
            />
          </div>
          <div className="flex justify-end gap-2 pt-2">
            <Button variant="outline" onClick={() => setShowAllocate(false)}>Cancelar</Button>
            <Button
              onClick={() => allocMut.mutate(allocForm)}
              disabled={!allocForm.lojaId || allocForm.quantity_allocated <= 0 || allocMut.isPending}
            >
              {allocMut.isPending ? 'Alocando...' : 'Alocar'}
            </Button>
          </div>
        </div>
      </Modal>

      {/* Supplier Order Modal */}
      <Modal open={showOrder} onClose={() => setShowOrder(false)} title="Pedido de reposição">
        <div className="space-y-4">
          <Input
            label="Quantidade a pedir *"
            type="number"
            value={orderForm.quantity_ordered}
            onChange={e => setOrderForm(f => ({ ...f, quantity_ordered: Number(e.target.value) }))}
          />
          <Input
            label="Fornecedor"
            value={orderForm.fornecedor_nome || ''}
            onChange={e => setOrderForm(f => ({ ...f, fornecedor_nome: e.target.value }))}
            placeholder="Nome do fornecedor"
          />
          <Input
            label="Observação"
            value={orderForm.observacao || ''}
            onChange={e => setOrderForm(f => ({ ...f, observacao: e.target.value }))}
            placeholder="Informações adicionais..."
          />
          <div className="flex justify-end gap-2 pt-2">
            <Button variant="outline" onClick={() => setShowOrder(false)}>Cancelar</Button>
            <Button
              onClick={() => orderMut.mutate(orderForm)}
              disabled={orderForm.quantity_ordered <= 0 || orderMut.isPending}
            >
              {orderMut.isPending ? 'Criando...' : 'Criar pedido'}
            </Button>
          </div>
        </div>
      </Modal>
    </div>
  )
}
