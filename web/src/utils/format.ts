/**
 * Formata centavos para moeda brasileira (BRL).
 */
export function formatCurrency(cents: number): string {
  return new Intl.NumberFormat('pt-BR', {
    style: 'currency',
    currency: 'BRL',
  }).format(cents / 100);
}

/**
 * Sanitiza texto para prevenir XSS ao exibir em HTML.
 */
export function sanitizeText(text: string): string {
  const div = document.createElement('div');
  div.textContent = text;
  return div.innerHTML;
}

/**
 * Gera ou recupera session ID do localStorage.
 */
export function getSessionId(): string {
  const key = 'sm_session_id';
  let sid = localStorage.getItem(key);
  if (!sid) {
    sid = crypto.randomUUID();
    localStorage.setItem(key, sid);
  }
  return sid;
}

/**
 * Retorna store_id da URL (subdomínio ou query param).
 * Padrão: usa query param ?store= ou localStorage.
 */
export function getStoreId(): string {
  const params = new URLSearchParams(window.location.search);
  const fromUrl = params.get('store');
  if (fromUrl) {
    localStorage.setItem('sm_store_id', fromUrl);
    return fromUrl;
  }
  return localStorage.getItem('sm_store_id') || '';
}

/**
 * Trunca texto com reticências.
 */
export function truncate(text: string, maxLen: number): string {
  if (text.length <= maxLen) return text;
  return text.slice(0, maxLen).trimEnd() + '...';
}
