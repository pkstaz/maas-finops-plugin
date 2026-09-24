import * as React from 'react';
import { Alert, Card, CardBody, CardTitle, Label } from '@patternfly/react-core';
import { Table, Thead, Tr, Th, Tbody, Td } from '@patternfly/react-table';
import FinOpsPage from '~/app/components/FinOpsPage';
import { getOverview, OverviewResponse, RangeKey } from '~/app/utils/api';
import { formatMoney, formatTokens, isExternalModel, originLabel } from '~/app/utils/format';

const OverviewPage: React.FC = () => {
  const [range, setRange] = React.useState<RangeKey>('24h');
  const [data, setData] = React.useState<OverviewResponse>();
  const [error, setError] = React.useState('');
  const [loading, setLoading] = React.useState(true);

  React.useEffect(() => {
    let cancelled = false;
    setLoading(true);
    getOverview(range)
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

  return (
    <FinOpsPage
      title="MaaS FinOps"
      range={range}
      onRangeChange={setRange}
      loading={loading}
      error={error}
      metricsError={data?.metricsError}
      source={data?.source}
    >
      {data ? (
        <>
          {data.unpricedModels > 0 ? (
            <Alert
              className="pf-v6-u-mb-md"
              variant="info"
              isInline
              title={`${data.unpricedModels} model(s) use the default price. Set Pricing for chargeback.`}
            />
          ) : null}
          <div className="mfp-cards">
            <Card>
              <CardTitle>Input tokens</CardTitle>
              <CardBody>
                <div className="mfp-stat">{formatTokens(data.totalTokensIn)}</div>
                <p className="mfp-muted">vllm:prompt_tokens in {data.range}</p>
              </CardBody>
            </Card>
            <Card>
              <CardTitle>Output tokens</CardTitle>
              <CardBody>
                <div className="mfp-stat">{formatTokens(data.totalTokensOut)}</div>
                <p className="mfp-muted">vllm:generation_tokens in {data.range}</p>
              </CardBody>
            </Card>
            <Card>
              <CardTitle>Total tokens</CardTitle>
              <CardBody>
                <div className="mfp-stat">{formatTokens(data.totalTokens)}</div>
                <p className="mfp-muted">input + output (or Limitador if vLLM is missing)</p>
              </CardBody>
            </Card>
            <Card>
              <CardTitle>Estimated cost</CardTitle>
              <CardBody>
                <div className="mfp-stat">{formatMoney(data.totalCost, data.currency)}</div>
                <p className="mfp-muted">tokens × price / 1M</p>
              </CardBody>
            </Card>
            <Card>
              <CardTitle>Requests</CardTitle>
              <CardBody>
                <div className="mfp-stat">{formatTokens(data.totalRequests)}</div>
                <p className="mfp-muted">authorized_calls</p>
              </CardBody>
            </Card>
            <Card>
              <CardTitle>Rate limited</CardTitle>
              <CardBody>
                <div className="mfp-stat">{formatTokens(data.totalLimited)}</div>
                <p className="mfp-muted">HTTP 429 / limited_calls</p>
              </CardBody>
            </Card>
          </div>
          <Card>
            <CardTitle>Top models by cost</CardTitle>
            <CardBody>
              <Table aria-label="Top models" variant="compact">
                <Thead>
                  <Tr>
                    <Th>Model</Th>
                    <Th>Input</Th>
                    <Th>Output</Th>
                    <Th>Total</Th>
                    <Th>Requests</Th>
                    <Th>Cost</Th>
                  </Tr>
                </Thead>
                <Tbody>
                  {(data.topModels || []).map((row) => (
                    <Tr key={row.name}>
                      <Td>
                        {row.displayName || row.name}
                        <div>
                          <Label color={isExternalModel(row) ? 'orange' : 'blue'}>{originLabel(row)}</Label>
                        </div>
                      </Td>
                      <Td>{formatTokens(row.tokensIn)}</Td>
                      <Td>{formatTokens(row.tokensOut)}</Td>
                      <Td>{formatTokens(row.tokens)}</Td>
                      <Td>{formatTokens(row.requests)}</Td>
                      <Td>{formatMoney(row.cost, data.currency)}</Td>
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

export default OverviewPage;
