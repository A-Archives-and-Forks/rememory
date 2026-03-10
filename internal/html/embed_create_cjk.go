//go:build !js && cjk

package html

import (
	_ "embed"
)

// Embed create-cjk.wasm when building with -tags cjk.
// This WASM was built with CJK font support, producing a larger (~21 MB)
// but CJK-capable maker.html. Build with: make wasm-cjk && make build-cjk

//go:embed assets/create-cjk.wasm
var createWASMEmbed []byte

func init() {
	createWASM = createWASMEmbed
}
