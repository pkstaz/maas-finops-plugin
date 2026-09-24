export const RANGES = [
  { key: '1h', label: '1 h' },
  { key: '6h', label: '6 h' },
  { key: '24h', label: '24 h' },
  { key: '7d', label: '7 d' },
] as const;

export function formatTokens(n: number): string {
  if (!n) {
    return '0';
  }
  if (n >= 1_000_000) {
    return `${(n / 1_000_000).toFixed(2)} M`;
  }
  if (n >= 1_000) {
    return `${(n / 1_000).toFixed(1)} k`;
  }
  return Math.round(n).toLocaleString();
}

export function formatMoney(n: number, currency = 'USD'): string {
  try {
    return new Intl.NumberFormat(undefined, {
      style: 'currency',
      currency,
      maximumFractionDigits: n >= 1 ? 2 : 4,
    }).format(n || 0);
  } catch {
    return `${currency} ${(n || 0).toFixed(4)}`;
  }
}

export function formatPrice(n: number, currency = 'USD'): string {
  return `${formatMoney(n, currency)} / 1M tok`;
}

export function isExternalModel(row: { origin?: string; kind?: string }): boolean {
  return row.origin === 'external' || row.kind === 'ExternalModel';
}

export function originLabel(row: { origin?: string; kind?: string }): string {
  return isExternalModel(row) ? 'External' : 'RHOAI';
}

export function formatModelPrice(
  row: { pricePerMillion?: number; inputPerMillion?: number; outputPerMillion?: number },
  currency = 'USD',
): string {
  if (row.inputPerMillion || row.outputPerMillion) {
    return `${formatMoney(row.inputPerMillion || 0, currency)} in / ${formatMoney(
      row.outputPerMillion || 0,
      currency,
    )} out / 1M`;
  }
  if (!row.pricePerMillion) {
    return '—';
  }
  return formatPrice(row.pricePerMillion, currency);
}
