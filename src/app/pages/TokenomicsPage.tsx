import * as React from 'react';
import { Alert, Card, CardBody, CardTitle } from '@patternfly/react-core';
import { Table, Thead, Tr, Th, Tbody, Td } from '@patternfly/react-table';
import FinOpsPage from '~/app/components/FinOpsPage';
import { TokenomicsResponse, TokenomicsVariant, getTokenomics, RangeKey } from '~/app/utils/api';
import { formatMoney, formatTokens, isExternalModel, originLabel } from '~/app/utils/format';

function variantFor(row: TokenomicsResponse['items'][number], provider: string): TokenomicsVariant | undefined {
  return row.variants.find((v) => v.provider === provider);
}

function cheapestProvider(data: TokenomicsResponse): { label: string; cost: number } | undefined {
  const tokenProviders = (data.providers || []).filter((p) => p.mode === 'tokens');
  return tokenProviders.reduce<{ label: string; cost: number } | undefined>(
    (best, p) => (!best || p.cost < best.cost ? { label: p.label, cost: p.cost } : best),
    undefined,
  );
}

const TokenomicsPage: React.FC = () => {
  const [range, setRange] = React.useState<RangeKey>('24h');
  const [data, setData] = React.useState<TokenomicsResponse>();
  const [error, setError] = React.useState('');
  const [loading, setLoading] = React.useState(true);

  React.useEffect(() => {
    let cancelled = false;
    setLoading(true);
    getTokenomics(range)
      .then((resp) => {
        if (!cancelled) {
          setData(resp);
          setError('');
        }
      })
      .catch((e) => {
        if (!cancelled) {
          setError(e.message);
        }
      })
      .finally(() => {
        if (!cancelled) {
          setLoading(false);
        }
      });
    return () => {
      cancelled = true;
    };
  }, [range]);

  const cheapest = data ? cheapestProvider(data) : undefined;
  const savings = data && cheapest ? cheapest.cost - data.maasCost : 0;

  return (
    <FinOpsPage
      title="Tokenomics"
      range={range}
      onRangeChange={setRange}
      loading={loading}
      error={error}
      metricsError={data?.metricsError}
      source={data?.source}
    >
      {data ? (
        <>
          <Alert
            className="pf-v6-u-mb-md"
            variant="info"
            isInline
            title="What would these tokens cost as-a-service?"
          >
            Consumed tokens are priced with the MaaS catalog. The reference columns use public list
            prices (USD / 1M tokens) of Claude Opus 4.8 (Anthropic), Gemini Flash (Google AI), and
            GPT-5 (OpenAI). When the in/out split is unknown (Limitador total only), costs use a
            blended 3:1 input:output rate. Check the vendor account for committed-use discounts.
          </Alert>
          <div className="mfp-cards">
            <Card>
              <CardTitle>Tokens consumed</CardTitle>
              <CardBody>
                <div className="mfp-stat">{formatTokens(data.tokens)}</div>
                <p className="mfp-muted">{formatTokens(data.requests)} requests in {data.range}</p>
              </CardBody>
            </Card>
            <Card>
              <CardTitle>MaaS showback</CardTitle>
              <CardBody>
                <div className="mfp-stat">{formatMoney(data.maasCost, data.currency)}</div>
                <p className="mfp-muted">catalog prices</p>
              </CardBody>
            </Card>
            <Card>
              <CardTitle>Cheapest as-a-service</CardTitle>
              <CardBody>
                <div className="mfp-stat">{formatMoney(cheapest?.cost || 0, data.currency)}</div>
                <p className="mfp-muted">{cheapest ? cheapest.label : 'no matching variants'}</p>
              </CardBody>
            </Card>
            <Card>
              <CardTitle>Delta vs MaaS</CardTitle>
              <CardBody>
                <div className="mfp-stat">{formatMoney(savings, data.currency)}</div>
                <p className="mfp-muted">
                  {savings <= 0 ? 'self-hosting is cheaper' : 'cheapest provider − MaaS showback'}
                </p>
              </CardBody>
            </Card>
          </div>
          <Card>
            <CardTitle>Tokens vs public reference models</CardTitle>
            <CardBody>
              <Table aria-label="Tokenomics comparison" variant="compact">
                <Thead>
                  <Tr>
                    <Th>Model</Th>
                    <Th>In</Th>
                    <Th>Out</Th>
                    <Th>Total</Th>
                    <Th>Requests</Th>
                    <Th>MaaS showback</Th>
                    {(data.providers || []).map((p) => (
                      <Th key={p.id} title={p.mode === 'requests' ? 'per-request pricing' : 'per-token pricing'}>
                        {p.label}
                      </Th>
                    ))}
                  </Tr>
                </Thead>
                <Tbody>
                  {data.items.map((row) => (
                    <Tr key={row.name}>
                      <Td>
                        {row.displayName || row.name}
                        <div>
                          {isExternalModel(row) ? (
                            <span className="mfp-muted">{originLabel(row)}</span>
                          ) : null}
                        </div>
                      </Td>
                      <Td>{formatTokens(row.tokensIn)}</Td>
                      <Td>{formatTokens(row.tokensOut)}</Td>
                      <Td>{formatTokens(row.tokens)}</Td>
                      <Td>{formatTokens(row.requests)}</Td>
                      <Td>{formatMoney(row.maasCost, data.currency)}</Td>
                      {(data.providers || []).map((p) => {
                        const v = variantFor(row, p.id);
                        return (
                          <Td key={p.id}>
                            {v ? (
                              <>
                                {formatMoney(v.cost, data.currency)}
                                <div className="mfp-muted">
                                  {v.model} · {v.mode === 'requests' ? 'per-request' : 'per-token'}
                                </div>
                              </>
                            ) : (
                              '—'
                            )}
                          </Td>
                        );
                      })}
                    </Tr>
                  ))}
                </Tbody>
              </Table>
            </CardBody>
          </Card>
        </>
      ) : null}
    </FinOpsPage>
  );
};

export default TokenomicsPage;
