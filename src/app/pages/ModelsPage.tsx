import * as React from 'react';
import {
  Button,
  Form,
  FormGroup,
  Label,
  Modal,
  ModalBody,
  ModalFooter,
  ModalHeader,
  TextInput,
} from '@patternfly/react-core';
import { Table, Thead, Tr, Th, Tbody, Td } from '@patternfly/react-table';
import FinOpsPage from '~/app/components/FinOpsPage';
import {
  getModels,
  getPricing,
  ModelRow,
  ModelsResponse,
  PricingCatalog,
  putPricing,
  RangeKey,
} from '~/app/utils/api';
import { formatMoney, formatModelPrice, formatTokens, isExternalModel, originLabel } from '~/app/utils/format';

const ModelsPage: React.FC = () => {
  const [range, setRange] = React.useState<RangeKey>('24h');
  const [data, setData] = React.useState<ModelsResponse>();
  const [catalog, setCatalog] = React.useState<PricingCatalog>();
  const [error, setError] = React.useState('');
  const [loading, setLoading] = React.useState(true);
  const [editing, setEditing] = React.useState<ModelRow>();
  const [price, setPrice] = React.useState('');
  const [priceIn, setPriceIn] = React.useState('');
  const [priceOut, setPriceOut] = React.useState('');
  const [saving, setSaving] = React.useState(false);

  const load = React.useCallback(() => {
    setLoading(true);
    Promise.all([getModels(range), getPricing()])
      .then(([models, pricing]) => {
        setData(models);
        setCatalog(pricing);
        setError('');
      })
      .catch((e) => setError(e.message))
      .finally(() => setLoading(false));
  }, [range]);

  React.useEffect(() => {
    load();
  }, [load]);

  const onSave = async () => {
    if (!editing || !catalog) {
      return;
    }
    const inV = Number(priceIn);
    const outV = Number(priceOut);
    const external = isExternalModel(editing);
    let value = Number(price);
    if (external && !(value > 0) && (inV > 0 || outV > 0)) {
      value = inV > 0 && outV > 0 ? (3 * inV + outV) / 4 : inV || outV;
    }
    if (Number.isNaN(value) || value < 0) {
      setError('Invalid price');
      return;
    }
    if (external && (Number.isNaN(inV) || Number.isNaN(outV) || inV < 0 || outV < 0)) {
      setError('Invalid input/output price');
      return;
    }
    setSaving(true);
    try {
      const next: PricingCatalog = {
        ...catalog,
        models: {
          ...catalog.models,
          [editing.name]: {
            displayName: editing.displayName,
            pricePerMillion: value,
            inputPerMillion: external && inV > 0 ? inV : undefined,
            outputPerMillion: external && outV > 0 ? outV : undefined,
            source: external ? 'manual-external' : 'manual',
          },
        },
      };
      await putPricing(next);
      setEditing(undefined);
      load();
    } catch (e) {
      setError((e as Error).message);
    } finally {
      setSaving(false);
    }
  };

  return (
    <>
      <FinOpsPage
        title="Models and price per token"
        range={range}
        onRangeChange={setRange}
        loading={loading}
        error={error}
        metricsError={data?.metricsError}
        source={data?.source}
      >
        <Table aria-label="MaaS models" variant="compact">
          <Thead>
            <Tr>
              <Th>Model</Th>
              <Th>Type</Th>
              <Th>Namespace</Th>
              <Th>Status</Th>
              <Th>Input</Th>
              <Th>Output</Th>
              <Th>Total</Th>
              <Th>Requests</Th>
              <Th>Price / 1M</Th>
              <Th>Cost</Th>
              <Th />
            </Tr>
          </Thead>
          <Tbody>
            {(data?.items || []).map((row) => (
              <Tr key={`${row.namespace}/${row.name}`}>
                <Td>
                  <strong>{row.displayName || row.name}</strong>
                  <div className="mfp-muted">{row.name}</div>
                </Td>
                <Td>
                  <Label color={isExternalModel(row) ? 'orange' : 'blue'}>{originLabel(row)}</Label>
                  {row.provider ? <div className="mfp-muted">{row.provider}</div> : null}
                </Td>
                <Td>{row.namespace}</Td>
                <Td>{row.phase || '—'}</Td>
                <Td>{formatTokens(row.tokensIn)}</Td>
                <Td>{formatTokens(row.tokensOut)}</Td>
                <Td>{formatTokens(row.tokens)}</Td>
                <Td>{formatTokens(row.requests)}</Td>
                <Td>
                  {formatModelPrice(row, data?.currency)}
                  {!row.priced ? <div className="mfp-muted">unpriced</div> : null}
                </Td>
                <Td>{formatMoney(row.cost, data?.currency)}</Td>
                <Td>
                  <Button
                    variant="link"
                    isInline
                    onClick={() => {
                      setEditing(row);
                      setPrice(String(row.pricePerMillion || ''));
                      setPriceIn(String(row.inputPerMillion || ''));
                      setPriceOut(String(row.outputPerMillion || ''));
                    }}
                  >
                    Price
                  </Button>
                </Td>
              </Tr>
            ))}
          </Tbody>
        </Table>
      </FinOpsPage>
      <Modal isOpen={!!editing} onClose={() => setEditing(undefined)} variant="small">
        <ModalHeader title={`Price: ${editing?.displayName || editing?.name || ''}`} />
        <ModalBody>
          <Form>
            {isExternalModel(editing || {}) ? (
              <>
                <FormGroup label="Input USD / 1M tokens" fieldId="price-in">
                  <TextInput id="price-in" type="number" value={priceIn} onChange={(_e, v) => setPriceIn(v)} />
                </FormGroup>
                <FormGroup label="Output USD / 1M tokens" fieldId="price-out">
                  <TextInput id="price-out" type="number" value={priceOut} onChange={(_e, v) => setPriceOut(v)} />
                </FormGroup>
                <FormGroup label="Blended USD / 1M (when only total tokens are counted)" fieldId="price-per-million">
                  <TextInput
                    id="price-per-million"
                    type="number"
                    value={price}
                    onChange={(_e, v) => setPrice(v)}
                  />
                </FormGroup>
                <p className="mfp-muted">
                  External models are billed by the vendor per token, not by cluster GPU hour.
                  {editing?.provider ? ` Provider: ${editing.provider}.` : ''}
                  {editing?.endpoint ? ` Endpoint: ${editing.endpoint}.` : ''}
                </p>
              </>
            ) : (
              <>
                <FormGroup
                  label="USD (or catalog currency) per 1 million tokens"
                  fieldId="price-per-million"
                >
                  <TextInput
                    id="price-per-million"
                    type="number"
                    value={price}
                    onChange={(_e, v) => setPrice(v)}
                  />
                </FormGroup>
                <p className="mfp-muted">
                  RHOAI-hosted models: cost is tokens × price / 1M. Recommend a GPU-based price on
                  Pricing.
                </p>
              </>
            )}
          </Form>
        </ModalBody>
        <ModalFooter>
          <Button variant="primary" onClick={onSave} isDisabled={saving}>
            Save
          </Button>
          <Button variant="link" onClick={() => setEditing(undefined)}>
            Cancel
          </Button>
        </ModalFooter>
      </Modal>
    </>
  );
};

export default ModelsPage;
