import { Link } from 'react-router-dom';
import type { Product } from '../../types';
import { formatCurrency } from '../../utils/format';
import { useTheme } from '../../contexts/ThemeContext';

interface ProductCardProps {
  product: Product;
}

export function ProductCard({ product }: ProductCardProps) {
  const theme = useTheme();
  const hasDiscount = product.compare_price_cents > product.price_cents;
  const mainPhoto = product.photos?.[0]?.url;

  return (
    <Link
      to={`/produto/${product.slug}`}
      state={{ product }}
      className="group block overflow-hidden bg-white shadow-sm hover:shadow-md transition-shadow"
      style={{ borderRadius: theme.borderRadius }}
    >
      {/* Imagem */}
      <div className="aspect-square bg-gray-100 overflow-hidden relative">
        {mainPhoto ? (
          <img
            src={mainPhoto}
            alt={product.name}
            className="w-full h-full object-cover group-hover:scale-105 transition-transform duration-300"
            loading="lazy"
          />
        ) : (
          <div className="w-full h-full flex items-center justify-center text-gray-300">
            <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={1} stroke="currentColor" className="w-16 h-16">
              <path strokeLinecap="round" strokeLinejoin="round" d="M2.25 15.75l5.159-5.159a2.25 2.25 0 013.182 0l5.159 5.159m-1.5-1.5l1.409-1.409a2.25 2.25 0 013.182 0l2.909 2.909M3.75 21h16.5A2.25 2.25 0 0022.5 18.75V5.25A2.25 2.25 0 0020.25 3H3.75A2.25 2.25 0 001.5 5.25v13.5A2.25 2.25 0 003.75 21z" />
            </svg>
          </div>
        )}
        {hasDiscount && (
          <span className="absolute top-2 left-2 bg-red-500 text-white text-xs font-bold px-2 py-1 rounded">
            -{Math.round(((product.compare_price_cents - product.price_cents) / product.compare_price_cents) * 100)}%
          </span>
        )}
        {product.stock_quantity === 0 && (
          <div className="absolute inset-0 bg-black/40 flex items-center justify-center">
            <span className="bg-white text-gray-800 font-bold px-4 py-2 rounded">Esgotado</span>
          </div>
        )}
      </div>

      {/* Info */}
      <div className="p-3 sm:p-4">
        <h3 className="font-medium text-sm sm:text-base line-clamp-2 mb-1">
          {product.name}
        </h3>
        <div className="flex items-baseline gap-2">
          <span className="text-lg font-bold" style={{ color: theme.primaryColor }}>
            {formatCurrency(product.price_cents)}
          </span>
          {hasDiscount && (
            <span className="text-sm text-gray-400 line-through">
              {formatCurrency(product.compare_price_cents)}
            </span>
          )}
        </div>
        {product.variations && product.variations.length > 0 && (
          <p className="text-xs text-gray-400 mt-1">
            {product.variations.length} variacao(oes)
          </p>
        )}
      </div>
    </Link>
  );
}
