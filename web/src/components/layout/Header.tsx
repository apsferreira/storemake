import { Link } from 'react-router-dom';
import { useStore } from '../../contexts/StoreContext';
import { useCart } from '../../contexts/CartContext';
import { useTheme } from '../../contexts/ThemeContext';

export function Header() {
  const { store } = useStore();
  const { toggleCart, itemCount } = useCart();
  const theme = useTheme();

  return (
    <header className={`${theme.headerBg} sticky top-0 z-40`}>
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="flex items-center justify-between h-16">
          {/* Logo / Nome da Loja */}
          <Link to="/" className="flex items-center gap-2">
            {store.logo_url && (
              <img
                src={store.logo_url}
                alt={store.name}
                className="h-8 w-8 object-contain"
              />
            )}
            <span className="text-lg font-bold truncate max-w-[200px]">
              {store.name}
            </span>
          </Link>

          {/* Carrinho */}
          <button
            onClick={toggleCart}
            className="relative p-2 rounded-lg hover:bg-black/5 transition-colors"
            aria-label="Abrir carrinho"
          >
            <svg
              xmlns="http://www.w3.org/2000/svg"
              fill="none"
              viewBox="0 0 24 24"
              strokeWidth={1.5}
              stroke="currentColor"
              className="w-6 h-6"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                d="M15.75 10.5V6a3.75 3.75 0 10-7.5 0v4.5m11.356-1.993l1.263 12c.07.665-.45 1.243-1.119 1.243H4.25a1.125 1.125 0 01-1.12-1.243l1.264-12A1.125 1.125 0 015.513 7.5h12.974c.576 0 1.059.435 1.119 1.007zM8.625 10.5a.375.375 0 11-.75 0 .375.375 0 01.75 0zm7.5 0a.375.375 0 11-.75 0 .375.375 0 01.75 0z"
              />
            </svg>
            {itemCount > 0 && (
              <span
                className="absolute -top-1 -right-1 min-w-[20px] h-5 flex items-center justify-center rounded-full text-xs font-bold text-white px-1"
                style={{ backgroundColor: theme.primaryColor }}
              >
                {itemCount}
              </span>
            )}
          </button>
        </div>
      </div>
    </header>
  );
}
