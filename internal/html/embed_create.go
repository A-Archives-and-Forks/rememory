//go:build !js && !cjk

package html

import (
	_ "embed"
)

// Embed create.wasm only in non-WASM builds (i.e., the CLI binary).
// This avoids circular dependency since create.wasm itself embeds the html package.
// The default build uses create.wasm (no CJK fonts, ~6 MB maker.html).
// Build with -tags cjk to embed create-cjk.wasm instead (see embed_create_cjk.go).

//go:embed assets/create.wasm
var createWASMEmbed []byte

func init() {
	createWASM = createWASMEmbed
}
