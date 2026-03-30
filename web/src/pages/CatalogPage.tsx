import { useState, useEffect, useCallback } from 'react';
import { useStore } from '../contexts/StoreContext';
import { useCart } from '../contexts/CartContext';
import { fetchCatalog } from '../api/catalog';
import { SEOHead } from '../components/seo/SEOHead';
import { SearchBar } from '../components/catalog/SearchBar';
import { CategoryFilter } from '../components/catalog/CategoryFilter';
import { ProductGrid } from '../components/catalog/ProductGrid';
import { Pagination } from '../components/catalog/Pagination';
import type { Product, Category } from '../types';

export function CatalogPage() {
  const { store } = useStore();
  const { refresh: refreshCart } = useCart();

  const [products, setProducts] = useState<Product[]>([]);
  const [categories, setCategories] = useState<Category[]>([]);
  const [loading, setLoading] = useState(true);
  const [page, setPage] = useState(1);
  const [totalPages, setTotalPages] = useState(1);
  const [selectedCategory, setSelectedCategory] = useState<string | null>(null);
  const [search, setSearch] = useState('');

  const loadCatalog = useCallback(async () => {
    if (!store.id) return;
    setLoading(true);
    try {
      const data = await fetchCatalog({
        storeId: store.id,
        page,
        perPage: 12,
        categoriaId: selectedCategory || undefined,
        search: search || undefined,
      });
      setProducts(data.data || []);
      setTotalPages(data.total_pages);

      // Extract unique categories from products for filter
      const catMap = new Map<string, Category>();
      (data.data || []).forEach((p) => {
        if (p.categoria_id) {
          // We don't have full category objects from the catalog endpoint,
          // so we just track IDs
          catMap.set(p.categoria_id, {
            id: p.categoria_id,
            loja_id: p.loja_id,
            name: p.categoria_id, // will be replaced if we had the endpoint
            slug: '',
            sort_order: 0,
            created_at: '',
          });
        }
      });
      // Keep existing categories if they were set
      if (categories.length === 0) {
        setCategories(Array.from(catMap.values()));
      }
    } catch {
      setProducts([]);
    } finally {
      setLoading(false);
    }
  }, [store.id, page, selectedCategory, search]);

  useEffect(() => {
    loadCatalog();
  }, [loadCatalog]);

  useEffect(() => {
    refreshCart();
  }, []);

  const handleSearch = (query: string) => {
    setSearch(query);
    setPage(1);
  };

  const handleCategoryChange = (catId: string | null) => {
    setSelectedCategory(catId);
    setPage(1);
  };

  return (
    <>
      <SEOHead
        title={store.name}
        description={`Confira os produtos da loja ${store.name}`}
      />

      <main className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-6 sm:py-8">
        {/* Header */}
        <div className="mb-6 sm:mb-8">
          {store.banner_url && (
            <div className="w-full h-32 sm:h-48 lg:h-64 rounded-lg overflow-hidden mb-6">
              <img
                src={store.banner_url}
                alt={`Banner ${store.name}`}
                className="w-full h-full object-cover"
              />
            </div>
          )}

          <h1 className="text-2xl sm:text-3xl font-bold mb-4">Produtos</h1>

          {/* Search + Filter */}
          <div className="space-y-4">
            <SearchBar onSearch={handleSearch} />
            <CategoryFilter
              categories={categories}
              selected={selectedCategory}
              onSelect={handleCategoryChange}
            />
          </div>
        </div>

        {/* Grid */}
        <ProductGrid products={products} loading={loading} />

        {/* Pagination */}
        <Pagination page={page} totalPages={totalPages} onPageChange={setPage} />
      </main>
    </>
  );
}
