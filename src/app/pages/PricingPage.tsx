import * as React from 'react';
import {
  Alert,
  Button,
  Card,
  CardBody,
  CardTitle,
  Form,
  FormGroup,
  Label,
  TextInput,
} from '@patternfly/react-core';
import { Table, Thead, Tr, Th, Tbody, Td, ActionsColumn } from '@patternfly/react-table';
import {
  applyRecommend,
  getPricing,
  getRecommend,
  HardwareConfig,
  ManualMachine,
  PricingCatalog,
  putPricing,
  RecommendResponse,
} from '~/app/utils/api';
import { formatMoney, formatModelPrice, formatPrice, isExternalModel, originLabel } from '~/app/utils/format';
import FinOpsPage from '~/app/components/FinOpsPage';

function emptyHardware(): HardwareConfig {
  return {
    provider: 'auto',
    utilization: 0.8,
    margin: 0,
    priceType: 'Consumption',
    machines: [],
  };
}

function defaultManual(provider?: string, sku?: string, region?: string): ManualMachine {
  let base: ManualMachine;
  switch ((provider || '').toLowerCase()) {
    case 'aws':
      base = {
        instanceType: 'g5.xlarge',
        region: 'us-east-1',
        hourlyUsd: 1.006,
        gpuProduct: 'A10G',
        gpuCount: 1,
      };
      break;
    case 'ibmcloud':
      base = {
        instanceType: 'gx3-24x120x1l40s',
        region: 'us-south',
        hourlyUsd: 3.88,
        gpuProduct: 'L40S',
        gpuCount: 1,
      };
      break;
    default:
      base = {
        instanceType: 'Standard_NC8as_T4_v3',
        region: 'eastus',
        hourlyUsd: 0.752,
        gpuProduct: 'Tesla T4',
        gpuCount: 1,
        gpuMemoryGib: 16,
      };
  }
  if (sku) {
    base = { ...base, instanceType: sku };
  }
  if (region) {
    base = { ...base, region };
  }
  return base;
}

const PricingPage: React.FC = () => {
  const [catalog, setCatalog] = React.useState<PricingCatalog>();
  const [rec, setRec] = React.useState<RecommendResponse>();
  const [error, setError] = React.useState('');
  const [saved, setSaved] = React.useState('');
  const [loading, setLoading] = React.useState(true);
  const [saving, setSaving] = React.useState(false);
  const [newName, setNewName] = React.useState('');
  const [newPrice, setNewPrice] = React.useState('0.15');
  const [manual, setManual] = React.useState<ManualMachine>({
    instanceType: 'Standard_NC8as_T4_v3',
    region: 'eastus',
    hourlyUsd: 0.752,
    gpuProduct: 'Tesla T4',
    gpuCount: 1,
    gpuMemoryGib: 16,
  });

  const load = () => {
    setLoading(true);
    Promise.all([getPricing(), getRecommend()])
      .then(([c, r]) => {
        setCatalog({ ...c, hardware: { ...emptyHardware(), ...(c.hardware || {}) } });
        setRec(r);
        const gpu = (r.inventory.nodes || []).find((n) => n.gpuCount > 0);
        setManual(
          defaultManual(r.inventory.provider, gpu?.instanceType, r.inventory.region || gpu?.region),
        );
        setError('');
      })
      .catch((e) => setError(e.message))
      .finally(() => setLoading(false));
  };

  React.useEffect(() => {
    load();
  }, []);

  const save = async (next: PricingCatalog) => {
    setSaving(true);
    setSaved('');
    try {
      const updated = await putPricing(next);
      setCatalog(updated);
      setSaved('Catalog saved');
      const refreshed = await getRecommend();
      setRec(refreshed);
    } catch (e) {
      setError((e as Error).message);
    } finally {
      setSaving(false);
    }
  };

  const gpuNodes = (rec?.inventory.nodes || []).filter((n) => n.gpuCount > 0);
  const rows = Object.entries(catalog?.models || {});
  const hw = catalog?.hardware || emptyHardware();

  return (
    <FinOpsPage title="Hardware and pricing" loading={loading} error={error}>
      {saved ? <Alert variant="success" isInline title={saved} className="pf-v6-u-mb-md" /> : null}
      {catalog ? (
        <div className="mfp-stack">
          <Card>
            <CardTitle>Detected cluster</CardTitle>
            <CardBody>
              <div className="mfp-cards">
                <div>
                  <div className="mfp-muted">Platform</div>
                  <strong>
                    {rec?.inventory.platform || rec?.inventory.provider || '—'}
                    {rec?.inventory.provider ? ` (${rec.inventory.provider})` : ''}{' '}
                    {rec?.inventory.region}
                  </strong>
                </div>
                <div>
                  <div className="mfp-muted">Source</div>
                  <strong>{rec?.inventory.source || '—'}</strong>
                </div>
              </div>
              {rec?.inventory.message ? (
                <Alert variant="info" isInline title={rec.inventory.message} />
              ) : null}
              <Table aria-label="GPU nodes" variant="compact">
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
                      <Td colSpan={4}>No GPU nodes. Add machine cost below.</Td>
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
            </CardBody>
          </Card>

          <Card>
            <CardTitle>Recommended price (80% of GPU capacity)</CardTitle>
            <CardBody>
              <Form isHorizontal className="mfp-inline">
                <FormGroup label="GPU utilization" fieldId="util">
                  <TextInput
                    id="util"
                    type="number"
                    min={0.1}
                    max={1}
                    step={0.05}
                    value={String(hw.utilization)}
                    onChange={(_e, v) =>
                      setCatalog({
                        ...catalog,
                        hardware: { ...hw, utilization: Number(v) || 0.8 },
                      })
                    }
                  />
                </FormGroup>
                <FormGroup label="Margin" fieldId="margin">
                  <TextInput
                    id="margin"
                    type="number"
                    min={0}
                    step={0.05}
                    value={String(hw.margin)}
                    onChange={(_e, v) =>
                      setCatalog({ ...catalog, hardware: { ...hw, margin: Number(v) || 0 } })
                    }
                  />
                </FormGroup>
                <Button variant="secondary" isDisabled={saving} onClick={() => save(catalog)}>
                  Recalculate
                </Button>
                <Button
                  variant="primary"
                  isDisabled={saving || !(rec?.models || []).some((m) => m.canApply)}
                  onClick={async () => {
                    setSaving(true);
                    try {
                      await save(catalog);
                      const resp = await applyRecommend();
                      setCatalog(resp.catalog);
                      setRec(resp.recommend);
                      setSaved(`Applied to ${resp.applied} model(s)`);
                    } catch (e) {
                      setError((e as Error).message);
                    } finally {
                      setSaving(false);
                    }
                  }}
                >
                  Apply to catalog
                </Button>
              </Form>
              {(rec?.notes || []).map((n) => (
                <p key={n} className="mfp-muted">
                  {n}
                </p>
              ))}
              <Table aria-label="Recommendations" variant="compact">
                <Thead>
                  <Tr>
                    <Th>Model</Th>
                    <Th>Type</Th>
                    <Th>Machine / GPU</Th>
                    <Th>tok/s @ util</Th>
                    <Th>USD/hour</Th>
                    <Th>Price / 1M</Th>
                    <Th>Cost source</Th>
                  </Tr>
                </Thead>
                <Tbody>
                  {(rec?.models || []).map((m) => (
                    <Tr key={m.name}>
                      <Td>
                        <strong>{m.displayName || m.name}</strong>
                        <div className="mfp-muted">
                          {isExternalModel(m)
                            ? m.provider || m.kind
                            : `${m.paramsB}B ${m.quant}`}
                        </div>
                      </Td>
                      <Td>
                        <Label color={isExternalModel(m) ? 'orange' : 'blue'}>{originLabel(m)}</Label>
                      </Td>
                      <Td>
                        {isExternalModel(m) ? (
                          <>
                            {m.provider || 'external'}
                            <div className="mfp-muted">{m.endpoint || 'vendor token price'}</div>
                          </>
                        ) : (
                          <>
                            {m.instanceType}
                            <div className="mfp-muted">{m.gpuProduct}</div>
                          </>
                        )}
                      </Td>
                      <Td>
                        {isExternalModel(m) ? (
                          '—'
                        ) : (
                          <>
                            {m.tokensPerSecUsed.toFixed(0)}
                            <div className="mfp-muted">{m.tokensPerSecFull.toFixed(0)} at 100%</div>
                          </>
                        )}
                      </Td>
                      <Td>{isExternalModel(m) ? '—' : formatMoney(m.hourlyUsd, rec?.currency)}</Td>
                      <Td>
                        {m.canApply || m.pricePerMillion || m.inputPerMillion
                          ? formatModelPrice(m, rec?.currency)
                          : '—'}
                        {m.warning ? <div className="mfp-muted">{m.warning}</div> : null}
                      </Td>
                      <Td>
                        <Label>{m.cost.kind}</Label>
                        <div className="mfp-muted">{m.cost.note}</div>
                      </Td>
                    </Tr>
                  ))}
                </Tbody>
              </Table>
            </CardBody>
          </Card>

          <Card>
            <CardTitle>Manual machine cost</CardTitle>
            <CardBody>
              <p className="mfp-muted">
                Use this for reservations, Savings Plans, EA, or SKUs not in the catalog. A matching
                SKU overrides list price.
              </p>
              <Table aria-label="Manual machines" variant="compact">
                <Thead>
                  <Tr>
                    <Th>SKU</Th>
                    <Th>Region</Th>
                    <Th>USD/hour</Th>
                    <Th>GPU</Th>
                    <Th />
                  </Tr>
                </Thead>
                <Tbody>
                  {(hw.machines || []).map((m, idx) => (
                    <Tr key={`${m.instanceType}-${idx}`}>
                      <Td>{m.instanceType}</Td>
                      <Td>{m.region || '—'}</Td>
                      <Td>{formatMoney(m.hourlyUsd)}</Td>
                      <Td>
                        {m.gpuProduct || '—'} {m.gpuCount ? `×${m.gpuCount}` : ''}
                      </Td>
                      <Td>
                        <ActionsColumn
                          items={[
                            {
                              title: 'Remove',
                              onClick: () => {
                                const machines = hw.machines.filter((_, i) => i !== idx);
                                save({ ...catalog, hardware: { ...hw, machines } });
                              },
                            },
                          ]}
                        />
                      </Td>
                    </Tr>
                  ))}
                </Tbody>
              </Table>
              <Form className="mfp-inline pf-v6-u-mt-md">
                <FormGroup label="SKU" fieldId="man-sku">
                  <TextInput
                    id="man-sku"
                    value={manual.instanceType}
                    onChange={(_e, v) => setManual({ ...manual, instanceType: v })}
                  />
                </FormGroup>
                <FormGroup label="Region" fieldId="man-region">
                  <TextInput
                    id="man-region"
                    value={manual.region || ''}
                    onChange={(_e, v) => setManual({ ...manual, region: v })}
                  />
                </FormGroup>
                <FormGroup label="USD/hour" fieldId="man-usd">
                  <TextInput
                    id="man-usd"
                    type="number"
                    value={String(manual.hourlyUsd)}
                    onChange={(_e, v) => setManual({ ...manual, hourlyUsd: Number(v) || 0 })}
                  />
                </FormGroup>
                <Button
                  variant="secondary"
                  isDisabled={saving || !manual.instanceType || manual.hourlyUsd <= 0}
                  onClick={() =>
                    save({
                      ...catalog,
                      hardware: { ...hw, machines: [...(hw.machines || []), manual] },
                    })
                  }
                >
                  Save cost
                </Button>
              </Form>
            </CardBody>
          </Card>

          <Card>
            <CardTitle>Default</CardTitle>
            <CardBody>
              <Form isHorizontal>
                <FormGroup label="Currency" fieldId="currency">
                  <TextInput
                    id="currency"
                    value={catalog.currency}
                    onChange={(_e, v) => setCatalog({ ...catalog, currency: v.toUpperCase() })}
                  />
                </FormGroup>
                <FormGroup label="Default price / 1M tokens" fieldId="default-price">
                  <TextInput
                    id="default-price"
                    type="number"
                    value={String(catalog.defaultPricePerMillion)}
                    onChange={(_e, v) =>
                      setCatalog({ ...catalog, defaultPricePerMillion: Number(v) || 0 })
                    }
                  />
                </FormGroup>
                <Button variant="primary" isDisabled={saving} onClick={() => save(catalog)}>
                  Save default
                </Button>
              </Form>
            </CardBody>
          </Card>
          <Card>
            <CardTitle>Price per model</CardTitle>
            <CardBody>
              <Table aria-label="Price per model" variant="compact">
                <Thead>
                  <Tr>
                    <Th>Model</Th>
                    <Th>Display name</Th>
                    <Th>Price / 1M</Th>
                    <Th>Source</Th>
                    <Th />
                  </Tr>
                </Thead>
                <Tbody>
                  {rows.map(([name, price]) => (
                    <Tr key={name}>
                      <Td>{name}</Td>
                      <Td>{price.displayName || '—'}</Td>
                      <Td>{formatPrice(price.pricePerMillion, catalog.currency)}</Td>
                      <Td>{price.source || 'manual'}</Td>
                      <Td>
                        <ActionsColumn
                          items={[
                            {
                              title: 'Remove',
                              onClick: () => {
                                const models = { ...catalog.models };
                                delete models[name];
                                save({ ...catalog, models });
                              },
                            },
                          ]}
                        />
                      </Td>
                    </Tr>
                  ))}
                </Tbody>
              </Table>
              <Form className="mfp-inline pf-v6-u-mt-md">
                <FormGroup label="New model" fieldId="new-model">
                  <TextInput
                    id="new-model"
                    placeholder="llama-31-instruct"
                    value={newName}
                    onChange={(_e, v) => setNewName(v)}
                  />
                </FormGroup>
                <FormGroup label="Price / 1M" fieldId="new-price">
                  <TextInput
                    id="new-price"
                    type="number"
                    value={newPrice}
                    onChange={(_e, v) => setNewPrice(v)}
                  />
                </FormGroup>
                <Button
                  variant="secondary"
                  isDisabled={saving || !newName}
                  onClick={() => {
                    save({
                      ...catalog,
                      models: {
                        ...catalog.models,
                        [newName.trim()]: {
                          pricePerMillion: Number(newPrice) || 0,
                          source: 'manual',
                        },
                      },
                    });
                    setNewName('');
                  }}
                >
                  Add
                </Button>
              </Form>
            </CardBody>
          </Card>
        </div>
      ) : null}
    </FinOpsPage>
  );
};

export default PricingPage;
