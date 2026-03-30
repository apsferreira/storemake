import { useState, useMemo } from 'react'
import { useQuery } from '@tanstack/react-query'
import { ordersApi, productsApi } from '@/api/endpoints'
import { Card, CardTitle } from '@/components/ui/Card'
import { Select } from '@/components/ui/Select'
import { formatCents } from '@/lib/utils'
import {
  BarChart,
  Bar,
  LineChart,
  Line,
  PieChart,
  Pie,
  Cell,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
  Legend,
} from 'recharts'
import type { Order, Product } from '@/types'

const COLORS = ['#6366f1', '#22c55e', '#f59e0b', '#ef4444', '#3b82f6', '#8b5cf6', '#ec4899', '#14b8a6']

type Period = '7d' | '30d' | '90d'

export function ReportsPage() {
  const [period, setPeriod] = useState<Period>('30d')

  const { data: ordersData } = useQuery({
    queryKey: ['orders', 'reports'],
    queryFn: () => ordersApi.list({ per_page: 500 }),
  })

  const { data: productsData } = useQuery({
    queryKey: ['products', 'reports'],
    queryFn: () => productsApi.list({ per_page: 500 }),
  })

  const orders = ordersData?.data || []
  const products = productsData?.data || []

  const periodDays = period === '7d' ? 7 : period === '30d' ? 30 : 90

  const filteredOrders = useMemo(() => {
    const cutoff = new Date()
    cutoff.setDate(cutoff.getDate() - periodDays)
    return orders.filter((o: Order) => new Date(o.created_at) >= cutoff)
  }, [orders, periodDays])

  // Sales by day
  const salesByDay = useMemo(() => {
    const map = new Map<string, number>()
    for (let i = 0; i < periodDays; i++) {
      const d = new Date()
      d.setDate(d.getDate() - (periodDays - 1 - i))
      const key = d.toISOString().slice(0, 10)
      map.set(key, 0)
    }
    filteredOrders.forEach((o: Order) => {
      if (o.status !== 'cancelado') {
        const key = o.created_at.slice(0, 10)
        map.set(key, (map.get(key) || 0) + o.total_cents)
      }
    })
    return Array.from(map.entries()).map(([date, total]) => ({
      date: new Date(date).toLocaleDateString('pt-BR', { day: '2-digit', month: '2-digit' }),
      vendas: total / 100,
    }))
  }, [filteredOrders, periodDays])

  // Top products
  const topProducts = useMemo(() => {
    const map = new Map<string, { name: string; quantity: number; revenue: number }>()
    filteredOrders.forEach((o: Order) => {
      if (o.items) {
        o.items.forEach((item) => {
          const existing = map.get(item.product_id) || { name: item.product_name, quantity: 0, revenue: 0 }
          existing.quantity += item.quantity
          existing.revenue += item.unit_price_cents * item.quantity
          map.set(item.product_id, existing)
        })
      }
    })
    return Array.from(map.values())
      .sort((a, b) => b.revenue - a.revenue)
      .slice(0, 8)
      .map((p) => ({ ...p, revenue: p.revenue / 100 }))
  }, [filteredOrders])

  // Status distribution
  const statusDistribution = useMemo(() => {
    const map = new Map<string, number>()
    filteredOrders.forEach((o: Order) => {
      map.set(o.status, (map.get(o.status) || 0) + 1)
    })
    return Array.from(map.entries()).map(([status, count]) => ({ status, count }))
  }, [filteredOrders])

  // Summary
  const totalRevenue = filteredOrders
    .filter((o: Order) => o.status !== 'cancelado')
    .reduce((s: number, o: Order) => s + o.total_cents, 0)
  const totalOrders = filteredOrders.length
  const completedOrders = filteredOrders.filter((o: Order) => o.status === 'entregue').length
  const conversionRate = totalOrders > 0 ? ((completedOrders / totalOrders) * 100).toFixed(1) : '0'
  const avgTicket = totalOrders > 0 ? totalRevenue / totalOrders : 0

  return (
    <div className="space-y-6">
      {/* Period selector */}
      <div className="flex items-center gap-4">
        <Select
          value={period}
          onChange={(e) => setPeriod(e.target.value as Period)}
          options={[
            { value: '7d', label: 'Ultimos 7 dias' },
            { value: '30d', label: 'Ultimos 30 dias' },
            { value: '90d', label: 'Ultimos 90 dias' },
          ]}
        />
      </div>

      {/* Summary cards */}
      <div className="grid grid-cols-2 lg:grid-cols-4 gap-4">
        <Card>
          <p className="text-sm text-muted">Receita Total</p>
          <p className="text-xl font-bold text-gray-900 mt-1">{formatCents(totalRevenue)}</p>
        </Card>
        <Card>
          <p className="text-sm text-muted">Total Pedidos</p>
          <p className="text-xl font-bold text-gray-900 mt-1">{totalOrders}</p>
        </Card>
        <Card>
          <p className="text-sm text-muted">Ticket Medio</p>
          <p className="text-xl font-bold text-gray-900 mt-1">{formatCents(avgTicket)}</p>
        </Card>
        <Card>
          <p className="text-sm text-muted">Taxa Conversao</p>
          <p className="text-xl font-bold text-gray-900 mt-1">{conversionRate}%</p>
        </Card>
      </div>

      {/* Sales chart */}
      <Card>
        <CardTitle>Vendas por Periodo</CardTitle>
        <div className="h-72 mt-4">
          <ResponsiveContainer width="100%" height="100%">
            <LineChart data={salesByDay}>
              <CartesianGrid strokeDasharray="3 3" stroke="#e5e7eb" />
              <XAxis dataKey="date" tick={{ fontSize: 11 }} />
              <YAxis tick={{ fontSize: 11 }} />
              <Tooltip
                formatter={(v: number) =>
                  new Intl.NumberFormat('pt-BR', { style: 'currency', currency: 'BRL' }).format(v)
                }
              />
              <Line type="monotone" dataKey="vendas" stroke="#6366f1" strokeWidth={2} dot={false} />
            </LineChart>
          </ResponsiveContainer>
        </div>
      </Card>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Top products */}
        <Card>
          <CardTitle>Top Produtos</CardTitle>
          <div className="h-72 mt-4">
            {topProducts.length === 0 ? (
              <p className="text-sm text-muted text-center py-12">Sem dados no periodo</p>
            ) : (
              <ResponsiveContainer width="100%" height="100%">
                <BarChart data={topProducts} layout="vertical">
                  <CartesianGrid strokeDasharray="3 3" stroke="#e5e7eb" />
                  <XAxis type="number" tick={{ fontSize: 11 }} />
                  <YAxis type="category" dataKey="name" width={120} tick={{ fontSize: 11 }} />
                  <Tooltip
                    formatter={(v: number) =>
                      new Intl.NumberFormat('pt-BR', { style: 'currency', currency: 'BRL' }).format(v)
                    }
                  />
                  <Bar dataKey="revenue" fill="#6366f1" radius={[0, 4, 4, 0]} />
                </BarChart>
              </ResponsiveContainer>
            )}
          </div>
        </Card>

        {/* Status distribution */}
        <Card>
          <CardTitle>Distribuicao de Pedidos</CardTitle>
          <div className="h-72 mt-4">
            {statusDistribution.length === 0 ? (
              <p className="text-sm text-muted text-center py-12">Sem dados no periodo</p>
            ) : (
              <ResponsiveContainer width="100%" height="100%">
                <PieChart>
                  <Pie
                    data={statusDistribution}
                    dataKey="count"
                    nameKey="status"
                    cx="50%"
                    cy="50%"
                    outerRadius={90}
                    label={({ status, count }) => `${status} (${count})`}
                  >
                    {statusDistribution.map((_, i) => (
                      <Cell key={i} fill={COLORS[i % COLORS.length]} />
                    ))}
                  </Pie>
                  <Tooltip />
                  <Legend />
                </PieChart>
              </ResponsiveContainer>
            )}
          </div>
        </Card>
      </div>

      {/* Product stats */}
      <Card>
        <CardTitle>Catalogo</CardTitle>
        <div className="grid grid-cols-2 sm:grid-cols-4 gap-4 mt-4">
          <div>
            <p className="text-2xl font-bold">{products.length}</p>
            <p className="text-xs text-muted">Total produtos</p>
          </div>
          <div>
            <p className="text-2xl font-bold">{products.filter((p: Product) => p.is_active).length}</p>
            <p className="text-xs text-muted">Ativos</p>
          </div>
          <div>
            <p className="text-2xl font-bold">
              {products.filter((p: Product) => p.stock_quantity <= p.stock_alert_threshold).length}
            </p>
            <p className="text-xs text-muted">Estoque baixo</p>
          </div>
          <div>
            <p className="text-2xl font-bold">
              {products.filter((p: Product) => p.stock_quantity === 0).length}
            </p>
            <p className="text-xs text-muted">Sem estoque</p>
          </div>
        </div>
      </Card>
    </div>
  )
}
