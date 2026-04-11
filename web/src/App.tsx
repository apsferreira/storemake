import { BrowserRouter, Routes, Route } from 'react-router-dom';
import { HelmetProvider } from 'react-helmet-async';
import { StoreProvider } from './contexts/StoreContext';
import { ThemeProvider } from './contexts/ThemeContext';
import { CartProvider } from './contexts/CartContext';
import { Header } from './components/layout/Header';
import { Footer } from './components/layout/Footer';
import { WhatsAppButton } from './components/layout/WhatsAppButton';
import { CartSidebar } from './components/cart/CartSidebar';
import { CatalogPage } from './pages/CatalogPage';
import { ProductPage } from './pages/ProductPage';
import { CheckoutPage } from './pages/CheckoutPage';
import { OrderConfirmedPage } from './pages/OrderConfirmedPage';
import { TermosUsoPage } from './pages/TermosUsoPage';

export default function App() {
  return (
    <HelmetProvider>
      <BrowserRouter>
        <StoreProvider>
          <ThemeProvider>
            <CartProvider>
              <Header />
              <CartSidebar />
              <Routes>
                <Route path="/" element={<CatalogPage />} />
                <Route path="/produto/:slug" element={<ProductPage />} />
                <Route path="/checkout" element={<CheckoutPage />} />
                <Route path="/pedido-confirmado" element={<OrderConfirmedPage />} />
                {/* BKL-655: Termos de Uso obrigatório para marketplace (CDC Art. 14) */}
                <Route path="/termos-de-uso" element={<TermosUsoPage />} />
              </Routes>
              <Footer />
              <WhatsAppButton />
            </CartProvider>
          </ThemeProvider>
        </StoreProvider>
      </BrowserRouter>
    </HelmetProvider>
  );
}
