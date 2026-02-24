// tlock-recover.ts — Time-lock decryption for drand's League of Entropy.
// Used by recover.html for time-locked bundle recovery.
// This module contains the HTTP drand client and decryption functions.
// In Phase 2 it's imported by app.ts behind __TLOCK__ guards, producing
// app-tlock.js (with HTTP code) vs app.js (without).

import { timelockDecrypt } from 'tlock-js';
import { roundTime, HttpCachingChain, HttpChainClient } from 'drand-client';
import type { ChainClient, ChainOptions } from 'drand-client';
import { DRAND_CONFIG } from './drand';
import type { TlockContainerMeta, TranslationFunction } from './types';

// Create an HTTP drand chain client for recovery, trying endpoints in order.
// This is the only path that makes network calls — needed for timelockDecrypt
// which must fetch the actual beacon signature for the target round.
async function createClient(): Promise<ChainClient> {
  const cfg = DRAND_CONFIG;
  const options: ChainOptions = {
    disableBeaconVerification: false,
    noCache: false,
    chainVerificationParams: {
      chainHash: cfg.chainHash,
      publicKey: cfg.publicKey,
    },
  };

  let lastError: Error | undefined;
  for (const endpoint of cfg.endpoints) {
    try {
      const url = `${endpoint}/${cfg.chainHash}`;
      const chain = new HttpCachingChain(url, options);
      const client = new HttpChainClient(chain, options);
      await chain.info();
      return client;
    } catch (e) {
      lastError = e instanceof Error ? e : new Error(String(e));
    }
  }
  throw new Error(`Could not connect to drand: ${lastError?.message ?? 'all endpoints failed'}`);
}

// Decrypt tlock ciphertext by fetching the beacon
export async function decrypt(ciphertext: Uint8Array): Promise<Uint8Array> {
  const client = await createClient();
  const armored = new TextDecoder().decode(ciphertext);
  const decrypted = await timelockDecrypt(armored, client);
  return new Uint8Array(decrypted);
}

// Check if a round's beacon is available (time has passed)
export async function isRoundAvailable(roundNumber: number): Promise<boolean> {
  try {
    const client = await createClient();
    const info = await client.chain().info();
    const rt = roundTime(info, roundNumber);
    return rt <= Date.now();
  } catch {
    return false;
  }
}

// Format a tlock unlock date for display. Shows time if within 24 hours, date-only otherwise.
export function formatTimelockDate(date: Date): string {
  const hoursUntil = (date.getTime() - Date.now()) / 3600000;
  return (hoursUntil > 0 && hoursUntil < 24)
    ? date.toLocaleString()
    : date.toLocaleDateString();
}

// Format an unlock date for the waiting UI, with relative time for near-future dates.
export function formatUnlockDate(date: Date, t: TranslationFunction): { text: string; relative: boolean } {
  const minutesUntil = (date.getTime() - Date.now()) / 60000;
  if (minutesUntil > 0 && minutesUntil < 60) {
    const m = Math.ceil(minutesUntil);
    return { text: t(m === 1 ? 'tlock_in_one_minute' : 'tlock_in_minutes', m), relative: true };
  }
  return { text: formatTimelockDate(date), relative: false };
}

// Internal timer state for waitAndDecrypt polling
let tlockTimer: ReturnType<typeof setInterval> | null = null;
let pendingCiphertext: Uint8Array | null = null;
let pendingMeta: TlockContainerMeta | null = null;

// Wait for the tlock unlock time to pass, then decrypt.
// onTick is called every 5 seconds with the current unlock date (for UI updates).
// onReady is called with the decrypted archive once the time lock passes.
// onError is called if decryption fails.
export function waitAndDecrypt(
  meta: TlockContainerMeta,
  ciphertext: Uint8Array,
  onTick: (unlockDate: Date) => void,
  onReady: (archive: Uint8Array) => void,
  onError: (err: Error) => void
): void {
  pendingCiphertext = ciphertext;
  pendingMeta = meta;

  async function check(): Promise<void> {
    if (!pendingMeta || !pendingCiphertext) return;

    const unlockDate = new Date(pendingMeta.unlock);
    if (unlockDate > new Date()) {
      // Still waiting — update the display
      onTick(unlockDate);
      return;
    }

    // Time passed — decrypt
    const ct = pendingCiphertext;
    stopWaiting();
    try {
      const archive = await decrypt(ct);
      onReady(archive);
    } catch (err) {
      onError(err instanceof Error ? err : new Error(String(err)));
    }
  }

  // Check immediately, then every 5 seconds
  check();
  if (!tlockTimer) {
    tlockTimer = setInterval(() => check(), 5000);
  }
}

// Stop the internal polling timer.
export function stopWaiting(): void {
  if (tlockTimer) {
    clearInterval(tlockTimer);
    tlockTimer = null;
  }
  pendingCiphertext = null;
  pendingMeta = null;
}
