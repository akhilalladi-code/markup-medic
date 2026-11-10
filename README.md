# MarkupMedic

[![CI](https://github.com/akhilalladi-code/markup-medic/actions/workflows/ci.yml/badge.svg)](https://github.com/akhilalladi-code/markup-medic/actions/workflows/ci.yml) [![Deploy](https://github.com/akhilalladi-code/markup-medic/actions/workflows/deploy.yml/badge.svg)](https://github.com/akhilalladi-code/markup-medic/actions/workflows/deploy.yml) [![License: MIT](https://img.shields.io/badge/License-MIT-lime.svg)](LICENSE)

**A browser-based HTML accessibility preflight for developers.** Inspect markup, find common accessibility issues, and export a report without uploading your source code.

**Live demo:** https://akhilalladi-code.github.io/markup-medic/

## Features

- Paste HTML or load/drop a local .html/.htm file (2 MB limit).
- Check image alternatives, accessible names, form labels, heading order, language, title, viewport, landmarks, duplicate IDs, iframe titles, SVG names and safer new-tab links.
- Export a portable JSON report.
- Responsive interface with keyboard-visible focus states and live status announcements.
- Go audit engine compiled to WebAssembly and executed locally in the browser.
- No account, backend, API key, upload, analytics, or runtime dependency.
- Files are processed in memory; source HTML is not persisted.

## Stack

- **Go 1.23:** primary language and deterministic audit engine.
- **WebAssembly (GOOS=js GOARCH=wasm):** runs the Go engine in the browser.
- **HTML5, CSS3 and JavaScript:** accessible interface, local file handling and WebAssembly bridge.
- **Go standard library + Node.js built-ins:** no third-party runtime packages.

## Run locally

Requires Go 1.23+ and Node.js 20+.

```bash
git clone https://github.com/akhilalladi-code/markup-medic.git
cd markup-medic
go test ./...
go vet ./...
npm run build
python -m http.server 8000 --directory dist
```

Open http://localhost:8000. Serve over HTTP; do not open the built page through file://.

## Usage

1. Open the live demo or serve the generated dist/ directory.
2. Paste HTML, choose a local HTML file, or drop one into the upload area.
3. Run the audit and review errors and warnings.
4. Export findings as JSON.

## Design decisions and what I learned

- Kept the app static and dependency-free for GitHub Pages.
- Implemented audit rules in Go and compiled them to WebAssembly.
- Separated DOM collection from the pure Go rule engine so rules can be tested natively.
- Used heuristic checks with explicit limitations rather than claiming automated WCAG certification.
- Kept file processing client-side to protect unpublished source code.

## Screenshots

Add genuine captures to `docs/screenshots/` after deployment. Suggested capture: the loaded sample and its complete audit report. Do not commit fabricated screenshots.

## Scope and limitations

MarkupMedic checks common structural patterns. It cannot reliably determine whether alt text is meaningful, whether color contrast is sufficient, whether focus order is logical, or whether a complete interaction is usable with a screen reader. Combine it with manual keyboard and assistive-technology testing. Findings are not a WCAG conformance statement.

## Deployment

The included GitHub Actions workflow compiles Go to WebAssembly, builds the static site into `dist/`, and deploys it to GitHub Pages.

1. Open **Settings → Pages** and select **GitHub Actions** as the source.
2. Push to `main`; the deployment workflow publishes `dist/`.

## License

MIT © 2026 Akhil Alladi. See [LICENSE](LICENSE).
