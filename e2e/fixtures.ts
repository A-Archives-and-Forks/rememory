// Custom Playwright fixtures that enforce offline-by-default for all tests.
//
// Every test gets HTTP/HTTPS requests blocked automatically. This guarantees
// that recovery bundles, maker.html, and static pages never phone home.
//
// Tests that legitimately need network access opt in with an allowlist:
//   test.use({ allowedHosts: ['127.0.0.1'] });        // localhost server tests
//   test.use({ allowedHosts: ['api.drand.sh'] });      // drand beacon tests

import { test as base } from '@playwright/test';

export { expect } from '@playwright/test';
export type { Page } from '@playwright/test';

type OfflineFixtures = {
  allowedHosts: string[];
};

export const test = base.extend<OfflineFixtures>({
  allowedHosts: [[], { option: true }],

  page: async ({ page, allowedHosts }, use) => {
    await page.route(/^https?:\/\//, route => {
      if (allowedHosts.length > 0) {
        const url = new URL(route.request().url());
        if (allowedHosts.includes(url.hostname)) {
          return route.continue();
        }
      }
      const reqUrl = route.request().url();
      const hint = allowedHosts.length > 0
        ? `Allowed hosts: ${allowedHosts.join(', ')}`
        : `Add test.use({ allowedHosts: ['hostname'] }) to the describe block.`;
      throw new Error(
        `Unexpected network request: ${reqUrl}\n` +
        `This test should be fully offline. ${hint}`
      );
    });
    await use(page);
  },
});
