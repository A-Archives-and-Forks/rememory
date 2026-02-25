// drand.ts — Drand quicknet chain parameters and offline client.
// All values come from Go via window.DRAND_CONFIG, injected at HTML generation time.
// The single source of truth is internal/core/tlock_common.go.
//
// This module contains only the config and the offline client — no HTTP imports.
// create-app.ts imports createOfflineClient() for encryption (zero network calls).
// tlock-recover.ts builds its own HTTP client from drand-client for decryption;
// it's only bundled into the network variant (app-tlock.js) via __TLOCK__ guards.

import type { ChainClient } from 'drand-client';

interface DrandConfig {
  chainHash: string;
  genesis: number;
  period: number;
  publicKey: string;
  endpoints: string[];
  schemeID: string;
  groupHash: string;
  beaconID: string;
}

const cfg = (window as any).DRAND_CONFIG as DrandConfig;
if (!cfg) {
  throw new Error('DRAND_CONFIG not found — drand configuration was not injected');
}

export const QUICKNET_CHAIN_HASH = cfg.chainHash;
export const QUICKNET_GENESIS = cfg.genesis;
export const QUICKNET_PERIOD = cfg.period;

// Re-export the full config for tlock-recover.ts, which builds its own HTTP client.
export const DRAND_CONFIG = cfg;

// Create an offline drand client using only embedded chain config.
// timelockEncrypt only calls chain().info() for the public key and scheme —
// it never fetches beacons — so this works without any network access.
export function createOfflineClient(): ChainClient {
  const info = {
    public_key: cfg.publicKey,
    period: cfg.period,
    genesis_time: cfg.genesis,
    hash: cfg.chainHash,
    groupHash: cfg.groupHash,
    schemeID: cfg.schemeID,
    metadata: { beaconID: cfg.beaconID },
  };

  return {
    chain() {
      return {
        info: () => Promise.resolve(info),
      };
    },
    get(_round: number) {
      return Promise.reject(new Error('Offline client cannot fetch beacons'));
    },
    latest() {
      return Promise.reject(new Error('Offline client cannot fetch beacons'));
    },
  } as ChainClient;
}
