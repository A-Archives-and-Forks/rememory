package html

import (
	"regexp"
	"strings"
	"testing"
)

// staticPages returns all static HTML pages for testing.
func staticPages() map[string]string {
	wasmStub := []byte{0x00, 0x61, 0x73, 0x6d} // WASM magic bytes
	ghURL := "https://github.com/eljojo/rememory"
	return map[string]string{
		"maker.html":   GenerateMakerHTML(wasmStub, "test", ghURL, MakerHTMLOptions{}),
		"recover.html": GenerateRecoverHTML("test", ghURL, nil),
		"index.html":   GenerateIndexHTML("test", ghURL),
		"docs.html":    GenerateDocsHTML("test", ghURL, "en"),
	}
}

// TestStaticHTMLHasNoServerCode verifies that static HTML builds (non-selfhosted)
// contain zero traces of server-interaction code. This ensures offline bundles
// never contain code that could phone home.
func TestStaticHTMLHasNoServerCode(t *testing.T) {
	// These patterns must NOT appear in static output.
	// Note: SELFHOSTED_CONFIG = null is allowed (it's inert).
	// rememoryLoadManifest is now always present (unified JS) but is inert when
	// SELFHOSTED_CONFIG is null — the fetch only activates on manifestURL.
	// rememoryOnBundlesCreated is still compile-flag-guarded in create-app.ts.
	forbidden := []string{
		"/api/bundle",
		"/api/setup",
		"rememoryOnBundlesCreated",
	}

	for name, content := range staticPages() {
		for _, pattern := range forbidden {
			if strings.Contains(content, pattern) {
				t.Errorf("static %s contains server code: found %q", name, pattern)
			}
		}
	}
}

// TestStaticHostedRecoverHTML verifies that static-hosted recover.html
// contains the right config for auto-fetching from a relative MANIFEST.age URL.
func TestStaticHostedRecoverHTML(t *testing.T) {
	ghURL := "https://github.com/eljojo/rememory"
	content := GenerateRecoverHTML("test", ghURL, nil, RecoverHTMLOptions{
		StaticHosted: true,
	})

	// Must contain manifestURL pointing to ./MANIFEST.age
	if !strings.Contains(content, `"manifestURL":"./MANIFEST.age"`) {
		t.Error("static-hosted recover.html missing manifestURL config")
	}

	// Must not contain server API endpoints
	if strings.Contains(content, "/api/bundle") {
		t.Error("static-hosted recover.html contains /api/bundle")
	}

	// Nav links should NOT be rewritten (no server routes)
	if strings.Contains(content, `href="/create"`) {
		t.Error("static-hosted recover.html has server nav rewrites")
	}
	if !strings.Contains(content, `href="index.html"`) {
		t.Error("static-hosted recover.html missing standard nav links")
	}
}

// TestStaticHTMLNoUnexpectedURLs scans static HTML output for every http:// and
// https:// URL and verifies it matches an allowed prefix. This catches accidental
// network calls that could phone home from what should be offline-capable bundles.
//
// The drand URLs are expected (tlock time-lock encryption needs the drand beacon).
// Project URLs (GitHub, GitHub Pages) are documentation links, not network calls.
// Everything else should be investigated before being added here.
func TestStaticHTMLNoUnexpectedURLs(t *testing.T) {
	allowed := []string{
		// drand beacon endpoints (tlock)
		"https://api.drand.sh",
		"https://api2.drand.sh",
		"https://api3.drand.sh",
		"https://pl-us.testnet.drand.sh",
		"https://drand.cloudflare.com",
		"https://docs.drand.love",

		// project URLs
		"https://github.com/eljojo/rememory",
		"https://eljojo.github.io/rememory",

		// docs: linked in user-facing documentation and index.html
		"https://github.com/FiloSottile/age", // age encryption library
		"https://www.youtube.com",            // index.html "Why I built this" documentary
		"https://www.cloudflare.com",         // docs: League of Entropy (tlock section)
		"https://cryptomator.org",            // docs: recommended encrypted vault tool
		"https://veracrypt.fr",               // docs: recommended encrypted vault tool

		// vendored JS: comments in bundled tlock/noble-curves/drand-client code
		"https://github.com/golang/go/issues",               // wasm_exec.js workaround comment
		"https://github.com/paulmillr/noble",                // noble-secp256k1 library reference
		"https://github.com/hyperledger/aries-framework-go", // BBS+ signature issue reference
		"https://eprint.iacr.org",                           // cryptography research papers
		"https://ethresear.ch",                              // BLS signature verification paper
		"https://www.rfc-editor.org",                        // RFC errata references
		"https://datatracker.ietf.org",                      // RFC 9380 hash-to-curve
		"https://crypto.stackexchange.com",                  // elliptic curve Q&A
		"https://bitcoin.stackexchange.com",                 // transaction script parsing Q&A
		"https://feross.org",                                // ieee754 library license attribution
		"https://hyperelliptic.org",                         // EFD curve operation formulas
		"https://developer.mozilla.org",                     // Web Crypto API JSDoc references

		// index.html: Shamir's Secret Sharing article (per-language translations)
		"https://en.wikipedia.org",
		"https://es.wikipedia.org",
		"https://de.wikipedia.org",
		"https://fr.wikipedia.org",
		"https://zh.wikipedia.org",
	}

	urlRe := regexp.MustCompile(`https?://[^\s"'<>,;)\\]+`)

	for name, content := range staticPages() {
		for _, url := range urlRe.FindAllString(content, -1) {
			if hasAllowedPrefix(url, allowed) {
				continue
			}
			t.Errorf("%s: unexpected URL %q", name, url)
		}
	}
}

func hasAllowedPrefix(url string, allowed []string) bool {
	for _, prefix := range allowed {
		if strings.HasPrefix(url, prefix) {
			return true
		}
	}
	return false
}
