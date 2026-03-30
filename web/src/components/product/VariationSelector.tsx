import type { ProductVariation } from '../../types';
import { useTheme } from '../../contexts/ThemeContext';
import { formatCurrency } from '../../utils/format';

interface VariationSelectorProps {
  variations: ProductVariation[];
  selected: string | null;
  onSelect: (id: string) => void;
}

export function VariationSelector({ variations, selected, onSelect }: VariationSelectorProps) {
  const theme = useTheme();

  if (!variations || variations.length === 0) return null;

  // Group variations by name (e.g., "Cor", "Tamanho")
  const groups = variations.reduce<Record<string, ProductVariation[]>>((acc, v) => {
    if (!acc[v.name]) acc[v.name] = [];
    acc[v.name].push(v);
    return acc;
  }, {});

  return (
    <div className="space-y-4">
      {Object.entries(groups).map(([groupName, items]) => (
        <div key={groupName}>
          <label className="text-sm font-medium mb-2 block">{groupName}</label>
          <div className="flex flex-wrap gap-2">
            {items.map((v) => (
              <button
                key={v.id}
                onClick={() => onSelect(v.id)}
                disabled={v.stock_quantity === 0}
                className={`px-4 py-2 text-sm rounded-lg border transition-colors ${
                  selected === v.id
                    ? theme.buttonStyle + ' border-transparent'
                    : v.stock_quantity === 0
                    ? 'border-gray-200 text-gray-300 cursor-not-allowed line-through'
                    : 'border-gray-300 hover:border-gray-500'
                }`}
              >
                {v.value}
                {v.price_adjustment_cents !== 0 && (
                  <span className="ml-1 text-xs opacity-70">
                    ({v.price_adjustment_cents > 0 ? '+' : ''}
                    {formatCurrency(v.price_adjustment_cents)})
                  </span>
                )}
              </button>
            ))}
          </div>
        </div>
      ))}
    </div>
  );
}
