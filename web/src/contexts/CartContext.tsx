import { createContext, useContext, useState, useCallback, type ReactNode } from 'react';
import type { CartResponse } from '../types';
import * as cartApi from '../api/cart';
import { useStore } from './StoreContext';

interface CartContextValue {
  cart: CartResponse | null;
  loading: boolean;
  isOpen: boolean;
  openCart: () => void;
  closeCart: () => void;
  toggleCart: () => void;
  refresh: () => Promise<void>;
  addItem: (productId: string, quantity: number, variantId?: string) => Promise<void>;
  updateItem: (itemId: string, quantity: number) => Promise<void>;
  removeItem: (itemId: string) => Promise<void>;
  itemCount: number;
}

const CartContext = createContext<CartContextValue>({
  cart: null,
  loading: false,
  isOpen: false,
  openCart: () => {},
  closeCart: () => {},
  toggleCart: () => {},
  refresh: async () => {},
  addItem: async () => {},
  updateItem: async () => {},
  removeItem: async () => {},
  itemCount: 0,
});

export function CartProvider({ children }: { children: ReactNode }) {
  const { store } = useStore();
  const [cart, setCart] = useState<CartResponse | null>(null);
  const [loading, setLoading] = useState(false);
  const [isOpen, setIsOpen] = useState(false);

  const refresh = useCallback(async () => {
    if (!store.id) return;
    setLoading(true);
    try {
      const data = await cartApi.fetchCart(store.id);
      setCart(data);
    } catch {
      // cart not found is normal
    } finally {
      setLoading(false);
    }
  }, [store.id]);

  const addItem = useCallback(
    async (productId: string, quantity: number, variantId?: string) => {
      if (!store.id) return;
      setLoading(true);
      try {
        await cartApi.addToCart(store.id, productId, quantity, variantId);
        await refresh();
        setIsOpen(true);
      } finally {
        setLoading(false);
      }
    },
    [store.id, refresh],
  );

  const updateItem = useCallback(
    async (itemId: string, quantity: number) => {
      if (!store.id) return;
      setLoading(true);
      try {
        await cartApi.updateCartItem(store.id, itemId, quantity);
        await refresh();
      } finally {
        setLoading(false);
      }
    },
    [store.id, refresh],
  );

  const removeItem = useCallback(
    async (itemId: string) => {
      if (!store.id) return;
      setLoading(true);
      try {
        await cartApi.removeCartItem(store.id, itemId);
        await refresh();
      } finally {
        setLoading(false);
      }
    },
    [store.id, refresh],
  );

  const itemCount = cart?.item_count ?? 0;

  return (
    <CartContext.Provider
      value={{
        cart,
        loading,
        isOpen,
        openCart: () => setIsOpen(true),
        closeCart: () => setIsOpen(false),
        toggleCart: () => setIsOpen((o) => !o),
        refresh,
        addItem,
        updateItem,
        removeItem,
        itemCount,
      }}
    >
      {children}
    </CartContext.Provider>
  );
}

export function useCart() {
  return useContext(CartContext);
}
