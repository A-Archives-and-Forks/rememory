package html

import (
	"strings"

	"github.com/eljojo/rememory/internal/translations"
)

// GenerateIndexHTML creates the landing page HTML with embedded CSS.
func GenerateIndexHTML(selfhosted bool) string {
	content := aboutHTMLTemplate

	// Embed language picker options
	content = strings.Replace(content, "{{LANG_OPTIONS}}", translations.LangSelectOptions(), 1)

	result := applyLayout(LayoutOptions{
		Title:      "ReMemory - Protect what matters with people you trust",
		BodyClass:  "landing",
		Selfhosted: selfhosted,
		HeadMeta: `<meta name="generator" content="ReMemory {{VERSION}}">
  <meta name="description" content="Protect your files by splitting a key among people you trust. No accounts, no servers. Recovery works offline, forever.">
  <!-- Open Graph / Facebook -->
  <meta property="og:type" content="website">
  <meta property="og:title" content="ReMemory - Protect what matters with people you trust">
  <meta property="og:description" content="Protect your files by splitting a key among people you trust. No accounts, no servers. Recovery works offline, forever.">
  <meta property="og:image" content="{{GITHUB_PAGES}}/screenshots/recovery-1.png">
  <!-- Twitter -->
  <meta name="twitter:card" content="summary_large_image">
  <meta name="twitter:title" content="ReMemory - Protect what matters with people you trust">
  <meta name="twitter:description" content="Protect your files by splitting a key among people you trust. No accounts, no servers. Recovery works offline, forever.">
  <meta name="twitter:image" content="{{GITHUB_PAGES}}/screenshots/recovery-1.png">`,
		PageStyles: indexCSS,
		Content:    content,
		FooterContent: `<p style="font-size: 0.8125rem; color: #8A8480;" data-i18n-html="footer_timelock">* <a href="docs.html#timelock" style="color: #8A8480;">Time-locked</a> archives need a brief internet connection at recovery time.</p>
    <p>
      <a href="{{GITHUB_REPO}}" target="_blank" data-i18n="footer_source">Source Code</a> &#xB7;
      <a href="{{GITHUB_URL}}" target="_blank" data-i18n="footer_download">Download</a> &#xB7;
      <a href="docs.html" data-i18n="footer_docs">Documentation</a>
    </p>
    <p class="version"><a href="{{GITHUB_REPO}}/blob/main/CHANGELOG.md" target="_blank" style="color: var(--text-muted); text-decoration: none;">{{VERSION}}</a></p>`,
		Scripts: `<script>document.querySelector('#nav-links-main a[href="about.html"]')?.remove();</script>

  <script>` + dataflowJS + `</script>` + i18nScript(I18nScriptOptions{
			Component:         "index",
			ExtraDeclarations: `const docsLangs = ` + DocsLanguagesJS() + `;`,
			SetLanguageExtra:  i18nIndexSetlangJS,
		}),
	})

	return result
}
