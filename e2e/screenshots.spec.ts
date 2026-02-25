/**
 * Automated guide screenshots.
 *
 * Captures 10 screenshots per guide language (en, es, de, fr) from the
 * creation and recovery flows, saving them to docs/screenshots/{lang}/.
 *
 * Three screenshots are not automatable (OS dialogs / camera) and stay
 * shared at docs/screenshots/:
 *   qr-camera-permission.png, qr-scanning.png, manifest-file-picker.png
 *
 * Usage:
 *   make screenshots          # or:
 *   REMEMORY_BIN=./rememory npx playwright test e2e/screenshots.spec.ts --project=chromium
 */

import { test, expect, Page } from './fixtures';
import { execFileSync } from 'child_process';
import * as fs from 'fs';
import * as path from 'path';
import * as os from 'os';
import {
  getRememoryBin,
  generateStandaloneHTML,
  extractBundle,
  extractWordsFromReadme,
  findReadmeFile,
  RecoveryPage,
  CreationPage,
} from './helpers';

// ---------------------------------------------------------------------------
// Config
// ---------------------------------------------------------------------------

const LANGUAGES = ['en', 'es', 'de', 'fr'] as const;
type Lang = typeof LANGUAGES[number];

const SCREENSHOTS_ROOT = path.resolve(__dirname, '..', 'docs', 'screenshots');

// Viewport: wide enough for clean layouts, tall enough to avoid clipping long cards.
const VIEWPORT = { width: 1280, height: 2000 };

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

/** Save a screenshot cropped tightly to visible cards. */
async function snap(page: Page, lang: Lang, name: string): Promise<void> {
  const dir = path.join(SCREENSHOTS_ROOT, lang);
  fs.mkdirSync(dir, { recursive: true });

  // Remove overflow:hidden from cards so content isn't clipped, then measure bounds.
  const bounds = await page.evaluate(() => {
    const cards = document.querySelectorAll('.container > .card');
    // Remove overflow clipping and force reflow
    for (const card of cards) {
      (card as HTMLElement).style.overflow = 'visible';
    }
    document.body.offsetHeight; // force reflow

    let minX = Infinity, minY = Infinity, maxX = 0, maxY = 0;
    for (const card of cards) {
      if (window.getComputedStyle(card).display === 'none') continue;
      const rect = card.getBoundingClientRect();
      if (rect.height === 0) continue;
      minX = Math.min(minX, rect.left);
      minY = Math.min(minY, rect.top);
      maxX = Math.max(maxX, rect.right);
      maxY = Math.max(maxY, rect.bottom);
    }
    return { x: minX, y: minY, width: maxX - minX, height: maxY - minY };
  });

  const pad = 16;
  await page.screenshot({
    path: path.join(dir, `${name}.png`),
    clip: {
      x: bounds.x - pad,
      y: bounds.y - pad,
      width: bounds.width + pad * 2,
      height: bounds.height + pad * 2,
    },
  });
}

/** Hide elements matching a CSS selector. */
async function hide(page: Page, selector: string): Promise<void> {
  await page.evaluate((sel) => {
    for (const el of document.querySelectorAll(sel)) {
      (el as HTMLElement).style.display = 'none';
    }
  }, selector);
}

/** Hide nav, footer, and page intro — common to all screenshots. */
async function hideChrome(page: Page): Promise<void> {
  await hide(page, '.site-nav, footer, .page-intro');
}

/**
 * Prepare a maker.html screenshot: hide chrome and all .card step sections
 * except the ones at the given indices (0-based).
 */
async function frameCreationStep(page: Page, steps: number[]): Promise<void> {
  await page.evaluate((steps) => {
    // Hide nav, footer, page intro
    for (const sel of ['.site-nav', 'footer', '.page-intro']) {
      const el = document.querySelector(sel) as HTMLElement;
      if (el) el.style.display = 'none';
    }
    // Show only the requested step cards
    const cards = document.querySelectorAll('.container > .card');
    cards.forEach((card, i) => {
      (card as HTMLElement).style.display = steps.includes(i) ? '' : 'none';
    });
  }, steps);
}

/**
 * Prepare a recover.html screenshot: hide chrome and all .card step sections
 * except the ones at the given indices (0-based).
 */
async function frameRecoveryStep(page: Page, steps: number[]): Promise<void> {
  await page.evaluate((steps) => {
    // Hide nav, footer, page intro
    for (const sel of ['.site-nav', 'footer', '.page-intro']) {
      const el = document.querySelector(sel) as HTMLElement;
      if (el) el.style.display = 'none';
    }
    // Show only the requested step cards
    const cards = document.querySelectorAll('.container > .card');
    cards.forEach((card, i) => {
      (card as HTMLElement).style.display = steps.includes(i) ? '' : 'none';
    });
  }, steps);
}

// ---------------------------------------------------------------------------
// Project setup — 5 friends, threshold 3
// ---------------------------------------------------------------------------

let projectDir: string;
let bundlesDir: string;
let makerHtmlPath: string;
let standaloneRecoverHtml: string;

// Per-language bundle dirs for translated word screenshots
const langBundlesDirs: Partial<Record<Lang, string>> = {};

test.beforeAll(async () => {
  const bin = getRememoryBin();
  if (!fs.existsSync(bin)) {
    test.skip();
    return;
  }

  // Create a project with 5 friends for richer screenshots.
  const tmpDir = fs.mkdtempSync(path.join(os.tmpdir(), 'rememory-screenshots-'));
  projectDir = path.join(tmpDir, 'screenshot-project');

  execFileSync(bin, [
    'init', projectDir, '--name', 'Family Recovery',
    '--threshold', '3',
    '--friend', 'Alice,alice@example.com',
    '--friend', 'Bob,bob@example.com',
    '--friend', 'Camila,camila@example.com',
    '--friend', 'David,david@example.com',
    '--friend', 'Elena,elena@example.com',
  ], { stdio: 'inherit' });

  // Add secret files
  const manifestDir = path.join(projectDir, 'manifest');
  fs.writeFileSync(path.join(manifestDir, 'passwords.txt'), 'bank: correct-horse-battery-staple');
  fs.writeFileSync(path.join(manifestDir, 'notes.txt'), 'Safe deposit box is at First National, box 4217.');

  // Seal and bundle
  execFileSync(bin, ['seal'], { cwd: projectDir, stdio: 'inherit' });
  execFileSync(bin, ['bundle'], { cwd: projectDir, stdio: 'inherit' });
  bundlesDir = path.join(projectDir, 'output', 'bundles');

  // Create per-language projects so word screenshots show translated BIP39 words
  for (const lang of LANGUAGES) {
    const langTmpDir = fs.mkdtempSync(path.join(os.tmpdir(), `rememory-ss-${lang}-`));
    const langProjectDir = path.join(langTmpDir, `words-${lang}`);
    execFileSync(bin, [
      'init', langProjectDir, '--name', 'Words',
      '--threshold', '2',
      '--friend', `Alice,alice@example.com,${lang}`,
      '--friend', `Bob,bob@example.com,${lang}`,
    ], { stdio: 'inherit' });
    const langManifestDir = path.join(langProjectDir, 'manifest');
    fs.writeFileSync(path.join(langManifestDir, 'secret.txt'), 'test secret');
    execFileSync(bin, ['seal'], { cwd: langProjectDir, stdio: 'inherit' });
    execFileSync(bin, ['bundle'], { cwd: langProjectDir, stdio: 'inherit' });
    langBundlesDirs[lang] = path.join(langProjectDir, 'output', 'bundles');
  }

  // Generate standalone HTML files
  const htmlTmpDir = fs.mkdtempSync(path.join(os.tmpdir(), 'rememory-ss-html-'));
  makerHtmlPath = generateStandaloneHTML(htmlTmpDir, 'create');
  standaloneRecoverHtml = generateStandaloneHTML(htmlTmpDir, 'recover');
});

test.afterAll(async () => {
  if (projectDir && fs.existsSync(projectDir)) {
    fs.rmSync(path.dirname(projectDir), { recursive: true, force: true });
  }
});

// ---------------------------------------------------------------------------
// Use retina-quality screenshots, Chromium only
// ---------------------------------------------------------------------------

test.use({
  viewport: VIEWPORT,
  deviceScaleFactor: 2,
});

// ---------------------------------------------------------------------------
// Screenshot capture — one describe block per language
// ---------------------------------------------------------------------------

for (const lang of LANGUAGES) {
  test.describe(`Screenshots — ${lang}`, () => {
    // Give WASM operations plenty of time
    test.setTimeout(180_000);

    // -----------------------------------------------------------------------
    // Creation flow (maker.html)
    // -----------------------------------------------------------------------

    test(`[${lang}] friends`, async ({ page }) => {
      const creation = new CreationPage(page, makerHtmlPath);
      await creation.open();
      await creation.setLanguage(lang);

      // Set up 5 friends
      await creation.setFriend(0, 'Alice', 'alice@example.com');
      await creation.setFriend(1, 'Bob', 'bob@example.com');
      await creation.addFriend();
      await creation.setFriend(2, 'Camila', 'camila@example.com');
      await creation.addFriend();
      await creation.setFriend(3, 'David', 'david@example.com');
      await creation.addFriend();
      await creation.setFriend(4, 'Elena', 'elena@example.com');

      await creation.setThreshold(3);

      // Frame: only Step 1 (friends)
      await frameCreationStep(page, [0]);

      await snap(page, lang, 'friends');
    });

    test(`[${lang}] files`, async ({ page }) => {
      const creation = new CreationPage(page, makerHtmlPath);
      await creation.open();
      await creation.setLanguage(lang);

      // Minimal friend setup (required for file preview)
      await creation.setFriend(0, 'Alice', 'alice@example.com');
      await creation.setFriend(1, 'Bob', 'bob@example.com');

      // Add files
      const tmpDir = fs.mkdtempSync(path.join(os.tmpdir(), 'ss-files-'));
      const testFiles = creation.createTestFiles(tmpDir, 'guide');
      await creation.addFiles(testFiles);

      // Frame: only Step 2 (files)
      await frameCreationStep(page, [1]);

      await snap(page, lang, 'files');
    });

    test(`[${lang}] bundles`, async ({ page }) => {
      const creation = new CreationPage(page, makerHtmlPath);
      await creation.open();
      await creation.setLanguage(lang);

      // Set up friends
      await creation.setFriend(0, 'Alice', 'alice@example.com');
      await creation.setFriend(1, 'Bob', 'bob@example.com');
      await creation.addFriend();
      await creation.setFriend(2, 'Camila', 'camila@example.com');

      await creation.setThreshold(2);

      // Add files
      const tmpDir = fs.mkdtempSync(path.join(os.tmpdir(), 'ss-bundles-'));
      const testFiles = creation.createTestFiles(tmpDir, 'bundle');
      await creation.addFiles(testFiles);

      // Generate and wait for completion
      await creation.generate();
      await expect(page.locator('#status-message.success')).toBeAttached({ timeout: 120000 });
      await expect(page.locator('#generate-btn.btn-secondary')).toBeAttached({ timeout: 5000 });

      // Frame: only Step 3 (bundles), hide progress bar
      await frameCreationStep(page, [2]);
      await hide(page, '#progress-bar');

      await snap(page, lang, 'bundles');
    });

    test(`[${lang}] multilingual-language-dropdown`, async ({ page }) => {
      const creation = new CreationPage(page, makerHtmlPath);
      await creation.open();
      await creation.setLanguage(lang);

      // Set up friends
      await creation.setFriend(0, 'Alice', 'alice@example.com');
      await creation.setFriend(1, 'Bob', 'bob@example.com');
      await creation.addFriend();
      await creation.setFriend(2, 'Camila', 'camila@example.com');

      // Enable custom language mode
      await page.locator('#custom-language-mode').check();

      // Set different languages for each friend
      const friendLangs = ['en', 'es', 'fr'];
      for (let i = 0; i < friendLangs.length; i++) {
        await page.locator('.friend-entry').nth(i).locator('.friend-language').selectOption(friendLangs[i]);
      }

      // Frame: only Step 1 (friends with language dropdowns)
      await frameCreationStep(page, [0]);

      await snap(page, lang, 'multilingual-language-dropdown');
    });

    test(`[${lang}] tlock-setup`, async ({ page }) => {
      const creation = new CreationPage(page, makerHtmlPath);
      await creation.open();
      await creation.setLanguage(lang);

      // Switch to Advanced mode and enable time lock
      await page.locator('.mode-tab[data-mode="advanced"]').click();
      await page.locator('#timelock-checkbox').check();

      // Wait for the date preview to appear
      await expect(page.locator('#timelock-date-preview')).not.toBeEmpty();

      // Frame: only Step 3 (generate bundles with tlock panel)
      await frameCreationStep(page, [2]);

      await snap(page, lang, 'tlock-setup');
    });

    // -----------------------------------------------------------------------
    // Recovery flow (recover.html from Alice's personalized bundle)
    // -----------------------------------------------------------------------

    test(`[${lang}] recovery-1`, async ({ page }) => {
      const aliceDir = extractBundle(bundlesDir, 'Alice');
      const bobDir = extractBundle(bundlesDir, 'Bob');
      const recovery = new RecoveryPage(page, aliceDir);

      await recovery.open();
      await page.locator('#lang-select').selectOption(lang);

      await recovery.expectShareCount(1);
      await recovery.expectManifestLoaded();

      // Add Bob's share (2 of 3 — still need one more)
      await recovery.addShares(bobDir);
      await recovery.expectShareCount(2);

      // Frame: only Step 1 (gather pieces)
      await frameRecoveryStep(page, [0]);

      await snap(page, lang, 'recovery-1');
    });

    test(`[${lang}] recovery-2`, async ({ page }) => {
      const aliceDir = extractBundle(bundlesDir, 'Alice');
      const bobDir = extractBundle(bundlesDir, 'Bob');
      const camilaDir = extractBundle(bundlesDir, 'Camila');
      const recovery = new RecoveryPage(page, aliceDir);

      await recovery.open();
      await page.locator('#lang-select').selectOption(lang);

      await recovery.expectShareCount(1);
      await recovery.expectManifestLoaded();

      // Add Bob's and Camila's shares (3/3 — auto-recovery)
      await recovery.addShares(bobDir);
      await recovery.addShares(camilaDir);

      // Wait for recovery success
      await expect(page.locator('#status-message.success')).toBeAttached({ timeout: 60000 });
      await page.waitForTimeout(500);

      // Frame: only Step 3 (files recovered)
      await frameRecoveryStep(page, [2]);

      await snap(page, lang, 'recovery-2');
    });

    test(`[${lang}] tlock-waiting`, async ({ page }) => {
      const recovery = new RecoveryPage(page, path.dirname(standaloneRecoverHtml));
      await recovery.openFile(standaloneRecoverHtml);
      await page.locator('#lang-select').selectOption(lang);

      // Simulate the tlock waiting state with tomorrow's date
      await page.evaluate((lang) => {
        const tomorrow = new Date();
        tomorrow.setDate(tomorrow.getDate() + 1);
        const dateStr = tomorrow.toLocaleDateString(lang, {
          year: 'numeric', month: 'long', day: 'numeric',
        });
        const boldDate = `<strong>${dateStr}</strong>`;
        const message = (window as any).t('tlock_waiting_message', boldDate);

        const el = document.getElementById('tlock-waiting');
        if (el) el.classList.remove('hidden');
        const dateEl = document.getElementById('tlock-waiting-date');
        if (dateEl) dateEl.innerHTML = message;

        // Hide recover button
        const btn = document.getElementById('recover-btn');
        if (btn) btn.classList.add('hidden');
      }, lang);

      // Frame: only Step 3 (recover card with tlock waiting)
      await frameRecoveryStep(page, [2]);

      await snap(page, lang, 'tlock-waiting');
    });

    test(`[${lang}] recovery-words-typing`, async ({ page }) => {
      const recovery = new RecoveryPage(page, path.dirname(standaloneRecoverHtml));
      await recovery.openFile(standaloneRecoverHtml);
      await page.locator('#lang-select').selectOption(lang);

      // Open paste area
      await recovery.clickPasteButton();
      await recovery.expectPasteAreaVisible();

      // Type a partial set of translated BIP39 words, stripped of accents
      // to show that tildes and umlauts aren't required.
      const aliceDir = extractBundle(langBundlesDirs[lang]!, 'Alice');
      const words = extractWordsFromReadme(findReadmeFile(aliceDir));
      const partialWords = words.split(' ').slice(0, 8).join(' ')
        .normalize('NFD').replace(/[\u0300-\u036f]/g, '');
      await recovery.pasteShare(partialWords);

      // Frame: only Step 1 (gather pieces with paste area)
      await frameRecoveryStep(page, [0]);

      await snap(page, lang, 'recovery-words-typing');
    });

    test(`[${lang}] recovery-words-recognized`, async ({ page }) => {
      const aliceDir = extractBundle(langBundlesDirs[lang]!, 'Alice');
      const recovery = new RecoveryPage(page, path.dirname(standaloneRecoverHtml));
      await recovery.openFile(standaloneRecoverHtml);
      await page.locator('#lang-select').selectOption(lang);

      // Submit full translated words
      const words = extractWordsFromReadme(findReadmeFile(aliceDir));
      await recovery.clickPasteButton();
      await recovery.expectPasteAreaVisible();
      await recovery.pasteShare(words);
      await recovery.submitPaste();

      // Wait for share to be recognized
      await recovery.expectShareCount(1);

      // Frame: only Step 1 (gather pieces with recognized share)
      await frameRecoveryStep(page, [0]);

      await snap(page, lang, 'recovery-words-recognized');
    });
  });
}

// ---------------------------------------------------------------------------
// README screenshot — English only, full maker overview
// ---------------------------------------------------------------------------

test.describe('README screenshot', () => {
  test.setTimeout(180_000);

  test('maker-overview', async ({ page }) => {
    const creation = new CreationPage(page, makerHtmlPath);
    await creation.open();

    // Hide footer and step 3, keep nav + intro + steps 1–2 (empty placeholders)
    await hide(page, 'footer');
    await page.evaluate(() => {
      const cards = document.querySelectorAll('.container > .card');
      cards.forEach((card, i) => {
        if (i > 1) (card as HTMLElement).style.display = 'none';
      });
    });

    // Measure the full container (nav + intro + step 1)
    const bounds = await page.evaluate(() => {
      const container = document.querySelector('.container');
      if (!container) return { x: 0, y: 0, width: 800, height: 600 };
      const rect = container.getBoundingClientRect();
      return { x: rect.left, y: rect.top, width: rect.width, height: rect.height };
    });

    const pad = 16;
    const dir = path.join(SCREENSHOTS_ROOT, 'en');
    fs.mkdirSync(dir, { recursive: true });
    await page.screenshot({
      path: path.join(dir, 'maker-overview.png'),
      clip: {
        x: bounds.x - pad, y: bounds.y - pad,
        width: bounds.width + pad * 2, height: bounds.height + pad * 2,
      },
    });
  });
});
