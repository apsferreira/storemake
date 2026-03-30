import { useState } from 'react';
import type { CartItem } from '../../types';
import { useCart } from '../../contexts/CartContext';
import { formatCurrency } from '../../utils/format';

interface CartItemRowProps {
  item: CartItem;
}

export function CartItemRow({ item }: CartItemRowProps) {
  const { updateItem, removeItem, loading } = useCart();
  const [removing, setRemoving] = useState(false);

  const handleQuantityChange = async (newQty: number) => {
    if (newQty < 1 || newQty > 999) return;
    await updateItem(item.id, newQty);
  };

  const handleRemove = async () => {
    setRemoving(true);
    await removeItem(item.id);
  };

  return (
    <div className={`flex gap-3 pb-4 border-b last:border-b-0 ${removing ? 'opacity-50' : ''}`}>
      <div className="flex-1 min-w-0">
        <h4 className="font-medium text-sm truncate">{item.product_name}</h4>
        {item.variant_name && (
          <p className="text-xs text-gray-500">{item.variant_name}</p>
        )}
        <p className="text-sm font-bold mt-1">
          {formatCurrency(item.unit_price_cents)}
        </p>

        {/* Qty controls */}
        <div className="flex items-center gap-2 mt-2">
          <button
            onClick={() => handleQuantityChange(item.quantity - 1)}
            disabled={loading || item.quantity <= 1}
            className="w-7 h-7 flex items-center justify-center rounded border hover:bg-gray-100 disabled:opacity-30 text-sm"
            aria-label="Diminuir quantidade"
          >
            -
          </button>
          <span className="text-sm font-medium w-6 text-center">{item.quantity}</span>
          <button
            onClick={() => handleQuantityChange(item.quantity + 1)}
            disabled={loading || item.quantity >= 999}
            className="w-7 h-7 flex items-center justify-center rounded border hover:bg-gray-100 disabled:opacity-30 text-sm"
            aria-label="Aumentar quantidade"
          >
            +
          </button>
        </div>
      </div>

      {/* Line total + remove */}
      <div className="flex flex-col items-end justify-between">
        <span className="text-sm font-bold">
          {formatCurrency(item.unit_price_cents * item.quantity)}
        </span>
        <button
          onClick={handleRemove}
          disabled={loading}
          className="text-red-400 hover:text-red-600 transition-colors p-1"
          aria-label="Remover item"
        >
          <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={1.5} stroke="currentColor" className="w-4 h-4">
            <path strokeLinecap="round" strokeLinejoin="round" d="M14.74 9l-.346 9m-4.788 0L9.26 9m9.968-3.21c.342.052.682.107 1.022.166m-1.022-.165L18.16 19.673a2.25 2.25 0 01-2.244 2.077H8.084a2.25 2.25 0 01-2.244-2.077L4.772 5.79m14.456 0a48.108 48.108 0 00-3.478-.397m-12 .562c.34-.059.68-.114 1.022-.165m0 0a48.11 48.11 0 013.478-.397m7.5 0v-.916c0-1.18-.91-2.164-2.09-2.201a51.964 51.964 0 00-3.32 0c-1.18.037-2.09 1.022-2.09 2.201v.916m7.5 0a48.667 48.667 0 00-7.5 0" />
          </svg>
        </button>
      </div>
    </div>
  );
}
