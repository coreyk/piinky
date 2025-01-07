import '@testing-library/jest-dom';
import { expect, afterEach } from 'vitest';
import { cleanup } from '@testing-library/react';
import * as matchers from '@testing-library/jest-dom/matchers';

// Extend Vitest's expect method with testing-library matchers
expect.extend(matchers);

// Mock timezone to UTC
process.env.TZ = 'UTC';

// Cleanup after each test case
afterEach(() => {
  cleanup();
});