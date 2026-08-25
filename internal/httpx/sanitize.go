package httpx

import (
	"regexp"
	"strings"

	"github.com/microcosm-cc/bluemonday"
)

var htmlTagRe = regexp.MustCompile(`(?i)<(/?)([a-z][\w]*)\b[^>]*>`)

var keepAsIs = map[string]bool{
	"br": true, "li": true, "ul": true, "ol": true, "p": true,
	"b": true, "i": true, "u": true, "em": true, "strong": true,
	"hr": true, "font": true,
}

var blockTags = map[string]bool{
	"maintext": true, "stats": true, "rules": true,
}

var leaveForSanitizer = map[string]bool{
	"script": true, "iframe": true, "object": true, "embed": true,
	"link": true, "meta": true, "style": true,
}

func descriptionPolicy() *bluemonday.Policy {
	p := bluemonday.NewPolicy()
	p.AllowElements("div", "span", "br", "li", "ul", "ol", "p", "b", "i", "u", "em", "strong", "hr", "font")
	p.AllowAttrs("class").OnElements("div", "span")
	p.AllowAttrs("color").OnElements("font")
	return p
}

func sanitizeDescription(p *bluemonday.Policy, raw string) string {
	return p.Sanitize(rewriteRiotTags(raw))
}

// rewriteRiotTags maps Data Dragon custom elements (mainText, attention, …)
// to span/div with a class, because HTML sanitizers drop unknown tags.
func rewriteRiotTags(s string) string {
	return htmlTagRe.ReplaceAllStringFunc(s, func(m string) string {
		parts := htmlTagRe.FindStringSubmatch(m)
		if len(parts) < 3 {
			return m
		}
		name := strings.ToLower(parts[2])
		if keepAsIs[name] || leaveForSanitizer[name] {
			return m
		}
		tag := "span"
		if blockTags[name] {
			tag = "div"
		}
		if parts[1] == "/" {
			return "</" + tag + ">"
		}
		return `<` + tag + ` class="` + name + `">`
	})
}
