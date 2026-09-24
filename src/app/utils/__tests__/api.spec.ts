import { getOverview } from '../api';

describe('api client', () => {
  const originalFetch = global.fetch;

  afterEach(() => {
    global.fetch = originalFetch;
    jest.resetAllMocks();
  });

  it('calls the BFF proxy path', async () => {
    global.fetch = jest.fn().mockResolvedValue({
      ok: true,
      json: async () => ({ range: '24h', totalCost: 1 }),
    }) as unknown as typeof fetch;

    await getOverview('24h');

    expect(global.fetch).toHaveBeenCalledWith(
      '/maas-finops/api/overview?range=24h',
      expect.objectContaining({ credentials: 'include' }),
    );
  });

  it('surfaces API errors', async () => {
    global.fetch = jest.fn().mockResolvedValue({
      ok: false,
      statusText: 'Forbidden',
      json: async () => ({ error: 'not admin' }),
    }) as unknown as typeof fetch;

    await expect(getOverview('1h')).rejects.toThrow('not admin');
  });
});
