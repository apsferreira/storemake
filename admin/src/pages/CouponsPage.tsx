import { useState, type FormEvent } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { couponsApi } from '@/api/endpoints'
import { Card } from '@/components/ui/Card'
import { Button } from '@/components/ui/Button'
import { Input } from '@/components/ui/Input'
import { Select } from '@/components/ui/Select'
import { Badge } from '@/components/ui/Badge'
import { Modal } from '@/components/ui/Modal'
import { Table, Thead, Tbody, Th, Tr, Td } from '@/components/ui/Table'
import { EmptyState } from '@/components/ui/EmptyState'
import { formatCents, sanitize } from '@/lib/utils'
import { Plus, Pencil, Trash2, Tag } from 'lucide-react'
import type { Coupon, CreateCouponRequest } from '@/types'

export function CouponsPage() {
  const qc = useQueryClient()
  const [modalOpen, setModalOpen] = useState(false)
  const [editCoupon, setEditCoupon] = useState<Coupon | null>(null)

  const { data, isLoading } = useQuery({
    queryKey: ['coupons'],
    queryFn: () => couponsApi.list(),
  })

  const createMut = useMutation({
    mutationFn: (d: CreateCouponRequest) => couponsApi.create(d),
    onSuccess: () => { qc.invalidateQueries({ queryKey: ['coupons'] }); setModalOpen(false) },
  })

  const updateMut = useMutation({
    mutationFn: ({ id, data }: { id: string; data: Partial<CreateCouponRequest> }) =>
      couponsApi.update(id, data),
    onSuccess: () => { qc.invalidateQueries({ queryKey: ['coupons'] }); setEditCoupon(null); setModalOpen(false) },
  })

  const deleteMut = useMutation({
    mutationFn: (id: string) => couponsApi.delete(id),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['coupons'] }),
  })

  const coupons = data?.data || []

  function handleSubmit(e: FormEvent<HTMLFormElement>) {
    e.preventDefault()
    const fd = new FormData(e.currentTarget)
    const payload: CreateCouponRequest = {
      code: sanitize(fd.get('code') as string).toUpperCase(),
      discount_type: fd.get('discount_type') as 'percent' | 'fixed',
      discount_value: parseInt(fd.get('discount_value') as string) || 0,
      min_order_cents: Math.round(parseFloat(fd.get('min_order') as string || '0') * 100),
      max_uses: parseInt(fd.get('max_uses') as string) || 0,
      valid_until: (fd.get('valid_until') as string) || undefined,
      is_active: fd.get('is_active') === 'true',
    }

    if (editCoupon) {
      updateMut.mutate({ id: editCoupon.id, data: payload })
    } else {
      createMut.mutate(payload)
    }
  }

  function openCreate() { setEditCoupon(null); setModalOpen(true) }
  function openEdit(c: Coupon) { setEditCoupon(c); setModalOpen(true) }

  return (
    <div className="space-y-4">
      <div className="flex justify-end">
        <Button onClick={openCreate}><Plus size={16} /> Novo Cupom</Button>
      </div>

      <Card className="p-0 overflow-hidden">
        {isLoading ? (
          <div className="flex items-center justify-center py-12">
            <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary" />
          </div>
        ) : coupons.length === 0 ? (
          <EmptyState
            icon={<Tag size={48} />}
            title="Nenhum cupom"
            description="Crie cupons de desconto para suas campanhas."
            action={<Button onClick={openCreate}><Plus size={16} /> Criar Cupom</Button>}
          />
        ) : (
          <Table>
            <Thead>
              <tr>
                <Th>Codigo</Th>
                <Th>Tipo</Th>
                <Th>Valor</Th>
                <Th className="hidden sm:table-cell">Min. Pedido</Th>
                <Th className="hidden md:table-cell">Uso</Th>
                <Th>Status</Th>
                <Th>Acoes</Th>
              </tr>
            </Thead>
            <Tbody>
              {coupons.map((c: Coupon) => (
                <Tr key={c.id}>
                  <Td className="font-mono font-medium">{c.code}</Td>
                  <Td>
                    <Badge variant="outline">
                      {c.discount_type === 'percent' ? 'Percentual' : 'Fixo'}
                    </Badge>
                  </Td>
                  <Td className="font-medium">
                    {c.discount_type === 'percent'
                      ? `${c.discount_value}%`
                      : formatCents(c.discount_value)}
                  </Td>
                  <Td className="hidden sm:table-cell">
                    {c.min_order_cents > 0 ? formatCents(c.min_order_cents) : '-'}
                  </Td>
                  <Td className="hidden md:table-cell">
                    {c.used_count}{c.max_uses > 0 ? `/${c.max_uses}` : ''}
                  </Td>
                  <Td>
                    <Badge variant={c.is_active ? 'success' : 'default'}>
                      {c.is_active ? 'Ativo' : 'Inativo'}
                    </Badge>
                  </Td>
                  <Td>
                    <div className="flex gap-1">
                      <button onClick={() => openEdit(c)} className="p-1.5 rounded hover:bg-gray-100 text-gray-500">
                        <Pencil size={16} />
                      </button>
                      <button
                        onClick={() => { if (confirm('Excluir cupom?')) deleteMut.mutate(c.id) }}
                        className="p-1.5 rounded hover:bg-red-50 text-gray-500 hover:text-red-500"
                      >
                        <Trash2 size={16} />
                      </button>
                    </div>
                  </Td>
                </Tr>
              ))}
            </Tbody>
          </Table>
        )}
      </Card>

      {/* Create/Edit Modal */}
      <Modal
        open={modalOpen}
        onClose={() => { setModalOpen(false); setEditCoupon(null) }}
        title={editCoupon ? 'Editar Cupom' : 'Novo Cupom'}
      >
        <form onSubmit={handleSubmit} className="space-y-4">
          <Input
            name="code"
            label="Codigo do Cupom"
            required
            defaultValue={editCoupon?.code}
            placeholder="DESCONTO10"
            className="uppercase"
          />
          <div className="grid grid-cols-2 gap-4">
            <Select
              name="discount_type"
              label="Tipo"
              defaultValue={editCoupon?.discount_type || 'percent'}
              options={[
                { value: 'percent', label: 'Percentual (%)' },
                { value: 'fixed', label: 'Valor Fixo (R$)' },
              ]}
            />
            <Input
              name="discount_value"
              label="Valor"
              type="number"
              min="0"
              required
              defaultValue={editCoupon?.discount_value ?? ''}
            />
          </div>
          <div className="grid grid-cols-2 gap-4">
            <Input
              name="min_order"
              label="Pedido minimo (R$)"
              type="number"
              step="0.01"
              min="0"
              defaultValue={editCoupon ? (editCoupon.min_order_cents / 100).toFixed(2) : '0'}
            />
            <Input
              name="max_uses"
              label="Usos maximos (0=ilimitado)"
              type="number"
              min="0"
              defaultValue={editCoupon?.max_uses ?? 0}
            />
          </div>
          <Input
            name="valid_until"
            label="Valido ate"
            type="datetime-local"
            defaultValue={editCoupon?.valid_until?.slice(0, 16) || ''}
          />
          <Select
            name="is_active"
            label="Status"
            defaultValue={editCoupon?.is_active !== false ? 'true' : 'false'}
            options={[
              { value: 'true', label: 'Ativo' },
              { value: 'false', label: 'Inativo' },
            ]}
          />
          <div className="flex justify-end gap-3 pt-2">
            <Button type="button" variant="ghost" onClick={() => { setModalOpen(false); setEditCoupon(null) }}>
              Cancelar
            </Button>
            <Button type="submit" loading={createMut.isPending || updateMut.isPending}>
              {editCoupon ? 'Salvar' : 'Criar'}
            </Button>
          </div>
        </form>
      </Modal>
    </div>
  )
}
