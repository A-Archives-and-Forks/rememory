//go:build js && !cjk

package pdf

// notoSansSCRegular and notoSansSCBold are nil in the default WASM build,
// so maker.html stays small (~6 MB). Build with -tags cjk (see fonts_cjk.go)
// to include full Noto Sans SC support in the WASM and produce a
// CJK-capable maker.html (~21 MB).
var notoSansSCRegular []byte
var notoSansSCBold []byte
