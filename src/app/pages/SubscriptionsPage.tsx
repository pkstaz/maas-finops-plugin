import * as React from 'react';
import { Label } from '@patternfly/react-core';
import { Table, Thead, Tr, Th, Tbody, Td } from '@patternfly/react-table';
import FinOpsPage from '~/app/components/FinOpsPage';
import { getSubscriptions, RangeKey, SubscriptionsResponse } from '~/app/utils/api';
import { formatMoney, formatTokens } from '~/app/utils/format';

const SubscriptionsPage: React.FC = () => {
  const [range, setRange] = React.useState<RangeKey>('24h');
  const [data, setData] = React.useState<SubscriptionsResponse>();
  const [error, setError] = React.useState('');
  const [loading, setLoading] = React.useState(true);

  React.useEffect(() => {
    setLoading(true);
    getSubscriptions(range)
      .then((resp) => {
        setData(resp);
        setError('');
      })
      .catch((e) => setError(e.message))
      .finally(() => setLoading(false));
  }, [range]);

  return (
    <FinOpsPage
      title="Tokens by subscription"
      range={range}
      onRangeChange={setRange}
      loading={loading}
      error={error}
      metricsError={data?.metricsError}
      source={data?.source}
    >
      <Table aria-label="Subscriptions MaaS" variant="compact">
        <Thead>
          <Tr>
            <Th>Subscription</Th>
            <Th>Priority</Th>
            <Th>Models</Th>
            <Th>Rate limits</Th>
            <Th>Input</Th>
            <Th>Output</Th>
            <Th>Total</Th>
            <Th>429</Th>
            <Th>Cost</Th>
          </Tr>
        </Thead>
        <Tbody>
          {(data?.items || []).map((row) => (
            <Tr key={`${row.namespace}/${row.name}`}>
              <Td>
                <strong>{row.name}</strong>
                <div className="mfp-muted">{row.namespace}</div>
              </Td>
              <Td>{row.priority}</Td>
              <Td>
                <div className="mfp-inline">
                  {(row.models || []).map((m) => (
                    <Label key={m}>{m}</Label>
                  ))}
                </div>
              </Td>
              <Td>{(row.rateLimits || []).join(' · ') || '—'}</Td>
              <Td>{formatTokens(row.tokensIn)}</Td>
              <Td>{formatTokens(row.tokensOut)}</Td>
              <Td>{formatTokens(row.tokens)}</Td>
              <Td>{formatTokens(row.limited)}</Td>
              <Td>{formatMoney(row.cost, data?.currency)}</Td>
            </Tr>
          ))}
        </Tbody>
      </Table>
    </FinOpsPage>
  );
};

export default SubscriptionsPage;
