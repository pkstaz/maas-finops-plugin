import { formatMoney, formatTokens, isExternalModel, originLabel } from '../format';

describe('format helpers', () => {
  it('formats token counts', () => {
    expect(formatTokens(0)).toBe('0');
    expect(formatTokens(500)).toBe('500');
    expect(formatTokens(1500)).toBe('1.5 k');
    expect(formatTokens(2_500_000)).toBe('2.50 M');
  });

  it('formats money with a fallback currency', () => {
    expect(formatMoney(1.5, 'USD')).toMatch(/1\.50/);
  });

  it('labels hosted vs external models', () => {
    expect(isExternalModel({ origin: 'external' })).toBe(true);
    expect(isExternalModel({ kind: 'LLMInferenceService' })).toBe(false);
    expect(originLabel({ origin: 'external' })).toBe('External');
    expect(originLabel({ kind: 'LLMInferenceService' })).toBe('RHOAI');
  });
});
