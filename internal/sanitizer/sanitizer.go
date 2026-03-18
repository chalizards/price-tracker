package sanitizer

import (
	"regexp"
	"strings"
)

const maxHTMLLength = 50000 // ~50KB limit for the LLM

var (
	commentRegex  = regexp.MustCompile(`<!--[\s\S]*?-->`)
	scriptRegex   = regexp.MustCompile(`(?i)<script[\s\S]*?</script>`)
	styleRegex    = regexp.MustCompile(`(?i)<style[\s\S]*?</style>`)
	noscriptRegex = regexp.MustCompile(`(?i)<noscript[\s\S]*?</noscript>`)
	svgRegex      = regexp.MustCompile(`(?i)<svg[\s\S]*?</svg>`)
	iframeRegex   = regexp.MustCompile(`(?i)<iframe[\s\S]*?</iframe>`)
	footerRegex   = regexp.MustCompile(`(?i)<footer[\s\S]*?</footer>`)
	imgRegex      = regexp.MustCompile(`(?i)<img[^>]*>`)
	tagRegex      = regexp.MustCompile(`<[^>]+>`)
	spaceRegex    = regexp.MustCompile(`\s{2,}`)
)

// SanitizeForLLM removes unnecessary content from HTML before sending
// it to the LLM, reducing token usage and noise.
func SanitizeForLLM(html string) string {
	// Remove HTML comments (prompt injection vector)
	html = commentRegex.ReplaceAllString(html, "")

	// Remove script tags and content
	html = scriptRegex.ReplaceAllString(html, "")

	// Remove style tags and content
	html = styleRegex.ReplaceAllString(html, "")

	// Remove noscript, svg, iframe tags
	html = noscriptRegex.ReplaceAllString(html, "")
	html = svgRegex.ReplaceAllString(html, "")
	html = iframeRegex.ReplaceAllString(html, "")

	// Remove footer (safe — never contains product price)
	html = footerRegex.ReplaceAllString(html, "")

	// Remove img tags (not needed for price extraction)
	html = imgRegex.ReplaceAllString(html, "")

	// Strip all HTML tags, keep only text content
	html = tagRegex.ReplaceAllString(html, " ")

	// Collapse whitespace
	html = spaceRegex.ReplaceAllString(html, " ")
	html = strings.TrimSpace(html)

	// Truncate to max length
	if len(html) > maxHTMLLength {
		html = html[:maxHTMLLength]
	}

	return html
}
