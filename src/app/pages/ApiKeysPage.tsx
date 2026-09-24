import * as React from 'react';
import { Alert, Label } from '@patternfly/react-core';
import { Table, Thead, Tr, Th, Tbody, Td } from '@patternfly/react-table';
import FinOpsPage from '~/app/components/FinOpsPage';
import { ApiKeysResponse, getApiKeys, RangeKey } from '~/app/utils/api';
import { formatMoney, formatTokens } from '~/app/utils/format';

const ApiKeysPage: React.FC = () => {
  const [range, setRange] = React.useState<RangeKey>('24h');
  const [data, setData] = React.useState<ApiKeysResponse>();
  const [error, setError] = React.useState('');
  const [loading, setLoading] = React.useState(true);

  React.useEffect(() => {
    setLoading(true);
    getApiKeys(range)
      .then((resp) => {
        setData(resp);
        setError('');
      })
      .catch((e) => setError(e.message))
      .finally(() => setLoading(false));
  }, [range]);

  return (
    <FinOpsPage
      title="Usage by API key / user"
      range={range}
      onRangeChange={setRange}
      loading={loading}
      error={error}
      metricsError={data?.metricsError}
      source={data?.source}
    >
      {data?.note ? (
        <Alert variant="info" isInline title="How an API key is attributed" className="pf-v6-u-mb-md">
          {data.note}
        </Alert>
      ) : null}
      <Table aria-label="Usage by API key" variant="compact">
        <Thead>
          <Tr>
            <Th>User / API key owner</Th>
            <Th>Subscriptions</Th>
            <Th>Models</Th>
            <Th>Input</Th>
            <Th>Output</Th>
            <Th>Total</Th>
            <Th>Requests</Th>
            <Th>429</Th>
            <Th>Cost</Th>
          </Tr>
        </Thead>
        <Tbody>
          {(data?.items || []).map((row) => (
            <Tr key={row.user}>
              <Td>
                <strong>{row.user}</strong>
              </Td>
              <Td>
                <div className="mfp-inline">
                  {(row.subscriptions || []).map((s) => (
                    <Label key={s}>{s}</Label>
                  ))}
                </div>
              </Td>
              <Td>
                <div className="mfp-inline">
                  {(row.models || []).map((m) => (
                    <Label key={m} color="blue">
                      {m}
                    </Label>
                  ))}
                </div>
              </Td>
              <Td>{formatTokens(row.tokensIn)}</Td>
              <Td>{formatTokens(row.tokensOut)}</Td>
              <Td>{formatTokens(row.tokens)}</Td>
              <Td>{formatTokens(row.requests)}</Td>
              <Td>{formatTokens(row.limited)}</Td>
              <Td>{formatMoney(row.cost, data?.currency)}</Td>
            </Tr>
          ))}
        </Tbody>
      </Table>
    </FinOpsPage>
  );
};

export default ApiKeysPage;
