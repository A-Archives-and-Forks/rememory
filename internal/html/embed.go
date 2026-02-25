package html

import (
	"embed"
)

// Shared layout template used by all pages
//
//go:embed assets/layout.html
var layoutHTMLTemplate string

// Embedded assets for the recovery HTML
// These files are embedded at compile time

//go:embed assets/recover.html
var recoverHTMLTemplate string

//go:embed assets/shared.js
var sharedJS string

// Recovery JS: two variants built from app.ts with different __TLOCK__ defines.
// app.js (__TLOCK__=false) — offline recovery, no tlock/drand code.
// app-tlock.js (__TLOCK__=true) — recovery with tlock HTTP decryption via drand.

//go:embed assets/app.js
var appJS string

//go:embed assets/app-tlock.js
var appTlockJS string

//go:embed assets/styles.css
var stylesCSS string

//go:embed assets/wasm_exec.js
var wasmExecJS string

// Embedded assets for the bundle creation HTML

//go:embed assets/maker.html
var makerHTMLTemplate string

//go:embed assets/create-app.js
var createAppJS string

// Selfhosted variant of the create-app script (with __SELFHOSTED__=true).
// Includes server integration for manifest upload.

//go:embed assets/create-app-selfhosted.js
var createAppSelfhostedJS string

// Note: tlock-create.js is gone — encryption is inline in create-app.js
// using the offline drand client. tlock-recover.ts still exists as a module
// but has no standalone output file — it's imported by app.ts behind
// __TLOCK__ guards, so it ends up in app-tlock.js only.

// Static page templates (no WASM needed)

//go:embed assets/about.html
var aboutHTMLTemplate string

//go:embed assets/docs-template.html
var docsHTMLTemplate string

//go:embed docs-content/*.md
var docsContentFS embed.FS

//go:embed assets/dataflow.js
var dataflowJS string

// Selfhosted page templates

//go:embed assets/home.html
var homeHTMLTemplate string

//go:embed assets/setup.html
var setupHTMLTemplate string

// Page-specific CSS (extracted from Go constants into .css files)

//go:embed assets/home.css
var homeCSS string

//go:embed assets/maker.css
var makerCSS string

//go:embed assets/index.css
var indexCSS string

//go:embed assets/setup.css
var setupCSS string

//go:embed assets/tlock-waiting.css
var tlockWaitingCSS string

// Page-specific JavaScript (extracted from Go to .js files)

//go:embed assets/i18n.js
var i18nJSTemplate string

//go:embed assets/i18n-index-setlang.js
var i18nIndexSetlangJS string

//go:embed assets/i18n-recover-init.js
var i18nRecoverInitJS string

//go:embed assets/home.js
var homeJS string

//go:embed assets/setup.js
var setupJS string

//go:embed assets/wasm-loader.js
var wasmLoaderJS string

// createWASM is set at build time for the CLI binary (not for WASM builds)
// This avoids circular dependency since create.wasm embeds the html package
var createWASM []byte

// GetCreateWASMBytes returns the full WASM binary with bundle creation.
// This larger WASM is used in maker.html for the creation tool.
// Note: Must be set via SetCreateWASMBytes before use (done in CLI init).
func GetCreateWASMBytes() []byte {
	return createWASM
}

// SetCreateWASMBytes sets the create.wasm bytes.
// Called by CLI initialization to avoid circular embedding.
func SetCreateWASMBytes(data []byte) {
	createWASM = data
}
