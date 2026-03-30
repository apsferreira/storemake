import { createContext, useContext, type ReactNode } from 'react';
import { getTheme, type ThemeConfig } from '../templates/themes';
import { useStore } from './StoreContext';

const ThemeContext = createContext<ThemeConfig>(getTheme('moda'));

export function ThemeProvider({ children }: { children: ReactNode }) {
  const { store } = useStore();
  const theme = getTheme(store.template);

  return (
    <ThemeContext.Provider value={theme}>
      <div
        style={{
          fontFamily: theme.fontFamily,
          backgroundColor: theme.bgColor,
          color: theme.textColor,
          minHeight: '100vh',
        }}
      >
        {children}
      </div>
    </ThemeContext.Provider>
  );
}

export function useTheme() {
  return useContext(ThemeContext);
}
