import { useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { useCart } from '../contexts/CartContext';
import { SEOHead } from '../components/seo/SEOHead';
import { CheckoutForm } from '../components/checkout/CheckoutForm';
import { formatCurrency } from '../utils/format';

export function CheckoutPage() {
  const { cart, refresh } = useCart();
  const navigate = useNavigate();

  useEffect(() => {
    refresh();
  }, []);

  const items = cart?.cart.items ?? [];

  return (
    <>
      <SEOHead title="Checkout" />

      <main className="max-w-3xl mx-auto px-4 sm:px-6 py-6 sm:py-8">
        <h1 className="text-2xl font-bold mb-6">Finalizar Compra</h1>

        {items.length === 0 ? (
          <div className="text-center py-16">
            <p className="text-gray-500 text-lg mb-4">Seu carrinho esta vazio</p>
            <button
              onClick={() => navigate('/')}
              className="text-blue-600 hover:underline"
            >
              Voltar ao catalogo
            </button>
          </div>
        ) : (
          <div className="grid grid-cols-1 lg:grid-cols-5 gap-8">
            {/* Form */}
            <div className="lg:col-span-3">
              <CheckoutForm />
            </div>

            {/* Order summary sidebar (desktop) */}
            <div className="lg:col-span-2 hidden lg:block">
              <div className="sticky top-24 bg-gray-50 rounded-lg p-4">
                <h3 className="font-bold mb-3">Resumo do Pedido</h3>
                <div className="space-y-3 max-h-80 overflow-y-auto">
                  {items.map((item) => (
                    <div key={item.id} className="flex justify-between text-sm">
                      <div className="flex-1 min-w-0">
                        <p className="truncate">{item.product_name}</p>
                        {item.variant_name && (
                          <p className="text-xs text-gray-500">{item.variant_name}</p>
                        )}
                        <p className="text-xs text-gray-400">Qtd: {item.quantity}</p>
                      </div>
                      <span className="font-medium ml-2">
                        {formatCurrency(item.unit_price_cents * item.quantity)}
                      </span>
                    </div>
                  ))}
                </div>
                <div className="border-t mt-3 pt-3 flex justify-between font-bold">
                  <span>Subtotal</span>
                  <span>{formatCurrency(cart?.subtotal_cents ?? 0)}</span>
                </div>
              </div>
            </div>
          </div>
        )}
      </main>
    </>
  );
}
