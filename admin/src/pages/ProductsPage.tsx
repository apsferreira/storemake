import { useState, type FormEvent } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { productsApi, categoriesApi } from '@/api/endpoints'
import { Card } from '@/components/ui/Card'
import { Button } from '@/components/ui/Button'
import { Input } from '@/components/ui/Input'
import { Select } from '@/components/ui/Select'
import { Badge } from '@/components/ui/Badge'
import { Modal } from '@/components/ui/Modal'
import { Table, Thead, Tbody, Th, Tr, Td } from '@/components/ui/Table'
import { Pagination } from '@/components/ui/Pagination'
import { EmptyState } from '@/components/ui/EmptyState'
import { formatCents, sanitize } from '@/lib/utils'
import { Plus, Pencil, Trash2, Search, Package, Upload } from 'lucide-react'
import type { Product, CreateProductRequest } from '@/types'

export function ProductsPage() {
  const qc = useQueryClient()
  const [page, setPage] = useState(1)
  const [search, setSearch] = useState('')
  const [modalOpen, setModalOpen] = useState(false)
  const [editProduct, setEditProduct] = useState<Product | null>(null)
  const [photoModal, setPhotoModal] = useState<string | null>(null)

  const { data, isLoading } = useQuery({
    queryKey: ['products', page, search],
    queryFn: () => productsApi.list({ page, per_page: 20, search: search || undefined }),
  })

  const { data: categoriesData } = useQuery({
    queryKey: ['categories'],
    queryFn: () => categoriesApi.list(),
  })

  const createMut = useMutation({
    mutationFn: (data: CreateProductRequest) => productsApi.create(data),
    onSuccess: () => { qc.invalidateQueries({ queryKey: ['products'] }); setModalOpen(false) },
  })

  const updateMut = useMutation({
    mutationFn: ({ id, data }: { id: string; data: Partial<CreateProductRequest> }) =>
      productsApi.update(id, data),
    onSuccess: () => { qc.invalidateQueries({ queryKey: ['products'] }); setEditProduct(null); setModalOpen(false) },
  })

  const deleteMut = useMutation({
    mutationFn: (id: string) => productsApi.delete(id),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['products'] }),
  })

  const uploadMut = useMutation({
    mutationFn: ({ id, formData }: { id: string; formData: FormData }) =>
      productsApi.uploadPhotos(id, formData),
    onSuccess: () => { qc.invalidateQueries({ queryKey: ['products'] }); setPhotoModal(null) },
  })

  const products = data?.data || []
  const totalPages = data?.total_pages || 1
  const categories = (categoriesData as { data?: { id: string; name: string }[] })?.data || []

  function handleSubmit(e: FormEvent<HTMLFormElement>) {
    e.preventDefault()
    const fd = new FormData(e.currentTarget)
    const payload: CreateProductRequest = {
      name: sanitize(fd.get('name') as string),
      slug: sanitize(fd.get('slug') as string),
      description: sanitize(fd.get('description') as string) || undefined,
      price_cents: Math.round(parseFloat(fd.get('price') as string) * 100),
      compare_price_cents: fd.get('compare_price')
        ? Math.round(parseFloat(fd.get('compare_price') as string) * 100)
        : 0,
      sku: sanitize(fd.get('sku') as string) || undefined,
      stock_quantity: parseInt(fd.get('stock_quantity') as string) || 0,
      stock_alert_threshold: parseInt(fd.get('stock_alert_threshold') as string) || 5,
      is_active: fd.get('is_active') === 'true',
      categoria_id: (fd.get('categoria_id') as string) || undefined,
    }

    if (editProduct) {
      updateMut.mutate({ id: editProduct.id, data: payload })
    } else {
      createMut.mutate(payload)
    }
  }

  function handlePhotoUpload(e: FormEvent<HTMLFormElement>) {
    e.preventDefault()
    if (!photoModal) return
    const fd = new FormData(e.currentTarget)
    const files = fd.getAll('photos')
    if (files.length === 0) return
    const uploadFd = new FormData()
    files.forEach((f) => uploadFd.append('photos', f))
    uploadMut.mutate({ id: photoModal, formData: uploadFd })
  }

  function openEdit(product: Product) {
    setEditProduct(product)
    setModalOpen(true)
  }

  function openCreate() {
    setEditProduct(null)
    setModalOpen(true)
  }

  return (
    <div className="space-y-4">
      {/* Search + Actions */}
      <div className="flex flex-col sm:flex-row gap-3 items-start sm:items-center justify-between">
        <div className="relative w-full sm:w-72">
          <Search size={16} className="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400" />
          <input
            type="text"
            placeholder="Buscar produtos..."
            value={search}
            onChange={(e) => { setSearch(e.target.value); setPage(1) }}
            className="w-full pl-9 pr-3 py-2 text-sm rounded-lg border border-border focus:outline-none focus:ring-2 focus:ring-primary/20"
          />
        </div>
        <Button onClick={openCreate}>
          <Plus size={16} /> Novo Produto
        </Button>
      </div>

      {/* Table */}
      <Card className="p-0 overflow-hidden">
        {isLoading ? (
          <div className="flex items-center justify-center py-12">
            <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary" />
          </div>
        ) : products.length === 0 ? (
          <EmptyState
            icon={<Package size={48} />}
            title="Nenhum produto"
            description="Crie seu primeiro produto para comecar a vender."
            action={<Button onClick={openCreate}><Plus size={16} /> Criar Produto</Button>}
          />
        ) : (
          <>
            <Table>
              <Thead>
                <tr>
                  <Th>Produto</Th>
                  <Th className="hidden sm:table-cell">SKU</Th>
                  <Th>Preco</Th>
                  <Th className="hidden md:table-cell">Estoque</Th>
                  <Th className="hidden md:table-cell">Status</Th>
                  <Th>Acoes</Th>
                </tr>
              </Thead>
              <Tbody>
                {products.map((p: Product) => (
                  <Tr key={p.id}>
                    <Td>
                      <div>
                        <p className="font-medium text-gray-900">{p.name}</p>
                        <p className="text-xs text-muted">{p.slug}</p>
                      </div>
                    </Td>
                    <Td className="hidden sm:table-cell">{p.sku || '-'}</Td>
                    <Td>
                      <div>
                        <p className="font-medium">{formatCents(p.price_cents)}</p>
                        {p.compare_price_cents > 0 && (
                          <p className="text-xs text-muted line-through">{formatCents(p.compare_price_cents)}</p>
                        )}
                      </div>
                    </Td>
                    <Td className="hidden md:table-cell">
                      <Badge variant={p.stock_quantity <= p.stock_alert_threshold ? 'warning' : 'default'}>
                        {p.stock_quantity}
                      </Badge>
                    </Td>
                    <Td className="hidden md:table-cell">
                      <Badge variant={p.is_active ? 'success' : 'default'}>
                        {p.is_active ? 'Ativo' : 'Inativo'}
                      </Badge>
                    </Td>
                    <Td>
                      <div className="flex gap-1">
                        <button onClick={() => setPhotoModal(p.id)} className="p-1.5 rounded hover:bg-gray-100 text-gray-500" title="Fotos">
                          <Upload size={16} />
                        </button>
                        <button onClick={() => openEdit(p)} className="p-1.5 rounded hover:bg-gray-100 text-gray-500" title="Editar">
                          <Pencil size={16} />
                        </button>
                        <button
                          onClick={() => { if (confirm('Excluir produto?')) deleteMut.mutate(p.id) }}
                          className="p-1.5 rounded hover:bg-red-50 text-gray-500 hover:text-red-500"
                          title="Excluir"
                        >
                          <Trash2 size={16} />
                        </button>
                      </div>
                    </Td>
                  </Tr>
                ))}
              </Tbody>
            </Table>
            <div className="px-4 pb-4">
              <Pagination page={page} totalPages={totalPages} onPageChange={setPage} />
            </div>
          </>
        )}
      </Card>

      {/* Create/Edit Modal */}
      <Modal
        open={modalOpen}
        onClose={() => { setModalOpen(false); setEditProduct(null) }}
        title={editProduct ? 'Editar Produto' : 'Novo Produto'}
        size="lg"
      >
        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
            <Input name="name" label="Nome" required defaultValue={editProduct?.name} />
            <Input name="slug" label="Slug" required defaultValue={editProduct?.slug} />
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Descricao</label>
            <textarea
              name="description"
              rows={3}
              defaultValue={editProduct?.description || ''}
              className="w-full rounded-lg border border-border px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-primary/20"
            />
          </div>
          <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
            <Input
              name="price"
              label="Preco (R$)"
              type="number"
              step="0.01"
              min="0"
              required
              defaultValue={editProduct ? (editProduct.price_cents / 100).toFixed(2) : ''}
            />
            <Input
              name="compare_price"
              label="Preco comparativo (R$)"
              type="number"
              step="0.01"
              min="0"
              defaultValue={editProduct?.compare_price_cents ? (editProduct.compare_price_cents / 100).toFixed(2) : ''}
            />
            <Input name="sku" label="SKU" defaultValue={editProduct?.sku || ''} />
          </div>
          <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
            <Input
              name="stock_quantity"
              label="Estoque"
              type="number"
              min="0"
              defaultValue={editProduct?.stock_quantity ?? 0}
            />
            <Input
              name="stock_alert_threshold"
              label="Alerta estoque"
              type="number"
              min="0"
              defaultValue={editProduct?.stock_alert_threshold ?? 5}
            />
            <Select
              name="is_active"
              label="Status"
              defaultValue={editProduct?.is_active !== false ? 'true' : 'false'}
              options={[
                { value: 'true', label: 'Ativo' },
                { value: 'false', label: 'Inativo' },
              ]}
            />
          </div>
          {categories.length > 0 && (
            <Select
              name="categoria_id"
              label="Categoria"
              defaultValue={editProduct?.categoria_id || ''}
              options={[
                { value: '', label: 'Sem categoria' },
                ...categories.map((c: { id: string; name: string }) => ({ value: c.id, label: c.name })),
              ]}
            />
          )}
          <div className="flex justify-end gap-3 pt-2">
            <Button type="button" variant="ghost" onClick={() => { setModalOpen(false); setEditProduct(null) }}>
              Cancelar
            </Button>
            <Button type="submit" loading={createMut.isPending || updateMut.isPending}>
              {editProduct ? 'Salvar' : 'Criar'}
            </Button>
          </div>
        </form>
      </Modal>

      {/* Upload photos modal */}
      <Modal
        open={!!photoModal}
        onClose={() => setPhotoModal(null)}
        title="Upload de Fotos"
      >
        <form onSubmit={handlePhotoUpload} className="space-y-4">
          <input
            type="file"
            name="photos"
            accept="image/jpeg,image/png,image/webp,image/gif"
            multiple
            className="block w-full text-sm text-gray-500 file:mr-4 file:py-2 file:px-4 file:rounded-lg file:border-0 file:text-sm file:font-medium file:bg-primary-light file:text-primary hover:file:bg-indigo-100"
          />
          <p className="text-xs text-muted">Formatos aceitos: JPEG, PNG, WebP, GIF. Max 5MB por foto.</p>
          <div className="flex justify-end gap-3">
            <Button type="button" variant="ghost" onClick={() => setPhotoModal(null)}>Cancelar</Button>
            <Button type="submit" loading={uploadMut.isPending}>Enviar</Button>
          </div>
        </form>
      </Modal>
    </div>
  )
}
