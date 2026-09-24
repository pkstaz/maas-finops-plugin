import { render, screen } from '@testing-library/react';
import App from '../App';

jest.mock('../pages/OverviewPage', () => {
  const MockPage = () => <div data-testid="overview-page">Overview Page</div>;
  MockPage.displayName = 'MockOverviewPage';
  return { __esModule: true, default: MockPage };
});

jest.mock('../pages/ModelsPage', () => {
  const MockPage = () => <div data-testid="models-page">Models Page</div>;
  MockPage.displayName = 'MockModelsPage';
  return { __esModule: true, default: MockPage };
});

jest.mock('../pages/SubscriptionsPage', () => {
  const MockPage = () => <div data-testid="subscriptions-page">Subscriptions Page</div>;
  MockPage.displayName = 'MockSubscriptionsPage';
  return { __esModule: true, default: MockPage };
});

jest.mock('../pages/ApiKeysPage', () => {
  const MockPage = () => <div data-testid="api-keys-page">API Keys Page</div>;
  MockPage.displayName = 'MockApiKeysPage';
  return { __esModule: true, default: MockPage };
});

jest.mock('../pages/PricingPage', () => {
  const MockPage = () => <div data-testid="pricing-page">Pricing Page</div>;
  MockPage.displayName = 'MockPricingPage';
  return { __esModule: true, default: MockPage };
});

jest.mock('../pages/SimulatorPage', () => {
  const MockPage = () => <div data-testid="simulator-page">Simulator Page</div>;
  MockPage.displayName = 'MockSimulatorPage';
  return { __esModule: true, default: MockPage };
});

describe('App Component', () => {
  it('should render the first route element', () => {
    render(<App />);
    expect(screen.getByTestId('routes')).toBeInTheDocument();
  });
});
