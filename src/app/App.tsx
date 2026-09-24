import React from 'react';
import { Routes, Route, Navigate } from 'react-router-dom';
import CommunityBanner from './components/CommunityBanner';
import OverviewPage from './pages/OverviewPage';
import ModelsPage from './pages/ModelsPage';
import SubscriptionsPage from './pages/SubscriptionsPage';
import ApiKeysPage from './pages/ApiKeysPage';
import PricingPage from './pages/PricingPage';
import SimulatorPage from './pages/SimulatorPage';

const App: React.FC = () => (
  <div className="community-plugin-layout">
    {/* [SHARED] Do not remove — all community plugins must display the CommunityBanner */}
    <CommunityBanner />
    <div className="community-plugin-content">
      <Routes>
        <Route path="/" element={<Navigate to="overview" replace />} />
        <Route path="overview/*" element={<OverviewPage />} />
        <Route path="models/*" element={<ModelsPage />} />
        <Route path="subscriptions/*" element={<SubscriptionsPage />} />
        <Route path="api-keys/*" element={<ApiKeysPage />} />
        <Route path="pricing/*" element={<PricingPage />} />
        <Route path="simulator/*" element={<SimulatorPage />} />
      </Routes>
    </div>
  </div>
);

export default App;
