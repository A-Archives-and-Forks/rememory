package core

import (
	"bytes"
	"encoding/hex"
	"os"
	"testing"
	"time"

	"github.com/drand/drand/v2/common/chain"
	"github.com/drand/drand/v2/crypto"
	tlockhttp "github.com/drand/tlock/networks/http"
)

func TestRoundForTime(t *testing.T) {
	genesis := time.Unix(QuicknetGenesis, 0)

	tests := []struct {
		name   string
		target time.Time
		want   uint64
	}{
		{
			name:   "at genesis",
			target: genesis,
			want:   1,
		},
		{
			name:   "before genesis",
			target: genesis.Add(-time.Hour),
			want:   1,
		},
		{
			name:   "3 seconds after genesis",
			target: genesis.Add(3 * time.Second),
			want:   2,
		},
		{
			name:   "6 seconds after genesis",
			target: genesis.Add(6 * time.Second),
			want:   3,
		},
		{
			name:   "1 second after genesis (rounds up)",
			target: genesis.Add(1 * time.Second),
			want:   2,
		},
		{
			name:   "exactly one period after genesis",
			target: genesis.Add(QuicknetPeriod),
			want:   2,
		},
		{
			name:   "30 days after genesis",
			target: genesis.Add(30 * 24 * time.Hour),
			want:   uint64(30*24*3600/3) + 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RoundForTime(tt.target)
			if got != tt.want {
				t.Errorf("RoundForTime(%v): got %d, want %d", tt.target, got, tt.want)
			}
		})
	}
}

func TestTimeForRound(t *testing.T) {
	genesis := time.Unix(QuicknetGenesis, 0)

	tests := []struct {
		name  string
		round uint64
		want  time.Time
	}{
		{
			name:  "round 1",
			round: 1,
			want:  genesis,
		},
		{
			name:  "round 0 (clamped)",
			round: 0,
			want:  genesis,
		},
		{
			name:  "round 2",
			round: 2,
			want:  genesis.Add(QuicknetPeriod),
		},
		{
			name:  "round 100",
			round: 100,
			want:  genesis.Add(99 * QuicknetPeriod),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := TimeForRound(tt.round)
			if !got.Equal(tt.want) {
				t.Errorf("TimeForRound(%d): got %v, want %v", tt.round, got, tt.want)
			}
		})
	}
}

func TestRoundTimeRoundtrip(t *testing.T) {
	// For any round, TimeForRound → RoundForTime should give back the same round.
	for _, round := range []uint64{1, 2, 100, 1000000, 12345678} {
		roundTime := TimeForRound(round)
		gotRound := RoundForTime(roundTime)
		if gotRound != round {
			t.Errorf("round %d → time %v → round %d", round, roundTime, gotRound)
		}
	}
}

func TestParseTimelockValue(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
		check   func(t *testing.T, result time.Time)
	}{
		{
			name:  "30 days",
			input: "30d",
			check: func(t *testing.T, result time.Time) {
				expected := time.Now().UTC().AddDate(0, 0, 30)
				diff := result.Sub(expected)
				if diff < -time.Second || diff > time.Second {
					t.Errorf("30d: expected ~%v, got %v", expected, result)
				}
			},
		},
		{
			name:  "2 weeks",
			input: "2w",
			check: func(t *testing.T, result time.Time) {
				expected := time.Now().UTC().AddDate(0, 0, 14)
				diff := result.Sub(expected)
				if diff < -time.Second || diff > time.Second {
					t.Errorf("2w: expected ~%v, got %v", expected, result)
				}
			},
		},
		{
			name:  "6 months",
			input: "6m",
			check: func(t *testing.T, result time.Time) {
				expected := time.Now().UTC().AddDate(0, 6, 0)
				diff := result.Sub(expected)
				if diff < -time.Second || diff > time.Second {
					t.Errorf("6m: expected ~%v, got %v", expected, result)
				}
			},
		},
		{
			name:  "1 year",
			input: "1y",
			check: func(t *testing.T, result time.Time) {
				expected := time.Now().UTC().AddDate(1, 0, 0)
				diff := result.Sub(expected)
				if diff < -time.Second || diff > time.Second {
					t.Errorf("1y: expected ~%v, got %v", expected, result)
				}
			},
		},
		{
			name:  "5 minutes",
			input: "5min",
			check: func(t *testing.T, result time.Time) {
				expected := time.Now().UTC().Add(5 * time.Minute)
				diff := result.Sub(expected)
				if diff < -time.Second || diff > time.Second {
					t.Errorf("5min: expected ~%v, got %v", expected, result)
				}
			},
		},
		{
			name:  "uppercase D",
			input: "30D",
			check: func(t *testing.T, result time.Time) {
				// Should be case-insensitive
				expected := time.Now().UTC().AddDate(0, 0, 30)
				diff := result.Sub(expected)
				if diff < -time.Second || diff > time.Second {
					t.Errorf("30D: expected ~%v, got %v", expected, result)
				}
			},
		},
		{
			name:  "absolute future datetime",
			input: "2099-01-01T00:00:00Z",
			check: func(t *testing.T, result time.Time) {
				if result.Year() != 2099 || result.Month() != 1 || result.Day() != 1 {
					t.Errorf("absolute: got %v", result)
				}
			},
		},
		{
			name:    "empty string",
			input:   "",
			wantErr: true,
		},
		{
			name:    "past date",
			input:   "2020-01-01T00:00:00Z",
			wantErr: true,
		},
		{
			name:    "invalid format",
			input:   "abc",
			wantErr: true,
		},
		{
			name:    "zero days",
			input:   "0d",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseTimelockValue(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error for %q", tt.input)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error for %q: %v", tt.input, err)
			}
			if tt.check != nil {
				tt.check(t, result)
			}
		})
	}
}

func TestTlockEncryptDecryptIntegration(t *testing.T) {
	if os.Getenv("REMEMORY_TEST_TLOCK") != "1" {
		t.Skip("set REMEMORY_TEST_TLOCK=1 to run tlock integration tests (requires internet)")
	}

	plaintext := []byte("the secret message for tlock integration test")

	// Encrypt to a recently-past round (should already be available)
	pastTime := time.Now().Add(-1 * time.Minute)
	round := RoundForTime(pastTime)

	var cipherBuf bytes.Buffer
	if err := TlockEncrypt(&cipherBuf, bytes.NewReader(plaintext), round); err != nil {
		t.Fatalf("TlockEncrypt: %v", err)
	}

	if cipherBuf.Len() == 0 {
		t.Fatal("TlockEncrypt produced empty output")
	}

	// Decrypt
	var decryptBuf bytes.Buffer
	if err := TlockDecrypt(&decryptBuf, bytes.NewReader(cipherBuf.Bytes())); err != nil {
		t.Fatalf("TlockDecrypt: %v", err)
	}

	if !bytes.Equal(decryptBuf.Bytes(), plaintext) {
		t.Errorf("decrypted data mismatch: got %q, want %q", decryptBuf.Bytes(), plaintext)
	}
}

func TestTlockFutureRoundCannotDecrypt(t *testing.T) {
	if os.Getenv("REMEMORY_TEST_TLOCK") != "1" {
		t.Skip("set REMEMORY_TEST_TLOCK=1 to run tlock integration tests (requires internet)")
	}

	plaintext := []byte("this should not be decryptable yet")

	// Encrypt to a far-future round (year 2099)
	futureTime := time.Date(2099, 1, 1, 0, 0, 0, 0, time.UTC)
	round := RoundForTime(futureTime)

	var cipherBuf bytes.Buffer
	if err := TlockEncrypt(&cipherBuf, bytes.NewReader(plaintext), round); err != nil {
		t.Fatalf("TlockEncrypt: %v", err)
	}

	// Decryption should fail (too early)
	var decryptBuf bytes.Buffer
	err := TlockDecrypt(&decryptBuf, bytes.NewReader(cipherBuf.Bytes()))
	if err == nil {
		t.Fatal("expected decryption to fail for future round")
	}

	if !IsTlockTooEarly(err) {
		t.Logf("error (not 'too early' but still an error, which is expected): %v", err)
	}
}

// TestQuicknetConstantsMatchChainHash validates that our hardcoded quicknet
// constants (public key, genesis, period, group hash, beacon ID) are
// internally consistent by recomputing the chain hash and comparing it to
// QuicknetChainHash. This catches any corrupted or drifted constant without
// needing network access.
//
// Both Go-side TlockEncrypt (via offlineNetwork) and browser-side
// createOfflineClient() depend on these constants. If they're wrong,
// encryption produces ciphertext that can never be decrypted.
func TestQuicknetConstantsMatchChainHash(t *testing.T) {
	sch, err := crypto.SchemeFromName(QuicknetSchemeID)
	if err != nil {
		t.Fatalf("QuicknetSchemeID %q is not a valid scheme: %v", QuicknetSchemeID, err)
	}

	pubKeyBytes, err := hex.DecodeString(QuicknetPublicKey)
	if err != nil {
		t.Fatalf("QuicknetPublicKey is not valid hex: %v", err)
	}

	pubKey := sch.KeyGroup.Point()
	if err := pubKey.UnmarshalBinary(pubKeyBytes); err != nil {
		t.Fatalf("QuicknetPublicKey does not unmarshal to a valid point: %v", err)
	}

	groupHash, err := hex.DecodeString(QuicknetGroupHash)
	if err != nil {
		t.Fatalf("QuicknetGroupHash is not valid hex: %v", err)
	}

	info := chain.Info{
		PublicKey:   pubKey,
		ID:          QuicknetBeaconID,
		Period:      QuicknetPeriod,
		Scheme:      QuicknetSchemeID,
		GenesisTime: QuicknetGenesis,
		GenesisSeed: groupHash,
	}

	got := hex.EncodeToString(info.Hash())
	if got != QuicknetChainHash {
		t.Errorf("chain hash mismatch: computed %s, want %s\n"+
			"One or more quicknet constants are wrong.", got, QuicknetChainHash)
	}
}

// TestQuicknetConstantsMatchNetwork validates our hardcoded constants against
// the real drand quicknet network. This catches the case where the drand
// network has been re-keyed or our constants have drifted from reality.
func TestQuicknetConstantsMatchNetwork(t *testing.T) {
	if os.Getenv("REMEMORY_TEST_TLOCK") != "1" {
		t.Skip("set REMEMORY_TEST_TLOCK=1 to validate constants against the live drand network")
	}

	network, err := tlockhttp.NewNetwork(DrandEndpoints[0], QuicknetChainHash)
	if err != nil {
		t.Fatalf("connecting to drand: %v", err)
	}

	// Verify scheme
	if network.Scheme().Name != QuicknetSchemeID {
		t.Errorf("scheme mismatch: network=%q, constant=%q", network.Scheme().Name, QuicknetSchemeID)
	}

	// Verify public key
	networkPubBytes, err := network.PublicKey().MarshalBinary()
	if err != nil {
		t.Fatalf("marshaling network public key: %v", err)
	}
	if hex.EncodeToString(networkPubBytes) != QuicknetPublicKey {
		t.Error("public key mismatch between network and QuicknetPublicKey constant")
	}

	// Verify chain hash itself (proves period, genesis, group hash, beacon ID all match)
	if network.ChainHash() != QuicknetChainHash {
		t.Errorf("chain hash mismatch: network=%q, constant=%q", network.ChainHash(), QuicknetChainHash)
	}
}

// TestOfflineEncryptProducesValidCiphertext verifies that offline encryption
// (using embedded constants, no HTTP) produces ciphertext that the network-
// connected decryptor can successfully decrypt.
func TestOfflineEncryptProducesValidCiphertext(t *testing.T) {
	if os.Getenv("REMEMORY_TEST_TLOCK") != "1" {
		t.Skip("set REMEMORY_TEST_TLOCK=1 to run tlock integration tests (requires internet)")
	}

	plaintext := []byte("offline encryption integration test")

	// Encrypt offline to a recently-past round
	pastTime := time.Now().Add(-1 * time.Minute)
	round := RoundForTime(pastTime)

	var cipherBuf bytes.Buffer
	if err := TlockEncrypt(&cipherBuf, bytes.NewReader(plaintext), round); err != nil {
		t.Fatalf("TlockEncrypt (offline): %v", err)
	}

	// Decrypt via network (proves offline encryption is compatible)
	var decryptBuf bytes.Buffer
	if err := TlockDecrypt(&decryptBuf, bytes.NewReader(cipherBuf.Bytes())); err != nil {
		t.Fatalf("TlockDecrypt: %v", err)
	}

	if !bytes.Equal(decryptBuf.Bytes(), plaintext) {
		t.Errorf("decrypted data mismatch: got %q, want %q", decryptBuf.Bytes(), plaintext)
	}
}
