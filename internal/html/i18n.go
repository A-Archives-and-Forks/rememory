package html

import (
	"strings"

	"github.com/eljojo/rememory/internal/translations"
)

// I18nScriptOptions controls the page-specific parts of the shared i18n script block.
type I18nScriptOptions struct {
	Component           string // translation component name: "maker", "recover", "index"
	UseNonce            bool   // whether to add nonce="{{CSP_NONCE}}" to the script tag
	ExtraDeclarations   string // JS inserted after translations (e.g. docsLangs)
	SetLanguageExtra    string // JS appended inside setLanguage() after core DOM updates
	DOMContentLoadedPre string // JS at start of DOMContentLoaded handler, before setLanguage()
}

// i18nScript generates the complete i18n <script> block shared by maker, recover, and index pages.
// The JS body comes from assets/i18n.js with placeholder substitution.
func i18nScript(opts I18nScriptOptions) string {
	nonceAttr := ""
	if opts.UseNonce {
		nonceAttr = ` nonce="{{CSP_NONCE}}"`
	}

	script := i18nJSTemplate
	script = strings.Replace(script, "{{TRANSLATIONS_JSON}}", translations.GetTranslationsJS(opts.Component), 1)
	script = strings.Replace(script, "{{LANG_DETECT_ARRAY}}", translations.LangDetectJS(), 1)
	script = strings.Replace(script, "{{EXTRA_DECLARATIONS}}", opts.ExtraDeclarations, 1)
	script = strings.Replace(script, "{{SET_LANGUAGE_EXTRA}}", opts.SetLanguageExtra, 1)
	script = strings.Replace(script, "{{DOM_CONTENT_LOADED_PRE}}", opts.DOMContentLoadedPre, 1)

	return "\n  <!-- Translations -->\n  <script" + nonceAttr + ">\n    " + script + "\n  </script>"
}
