import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { modulesApi } from '@/api/endpoints'
import { Card } from '@/components/ui/Card'
import { Badge } from '@/components/ui/Badge'
import { Modal } from '@/components/ui/Modal'
import { Button } from '@/components/ui/Button'
import {
  ShoppingBag,
  Warehouse,
  Users,
  MessageSquare,
  CheckCircle2,
  XCircle,
  AlertTriangle,
} from 'lucide-react'

const MODULE_META: Record<string, {
  label: string
  description: string
  icon: React.ReactNode
  warning?: string
}> = {
  storefront: {
    label: 'Loja Virtual',
    description: 'Catálogo público, checkout e pedidos. Desativar oculta a loja ao público mas preserva todos os dados.',
    icon: <ShoppingBag size={20} />,
    warning: 'Desativar a loja virtual tornará o catálogo inacessível ao público. Pedidos e produtos existentes são preservados.',
  },
  inventory: {
    label: 'Estoque',
    description: 'Gestão centralizada de inventário multi-loja, SKUs, alocações e alertas de reposição.',
    icon: <Warehouse size={20} />,
  },
  crm: {
    label: 'CRM',
    description: 'Gestão de clientes, histórico de compras e segmentação.',
    icon: <Users size={20} />,
  },
  whatsapp: {
    label: 'WhatsApp',
    description: 'Notificações automáticas de pedidos via WhatsApp Business.',
    icon: <MessageSquare size={20} />,
  },
}

export function ModulesPage() {
  const qc = useQueryClient()
  const [confirmModule, setConfirmModule] = useState<{ name: string; enabling: boolean } | null>(null)

  const { data, isLoading } = useQuery({
    queryKey: ['modules'],
    queryFn: () => modulesApi.list(),
  })

  const updateMut = useMutation({
    mutationFn: ({ module, enabled }: { module: string; enabled: boolean }) =>
      modulesApi.update(module, enabled),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['modules'] })
      setConfirmModule(null)
    },
  })

  const modules = data?.modules || []

  function handleToggle(moduleName: string, currentEnabled: boolean) {
    const meta = MODULE_META[moduleName]
    // Se tem warning e está desativando, pede confirmação
    if (!currentEnabled === false && meta?.warning) {
      setConfirmModule({ name: moduleName, enabling: false })
      return
    }
    updateMut.mutate({ module: moduleName, enabled: !currentEnabled })
  }

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-gray-900">Módulos</h1>
        <p className="text-sm text-gray-500 mt-1">
          Ative ou desative funcionalidades da sua loja. Dados nunca são excluídos ao desativar um módulo.
        </p>
      </div>

      {isLoading ? (
        <Card><div className="p-8 text-center text-gray-400 text-sm">Carregando módulos...</div></Card>
      ) : (
        <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
          {modules.map(mod => {
            const meta = MODULE_META[mod.module] || {
              label: mod.module,
              description: '',
              icon: <CheckCircle2 size={20} />,
            }

            return (
              <Card key={mod.module}>
                <div className="p-5">
                  <div className="flex items-start justify-between gap-3">
                    <div className="flex items-start gap-3">
                      <div className={`mt-0.5 shrink-0 ${mod.enabled ? 'text-indigo-600' : 'text-gray-300'}`}>
                        {meta.icon}
                      </div>
                      <div>
                        <div className="flex items-center gap-2">
                          <h3 className="text-sm font-semibold text-gray-900">{meta.label}</h3>
                          {mod.enabled ? (
                            <Badge variant="success">Ativo</Badge>
                          ) : (
                            <Badge variant="default">Inativo</Badge>
                          )}
                        </div>
                        <p className="text-xs text-gray-500 mt-1">{meta.description}</p>
                      </div>
                    </div>

                    {/* Toggle switch */}
                    <button
                      onClick={() => handleToggle(mod.module, mod.enabled)}
                      disabled={updateMut.isPending}
                      className={`relative inline-flex h-6 w-11 items-center rounded-full transition-colors focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:ring-offset-2 shrink-0 ${
                        mod.enabled ? 'bg-indigo-600' : 'bg-gray-200'
                      } disabled:opacity-50`}
                      aria-label={`${mod.enabled ? 'Desativar' : 'Ativar'} módulo ${meta.label}`}
                    >
                      <span
                        className={`inline-block h-4 w-4 transform rounded-full bg-white shadow transition-transform ${
                          mod.enabled ? 'translate-x-6' : 'translate-x-1'
                        }`}
                      />
                    </button>
                  </div>

                  <div className="mt-3 pt-3 border-t border-gray-50 flex items-center justify-between">
                    <span className="text-xs text-gray-400">
                      {mod.enabled ? (
                        <span className="flex items-center gap-1 text-green-600">
                          <CheckCircle2 size={12} /> Funcionalidade disponível
                        </span>
                      ) : (
                        <span className="flex items-center gap-1 text-gray-400">
                          <XCircle size={12} /> Funcionalidade desativada
                        </span>
                      )}
                    </span>
                    <span className="text-xs text-gray-300">
                      {new Date(mod.updated_at).toLocaleDateString('pt-BR')}
                    </span>
                  </div>
                </div>
              </Card>
            )
          })}
        </div>
      )}

      {/* Aviso sobre dados */}
      <div className="flex gap-3 p-4 bg-blue-50 rounded-xl border border-blue-100">
        <AlertTriangle size={16} className="text-blue-500 mt-0.5 shrink-0" />
        <div>
          <p className="text-sm font-medium text-blue-800">Dados preservados</p>
          <p className="text-xs text-blue-600 mt-0.5">
            Desativar um módulo apenas oculta a funcionalidade. Todos os dados (produtos, pedidos, clientes) são mantidos e podem ser restaurados reativando o módulo.
          </p>
        </div>
      </div>

      {/* Confirm Modal para desativação com warning */}
      <Modal
        open={!!confirmModule}
        onClose={() => setConfirmModule(null)}
        title="Confirmar desativação"
      >
        {confirmModule && (
          <div className="space-y-4">
            <div className="flex gap-3 p-4 bg-amber-50 rounded-lg border border-amber-200">
              <AlertTriangle size={18} className="text-amber-500 mt-0.5 shrink-0" />
              <p className="text-sm text-amber-800">
                {MODULE_META[confirmModule.name]?.warning}
              </p>
            </div>
            <div className="flex justify-end gap-2">
              <Button variant="outline" onClick={() => setConfirmModule(null)}>Cancelar</Button>
              <Button
                variant="danger"
                onClick={() => updateMut.mutate({ module: confirmModule.name, enabled: false })}
                disabled={updateMut.isPending}
              >
                {updateMut.isPending ? 'Desativando...' : 'Sim, desativar'}
              </Button>
            </div>
          </div>
        )}
      </Modal>
    </div>
  )
}
