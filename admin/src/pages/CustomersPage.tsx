import { useState, useMemo } from 'react'
import { useQuery } from '@tanstack/react-query'
import { ordersApi } from '@/api/endpoints'
import { Card } from '@/components/ui/Card'
import { Table, Thead, Tbody, Th, Tr, Td } from '@/components/ui/Table'
import { Badge } from '@/components/ui/Badge'
import { Modal } from '@/components/ui/Modal'
import { EmptyState } from '@/components/ui/EmptyState'
import { formatCents, formatDate } from '@/lib/utils'
import { Users, Search } from 'lucide-react'
import type { Order } from '@/types'

interface CustomerSummary {
  email: string
  name: string
  phone: string
  ordersCount: number
  totalSpent: number
  lastOrder: string
  orders: Order[]
}

export function CustomersPage() {
  const [search, setSearch] = useState('')
  const [selectedCustomer, setSelectedCustomer] = useState<CustomerSummary | null>(null)

  const { data, isLoading } = useQuery({
    queryKey: ['orders', 'all-customers'],
    queryFn: () => ordersApi.list({ per_page: 500 }),
  })

  const customers = useMemo(() => {
    const orders = data?.data || []
    const map = new Map<string, CustomerSummary>()

    orders.forEach((order: Order) => {
      const key = order.customer_email.toLowerCase()
      if (!map.has(key)) {
        map.set(key, {
          email: order.customer_email,
          name: order.customer_name,
          phone: order.customer_phone,
          ordersCount: 0,
          totalSpent: 0,
          lastOrder: order.created_at,
          orders: [],
        })
      }
      const c = map.get(key)!
      c.ordersCount++
      c.totalSpent += order.total_cents
      c.orders.push(order)
      if (new Date(order.created_at) > new Date(c.lastOrder)) {
        c.lastOrder = order.created_at
        c.name = order.customer_name
      }
    })

    return Array.from(map.values()).sort((a, b) => b.totalSpent - a.totalSpent)
  }, [data])

  const filtered = customers.filter(
    (c) =>
      c.name.toLowerCase().includes(search.toLowerCase()) ||
      c.email.toLowerCase().includes(search.toLowerCase()) ||
      c.phone.includes(search)
  )

  return (
    <div className="space-y-4">
      <div className="relative w-full sm:w-72">
        <Search size={16} className="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400" />
        <input
          type="text"
          placeholder="Buscar clientes..."
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          className="w-full pl-9 pr-3 py-2 text-sm rounded-lg border border-border focus:outline-none focus:ring-2 focus:ring-primary/20"
        />
      </div>

      <Card className="p-0 overflow-hidden">
        {isLoading ? (
          <div className="flex items-center justify-center py-12">
            <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary" />
          </div>
        ) : filtered.length === 0 ? (
          <EmptyState
            icon={<Users size={48} />}
            title="Nenhum cliente"
            description="Os clientes aparecerao aqui quando realizarem pedidos."
          />
        ) : (
          <Table>
            <Thead>
              <tr>
                <Th>Cliente</Th>
                <Th className="hidden sm:table-cell">Telefone</Th>
                <Th>Pedidos</Th>
                <Th>Total Gasto</Th>
                <Th className="hidden md:table-cell">Ultimo Pedido</Th>
              </tr>
            </Thead>
            <Tbody>
              {filtered.map((c) => (
                <Tr key={c.email} onClick={() => setSelectedCustomer(c)}>
                  <Td>
                    <div>
                      <p className="font-medium text-gray-900">{c.name}</p>
                      <p className="text-xs text-muted">{c.email}</p>
                    </div>
                  </Td>
                  <Td className="hidden sm:table-cell">{c.phone || '-'}</Td>
                  <Td><Badge variant="outline">{c.ordersCount}</Badge></Td>
                  <Td className="font-medium">{formatCents(c.totalSpent)}</Td>
                  <Td className="hidden md:table-cell text-sm text-muted">{formatDate(c.lastOrder)}</Td>
                </Tr>
              ))}
            </Tbody>
          </Table>
        )}
      </Card>

      {/* Customer Detail Modal */}
      <Modal
        open={!!selectedCustomer}
        onClose={() => setSelectedCustomer(null)}
        title={`Cliente: ${selectedCustomer?.name || ''}`}
        size="lg"
      >
        {selectedCustomer && (
          <div className="space-y-4">
            <div className="grid grid-cols-2 sm:grid-cols-3 gap-4">
              <div>
                <p className="text-xs text-muted">Email</p>
                <p className="text-sm font-medium">{selectedCustomer.email}</p>
              </div>
              <div>
                <p className="text-xs text-muted">Telefone</p>
                <p className="text-sm font-medium">{selectedCustomer.phone || '-'}</p>
              </div>
              <div>
                <p className="text-xs text-muted">Total Gasto</p>
                <p className="text-sm font-bold text-primary">{formatCents(selectedCustomer.totalSpent)}</p>
              </div>
            </div>

            <div>
              <h4 className="text-sm font-medium text-gray-900 mb-2">
                Historico de Pedidos ({selectedCustomer.ordersCount})
              </h4>
              <div className="space-y-2 max-h-64 overflow-y-auto">
                {selectedCustomer.orders
                  .sort((a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime())
                  .map((order) => (
                    <div
                      key={order.id}
                      className="flex items-center justify-between p-3 bg-gray-50 rounded-lg"
                    >
                      <div>
                        <p className="text-sm font-medium">{formatCents(order.total_cents)}</p>
                        <p className="text-xs text-muted">{formatDate(order.created_at)}</p>
                      </div>
                      <Badge
                        variant={
                          order.status === 'entregue' ? 'success' :
                          order.status === 'cancelado' ? 'danger' :
                          order.status === 'pendente' ? 'warning' : 'info'
                        }
                      >
                        {order.status}
                      </Badge>
                    </div>
                  ))}
              </div>
            </div>
          </div>
        )}
      </Modal>
    </div>
  )
}
