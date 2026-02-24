import { test, expect } from './fixtures';
import { ChildProcess, spawn } from 'child_process';
import * as fs from 'fs';
import * as path from 'path';
import * as os from 'os';
import * as net from 'net';
import {
  getRememoryBin,
  createTestProject,
  cleanupProject,
  extractBundles,
  findReadmeFile,
  RecoveryPage,
} from './helpers';

// Find an available port
async function getAvailablePort(): Promise<number> {
  return new Promise((resolve, reject) => {
    const srv = net.createServer();
    srv.listen(0, '127.0.0.1', () => {
      const addr = srv.address() as net.AddressInfo;
      srv.close(() => resolve(addr.port));
    });
    srv.on('error', reject);
  });
}

test.describe('Static Pages', () => {
  test.use({ allowedHosts: ['127.0.0.1'] });
  let projectDir: string;
  let pagesDir: string;
  let bundlesDir: string;
  let serverProc: ChildProcess | null = null;
  let baseURL: string;

  test.beforeAll(async () => {
    const bin = getRememoryBin();
    if (!fs.existsSync(bin)) {
      test.skip();
      return;
    }

    // Create a sealed project with --pages flag
    const tmpDir = fs.mkdtempSync(path.join(os.tmpdir(), 'rememory-pages-e2e-'));
    projectDir = path.join(tmpDir, 'test-project');

    const { execFileSync } = require('child_process');

    execFileSync(bin, [
      'init', projectDir, '--name', 'Pages E2E Test', '--threshold', '2',
      '--friend', 'Alice,alice@test.com', '--friend', 'Bob,bob@test.com', '--friend', 'Carol,carol@test.com',
    ], { stdio: 'inherit' });

    // Add secret content
    const manifestDir = path.join(projectDir, 'manifest');
    fs.writeFileSync(path.join(manifestDir, 'secret.txt'), 'The secret password is: static-pages-work');
    fs.writeFileSync(path.join(manifestDir, 'notes.txt'), 'Pages mode recovery test');

    // Seal with --pages flag
    execFileSync(bin, ['seal', '--pages'], { cwd: projectDir, stdio: 'inherit' });

    pagesDir = path.join(projectDir, 'output', 'pages');
    bundlesDir = path.join(projectDir, 'output', 'bundles');
  });

  test.afterAll(async () => {
    if (serverProc) {
      serverProc.kill();
      await new Promise(r => setTimeout(r, 500));
    }
    if (projectDir && fs.existsSync(path.dirname(projectDir))) {
      fs.rmSync(path.dirname(projectDir), { recursive: true, force: true });
    }
  });

  test('pages directory contains recover.html and MANIFEST.age', async () => {
    expect(fs.existsSync(path.join(pagesDir, 'recover.html'))).toBe(true);
    expect(fs.existsSync(path.join(pagesDir, 'MANIFEST.age'))).toBe(true);

    // MANIFEST.age in pages should match the sealed manifest
    const pagesManifest = fs.readFileSync(path.join(pagesDir, 'MANIFEST.age'));
    const sealedManifest = fs.readFileSync(path.join(projectDir, 'output', 'MANIFEST.age'));
    expect(pagesManifest.equals(sealedManifest)).toBe(true);
  });

  test('recover.html contains manifestURL config', async () => {
    const html = fs.readFileSync(path.join(pagesDir, 'recover.html'), 'utf-8');
    expect(html).toContain('"manifestURL":"./MANIFEST.age"');
    expect(html).not.toContain('/api/bundle');
    // Nav links should not be rewritten
    expect(html).toContain('href="index.html"');
  });

  test('static pages recovery: manifest auto-loads and recovery completes', async ({ page }, testInfo) => {
    testInfo.setTimeout(120000);

    // Serve the pages directory with a minimal static HTTP server
    const port = await getAvailablePort();
    baseURL = `http://127.0.0.1:${port}`;

    serverProc = spawn('python3', ['-m', 'http.server', String(port), '--bind', '127.0.0.1'], {
      cwd: pagesDir,
      stdio: ['ignore', 'pipe', 'pipe'],
    });

    // Wait for server to be ready
    const deadline = Date.now() + 10000;
    while (Date.now() < deadline) {
      try {
        const resp = await fetch(`${baseURL}/recover.html`);
        if (resp.ok) break;
      } catch {
        // not ready yet
      }
      await new Promise(r => setTimeout(r, 200));
    }

    // Navigate to recover.html
    await page.goto(`${baseURL}/recover.html`);
    await page.waitForFunction(
      () => (window as any).rememoryAppReady === true,
      { timeout: 30000 }
    );

    // Manifest should auto-load from the static ./MANIFEST.age
    await expect(page.locator('#manifest-status')).toHaveClass(/loaded/, { timeout: 15000 });

    // Add shares from bundle ZIPs
    const [aliceDir, bobDir] = extractBundles(bundlesDir, ['Alice', 'Bob']);
    const readmePaths = [aliceDir, bobDir].map(dir => findReadmeFile(dir, '.txt'));
    await page.locator('#share-file-input').setInputFiles(readmePaths);

    // Should have 2 shares
    await expect(page.locator('.share-item')).toHaveCount(2);

    // Recovery should complete automatically (threshold is 2)
    await expect(page.locator('#status-message')).toContainText('files are ready', { timeout: 60000 });
    await expect(page.locator('#download-all-btn')).toBeVisible();

    // Clean up server
    serverProc.kill();
    serverProc = null;
  });
});
