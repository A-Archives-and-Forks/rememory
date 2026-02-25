package docs

import "embed"

// ScreenshotsFS holds the embedded screenshot images for the docs pages.
// Only English and language-neutral screenshots are included to keep the
// binary small. The server rewrites other language paths to English.
//
//go:embed screenshots/en screenshots/demo-pdf screenshots/demo-pdf-es screenshots/*.png
var ScreenshotsFS embed.FS
