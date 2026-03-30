import { useState, useEffect } from 'react';
import { useParams, useLocation, Link } from 'react-router-dom';
import { useStore } from '../contexts/StoreContext';
import { useCart } from '../contexts/CartContext';
import { useTheme } from '../contexts/ThemeContext';
import { fetchCatalog } from '../api/catalog';
import { SEOHead } from '../components/seo/SEOHead';
import { ProductGallery } from '../components/product/ProductGallery';
import { VariationSelector } from '../components/product/VariationSelector';
import { formatCurrency } from '../utils/format';
import DOMPurify from 'dompurify';
import type { Product } from '../types';

export function ProductPage() {
  const { slug } = useParams<{ slug: string }>();
  const location = useLocation();
  const { store } = useStore();
  const { addItem, loading: cartLoading } = useCart();
  const theme = useTheme();

  const [product, setProduct] = useState<Product | null>(
    (location.state as { product?: Product })?.product ?? null,
  );
  const [loading, setLoading] = useState(!product);
  const [selectedVariant, setSelectedVariant] = useState<string | null>(null);
  const [quantity, setQuantity] = useState(1);
  const [addedMessage, setAddedMessage] = useState('');

  // Fetch product if not passed via state
  useEffect(() => {
    if (product) return;
    if (!store.id || !slug) return;

    setLoading(true);
    fetchCatalog({ storeId: store.id, search: slug, perPage: 50 })
      .then((data) => {
        const found = (data.data || []).find((p) => p.slug === slug);
        if (found) setProduct(found);
      })
      .finally(() => setLoading(false));
  }, [store.id, slug, product]);

  if (loading) {
    return (
      <main className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <div className="animate-pulse grid grid-cols-1 md:grid-cols-2 gap-8">
          <div className="aspect-square bg-gray-200 rounded-lg" />
          <div className="space-y-4">
            <div className="h-8 bg-gray-200 rounded w-3/4" />
            <div className="h-6 bg-gray-200 rounded w-1/2" />
            <div className="h-20 bg-gray-200 rounded" />
          </div>
        </div>
      </main>
    );
  }

  if (!product) {
    return (
      <main className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-16 text-center">
        <h1 className="text-2xl font-bold mb-4">Produto nao encontrado</h1>
        <Link to="/" className="text-blue-600 hover:underline">
          Voltar ao catalogo
        </Link>
      </main>
    );
  }

  const selectedVariation = product.variations?.find((v) => v.id === selectedVariant);
  const finalPrice = product.price_cents + (selectedVariation?.price_adjustment_cents ?? 0);
  const hasDiscount = product.compare_price_cents > product.price_cents;
  const inStock =
    selectedVariation !== undefined
      ? selectedVariation.stock_quantity > 0
      : product.stock_quantity > 0;

  const handleAddToCart = async () => {
    if (!inStock) return;
    if (product.variations?.length && !selectedVariant) return;

    try {
      await addItem(product.id, quantity, selectedVariant || undefined);
      setAddedMessage('Produto adicionado ao carrinho!');
      setTimeout(() => setAddedMessage(''), 3000);
    } catch {
      setAddedMessage('Erro ao adicionar ao carrinho');
      setTimeout(() => setAddedMessage(''), 3000);
    }
  };

  return (
    <>
      <SEOHead
        title={product.name}
        description={product.description || `${product.name} - ${formatCurrency(product.price_cents)}`}
        image={product.photos?.[0]?.url}
        type="product"
      />

      <main className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-6 sm:py-8">
        {/* Breadcrumb */}
        <nav className="text-sm text-gray-500 mb-6">
          <Link to="/" className="hover:text-gray-700">Produtos</Link>
          <span className="mx-2">/</span>
          <span>{product.name}</span>
        </nav>

        <div className="grid grid-cols-1 md:grid-cols-2 gap-6 lg:gap-12">
          {/* Gallery */}
          <ProductGallery photos={product.photos || []} productName={product.name} />

          {/* Details */}
          <div className="space-y-5">
            <h1 className="text-2xl sm:text-3xl font-bold">{product.name}</h1>

            {/* Price */}
            <div className="flex items-baseline gap-3">
              <span className="text-2xl sm:text-3xl font-bold" style={{ color: theme.primaryColor }}>
                {formatCurrency(finalPrice)}
              </span>
              {hasDiscount && (
                <span className="text-lg text-gray-400 line-through">
                  {formatCurrency(product.compare_price_cents)}
                </span>
              )}
            </div>

            {/* Stock */}
            {!inStock ? (
              <p className="text-red-500 font-medium">Produto esgotado</p>
            ) : (
              <p className="text-green-600 text-sm">Em estoque</p>
            )}

            {/* Variations */}
            {product.variations && product.variations.length > 0 && (
              <VariationSelector
                variations={product.variations}
                selected={selectedVariant}
                onSelect={setSelectedVariant}
              />
            )}

            {/* Quantity */}
            <div>
              <label className="text-sm font-medium mb-2 block">Quantidade</label>
              <div className="flex items-center gap-3">
                <button
                  onClick={() => setQuantity((q) => Math.max(1, q - 1))}
                  className="w-10 h-10 flex items-center justify-center rounded-lg border hover:bg-gray-100 transition-colors"
                  aria-label="Diminuir quantidade"
                >
                  -
                </button>
                <span className="text-lg font-medium w-8 text-center">{quantity}</span>
                <button
                  onClick={() => setQuantity((q) => Math.min(999, q + 1))}
                  className="w-10 h-10 flex items-center justify-center rounded-lg border hover:bg-gray-100 transition-colors"
                  aria-label="Aumentar quantidade"
                >
                  +
                </button>
              </div>
            </div>

            {/* Add to cart */}
            <button
              onClick={handleAddToCart}
              disabled={!inStock || cartLoading || (product.variations?.length > 0 && !selectedVariant)}
              className={`w-full py-3.5 rounded-lg font-bold text-center transition-colors ${theme.buttonStyle} disabled:opacity-50 disabled:cursor-not-allowed`}
            >
              {cartLoading
                ? 'Adicionando...'
                : !inStock
                ? 'Esgotado'
                : product.variations?.length > 0 && !selectedVariant
                ? 'Selecione uma variacao'
                : 'Adicionar ao Carrinho'}
            </button>

            {addedMessage && (
              <p className={`text-sm text-center ${addedMessage.includes('Erro') ? 'text-red-500' : 'text-green-600'}`}>
                {addedMessage}
              </p>
            )}

            {/* Description */}
            {product.description && (
              <div className="pt-4 border-t">
                <h2 className="text-base font-bold mb-2">Descricao</h2>
                <div
                  className="text-sm text-gray-600 leading-relaxed prose prose-sm max-w-none"
                  dangerouslySetInnerHTML={{
                    __html: DOMPurify.sanitize(product.description),
                  }}
                />
              </div>
            )}

            {/* SKU */}
            {product.sku && (
              <p className="text-xs text-gray-400">SKU: {product.sku}</p>
            )}
          </div>
        </div>
      </main>
    </>
  );
}
