import {
  maasFinopsAreaExtension,
  communityPluginsSectionExtension,
  maasFinopsSectionExtension,
  overviewNavExtension,
  modelsNavExtension,
  subscriptionsNavExtension,
  apiKeysNavExtension,
  pricingNavExtension,
  simulatorNavExtension,
  maasFinopsRouteExtension,
  extensions,
} from '../extensions';

describe('RHOAI Plugin Extensions', () => {
  describe('maasFinopsAreaExtension', () => {
    it('should have the correct type and id', () => {
      expect(maasFinopsAreaExtension.type).toBe('app.area');
      expect(maasFinopsAreaExtension.properties.id).toBe('maas-finops');
    });

    it('should have an empty featureFlags array', () => {
      expect(maasFinopsAreaExtension.properties.featureFlags).toEqual([]);
    });
  });

  describe('communityPluginsSectionExtension', () => {
    it('should define the community-plugins section', () => {
      expect(communityPluginsSectionExtension.type).toBe('app.navigation/section');
      expect(communityPluginsSectionExtension.properties.id).toBe('community-plugins');
      expect(communityPluginsSectionExtension.properties.title).toBe('Community plugins');
      expect(communityPluginsSectionExtension.properties.group).toBe('9_plugins');
    });

    it('should have an iconRef function', () => {
      expect(typeof communityPluginsSectionExtension.properties.iconRef).toBe('function');
    });
  });

  describe('maasFinopsSectionExtension', () => {
    it('should define a subsection nested under community-plugins', () => {
      expect(maasFinopsSectionExtension.type).toBe('app.navigation/section');
      expect(maasFinopsSectionExtension.properties.id).toBe('maas-finops');
      expect(maasFinopsSectionExtension.properties.title).toBe('MaaS FinOps');
      expect(maasFinopsSectionExtension.properties.group).toBe('1_maas_finops');
      expect(maasFinopsSectionExtension.properties.section).toBe('community-plugins');
      expect(typeof maasFinopsSectionExtension.properties.iconRef).toBe('function');
    });
  });

  describe('navigation extensions', () => {
    it.each([
      [overviewNavExtension, 'maas-finops-overview', 'Overview', '/maas-finops/overview'],
      [modelsNavExtension, 'maas-finops-models', 'Models', '/maas-finops/models'],
      [subscriptionsNavExtension, 'maas-finops-subscriptions', 'Subscriptions', '/maas-finops/subscriptions'],
      [apiKeysNavExtension, 'maas-finops-api-keys', 'API keys', '/maas-finops/api-keys'],
      [pricingNavExtension, 'maas-finops-pricing', 'Pricing', '/maas-finops/pricing'],
      [simulatorNavExtension, 'maas-finops-simulator', 'Pricing Simulator', '/maas-finops/simulator'],
    ])('should define %s nav item', (ext, id, title, href) => {
      expect(ext.type).toBe('app.navigation/href');
      expect(ext.properties.id).toBe(id);
      expect(ext.properties.title).toBe(title);
      expect(ext.properties.href).toBe(href);
      expect(ext.properties.section).toBe('maas-finops');
      expect(ext.properties.path).toBe(`${href}/*`);
    });
  });

  describe('route extension', () => {
    it('should define a single wildcard route with lazy component', () => {
      expect(maasFinopsRouteExtension.type).toBe('app.route');
      expect(maasFinopsRouteExtension.properties.path).toBe('/maas-finops/*');
      expect(typeof maasFinopsRouteExtension.properties.component).toBe('function');
      expect(maasFinopsRouteExtension.properties.component()).toBeInstanceOf(Promise);
    });
  });

  describe('extensions array', () => {
    it('should contain all ten extensions', () => {
      expect(extensions).toHaveLength(10);
    });

    it('should include all extensions in the correct order', () => {
      expect(extensions).toEqual([
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
      ]);
    });
  });
});
