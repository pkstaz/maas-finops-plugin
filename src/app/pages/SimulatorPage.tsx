import * as React from 'react';
import {
  Alert,
  Card,
  CardBody,
  CardTitle,
  Form,
  FormGroup,
  FormSelect,
  FormSelectOption,
  Label,
  Spinner,
  TextInput,
  ToggleGroup,
  ToggleGroupItem,
} from '@patternfly/react-core';
import { Table, Thead, Tr, Th, Tbody, Td } from '@patternfly/react-table';
import {
  CatalogSKU,
  getSimulatorCatalog,
  getSimulatorQuote,
  SimulatorCatalogResponse,
  SimulatorQuote,
} from '~/app/utils/api';
import { formatMoney, formatPrice } from '~/app/utils/format';
import FinOpsPage from '~/app/components/FinOpsPage';

const DEFAULT_MODEL = 'llama-3.1-8b-instruct-w4a16';

function providerSKUs(data: SimulatorCatalogResponse | undefined, provider: string): CatalogSKU[] {
  return data?.providers.find((p) => p.id === provider)?.skus || [];
}

function withCurrentSKU(skus: CatalogSKU[], data: SimulatorCatalogResponse, provider: string): CatalogSKU[] {
  const current = data.current;
  if (!current.sku || provider !== current.provider) {
    return skus;
  }
  if (skus.some((s) => s.instanceType === current.sku)) {
    return skus;
  }
  return [
    {
      instanceType: current.sku,
      hourlyUsd: 0,
      gpuProduct: current.gpuProduct,
      gpuCount: current.gpuCount || 1,
      gpuMemoryGib: current.gpuMemoryGib,
      region: current.region,
      note: 'detected on this cluster',
    },
    ...skus,
  ];
}

const SimulatorPage: React.FC = () => {
  const [data, setData] = React.useState<SimulatorCatalogResponse>();
  const [error, setError] = React.useState('');
  const [loading, setLoading] = React.useState(true);
  const [scope, setScope] = React.useState<'this' | 'other'>('this');
  const [provider, setProvider] = React.useState('azure');
  const [sku, setSku] = React.useState('');
  const [model, setModel] = React.useState(DEFAULT_MODEL);
  const [utilization, setUtilization] = React.useState('0.8');
  const [hourlyUsd, setHourlyUsd] = React.useState('');
  const [quote, setQuote] = React.useState<SimulatorQuote>();
  const [quoting, setQuoting] = React.useState(false);

  React.useEffect(() => {
    getSimulatorCatalog()
      .then((resp) => {
        setData(resp);
        const currentProvider = (resp.current.provider || '').toLowerCase();
        const known = resp.providers.some((p) => p.id === currentProvider);
        if (known) {
          setProvider(currentProvider);
          setScope('this');
        } else {
          setProvider('azure');
          setScope('other');
        }
        const skus = withCurrentSKU(providerSKUs(resp, known ? currentProvider : 'azure'), resp, known ? currentProvider : 'azure');
        const detected = resp.current.sku;
        setSku(detected && skus.some((s) => s.instanceType === detected) ? detected : skus[0]?.instanceType || '');
        const hosted = (resp.inventory.models || []).filter((m) => m.origin !== 'external' && m.kind !== 'ExternalModel');
        const served = hosted
          .map((m) => `${m.name} ${m.displayName} ${m.uri}`.toLowerCase())
          .join(' ');
        const match = resp.models.find((m) => served.includes(m.id.replace(/-/g, ' ')) || served.includes(m.displayName.toLowerCase()));
        setModel(match?.id || DEFAULT_MODEL);
        setError('');
      })
      .catch((e) => setError(e.message))
      .finally(() => setLoading(false));
  }, []);

  const skus = React.useMemo(() => {
    if (!data) {
      return [];
    }
    const list = providerSKUs(data, provider);
    return scope === 'this' ? withCurrentSKU(list, data, provider) : list;
  }, [data, provider, scope]);

  React.useEffect(() => {
    if (!sku || skus.length === 0) {
      return;
    }
    if (!skus.some((s) => s.instanceType === sku)) {
      setSku(skus[0].instanceType);
    }
  }, [skus, sku]);

  React.useEffect(() => {
    if (!provider || !sku || !model) {
      return;
    }
    if (provider === 'baremetal' && !(Number(hourlyUsd) > 0)) {
      setQuote(undefined);
      return;
    }
    let cancelled = false;
    setQuoting(true);
    const selected = skus.find((s) => s.instanceType === sku);
    const useDetectedGPU = scope === 'this' && sku === data?.current.sku;
    getSimulatorQuote({
      provider,
      sku,
      model,
      utilization: Number(utilization) || 0.8,
      hourlyUsd: Number(hourlyUsd) || undefined,
      gpuProduct: useDetectedGPU ? data?.current.gpuProduct : selected?.gpuProduct,
      gpuCount: useDetectedGPU ? data?.current.gpuCount : selected?.gpuCount,
      gpuMemoryGib: useDetectedGPU ? data?.current.gpuMemoryGib : selected?.gpuMemoryGib,
    })
      .then((resp) => {
        if (!cancelled) {
          setQuote(resp);
          setError('');
        }
      })
      .catch((e) => {
        if (!cancelled) {
          setQuote(undefined);
          setError(e.message);
        }
      })
      .finally(() => {
        if (!cancelled) {
          setQuoting(false);
        }
      });
    return () => {
      cancelled = true;
    };
  }, [provider, sku, model, utilization, hourlyUsd, scope, data, skus]);

  const onScopeThis = () => {
    setScope('this');
    if (data?.current.provider && data.providers.some((p) => p.id === data.current.provider)) {
      setProvider(data.current.provider);
      if (data.current.sku) {
        setSku(data.current.sku);
      }
    }
  };

  const onProviderChange = (_e: unknown, value: string) => {
    setScope('other');
    setProvider(value);
    setHourlyUsd('');
    const next = providerSKUs(data, value);
    setSku(next[0]?.instanceType || '');
  };

  const gpuNodes = (data?.inventory.nodes || []).filter((n) => n.gpuCount > 0);
  const selectedSKU = skus.find((s) => s.instanceType === sku);
  const externalModels = (data?.inventory.models || []).filter(
    (m) => m.origin === 'external' || m.kind === 'ExternalModel',
  );

  return (
    <FinOpsPage title="Pricing Simulator" loading={loading} error={error}>
      <p className="mfp-muted">
        Estimate USD per 1M tokens from list GPU cost, a Red Hat model, and {Math.round((Number(utilization) || 0.8) * 100)}% of
        estimated throughput. This is for models you host on RHOAI. External models (Azure OpenAI, other MaaS) use vendor
        token rates on Pricing.
      </p>
      {data ? (
        <div className="mfp-stack">
          <Card>
            <CardTitle>Infrastructure</CardTitle>
            <CardBody>
              <ToggleGroup aria-label="Infrastructure source" className="pf-v6-u-mb-md">
                <ToggleGroupItem
                  text="This cluster"
                  isSelected={scope === 'this'}
                  onChange={onScopeThis}
                />
                <ToggleGroupItem
                  text="Another infrastructure"
                  isSelected={scope === 'other'}
                  onChange={() => setScope('other')}
                />
              </ToggleGroup>
              {scope === 'this' ? (
                <>
                  <div className="mfp-cards">
                    <div>
                      <div className="mfp-muted">Provider</div>
                      <strong>
                        {data.inventory.platform || data.current.provider || '—'}
                        {data.current.region ? ` · ${data.current.region}` : ''}
                      </strong>
                    </div>
                    <div>
                      <div className="mfp-muted">Detected SKU</div>
                      <strong>{data.current.sku || '—'}</strong>
                    </div>
                    <div>
                      <div className="mfp-muted">GPU</div>
                      <strong>
                        {data.current.gpuProduct || '—'}
                        {data.current.gpuCount ? ` ×${data.current.gpuCount}` : ''}
                        {data.current.gpuMemoryGib ? ` · ${data.current.gpuMemoryGib} GiB` : ''}
                      </strong>
                    </div>
                  </div>
                  <Table aria-label="GPU nodes on this cluster" variant="compact">
                    <Thead>
                      <Tr>
                        <Th>Node</Th>
                        <Th>SKU</Th>
                        <Th>GPU</Th>
                        <Th>VRAM</Th>
                      </Tr>
                    </Thead>
                    <Tbody>
                      {gpuNodes.length === 0 ? (
                        <Tr>
                          <Td colSpan={4}>No GPU nodes detected. Pick a SKU from the catalog below.</Td>
                        </Tr>
                      ) : (
                        gpuNodes.map((n) => (
                          <Tr key={n.name}>
                            <Td>
                              {n.name}
                              <div className="mfp-muted">{n.role}</div>
                            </Td>
                            <Td>{n.instanceType}</Td>
                            <Td>
                              {n.gpuProduct || '—'} ×{n.gpuCount}
                            </Td>
                            <Td>{n.gpuMemoryMib ? `${Math.round(n.gpuMemoryMib / 1024)} GiB` : '—'}</Td>
                          </Tr>
                        ))
                      )}
                    </Tbody>
                  </Table>
                </>
              ) : (
                <Form isHorizontal>
                  <FormGroup label="Provider" fieldId="sim-provider">
                    <FormSelect id="sim-provider" value={provider} onChange={onProviderChange} aria-label="Cloud provider">
                      {data.providers.map((p) => (
                        <FormSelectOption key={p.id} value={p.id} label={p.label} />
                      ))}
                    </FormSelect>
                  </FormGroup>
                </Form>
              )}
            </CardBody>
          </Card>

          {externalModels.length > 0 ? (
            <Alert
              variant="info"
              isInline
              title="External models on this cluster are not simulated here"
            >
              {externalModels.map((m) => m.displayName || m.name).join(', ')} — token price is the
              vendor rate (Azure OpenAI, other MaaS, …), not GPU USD/hour. Set it on Pricing.
            </Alert>
          ) : null}

          <Card>
            <CardTitle>
              Hardware catalog · {data.providers.find((p) => p.id === provider)?.label || provider}
            </CardTitle>
            <CardBody>
              <Table aria-label="Hardware catalog" variant="compact">
                <Thead>
                  <Tr>
                    <Th>SKU</Th>
                    <Th>GPU</Th>
                    <Th>VRAM</Th>
                    <Th>USD / hour</Th>
                  </Tr>
                </Thead>
                <Tbody>
                  {skus.map((s) => (
                    <Tr
                      key={s.instanceType}
                      onClick={() => setSku(s.instanceType)}
                    >
                      <Td>
                        <strong>{s.instanceType}</strong>
                        {s.instanceType === sku ? (
                          <div>
                            <Label>Selected</Label>
                          </div>
                        ) : null}
                        {s.note ? <div className="mfp-muted">{s.note}</div> : null}
                      </Td>
                      <Td>
                        {s.gpuProduct} ×{s.gpuCount}
                      </Td>
                      <Td>{s.gpuMemoryGib ? `${s.gpuMemoryGib * (s.gpuCount || 1)} GiB` : '—'}</Td>
                      <Td>{s.hourlyUsd > 0 ? formatMoney(s.hourlyUsd) : s.note || 'enter USD/hour'}</Td>
                    </Tr>
                  ))}
                </Tbody>
              </Table>
            </CardBody>
          </Card>

          <Card>
            <CardTitle>Simulation</CardTitle>
            <CardBody>
              <Form className="mfp-inline">
                <FormGroup label="SKU" fieldId="sim-sku">
                  <FormSelect id="sim-sku" value={sku} onChange={(_e, v) => setSku(v)} aria-label="SKU">
                    {skus.map((s) => (
                      <FormSelectOption
                        key={s.instanceType}
                        value={s.instanceType}
                        label={`${s.instanceType} · ${s.gpuProduct}`}
                      />
                    ))}
                  </FormSelect>
                </FormGroup>
                <FormGroup label="Red Hat model" fieldId="sim-model">
                  <FormSelect id="sim-model" value={model} onChange={(_e, v) => setModel(v)} aria-label="Red Hat model">
                    {(data.models || []).map((m) => (
                      <FormSelectOption
                        key={m.id}
                        value={m.id}
                        label={`${m.displayName} · ${m.paramsB}B ${m.quant.toUpperCase()}`}
                      />
                    ))}
                  </FormSelect>
                </FormGroup>
                <FormGroup label="GPU utilization" fieldId="sim-util">
                  <TextInput
                    id="sim-util"
                    type="number"
                    min={0.1}
                    max={1}
                    step={0.05}
                    value={utilization}
                    onChange={(_e, v) => setUtilization(v)}
                  />
                </FormGroup>
                <FormGroup label={provider === 'baremetal' ? 'USD/hour (required)' : 'USD/hour override'} fieldId="sim-hourly">
                  <TextInput
                    id="sim-hourly"
                    type="number"
                    placeholder={selectedSKU?.hourlyUsd ? String(selectedSKU.hourlyUsd) : '0.752'}
                    value={hourlyUsd}
                    onChange={(_e, v) => setHourlyUsd(v)}
                  />
                </FormGroup>
              </Form>
              {provider === 'baremetal' && !(Number(hourlyUsd) > 0) ? (
                <Alert variant="info" isInline title="Enter USD/hour for this GPU server to get a token price." />
              ) : null}
              {quoting && !quote ? (
                <div className="mfp-loading">
                  <Spinner size="lg" />
                </div>
              ) : null}
              {quote ? (
                <>
                  <div className="mfp-cards pf-v6-u-mt-md">
                    <div>
                      <div className="mfp-muted">Token price</div>
                      <div className="mfp-quote">
                        {quote.pricePerMillion > 0 ? formatPrice(quote.pricePerMillion, data.currency) : '—'}
                      </div>
                      <div className="mfp-muted mfp-quote-sub">
                        {quote.model.displayName} on {quote.sku.instanceType}
                      </div>
                    </div>
                    <div>
                      <div className="mfp-muted">Machine</div>
                      <strong>{formatMoney(quote.hourlyUsd, data.currency)} / hour</strong>
                      <div className="mfp-muted">
                        <Label>{quote.cost.kind}</Label> {quote.cost.note}
                      </div>
                    </div>
                    <div>
                      <div className="mfp-muted">Throughput @ {(quote.utilization * 100).toFixed(0)}%</div>
                      <strong>{quote.tokensPerSecUsed.toFixed(0)} tok/s</strong>
                      <div className="mfp-muted">{quote.tokensPerSecFull.toFixed(0)} tok/s at 100%</div>
                    </div>
                    <div>
                      <div className="mfp-muted">VRAM</div>
                      <strong>~{quote.vramGib.toFixed(1)} GiB weights+KV</strong>
                      <div className="mfp-muted">
                        {quote.fits ? 'Fits on this SKU' : 'Likely does not fit'}
                      </div>
                    </div>
                  </div>
                  {quote.warning ? <Alert variant="warning" isInline title={quote.warning} /> : null}
                  <ul className="mfp-muted">
                    {quote.assumptions.map((a) => (
                      <li key={a}>{a}</li>
                    ))}
                  </ul>
                </>
              ) : null}
            </CardBody>
          </Card>
        </div>
      ) : null}
    </FinOpsPage>
  );
};

export default SimulatorPage;
