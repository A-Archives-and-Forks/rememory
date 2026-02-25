package html

import (
	"encoding/json"
	"strings"

	"github.com/eljojo/rememory/internal/translations"
)

// GenerateHomeHTML creates the selfhosted home page with bundle data.
func GenerateHomeHTML(bundlesJSON string) string {
	content := homeHTMLTemplate
	content = strings.Replace(content, "{{BUNDLES_JSON}}", bundlesJSON, 1)

	// Language selector for nav
	navExtras := `<select class="lang-select" id="lang-select">
        ` + translations.LangSelectOptions() + `
      </select>`

	// Build scripts: i18n first, then home logic
	var scripts strings.Builder
	scripts.WriteString(i18nScript(I18nScriptOptions{
		Component:         "home",
		ExtraDeclarations: `const docsLangs = ` + DocsLanguagesJS() + `;`,
	}))
	scripts.WriteString("\n  <script>" + strings.Replace(homeJS, "{{BUNDLES_JSON}}", bundlesJSON, 1) + "</script>")

	result := applyLayout(LayoutOptions{
		Title:         "ReMemory",
		Selfhosted:    true,
		PageStyles:    homeCSS,
		NavExtras:     navExtras,
		Content:       content,
		FooterContent: `<p>ReMemory</p><p class="version">{{VERSION}}</p>`,
		Scripts:       scripts.String(),
	})

	return result
}

// HomeBundlesJSON serializes bundle metadata to JSON for the home page.
func HomeBundlesJSON(bundles any) string {
	data, _ := json.Marshal(bundles)
	return string(data)
}
