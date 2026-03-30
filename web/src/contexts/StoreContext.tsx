import { createContext, useContext, useState, useEffect, type ReactNode } from 'react';
import type { StoreConfig, TemplateTheme } from '../types';
import { getStoreId } from '../utils/format';

interface StoreContextValue {
  store: StoreConfig;
  setStore: (s: StoreConfig) => void;
}

const defaultStore: StoreConfig = {
  id: '',
  name: 'Minha Loja',
  whatsapp: '',
  template: 'moda',
};

const StoreContext = createContext<StoreContextValue>({
  store: defaultStore,
  setStore: () => {},
});

export function StoreProvider({ children }: { children: ReactNode }) {
  const [store, setStore] = useState<StoreConfig>(() => {
    const saved = localStorage.getItem('sm_store_config');
    if (saved) {
      try {
        return JSON.parse(saved);
      } catch {
        // ignore
      }
    }
    const storeId = getStoreId();
    return { ...defaultStore, id: storeId };
  });

  useEffect(() => {
    localStorage.setItem('sm_store_config', JSON.stringify(store));
  }, [store]);

  // Read store config from URL params on mount
  useEffect(() => {
    const params = new URLSearchParams(window.location.search);
    const storeId = params.get('store');
    const storeName = params.get('store_name');
    const whatsapp = params.get('whatsapp');
    const template = params.get('template') as TemplateTheme | null;

    if (storeId) {
      setStore((prev) => ({
        ...prev,
        id: storeId,
        ...(storeName ? { name: storeName } : {}),
        ...(whatsapp ? { whatsapp } : {}),
        ...(template && ['moda', 'semi-joias', 'festas', 'artesanato'].includes(template)
          ? { template }
          : {}),
      }));
    }
  }, []);

  return (
    <StoreContext.Provider value={{ store, setStore }}>
      {children}
    </StoreContext.Provider>
  );
}

export function useStore() {
  return useContext(StoreContext);
}
