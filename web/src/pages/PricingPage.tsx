import { useEffect, useState } from 'react';
import { Check, X } from 'lucide-react';

interface Plano {
  id: string;
  slug: string;
  name: string;
  price: string;
  price_cents: number;
  max_products: number;
  custom_domain: boolean;
  support_level: string;
  features: Record<string, boolean>;
}

export function PricingPage() {
  const [planos, setPlanos] = useState<Plano[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const fetchPlanos = async () => {
      try {
        const response = await fetch('/api/v1/public/plans');
        const data = await response.json();
        setPlanos(data.data || []);
      } catch (error) {
        console.error('Erro ao buscar planos:', error);
      } finally {
        setLoading(false);
      }
    };

    fetchPlanos();
  }, []);

  const featuresList = [
    { key: 'frete', label: 'Cálculo de frete integrado' },
    { key: 'cupons', label: 'Cupons de desconto' },
    { key: 'whatsapp', label: 'Integração WhatsApp' },
    { key: 'crm', label: 'Sistema de CRM' },
  ];

  if (loading) {
    return <div className="min-h-screen flex items-center justify-center">Carregando planos...</div>;
  }

  return (
    <div className="min-h-screen bg-gray-50 py-12 px-4 sm:px-6 lg:px-8">
      <div className="max-w-7xl mx-auto">
        {/* Header */}
        <div className="text-center mb-12">
          <h1 className="text-4xl font-bold text-gray-900 mb-4">
            Planos para sua loja virtual
          </h1>
          <p className="text-xl text-gray-600">
            Escolha o plano perfeito para começar a vender online
          </p>
        </div>

        {/* Pricing Grid */}
        <div className="grid md:grid-cols-3 gap-8 mb-12">
          {planos.map((plano) => {
            const isPopular = plano.slug === 'starter';
            const isFree = plano.slug === 'free';
            const isPro = plano.slug === 'pro';

            return (
              <div
                key={plano.id}
                className={`relative rounded-lg shadow-lg overflow-hidden transition-transform hover:scale-105 ${
                  isPopular ? 'ring-2 ring-blue-500 md:scale-105' : ''
                } ${isFree ? 'bg-gray-100' : 'bg-white'} ${
                  isPro ? 'bg-gradient-to-br from-blue-50 to-blue-100' : ''
                }`}
              >
                {isPopular && (
                  <div className="absolute top-0 right-0 bg-blue-500 text-white px-4 py-1 rounded-bl-lg text-sm font-bold">
                    Mais Popular
                  </div>
                )}

                <div className="p-8">
                  {/* Plan Name */}
                  <h3 className="text-2xl font-bold text-gray-900 mb-2">
                    {plano.name}
                  </h3>

                  {/* Price */}
                  <div className="mb-6">
                    <span className="text-5xl font-bold text-gray-900">
                      {isFree ? 'Grátis' : `${plano.price}`}
                    </span>
                    {!isFree && (
                      <span className="text-gray-600 ml-2">/mês</span>
                    )}
                  </div>

                  {/* CTA Button */}
                  <button
                    className={`w-full py-2 px-4 rounded-lg font-semibold mb-8 transition-colors ${
                      isPopular
                        ? 'bg-blue-500 text-white hover:bg-blue-600'
                        : isFree
                          ? 'bg-gray-300 text-gray-900 hover:bg-gray-400'
                          : 'bg-blue-600 text-white hover:bg-blue-700'
                    }`}
                  >
                    {isFree ? 'Começar Grátis' : 'Começar'}
                  </button>

                  {/* Main Features */}
                  <div className="mb-8 pb-8 border-b border-gray-200">
                    <div className="text-sm font-semibold text-gray-700 mb-3">
                      Até {plano.max_products.toLocaleString()} produtos
                    </div>
                    {plano.custom_domain && (
                      <div className="text-sm font-semibold text-gray-700">
                        ✓ Domínio customizado
                      </div>
                    )}
                  </div>

                  {/* Support Level */}
                  <div className="mb-6 p-3 bg-gray-100 rounded">
                    <p className="text-sm text-gray-700">
                      <span className="font-semibold">Suporte:</span>{' '}
                      <span className="capitalize">{plano.support_level}</span>
                    </p>
                  </div>

                  {/* Features List */}
                  <div className="space-y-3">
                    {featuresList.map((feature) => {
                      const hasFeature = plano.features?.[feature.key] || false;
                      return (
                        <div
                          key={feature.key}
                          className="flex items-start"
                        >
                          {hasFeature ? (
                            <Check className="w-5 h-5 text-green-500 mr-3 flex-shrink-0 mt-0.5" />
                          ) : (
                            <X className="w-5 h-5 text-gray-300 mr-3 flex-shrink-0 mt-0.5" />
                          )}
                          <span
                            className={`text-sm ${
                              hasFeature ? 'text-gray-700' : 'text-gray-400'
                            }`}
                          >
                            {feature.label}
                          </span>
                        </div>
                      );
                    })}
                  </div>
                </div>
              </div>
            );
          })}
        </div>

        {/* Comparison Table */}
        <div className="overflow-x-auto">
          <table className="w-full border-collapse border border-gray-300 bg-white rounded-lg overflow-hidden">
            <thead className="bg-gray-100">
              <tr>
                <th className="border border-gray-300 px-4 py-3 text-left font-semibold">
                  Funcionalidade
                </th>
                {planos.map((plano) => (
                  <th
                    key={plano.id}
                    className="border border-gray-300 px-4 py-3 text-center font-semibold"
                  >
                    {plano.name}
                  </th>
                ))}
              </tr>
            </thead>
            <tbody>
              {featuresList.map((feature) => (
                <tr key={feature.key} className="hover:bg-gray-50">
                  <td className="border border-gray-300 px-4 py-3 font-medium">
                    {feature.label}
                  </td>
                  {planos.map((plano) => (
                    <td
                      key={`${plano.id}-${feature.key}`}
                      className="border border-gray-300 px-4 py-3 text-center"
                    >
                      {plano.features?.[feature.key] ? (
                        <Check className="w-5 h-5 text-green-500 mx-auto" />
                      ) : (
                        <X className="w-5 h-5 text-gray-300 mx-auto" />
                      )}
                    </td>
                  ))}
                </tr>
              ))}
            </tbody>
          </table>
        </div>

        {/* FAQ Section */}
        <div className="mt-16 max-w-2xl mx-auto">
          <h2 className="text-2xl font-bold text-center mb-8">Dúvidas Frequentes</h2>
          <div className="space-y-6">
            <div>
              <h3 className="font-semibold text-gray-900 mb-2">
                Posso mudar de plano a qualquer momento?
              </h3>
              <p className="text-gray-600">
                Sim! Você pode fazer upgrade ou downgrade a qualquer momento. As mudanças entram em vigor no próximo período de cobrança.
              </p>
            </div>
            <div>
              <h3 className="font-semibold text-gray-900 mb-2">
                O plano Free tem alguma restrição?
              </h3>
              <p className="text-gray-600">
                O plano Free permite até 10 produtos e suporte da comunidade. Para mais funcionalidades, escolha Starter ou Pro.
              </p>
            </div>
            <div>
              <h3 className="font-semibold text-gray-900 mb-2">
                Qual é o diferencial do plano Starter?
              </h3>
              <p className="text-gray-600">
                O plano Starter (R$ 79/mês) oferece 200 produtos, domínio customizado e suporte por email. É perfeito para pequenos lojistas que querem crescer.
              </p>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
