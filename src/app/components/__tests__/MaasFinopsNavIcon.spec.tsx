import React from 'react';
import { render } from '@testing-library/react';
import MaasFinopsIcon from '../MaasFinopsNavIcon';

describe('MaasFinopsNavIcon (MaasFinopsIcon) Component', () => {
  it('should render the SVG icon', () => {
    const { container } = render(<MaasFinopsIcon />);
    const svg = container.querySelector('svg');
    expect(svg).toBeInTheDocument();
  });

  it('should use PatternFly SVG conventions', () => {
    const { container } = render(<MaasFinopsIcon />);
    const svg = container.querySelector('svg');
    expect(svg).toHaveAttribute('width', '1em');
    expect(svg).toHaveAttribute('height', '1em');
    expect(svg).toHaveAttribute('viewBox', '0 0 32 32');
    expect(svg).toHaveClass('pf-v6-svg');
    expect(svg).toHaveAttribute('role', 'img');
    expect(svg).toHaveAttribute('aria-hidden', 'true');
  });

  it('should render the MF text', () => {
    const { container } = render(<MaasFinopsIcon />);
    const textElement = container.querySelector('text');
    expect(textElement).toBeInTheDocument();
    expect(textElement?.textContent).toBe('MF');
  });

  it('should render the rectangle background with purple fill', () => {
    const { container } = render(<MaasFinopsIcon />);
    const rect = container.querySelector('rect');
    expect(rect).toBeInTheDocument();
    expect(rect).toHaveAttribute('fill', '#c9190b');
  });

  it('should have white bold centered text', () => {
    const { container } = render(<MaasFinopsIcon />);
    const text = container.querySelector('text');
    expect(text).toHaveAttribute('fill', 'white');
    expect(text).toHaveAttribute('font-weight', 'bold');
    expect(text).toHaveAttribute('text-anchor', 'middle');
  });
});
