package audit

import (
	"fmt"
	"strings"
)

type Element struct {
	Tag   string            `json:"tag"`
	Attrs map[string]string `json:"attrs"`
	Text  string            `json:"text"`
}

type Metadata struct {
	Title       bool     `json:"title"`
	Viewport    bool     `json:"viewport"`
	LabelledIDs []string `json:"labelledIds"`
}

type Finding struct {
	Severity string `json:"severity"`
	Rule     string `json:"rule"`
	Title    string `json:"title"`
	Detail   string `json:"detail"`
	Index    *int   `json:"index"`
}

type Summary struct {
	Errors   int `json:"errors"`
	Warnings int `json:"warnings"`
	Passed   int `json:"passed"`
	Total    int `json:"total"`
	Elements int `json:"elements"`
}

type Report struct {
	Findings []Finding `json:"findings"`
	Summary  Summary   `json:"summary"`
}

func has(value string) bool               { return strings.TrimSpace(value) != "" }
func (e Element) attr(name string) string { return e.Attrs[name] }

// Run evaluates common, deterministic HTML accessibility preflight rules.
func Run(elements []Element, metadata Metadata) Report {
	findings := make([]Finding, 0)
	ids := map[string]int{}
	labelled := map[string]bool{}
	for _, id := range metadata.LabelledIDs {
		labelled[id] = true
	}
	type heading struct{ level, index int }
	headings := make([]heading, 0)
	hasH1, hasMain := false, false
	hasTableHeader := false
	for _, el := range elements {
		if strings.EqualFold(el.Tag, "th") {
			hasTableHeader = true
		}
	}
	add := func(severity, rule, title, detail string, index *int) {
		findings = append(findings, Finding{Severity: severity, Rule: rule, Title: title, Detail: detail, Index: index})
	}
	for i, el := range elements {
		tag := strings.ToLower(el.Tag)
		idx := i
		id := el.attr("id")
		if id != "" {
			ids[id]++
		}
		if len(tag) == 2 && tag[0] == 'h' && tag[1] >= '1' && tag[1] <= '6' {
			level := int(tag[1] - '0')
			headings = append(headings, heading{level, i})
			if level == 1 {
				hasH1 = true
			}
		}
		if tag == "main" {
			hasMain = true
		}
		switch tag {
		case "img":
			if _, ok := el.Attrs["alt"]; !ok {
				add("error", "img-alt", "Image is missing alternative text", "Add an alt attribute. Use alt=\"\" only when the image is decorative.", &idx)
			}
		case "a":
			if !has(el.Text) && !has(el.attr("aria-label")) && !has(el.attr("aria-labelledby")) {
				add("error", "link-name", "Link has no accessible name", "Add descriptive link text or an accessible label.", &idx)
			}
			if el.attr("target") == "_blank" && !containsToken(el.attr("rel"), "noopener") {
				add("warning", "blank-noopener", "New-tab link should use rel=\"noopener\"", "Add rel=\"noopener\" (usually with noreferrer) when opening a new tab.", &idx)
			}
		case "button":
			if !has(el.Text) && !has(el.attr("aria-label")) && !has(el.attr("aria-labelledby")) {
				add("error", "button-name", "Button has no accessible name", "Add visible button text or an accessible label.", &idx)
			}
		case "input", "select", "textarea":
			if el.attr("type") != "hidden" {
				id := el.attr("id")
				labelOK := has(el.attr("aria-label")) || has(el.attr("aria-labelledby")) || has(el.attr("title")) || (id != "" && labelled[id])
				if !labelOK {
					add("error", "form-label", "Form control has no programmatic label", "Add a matching <label>, aria-label, or aria-labelledby.", &idx)
				}
			}
		case "html":
			if !has(el.attr("lang")) {
				add("error", "html-lang", "Document language is missing", "Set a valid lang attribute on the <html> element.", &idx)
			}
		case "table":
			if !hasTableHeader {
				add("warning", "table-headers", "Table has no header cells", "Use <th> elements and appropriate scope attributes for data tables.", &idx)
			}
		case "iframe":
			if !has(el.attr("title")) {
				add("error", "iframe-title", "Embedded frame has no title", "Give each iframe a concise, descriptive title.", &idx)
			}
		case "svg":
			if !has(el.attr("aria-label")) && el.attr("aria-hidden") != "true" && !has(el.attr("role")) {
				add("warning", "svg-name", "SVG may need an accessible name", "Add a title/accessible label or mark decorative SVGs aria-hidden=\"true\".", &idx)
			}
		}
	}
	for id, count := range ids {
		if count > 1 {
			add("error", "duplicate-id", fmt.Sprintf("Duplicate id: %s", id), fmt.Sprintf("This id appears %d times. IDs must be unique.", count), nil)
		}
	}
	for i := 1; i < len(headings); i++ {
		if headings[i].level > headings[i-1].level+1 {
			idx := headings[i].index
			add("warning", "heading-order", "Heading level skips a step", fmt.Sprintf("Heading level jumps from H%d to H%d.", headings[i-1].level, headings[i].level), &idx)
		}
	}
	if !hasH1 {
		add("warning", "missing-h1", "No level-one heading found", "Add a clear page-level H1 that describes the main content.", nil)
	}
	if !hasMain {
		add("warning", "missing-main", "No main landmark found", "Wrap the primary page content in a <main> element.", nil)
	}
	if !metadata.Title {
		add("error", "document-title", "Document title is missing", "Add a meaningful <title> inside the document head.", nil)
	}
	if !metadata.Viewport {
		add("warning", "viewport-meta", "Viewport meta tag is missing", "Add a responsive viewport meta tag for mobile browsers.", nil)
	}
	errors, warnings := 0, 0
	for _, finding := range findings {
		if finding.Severity == "error" {
			errors++
		} else if finding.Severity == "warning" {
			warnings++
		}
	}
	return Report{Findings: findings, Summary: Summary{Errors: errors, Warnings: warnings, Passed: max(0, 12-len(findings)), Total: len(findings), Elements: len(elements)}}
}

func containsToken(value, token string) bool {
	for _, item := range strings.Fields(strings.ToLower(value)) {
		if item == token {
			return true
		}
	}
	return false
}
