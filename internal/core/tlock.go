//go:build !js

package core

import (
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"time"

	"filippo.io/age/armor"
	chain "github.com/drand/drand/v2/common"
	"github.com/drand/drand/v2/crypto"
	"github.com/drand/kyber"
	"github.com/drand/tlock"
	tlockhttp "github.com/drand/tlock/networks/http"
)

// IsTlockTooEarly returns true if the error wraps tlock.ErrTooEarly,
// meaning the drand round has not been reached yet.
func IsTlockTooEarly(err error) bool {
	return errors.Is(err, tlock.ErrTooEarly)
}

// TlockEncrypt encrypts src to a specific drand round number using tlock.
// Encryption is offline — it uses only the embedded chain parameters
// (public key, scheme) and never contacts the drand network.
// The output is ASCII-armored age format for compatibility with tlock-js
// (which expects armored input in timelockDecrypt). The Go TlockDecrypt
// handles both armored and binary age format, so this is safe.
func TlockEncrypt(dst io.Writer, src io.Reader, roundNumber uint64) error {
	network, err := offlineNetwork()
	if err != nil {
		return fmt.Errorf("tlock encrypt: %w", err)
	}

	aw := armor.NewWriter(dst)
	if err := tlock.New(network).Encrypt(aw, src, roundNumber); err != nil {
		return fmt.Errorf("tlock encrypt: %w", err)
	}
	if err := aw.Close(); err != nil {
		return fmt.Errorf("tlock encrypt (armor): %w", err)
	}

	return nil
}

// TlockDecrypt decrypts tlock-encrypted ciphertext by fetching the drand
// beacon signature for the round embedded in the ciphertext.
// Returns tlock.ErrTooEarly if the round has not been reached yet.
func TlockDecrypt(dst io.Writer, src io.Reader) error {
	network, err := connectDrand()
	if err != nil {
		return fmt.Errorf("tlock decrypt: %w", err)
	}

	if err := tlock.New(network).Decrypt(dst, src); err != nil {
		return fmt.Errorf("tlock decrypt: %w", err)
	}

	return nil
}

// offlineNetwork constructs a tlock.Network from embedded constants.
// Encryption only needs the public key and scheme — it never fetches beacons.
// This is the same approach used by the browser-side createOfflineClient().
func offlineNetwork() (tlock.Network, error) {
	sch, err := crypto.SchemeFromName(QuicknetSchemeID)
	if err != nil {
		return nil, fmt.Errorf("invalid scheme %q: %w", QuicknetSchemeID, err)
	}

	pubKeyBytes, err := hex.DecodeString(QuicknetPublicKey)
	if err != nil {
		return nil, fmt.Errorf("decoding public key: %w", err)
	}

	pubKey := sch.KeyGroup.Point()
	if err := pubKey.UnmarshalBinary(pubKeyBytes); err != nil {
		return nil, fmt.Errorf("unmarshaling public key: %w", err)
	}

	return &offlineNet{
		chainHash: QuicknetChainHash,
		publicKey: pubKey,
		scheme:    *sch,
	}, nil
}

// offlineNet implements tlock.Network using only embedded constants.
// Signature() is not supported — encryption never calls it.
type offlineNet struct {
	chainHash string
	publicKey kyber.Point
	scheme    crypto.Scheme
}

func (n *offlineNet) ChainHash() string              { return n.chainHash }
func (n *offlineNet) PublicKey() kyber.Point         { return n.publicKey }
func (n *offlineNet) Scheme() crypto.Scheme          { return n.scheme }
func (n *offlineNet) SwitchChainHash(_ string) error { return errors.New("offline network") }
func (n *offlineNet) Current(t time.Time) uint64 {
	return chain.CurrentRound(t.Unix(), QuicknetPeriod, QuicknetGenesis)
}
func (n *offlineNet) Signature(_ uint64) ([]byte, error) {
	return nil, errors.New("offline network cannot fetch beacon signatures")
}

// connectDrand tries each drand endpoint until one connects.
// Used only for decryption (which needs to fetch beacon signatures).
func connectDrand() (tlock.Network, error) {
	var lastErr error
	for _, endpoint := range DrandEndpoints {
		network, err := tlockhttp.NewNetwork(endpoint, QuicknetChainHash)
		if err != nil {
			lastErr = err
			continue
		}
		return network, nil
	}
	return nil, fmt.Errorf("connecting to drand: all endpoints failed (last error: %w)", lastErr)
}
