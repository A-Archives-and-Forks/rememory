//go:build !js || cjk

package pdf

import _ "embed"

// Noto Sans SC font files (SIL Open Font License - see fonts/LICENSE-NotoSansSC.txt)
// These provide CJK (Chinese, Japanese, Korean) glyph coverage.
//
// They are compiled in for all non-WASM builds (CLI, tests) and for WASM
// built with -tags cjk (produces a larger but CJK-capable maker.html).
// The default WASM build excludes them to keep maker.html at ~6 MB.

//go:embed fonts/NotoSansSC-Regular.ttf
var notoSansSCRegular []byte

//go:embed fonts/NotoSansSC-Bold.ttf
var notoSansSCBold []byte
