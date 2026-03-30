import { Link, useLocation } from 'react-router-dom';
import { useTheme } from '../contexts/ThemeContext';
import { SEOHead } from '../components/seo/SEOHead';
import { formatCurrency } from '../utils/format';
import type { Order } from '../types';

export function OrderConfirmedPage() {
  const location = useLocation();
  const theme = useTheme();
  const order = (location.state as { order?: Order })?.order;

  return (
    <>
      <SEOHead title="Pedido Confirmado" />

      <main className="max-w-lg mx-auto px-4 py-12 text-center">
        <div className="mb-6">
          <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={1.5} stroke="currentColor" className="w-20 h-20 mx-auto text-green-500">
            <path strokeLinecap="round" strokeLinejoin="round" d="M9 12.75L11.25 15 15 9.75M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
          </svg>
        </div>

        <h1 className="text-2xl font-bold mb-2">Pedido confirmado!</h1>
        <p className="text-gray-500 mb-6">
          Obrigado pela sua compra. Voce recebera atualizacoes por email.
        </p>

        {order && (
          <div className="bg-gray-50 rounded-lg p-4 text-left mb-6 space-y-2 text-sm">
            <div className="flex justify-between">
              <span className="text-gray-500">Pedido</span>
              <span className="font-mono text-xs">{order.id.slice(0, 8)}...</span>
            </div>
            <div className="flex justify-between">
              <span className="text-gray-500">Total</span>
              <span className="font-bold">{formatCurrency(order.total_cents)}</span>
            </div>
            <div className="flex justify-between">
              <span className="text-gray-500">Pagamento</span>
              <span className="capitalize">{order.payment_method}</span>
            </div>
            <div className="flex justify-between">
              <span className="text-gray-500">Status</span>
              <span className="capitalize">{order.status}</span>
            </div>
          </div>
        )}

        <Link
          to="/"
          className={`inline-block px-6 py-3 rounded-lg font-bold transition-colors ${theme.buttonStyle}`}
        >
          Continuar comprando
        </Link>
      </main>
    </>
  );
}
