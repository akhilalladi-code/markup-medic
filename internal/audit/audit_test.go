package audit

import "testing"

func el(tag string, attrs map[string]string, text string) Element {
	return Element{Tag: tag, Attrs: attrs, Text: text}
}

func rules(report Report) map[string]bool {
	out := map[string]bool{}
	for _, finding := range report.Findings {
		out[finding.Rule] = true
	}
	return out
}

func TestFindsMissingAccessibleNames(t *testing.T) {
	report := Run([]Element{el("img", nil, ""), el("a", nil, ""), el("button", nil, "")}, Metadata{})
	got := rules(report)
	for _, name := range []string{"img-alt", "link-name", "button-name"} {
		if !got[name] {
			t.Errorf("expected rule %s", name)
		}
	}
}

func TestAcceptsDecorativeImageAndLabelledControl(t *testing.T) {
	report := Run([]Element{
		el("img", map[string]string{"alt": ""}, ""),
		el("input", map[string]string{"id": "email"}, ""),
	}, Metadata{LabelledIDs: []string{"email"}, Title: true, Viewport: true})
	got := rules(report)
	if got["img-alt"] || got["form-label"] {
		t.Fatalf("unexpected findings: %#v", report.Findings)
	}
}

func TestDuplicateIDsAndHeadingSkips(t *testing.T) {
	report := Run([]Element{
		el("h1", nil, "Title"),
		el("h3", nil, "Subheading"),
		el("div", map[string]string{"id": "x"}, ""),
		el("div", map[string]string{"id": "x"}, ""),
	}, Metadata{Title: true, Viewport: true})
	got := rules(report)
	if !got["duplicate-id"] || !got["heading-order"] {
		t.Fatalf("expected duplicate id and heading order findings: %#v", report.Findings)
	}
}

func TestDocumentMetadataAndLandmarks(t *testing.T) {
	report := Run([]Element{el("html", nil, ""), el("body", nil, "")}, Metadata{})
	got := rules(report)
	for _, name := range []string{"html-lang", "document-title", "viewport-meta", "missing-main", "missing-h1"} {
		if !got[name] {
			t.Errorf("expected rule %s", name)
		}
	}
}

func TestCleanDocument(t *testing.T) {
	report := Run([]Element{
		el("html", map[string]string{"lang": "en"}, ""),
		el("main", nil, ""),
		el("h1", nil, "Hello"),
		el("img", map[string]string{"alt": "A landscape"}, ""),
	}, Metadata{Title: true, Viewport: true})
	if report.Summary.Errors != 0 || report.Summary.Warnings != 0 || report.Summary.Elements != 4 {
		t.Fatalf("unexpected summary: %#v", report.Summary)
	}
}

func TestNewTabAndEmbeddedContentWarnings(t *testing.T) {
	report := Run([]Element{
		el("a", map[string]string{"target": "_blank", "href": "/docs"}, "Read docs"),
		el("iframe", nil, ""),
		el("svg", nil, ""),
	}, Metadata{Title: true, Viewport: true})
	got := rules(report)
	for _, name := range []string{"blank-noopener", "iframe-title", "svg-name"} {
		if !got[name] {
			t.Errorf("expected rule %s", name)
		}
	}
}
