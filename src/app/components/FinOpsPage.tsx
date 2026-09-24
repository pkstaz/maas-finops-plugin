import React from 'react';
import { Alert, PageSection, Spinner, Title, ToggleGroup, ToggleGroupItem } from '@patternfly/react-core';
import { RangeKey } from '~/app/utils/api';
import { RANGES } from '~/app/utils/format';
import './finops.css';

type Props = {
  title: string;
  range?: RangeKey;
  onRangeChange?: (range: RangeKey) => void;
  loading?: boolean;
  error?: string;
  metricsError?: string;
  source?: string;
  children: React.ReactNode;
};

const FinOpsPage: React.FC<Props> = ({
  title,
  range,
  onRangeChange,
  loading,
  error,
  metricsError,
  source,
  children,
}) => (
  <>
    <PageSection>
      <div className="mfp-header">
        <div>
          <Title headingLevel="h1">{title}</Title>
          {source ? (
            <p className="mfp-muted">
              Source: {source === 'mock' ? 'demo data (MOCK_DATA)' : 'cluster / Prometheus'}
            </p>
          ) : null}
        </div>
        {range && onRangeChange ? (
          <ToggleGroup aria-label="Metrics range">
            {RANGES.map((item) => (
              <ToggleGroupItem
                key={item.key}
                text={item.label}
                isSelected={range === item.key}
                onChange={() => onRangeChange(item.key)}
              />
            ))}
          </ToggleGroup>
        ) : null}
      </div>
    </PageSection>
    <PageSection>
      {error ? (
        <Alert variant="danger" title="Could not load metrics" isInline>
          {error}
        </Alert>
      ) : null}
      {metricsError ? (
        <Alert variant="warning" title="Thanos did not return all token series" isInline className="pf-v6-u-mb-md">
          {metricsError}
        </Alert>
      ) : null}
      {loading ? (
        <div className="mfp-loading">
          <Spinner />
        </div>
      ) : (
        children
      )}
    </PageSection>
  </>
);

export default FinOpsPage;
