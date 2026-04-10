import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useNavigate } from 'react-router-dom'
import { inventoryApi } from '@/api/endpoints'
import { Card } from '@/components/ui/Card'
import { Badge } from '@/components/ui/Badge'
import { Button } from '@/components/ui/Button'
import { Modal } from '@/components/ui/Modal'
import { Input } from '@/components/ui/Input'
import { EmptyState } from '@/components/ui/EmptyState'

import { Package, AlertTriangle, Plus, Search } from 'lucide-react'
import type { InventoryMaster, CreateInventoryMasterRequest } from '@/types'

function StockBadge({ master }: { master: InventoryMaster }) {
  const available = master.quantity_total - master.quantity_reserved
  if (available <= 0) return <Badge variant="danger">Sem estoque</Badge>
  if (available <= master.reorder_point) return <Badge variant="warning">Estoque baixo</Badge>
  return <Badge variant="success">{available} disp.</Badge>
}

export function InventoryPage() {
  const qc = useQueryClient()
  const navigate = useNavigate()
  const [search, setSearch] = useState('')
  const [showCreate, setShowCreate] = useState(false)
  const [form, setForm] = useState<Partial<CreateInventoryMasterRequest>>({})

  const { data, isLoading } = useQuery({
    queryKey: ['inventory'],
    queryFn: () => inventoryApi.list(),
  })

  const createMut = useMutation({
    mutationFn: (req: CreateInventoryMasterRequest) => inventoryApi.create(req),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['inventory'] })
      setShowCreate(false)
      setForm({})
    },
  })

  const items = (data?.items || []).filter(m =>
    !search || m.nome.toLowerCase().includes(search.toLowerCase()) || m.sku_global.toLowerCase().includes(search.toLowerCase())
  )

  const lowStockCount = (data?.items || []).filter(
    m => m.quantity_total - m.quantity_reserved <= m.reorder_point
  ).length

  function handleCreate() {
    if (!form.nome || !form.quantity_total) return
    createMut.mutate({
      nome: form.nome,
      sku_global: form.sku_global,
      descricao: form.descricao,
      unidade: form.unidade || 'un',
      custo_unitario_cents: form.custo_unitario_cents || 0,
      quantity_total: form.quantity_total,
      reorder_point: form.reorder_point || 5,
      reorder_quantity: form.reorder_quantity || 10,
    })
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Estoque</h1>
          <p className="text-sm text-gray-500 mt-1">Gerenciamento centralizado de inventário multi-loja</p>
        </div>
        <div className="flex gap-2">
          {lowStockCount > 0 && (
            <Button variant="outline" onClick={() => navigate('/estoque/alertas')}>
              <AlertTriangle size={16} className="text-amber-500" />
              {lowStockCount} alertas
            </Button>
          )}
          <Button onClick={() => setShowCreate(true)}>
            <Plus size={16} />
            Novo SKU
          </Button>
        </div>
      </div>

      {/* Search */}
      <div className="relative w-full sm:w-72">
        <Search size={16} className="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400" />
        <input
          value={search}
          onChange={e => setSearch(e.target.value)}
          placeholder="Buscar por nome ou SKU..."
          className="w-full pl-9 pr-4 py-2 rounded-lg border border-gray-200 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500"
        />
      </div>

      {/* Table */}
      {isLoading ? (
        <Card><div className="p-8 text-center text-gray-400 text-sm">Carregando...</div></Card>
      ) : items.length === 0 ? (
        <EmptyState icon={<Package size={48} />} title="Nenhum SKU cadastrado" description="Crie o primeiro item de inventário" />
      ) : (
        <Card>
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b border-gray-100 text-left">
                  <th className="px-4 py-3 font-semibold text-gray-600">Nome / SKU</th>
                  <th className="px-4 py-3 font-semibold text-gray-600 text-center">Total</th>
                  <th className="px-4 py-3 font-semibold text-gray-600 text-center">Reservado</th>
                  <th className="px-4 py-3 font-semibold text-gray-600 text-center">Disponível</th>
                  <th className="px-4 py-3 font-semibold text-gray-600 text-center">Reorder</th>
                  <th className="px-4 py-3 font-semibold text-gray-600 text-center">Status</th>
                  <th className="px-4 py-3" />
                </tr>
              </thead>
              <tbody>
                {items.map(item => {
                  const available = item.quantity_total - item.quantity_reserved
                  return (
                    <tr key={item.id} className="border-b border-gray-50 hover:bg-gray-50">
                      <td className="px-4 py-3">
                        <p className="font-medium text-gray-900">{item.nome}</p>
                        <p className="text-xs text-gray-400">{item.sku_global || '—'}</p>
                      </td>
                      <td className="px-4 py-3 text-center text-gray-700">{item.quantity_total}</td>
                      <td className="px-4 py-3 text-center text-gray-500">{item.quantity_reserved}</td>
                      <td className="px-4 py-3 text-center">
                        <span className={available <= item.reorder_point ? 'text-amber-600 font-semibold' : 'text-gray-700'}>
                          {available}
                        </span>
                        {available <= item.reorder_point && (
                          <AlertTriangle size={14} className="inline ml-1 text-amber-500" />
                        )}
                      </td>
                      <td className="px-4 py-3 text-center text-gray-500">{item.reorder_point}</td>
                      <td className="px-4 py-3 text-center">
                        <StockBadge master={item} />
                      </td>
                      <td className="px-4 py-3 text-right">
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => navigate(`/estoque/${item.id}`)}
                        >
                          Detalhes
                        </Button>
                      </td>
                    </tr>
                  )
                })}
              </tbody>
            </table>
          </div>
        </Card>
      )}

      {/* Create Modal */}
      <Modal open={showCreate} onClose={() => setShowCreate(false)} title="Novo SKU">
        <div className="space-y-4">
          <Input
            label="Nome *"
            value={form.nome || ''}
            onChange={e => setForm(f => ({ ...f, nome: e.target.value }))}
            placeholder="Ex: Camiseta P/M/G"
          />
          <Input
            label="SKU global"
            value={form.sku_global || ''}
            onChange={e => setForm(f => ({ ...f, sku_global: e.target.value }))}
            placeholder="Ex: CAM-PMG-001"
          />
          <div className="grid grid-cols-2 gap-4">
            <Input
              label="Quantidade inicial *"
              type="number"
              value={form.quantity_total || ''}
              onChange={e => setForm(f => ({ ...f, quantity_total: Number(e.target.value) }))}
            />
            <Input
              label="Unidade"
              value={form.unidade || ''}
              onChange={e => setForm(f => ({ ...f, unidade: e.target.value }))}
              placeholder="un, kg, cx..."
            />
          </div>
          <div className="grid grid-cols-2 gap-4">
            <Input
              label="Ponto de reposição"
              type="number"
              value={form.reorder_point || ''}
              onChange={e => setForm(f => ({ ...f, reorder_point: Number(e.target.value) }))}
              placeholder="5"
            />
            <Input
              label="Qtd. reposição"
              type="number"
              value={form.reorder_quantity || ''}
              onChange={e => setForm(f => ({ ...f, reorder_quantity: Number(e.target.value) }))}
              placeholder="10"
            />
          </div>
          <div className="flex justify-end gap-2 pt-2">
            <Button variant="outline" onClick={() => setShowCreate(false)}>Cancelar</Button>
            <Button onClick={handleCreate} disabled={!form.nome || !form.quantity_total || createMut.isPending}>
              {createMut.isPending ? 'Criando...' : 'Criar SKU'}
            </Button>
          </div>
        </div>
      </Modal>
    </div>
  )
}
