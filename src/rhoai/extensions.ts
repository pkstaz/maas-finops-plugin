// [SHARED] Common section for all community plugins — never changes across plugins.
// Do not change the id or name: all community plugins share this section
// so they appear grouped together in the dashboard sidebar.
export const communityPluginsSectionExtension = {
  type: 'app.navigation/section' as const,
  properties: {
    id: 'community-plugins', // [SHARED] common section for all community plugins
    title: 'Community plugins', // [SHARED]
    group: '9_plugins', // [SHARED]
    iconRef: () => import(/* webpackMode: "eager" */ './CommunityNavIcon'),
  },
};

// [PLUGIN-SPECIFIC] Everything below is specific to this plugin

export const maasFinopsAreaExtension = {
  type: 'app.area' as const,
  properties: {
    id: 'maas-finops', // [PLUGIN-SPECIFIC] unique area ID
    featureFlags: [] as string[],
  },
};

export const maasFinopsSectionExtension = {
  type: 'app.navigation/section' as const,
  properties: {
    id: 'maas-finops', // [PLUGIN-SPECIFIC] unique nav section ID
    title: 'MaaS FinOps', // [PLUGIN-SPECIFIC] display name in sidebar
    group: '1_maas_finops', // [PLUGIN-SPECIFIC] sort key within community-plugins
    section: 'community-plugins', // [SHARED] must match communityPluginsSectionExtension.id — do not change
    iconRef: () => import(/* webpackMode: "eager" */ '~/app/components/MaasFinopsNavIcon'),
  },
};

export const overviewNavExtension = {
  type: 'app.navigation/href' as const,
  properties: {
    id: 'maas-finops-overview',
    title: 'Overview',
    href: '/maas-finops/overview',
    section: 'maas-finops',
    path: '/maas-finops/overview/*',
  },
};

export const modelsNavExtension = {
  type: 'app.navigation/href' as const,
  properties: {
    id: 'maas-finops-models',
    title: 'Models',
    href: '/maas-finops/models',
    section: 'maas-finops',
    path: '/maas-finops/models/*',
  },
};

export const subscriptionsNavExtension = {
  type: 'app.navigation/href' as const,
  properties: {
    id: 'maas-finops-subscriptions',
    title: 'Subscriptions',
    href: '/maas-finops/subscriptions',
    section: 'maas-finops',
    path: '/maas-finops/subscriptions/*',
  },
};

export const apiKeysNavExtension = {
  type: 'app.navigation/href' as const,
  properties: {
    id: 'maas-finops-api-keys',
    title: 'API keys',
    href: '/maas-finops/api-keys',
    section: 'maas-finops',
    path: '/maas-finops/api-keys/*',
  },
};

export const pricingNavExtension = {
  type: 'app.navigation/href' as const,
  properties: {
    id: 'maas-finops-pricing',
    title: 'Pricing',
    href: '/maas-finops/pricing',
    section: 'maas-finops',
    path: '/maas-finops/pricing/*',
  },
};

export const simulatorNavExtension = {
  type: 'app.navigation/href' as const,
  properties: {
    id: 'maas-finops-simulator',
    title: 'Pricing Simulator',
    href: '/maas-finops/simulator',
    section: 'maas-finops',
    path: '/maas-finops/simulator/*',
  },
};

export const maasFinopsRouteExtension = {
  type: 'app.route' as const,
  properties: {
    path: '/maas-finops/*',
    component: () => import(/* webpackMode: "eager" */ '~/app/App'),
  },
};

export const extensions = [
  communityPluginsSectionExtension,
  maasFinopsAreaExtension,
  maasFinopsSectionExtension,
  overviewNavExtension,
  modelsNavExtension,
  subscriptionsNavExtension,
  apiKeysNavExtension,
  pricingNavExtension,
  simulatorNavExtension,
  maasFinopsRouteExtension,
];

export default extensions;
