import { Link } from 'react-router-dom';
import { useStore } from '../../contexts/StoreContext';

export function Footer() {
  const { store } = useStore();

  return (
    <footer className="border-t mt-12 py-8 px-4 text-center text-sm opacity-60">
      <p>&copy; {new Date().getFullYear()} {store.name}. Todos os direitos reservados.</p>
      <p className="mt-1">Powered by StoreMaker</p>
      {/* BKL-655: link obrigatório para Termos de Uso (CDC Art. 14) */}
      <p className="mt-2">
        <Link to="/termos-de-uso" className="underline hover:opacity-100">
          Termos de Uso
        </Link>
      </p>
    </footer>
  );
}
