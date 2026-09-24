export type RangeKey = '1h' | '6h' | '24h' | '7d';

export interface UserInfo {
  username: string;
  groups: string[];
  is_admin: boolean;
}

export interface ModelPrice {
  displayName?: string;
  pricePerMillion: number;
  inputPerMillion?: number;
  outputPerMillion?: number;
  source?: string;
}

export interface ManualMachine {
  instanceType: string;
  region?: string;
  hourlyUsd: number;
  gpuProduct?: string;
  gpuCount?: number;
  gpuMemoryGib?: number;
  notes?: string;
}

export interface HardwareConfig {
  provider: string;
  utilization: number;
  margin: number;
  priceType: string;
  machines: ManualMachine[];
}

export interface PricingCatalog {
  currency: string;
  defaultPricePerMillion: number;
  models: Record<string, ModelPrice>;
  hardware: HardwareConfig;
}

export interface ModelRow {
  name: string;
  namespace: string;
  displayName: string;
  phase: string;
  kind: string;
  origin?: string;
  provider?: string;
  endpoint?: string;
  targetModel?: string;
  tokensIn: number;
  tokensOut: number;
  tokens: number;
  requests: number;
  limited: number;
  pricePerMillion: number;
  inputPerMillion?: number;
  outputPerMillion?: number;
  cost: number;
  priced: boolean;
}

export interface SubscriptionRow {
  name: string;
  namespace: string;
  priority: number;
  models: string[];
  rateLimits: string[];
  tokensIn: number;
  tokensOut: number;
  tokens: number;
  requests: number;
  limited: number;
  cost: number;
}

export interface ApiKeyRow {
  user: string;
  tokensIn: number;
  tokensOut: number;
  tokens: number;
  requests: number;
  limited: number;
  cost: number;
  models: string[];
  subscriptions: string[];
}

export interface UsageBreakdown {
  user?: string;
  subscription?: string;
  model?: string;
  tokensIn: number;
  tokensOut: number;
  tokens: number;
  requests: number;
  limited: number;
  cost: number;
}

export interface OverviewResponse {
  range: string;
  currency: string;
  totalTokensIn: number;
  totalTokensOut: number;
  totalTokens: number;
  totalCost: number;
  totalRequests: number;
  totalLimited: number;
  unpricedModels: number;
  source: string;
  metricsError?: string;
  topModels: ModelRow[];
  byModel: UsageBreakdown[];
  bySubscription: UsageBreakdown[];
  byUser: UsageBreakdown[];
}

export interface ModelsResponse {
  range: string;
  currency: string;
  source: string;
  metricsError?: string;
  items: ModelRow[];
}

export interface SubscriptionsResponse {
  range: string;
  currency: string;
  source: string;
  metricsError?: string;
  items: SubscriptionRow[];
}

export interface ApiKeysResponse {
  range: string;
  currency: string;
  source: string;
  note: string;
  metricsError?: string;
  items: ApiKeyRow[];
}

export interface NodeInventory {
  name: string;
  role: string;
  instanceType: string;
  region: string;
  zone: string;
  gpuProduct: string;
  gpuFamily: string;
  gpuCount: number;
  gpuMemoryMib: number;
}

export interface ServingModel {
  name: string;
  namespace: string;
  displayName: string;
  uri: string;
  quant: string;
  paramsB: number;
  gpuRequest: number;
  maxModelLen: number;
  nodeName: string;
  kind?: string;
  origin?: string;
  provider?: string;
  endpoint?: string;
  targetModel?: string;
}

export interface ClusterInventory {
  provider: string;
  platform: string;
  region: string;
  cloud: string;
  source: string;
  supported: boolean;
  message?: string;
  nodes: NodeInventory[];
  models: ServingModel[];
}

export interface CostSource {
  hourlyUsd: number;
  kind: string;
  sku: string;
  region: string;
  meter?: string;
  note?: string;
}

export interface ModelRecommendation {
  name: string;
  displayName: string;
  namespace: string;
  origin?: string;
  kind?: string;
  provider?: string;
  endpoint?: string;
  instanceType: string;
  gpuProduct: string;
  quant: string;
  paramsB: number;
  tokensPerSecFull: number;
  tokensPerSecUsed: number;
  tokensPerHour: number;
  hourlyUsd: number;
  pricePerMillion: number;
  inputPerMillion?: number;
  outputPerMillion?: number;
  cost: CostSource;
  assumptions: string[];
  canApply: boolean;
  warning?: string;
}

export interface RecommendResponse {
  inventory: ClusterInventory;
  utilization: number;
  margin: number;
  currency: string;
  models: ModelRecommendation[];
  needsManual: boolean;
  notes: string[];
}

const API_BASE = '/maas-finops/api';

async function request<T>(path: string, options?: RequestInit): Promise<T> {
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
    ...(options?.headers as Record<string, string>),
  };
  const resp = await fetch(`${API_BASE}${path}`, {
    credentials: 'include',
    ...options,
    headers,
  });
  if (!resp.ok) {
    const err = await resp.json().catch(() => ({ error: resp.statusText }));
    throw new Error(err.error || resp.statusText);
  }
  return resp.json();
}

export const getMe = () => request<UserInfo>('/auth/me');
export const getOverview = (range: RangeKey) => request<OverviewResponse>(`/overview?range=${range}`);
export const getModels = (range: RangeKey) => request<ModelsResponse>(`/models?range=${range}`);
export const getSubscriptions = (range: RangeKey) =>
  request<SubscriptionsResponse>(`/subscriptions?range=${range}`);
export const getApiKeys = (range: RangeKey) => request<ApiKeysResponse>(`/apikeys?range=${range}`);
export const getPricing = () => request<PricingCatalog>('/pricing');
export const putPricing = (catalog: PricingCatalog) =>
  request<PricingCatalog>('/pricing', { method: 'PUT', body: JSON.stringify(catalog) });
export const getRecommend = (utilization?: number) =>
  request<RecommendResponse>(
    `/recommend${utilization ? `?utilization=${utilization}` : ''}`,
  );
export const applyRecommend = () =>
  request<{ applied: number; catalog: PricingCatalog; recommend: RecommendResponse }>(
    '/recommend/apply',
    { method: 'POST' },
  );

export interface CatalogSKU {
  instanceType: string;
  hourlyUsd: number;
  gpuProduct: string;
  gpuCount: number;
  gpuMemoryGib: number;
  region?: string;
  note?: string;
}

export interface ProviderCatalog {
  id: string;
  label: string;
  skus: CatalogSKU[];
}

export interface CatalogModel {
  id: string;
  displayName: string;
  family: string;
  paramsB: number;
  quant: string;
  uri: string;
  source: string;
}

export interface SimulatorCurrent {
  provider: string;
  sku: string;
  region: string;
  gpuProduct: string;
  gpuCount: number;
  gpuMemoryGib: number;
  thisCluster: boolean;
}

export interface SimulatorCatalogResponse {
  inventory: ClusterInventory;
  current: SimulatorCurrent;
  providers: ProviderCatalog[];
  models: CatalogModel[];
  currency: string;
}

export interface SimulatorQuote {
  provider: string;
  sku: CatalogSKU;
  model: CatalogModel;
  utilization: number;
  tokensPerSecFull: number;
  tokensPerSecUsed: number;
  tokensPerHour: number;
  hourlyUsd: number;
  pricePerMillion: number;
  vramGib: number;
  fits: boolean;
  cost: CostSource;
  assumptions: string[];
  warning?: string;
}

export const getSimulatorCatalog = () => request<SimulatorCatalogResponse>('/simulator');

export const getSimulatorQuote = (params: {
  provider: string;
  sku: string;
  model: string;
  utilization?: number;
  hourlyUsd?: number;
  gpuProduct?: string;
  gpuCount?: number;
  gpuMemoryGib?: number;
}) => {
  const q = new URLSearchParams();
  q.set('provider', params.provider);
  q.set('sku', params.sku);
  q.set('model', params.model);
  if (params.utilization) {
    q.set('utilization', String(params.utilization));
  }
  if (params.hourlyUsd && params.hourlyUsd > 0) {
    q.set('hourlyUsd', String(params.hourlyUsd));
  }
  if (params.gpuProduct) {
    q.set('gpuProduct', params.gpuProduct);
  }
  if (params.gpuCount) {
    q.set('gpuCount', String(params.gpuCount));
  }
  if (params.gpuMemoryGib) {
    q.set('gpuMemoryGib', String(params.gpuMemoryGib));
  }
  return request<SimulatorQuote>(`/simulator/quote?${q.toString()}`);
};
