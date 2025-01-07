import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import App from './App';

describe('App', () => {
  it('renders without crashing', () => {
    render(<App />);
    // The Calendar component should be rendered within the App
    expect(document.querySelector('.flex')).toBeInTheDocument();
  });

  it('has the correct layout classes', () => {
    render(<App />);
    const mainContainer = document.querySelector('.p-0');
    const flexContainer = document.querySelector('.flex');

    expect(mainContainer).toBeInTheDocument();
    expect(flexContainer).toHaveClass('justify-between', 'items-center', 'mb-4');
  });
});