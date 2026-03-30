import { useState, type FormEvent } from 'react';
import { useNavigate } from 'react-router-dom';
import { useStore } from '../../contexts/StoreContext';
import { useCart } from '../../contexts/CartContext';
import { useTheme } from '../../contexts/ThemeContext';
import { formatCurrency } from '../../utils/format';
import { validateEmail, validatePhone, validateCep, validateRequired } from '../../utils/validation';
import { submitCheckout, validateCoupon } from '../../api/checkout';
import type { Address, CheckoutRequest, ValidateCouponResponse } from '../../types';

interface FormData {
  name: string;
  email: string;
  phone: string;
  street: string;
  number: string;
  complement: string;
  neighborhood: string;
  city: string;
  state: string;
  zipCode: string;
  paymentMethod: string;
  notes: string;
  couponCode: string;
}

const initialForm: FormData = {
  name: '',
  email: '',
  phone: '',
  street: '',
  number: '',
  complement: '',
  neighborhood: '',
  city: '',
  state: '',
  zipCode: '',
  paymentMethod: 'pix',
  notes: '',
  couponCode: '',
};

export function CheckoutForm() {
  const { store } = useStore();
  const { cart, refresh } = useCart();
  const theme = useTheme();
  const navigate = useNavigate();

  const [form, setForm] = useState<FormData>(initialForm);
  const [errors, setErrors] = useState<Partial<Record<keyof FormData, string>>>({});
  const [couponResult, setCouponResult] = useState<ValidateCouponResponse | null>(null);
  const [submitting, setSubmitting] = useState(false);
  const [submitError, setSubmitError] = useState('');
  const [couponLoading, setCouponLoading] = useState(false);

  const subtotal = cart?.subtotal_cents ?? 0;
  const discount = couponResult?.valid ? couponResult.discount_cents : 0;
  const shipping = 0; // frete a ser calculado
  const total = Math.max(0, subtotal + shipping - discount);

  const handleChange = (field: keyof FormData, value: string) => {
    setForm((prev) => ({ ...prev, [field]: value }));
    setErrors((prev) => ({ ...prev, [field]: undefined }));
  };

  const validate = (): boolean => {
    const newErrors: Partial<Record<keyof FormData, string>> = {};

    const nameErr = validateRequired(form.name, 'Nome');
    if (nameErr) newErrors.name = nameErr;

    if (!validateEmail(form.email)) newErrors.email = 'Email invalido';
    if (form.phone && !validatePhone(form.phone)) newErrors.phone = 'Telefone invalido';

    const streetErr = validateRequired(form.street, 'Rua');
    if (streetErr) newErrors.street = streetErr;

    const numErr = validateRequired(form.number, 'Numero');
    if (numErr) newErrors.number = numErr;

    const neighErr = validateRequired(form.neighborhood, 'Bairro');
    if (neighErr) newErrors.neighborhood = neighErr;

    const cityErr = validateRequired(form.city, 'Cidade');
    if (cityErr) newErrors.city = cityErr;

    const stateErr = validateRequired(form.state, 'Estado');
    if (stateErr) newErrors.state = stateErr;

    if (!validateCep(form.zipCode)) newErrors.zipCode = 'CEP invalido';

    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const handleApplyCoupon = async () => {
    if (!form.couponCode.trim()) return;
    setCouponLoading(true);
    try {
      const result = await validateCoupon(store.id, form.couponCode.trim(), subtotal);
      setCouponResult(result);
    } catch (err) {
      setCouponResult({ valid: false, message: 'Erro ao validar cupom', discount_cents: 0 });
    } finally {
      setCouponLoading(false);
    }
  };

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
    setSubmitError('');

    if (!validate()) return;
    if (!cart || cart.cart.items.length === 0) {
      setSubmitError('Carrinho vazio');
      return;
    }

    setSubmitting(true);
    try {
      const address: Address = {
        street: form.street.trim(),
        number: form.number.trim(),
        complement: form.complement.trim(),
        neighborhood: form.neighborhood.trim(),
        city: form.city.trim(),
        state: form.state.trim(),
        zip_code: form.zipCode.replace(/\D/g, ''),
      };

      const req: CheckoutRequest = {
        store_id: store.id,
        customer_name: form.name.trim(),
        customer_email: form.email.trim(),
        customer_phone: form.phone.replace(/\D/g, ''),
        address,
        shipping_cents: shipping,
        payment_method: form.paymentMethod,
        notes: form.notes.trim(),
        ...(couponResult?.valid ? { coupon_code: form.couponCode.trim() } : {}),
      };

      const result = await submitCheckout(req);
      await refresh();
      navigate('/pedido-confirmado', { state: { order: result.order } });
    } catch (err) {
      setSubmitError(err instanceof Error ? err.message : 'Erro ao processar pedido');
    } finally {
      setSubmitting(false);
    }
  };

  const inputClass = (field: keyof FormData) =>
    `w-full border rounded-lg px-3 py-2.5 text-sm focus:outline-none focus:ring-2 focus:ring-gray-300 ${
      errors[field] ? 'border-red-400' : 'border-gray-300'
    }`;

  return (
    <form onSubmit={handleSubmit} className="space-y-6" noValidate>
      {/* Dados pessoais */}
      <section>
        <h3 className="text-base font-bold mb-3">Dados pessoais</h3>
        <div className="space-y-3">
          <div>
            <input
              type="text"
              placeholder="Nome completo *"
              value={form.name}
              onChange={(e) => handleChange('name', e.target.value)}
              maxLength={255}
              className={inputClass('name')}
            />
            {errors.name && <p className="text-red-500 text-xs mt-1">{errors.name}</p>}
          </div>
          <div>
            <input
              type="email"
              placeholder="Email *"
              value={form.email}
              onChange={(e) => handleChange('email', e.target.value)}
              maxLength={255}
              className={inputClass('email')}
            />
            {errors.email && <p className="text-red-500 text-xs mt-1">{errors.email}</p>}
          </div>
          <div>
            <input
              type="tel"
              placeholder="Telefone (DDD + numero)"
              value={form.phone}
              onChange={(e) => handleChange('phone', e.target.value)}
              maxLength={20}
              className={inputClass('phone')}
            />
            {errors.phone && <p className="text-red-500 text-xs mt-1">{errors.phone}</p>}
          </div>
        </div>
      </section>

      {/* Endereco */}
      <section>
        <h3 className="text-base font-bold mb-3">Endereco de entrega</h3>
        <div className="space-y-3">
          <div className="grid grid-cols-3 gap-3">
            <div className="col-span-2">
              <input
                type="text"
                placeholder="CEP *"
                value={form.zipCode}
                onChange={(e) => handleChange('zipCode', e.target.value)}
                maxLength={9}
                className={inputClass('zipCode')}
              />
              {errors.zipCode && <p className="text-red-500 text-xs mt-1">{errors.zipCode}</p>}
            </div>
            <div>
              <input
                type="text"
                placeholder="UF *"
                value={form.state}
                onChange={(e) => handleChange('state', e.target.value.toUpperCase())}
                maxLength={2}
                className={inputClass('state')}
              />
              {errors.state && <p className="text-red-500 text-xs mt-1">{errors.state}</p>}
            </div>
          </div>
          <div>
            <input
              type="text"
              placeholder="Rua *"
              value={form.street}
              onChange={(e) => handleChange('street', e.target.value)}
              maxLength={255}
              className={inputClass('street')}
            />
            {errors.street && <p className="text-red-500 text-xs mt-1">{errors.street}</p>}
          </div>
          <div className="grid grid-cols-3 gap-3">
            <div>
              <input
                type="text"
                placeholder="Numero *"
                value={form.number}
                onChange={(e) => handleChange('number', e.target.value)}
                maxLength={20}
                className={inputClass('number')}
              />
              {errors.number && <p className="text-red-500 text-xs mt-1">{errors.number}</p>}
            </div>
            <div className="col-span-2">
              <input
                type="text"
                placeholder="Complemento"
                value={form.complement}
                onChange={(e) => handleChange('complement', e.target.value)}
                maxLength={255}
                className={inputClass('complement')}
              />
            </div>
          </div>
          <div>
            <input
              type="text"
              placeholder="Bairro *"
              value={form.neighborhood}
              onChange={(e) => handleChange('neighborhood', e.target.value)}
              maxLength={255}
              className={inputClass('neighborhood')}
            />
            {errors.neighborhood && <p className="text-red-500 text-xs mt-1">{errors.neighborhood}</p>}
          </div>
          <div>
            <input
              type="text"
              placeholder="Cidade *"
              value={form.city}
              onChange={(e) => handleChange('city', e.target.value)}
              maxLength={255}
              className={inputClass('city')}
            />
            {errors.city && <p className="text-red-500 text-xs mt-1">{errors.city}</p>}
          </div>
        </div>
      </section>

      {/* Cupom */}
      <section>
        <h3 className="text-base font-bold mb-3">Cupom de desconto</h3>
        <div className="flex gap-2">
          <input
            type="text"
            placeholder="Codigo do cupom"
            value={form.couponCode}
            onChange={(e) => {
              handleChange('couponCode', e.target.value);
              setCouponResult(null);
            }}
            maxLength={50}
            className="flex-1 border border-gray-300 rounded-lg px-3 py-2.5 text-sm focus:outline-none focus:ring-2 focus:ring-gray-300"
          />
          <button
            type="button"
            onClick={handleApplyCoupon}
            disabled={couponLoading || !form.couponCode.trim()}
            className={`px-4 py-2.5 rounded-lg text-sm font-medium transition-colors ${theme.buttonStyle} disabled:opacity-50`}
          >
            {couponLoading ? '...' : 'Aplicar'}
          </button>
        </div>
        {couponResult && (
          <p className={`text-sm mt-2 ${couponResult.valid ? 'text-green-600' : 'text-red-500'}`}>
            {couponResult.valid
              ? `Desconto aplicado: -${formatCurrency(couponResult.discount_cents)}`
              : couponResult.message}
          </p>
        )}
      </section>

      {/* Pagamento */}
      <section>
        <h3 className="text-base font-bold mb-3">Forma de pagamento</h3>
        <div className="space-y-2">
          {[
            { value: 'pix', label: 'PIX' },
            { value: 'cartao', label: 'Cartao de Credito' },
            { value: 'boleto', label: 'Boleto Bancario' },
          ].map((opt) => (
            <label
              key={opt.value}
              className={`flex items-center gap-3 p-3 rounded-lg border cursor-pointer transition-colors ${
                form.paymentMethod === opt.value ? 'border-gray-800 bg-gray-50' : 'border-gray-200'
              }`}
            >
              <input
                type="radio"
                name="payment"
                value={opt.value}
                checked={form.paymentMethod === opt.value}
                onChange={(e) => handleChange('paymentMethod', e.target.value)}
                className="accent-gray-800"
              />
              <span className="text-sm">{opt.label}</span>
            </label>
          ))}
        </div>
      </section>

      {/* Observacoes */}
      <section>
        <h3 className="text-base font-bold mb-3">Observacoes</h3>
        <textarea
          placeholder="Alguma observacao sobre o pedido?"
          value={form.notes}
          onChange={(e) => handleChange('notes', e.target.value)}
          maxLength={500}
          rows={3}
          className="w-full border border-gray-300 rounded-lg px-3 py-2.5 text-sm focus:outline-none focus:ring-2 focus:ring-gray-300 resize-none"
        />
      </section>

      {/* Resumo */}
      <section className="bg-gray-50 rounded-lg p-4 space-y-2">
        <div className="flex justify-between text-sm">
          <span>Subtotal</span>
          <span>{formatCurrency(subtotal)}</span>
        </div>
        {discount > 0 && (
          <div className="flex justify-between text-sm text-green-600">
            <span>Desconto</span>
            <span>-{formatCurrency(discount)}</span>
          </div>
        )}
        <div className="flex justify-between text-sm">
          <span>Frete</span>
          <span>{shipping === 0 ? 'A calcular' : formatCurrency(shipping)}</span>
        </div>
        <div className="flex justify-between text-base font-bold pt-2 border-t">
          <span>Total</span>
          <span>{formatCurrency(total)}</span>
        </div>
      </section>

      {submitError && (
        <div className="bg-red-50 text-red-600 p-3 rounded-lg text-sm">{submitError}</div>
      )}

      <button
        type="submit"
        disabled={submitting}
        className={`w-full py-3.5 rounded-lg font-bold text-center transition-colors ${theme.buttonStyle} disabled:opacity-50`}
      >
        {submitting ? 'Processando...' : `Confirmar Pedido - ${formatCurrency(total)}`}
      </button>
    </form>
  );
}
