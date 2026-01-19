// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: BUSL-1.1

package report

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"os"
	"strings"

	"github.com/microcosm-cc/bluemonday"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/renderer/html"
)

// HTMLExporter exports reports to HTML format.
type HTMLExporter struct{}

// NewHTMLExporter creates a new HTML exporter.
func NewHTMLExporter() *HTMLExporter {
	return &HTMLExporter{}
}

// Export writes the report to an HTML file.
func (e *HTMLExporter) Export(report *Report, path string) error {
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create HTML file: %w", err)
	}
	defer func() {
		_ = file.Close() //nolint:errcheck // Best effort close
	}()

	return e.ExportToWriter(report, file)
}

// htmlTemplateData contains all data passed to the HTML template.
type htmlTemplateData struct {
	Report         *Report
	Summary        *Summary
	FailedRows     []Row
	PassedRows     []Row
	SkippedRows    []Row
	PassPercent    float64
	FailPercent    float64
	SkipPercent    float64
	SeverityCounts map[string]int
}

// ExportToWriter writes the report to an io.Writer in HTML format.
func (e *HTMLExporter) ExportToWriter(report *Report, w io.Writer) error {
	summary := report.Summary()

	// Separate rows by status
	var failedRows, passedRows, skippedRows []Row
	for _, row := range report.Rows {
		switch row.CheckStatus {
		case StatusFail, StatusFailed:
			failedRows = append(failedRows, row)
		case StatusPass, StatusPassed:
			passedRows = append(passedRows, row)
		case StatusSkip, StatusSkipped:
			skippedRows = append(skippedRows, row)
		}
	}

	// Calculate percentages
	total := float64(summary.TotalChecks)
	var passPercent, failPercent, skipPercent float64
	if total > 0 {
		passPercent = float64(summary.PassedChecks) / total * 100
		failPercent = float64(summary.FailedChecks) / total * 100
		skipPercent = float64(summary.SkippedChecks) / total * 100
	}

	data := htmlTemplateData{
		Report:         report,
		Summary:        summary,
		FailedRows:     failedRows,
		PassedRows:     passedRows,
		SkippedRows:    skippedRows,
		PassPercent:    passPercent,
		FailPercent:    failPercent,
		SkipPercent:    skipPercent,
		SeverityCounts: summary.FailedBySeverity,
	}

	tmpl, err := template.New("report").Funcs(template.FuncMap{
		"lower":          strings.ToLower,
		"severityClass":  severityClass,
		"statusClass":    statusClass,
		"gradeClass":     gradeClass,
		"renderMarkdown": renderMarkdown,
	}).Parse(htmlTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse HTML template: %w", err)
	}

	if err := tmpl.Execute(w, data); err != nil {
		return fmt.Errorf("failed to execute HTML template: %w", err)
	}

	return nil
}

// severityClass returns a CSS class for a severity level.
func severityClass(severity string) string {
	switch strings.ToLower(severity) {
	case "critical":
		return "severity-critical"
	case "high":
		return "severity-high"
	case "medium":
		return "severity-medium"
	case "low":
		return "severity-low"
	default:
		return "severity-info"
	}
}

// statusClass returns a CSS class for a check status.
func statusClass(status string) string {
	switch strings.ToLower(status) {
	case "pass", "passed":
		return "status-pass"
	case "fail", "failed":
		return "status-fail"
	case "skip", "skipped":
		return "status-skip"
	default:
		return "status-unknown"
	}
}

// gradeClass returns a CSS class for a grade.
func gradeClass(grade string) string {
	switch strings.ToUpper(grade) {
	case "A":
		return "grade-a"
	case "B":
		return "grade-b"
	case "C":
		return "grade-c"
	case "D":
		return "grade-d"
	default:
		return "grade-f"
	}
}

// mdRenderer is the configured goldmark markdown parser with GFM extensions.
var mdRenderer = goldmark.New(
	goldmark.WithExtensions(
		extension.GFM, // GitHub Flavored Markdown (tables, strikethrough, autolinks, task lists)
	),
	goldmark.WithRendererOptions(
		html.WithHardWraps(), // Convert newlines to <br>
	),
)

// htmlSanitizer is the configured bluemonday policy for sanitizing HTML output.
// Uses UGCPolicy which allows safe HTML for user-generated content.
var htmlSanitizer = bluemonday.UGCPolicy()

// renderMarkdown converts GitHub Flavored Markdown to sanitized HTML.
// Uses goldmark for parsing and bluemonday for XSS protection.
func renderMarkdown(md string) template.HTML {
	if md == "" {
		return ""
	}

	// Convert markdown to HTML using goldmark
	var buf bytes.Buffer
	if err := mdRenderer.Convert([]byte(md), &buf); err != nil {
		// On error, return escaped plain text
		return template.HTML("<p>" + template.HTMLEscapeString(md) + "</p>") //nolint:gosec // Escaped
	}

	// Sanitize HTML output to prevent XSS
	sanitized := htmlSanitizer.SanitizeBytes(buf.Bytes())

	return template.HTML(sanitized) //nolint:gosec // Sanitized by bluemonday
}

const htmlTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>kspec Security Report - {{.Report.Metadata.ScanTarget}}</title>
    <style>
        :root {
            --color-bg: #0d1117;
            --color-bg-secondary: #161b22;
            --color-bg-tertiary: #21262d;
            --color-border: #30363d;
            --color-text: #c9d1d9;
            --color-text-muted: #8b949e;
            --color-text-heading: #f0f6fc;
            --color-pass: #3fb950;
            --color-fail: #f85149;
            --color-skip: #8b949e;
            --color-critical: #ff7b72;
            --color-high: #ffa657;
            --color-medium: #d29922;
            --color-low: #58a6ff;
            --color-info: #8b949e;
        }

        * {
            box-sizing: border-box;
            margin: 0;
            padding: 0;
        }

        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', 'Noto Sans', Helvetica, Arial, sans-serif;
            background-color: var(--color-bg);
            color: var(--color-text);
            line-height: 1.5;
            padding: 2rem;
        }

        .container {
            max-width: 1400px;
            margin: 0 auto;
        }

        header {
            margin-bottom: 2rem;
            padding-bottom: 1rem;
            border-bottom: 1px solid var(--color-border);
        }

        h1 {
            color: var(--color-text-heading);
            font-size: 1.75rem;
            font-weight: 600;
            margin-bottom: 0.5rem;
        }

        .subtitle {
            color: var(--color-text-muted);
            font-size: 0.875rem;
        }

        .meta-info {
            display: flex;
            gap: 2rem;
            margin-top: 1rem;
            font-size: 0.875rem;
        }

        .meta-item {
            display: flex;
            align-items: center;
            gap: 0.5rem;
        }

        .meta-label {
            color: var(--color-text-muted);
        }

        .meta-value {
            color: var(--color-text);
            font-weight: 500;
        }

        .summary-grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
            gap: 1rem;
            margin-bottom: 2rem;
        }

        .summary-card {
            background-color: var(--color-bg-secondary);
            border: 1px solid var(--color-border);
            border-radius: 6px;
            padding: 1.25rem;
        }

        .summary-card h3 {
            color: var(--color-text-muted);
            font-size: 0.75rem;
            font-weight: 500;
            text-transform: uppercase;
            letter-spacing: 0.05em;
            margin-bottom: 0.5rem;
        }

        .summary-value {
            font-size: 2rem;
            font-weight: 600;
            color: var(--color-text-heading);
        }

        .summary-value.pass { color: var(--color-pass); }
        .summary-value.fail { color: var(--color-fail); }
        .summary-value.skip { color: var(--color-skip); }

        .summary-value.grade-a { color: var(--color-pass); }
        .summary-value.grade-b { color: #58a6ff; }
        .summary-value.grade-c { color: var(--color-medium); }
        .summary-value.grade-d { color: var(--color-high); }
        .summary-value.grade-f { color: var(--color-fail); }

        .score-card {
            background: linear-gradient(135deg, var(--color-bg-secondary) 0%, var(--color-bg-tertiary) 100%);
            border: 1px solid var(--color-border);
            border-radius: 6px;
            padding: 1.25rem;
            grid-column: span 2;
        }

        .score-card .score-header {
            display: flex;
            justify-content: space-between;
            align-items: flex-start;
        }

        .score-card .score-main {
            display: flex;
            align-items: baseline;
            gap: 0.5rem;
            margin-bottom: 0.5rem;
        }

        .score-card .score-value {
            font-size: 3rem;
            font-weight: 700;
        }

        .score-card .score-max {
            font-size: 1.5rem;
            color: var(--color-text-muted);
        }

        .score-card .score-details {
            display: flex;
            gap: 1.5rem;
            font-size: 0.875rem;
        }

        .score-card .score-detail {
            display: flex;
            align-items: center;
            gap: 0.5rem;
        }

        .score-card .score-label {
            color: var(--color-text-muted);
        }

        .score-gauge {
            width: 100%;
            height: 12px;
            background: linear-gradient(to right,
                var(--color-fail) 0%,
                var(--color-fail) 40%,
                var(--color-high) 40%,
                var(--color-high) 60%,
                var(--color-medium) 60%,
                var(--color-medium) 70%,
                #58a6ff 70%,
                #58a6ff 90%,
                var(--color-pass) 90%,
                var(--color-pass) 100%
            );
            border-radius: 6px;
            margin: 1rem 0 0.5rem 0;
            position: relative;
        }

        .score-gauge-marker {
            position: absolute;
            top: -4px;
            width: 4px;
            height: 20px;
            background: white;
            border-radius: 2px;
            box-shadow: 0 0 4px rgba(0,0,0,0.5);
            transform: translateX(-50%);
        }

        .score-legend {
            display: flex;
            justify-content: space-between;
            font-size: 0.7rem;
            color: var(--color-text-muted);
            margin-bottom: 1rem;
        }

        .score-legend-item {
            display: flex;
            flex-direction: column;
            align-items: center;
            gap: 0.25rem;
        }

        .score-legend-dot {
            width: 10px;
            height: 10px;
            border-radius: 50%;
        }

        .score-legend-dot.critical { background-color: var(--color-fail); }
        .score-legend-dot.high { background-color: var(--color-high); }
        .score-legend-dot.medium { background-color: var(--color-medium); }
        .score-legend-dot.low { background-color: #58a6ff; }
        .score-legend-dot.none { background-color: var(--color-pass); }

        .findings-badges {
            display: flex;
            gap: 0.5rem;
            flex-wrap: wrap;
            margin-top: 0.75rem;
        }

        .findings-badge {
            display: inline-flex;
            align-items: center;
            gap: 0.25rem;
            padding: 0.25rem 0.5rem;
            border-radius: 4px;
            font-size: 0.75rem;
            font-weight: 500;
        }

        .findings-badge.critical {
            background-color: rgba(255, 123, 114, 0.15);
            color: var(--color-critical);
        }

        .findings-badge.high {
            background-color: rgba(255, 166, 87, 0.15);
            color: var(--color-high);
        }

        .findings-badge.medium {
            background-color: rgba(210, 153, 34, 0.15);
            color: var(--color-medium);
        }

        .findings-badge.low {
            background-color: rgba(88, 166, 255, 0.15);
            color: var(--color-low);
        }

        .findings-badge.info {
            background-color: rgba(139, 148, 158, 0.15);
            color: var(--color-info);
        }

        .scoring-explanation {
            background-color: var(--color-bg-secondary);
            border: 1px solid var(--color-border);
            border-radius: 6px;
            padding: 1.5rem;
            margin-top: 2rem;
        }

        .scoring-explanation h3 {
            color: var(--color-text-heading);
            font-size: 1rem;
            margin-bottom: 1rem;
        }

        .scoring-explanation p {
            color: var(--color-text-muted);
            font-size: 0.875rem;
            line-height: 1.6;
            margin-bottom: 1rem;
        }

        .scoring-explanation p:last-child {
            margin-bottom: 0;
        }

        .scoring-explanation .scoring-method {
            background-color: var(--color-bg-tertiary);
            border-radius: 4px;
            padding: 1rem;
            margin-top: 1rem;
        }

        .scoring-explanation .scoring-method h4 {
            color: var(--color-text-heading);
            font-size: 0.875rem;
            margin-bottom: 0.5rem;
        }

        .scoring-explanation .scoring-method ul {
            list-style: none;
            padding: 0;
            margin: 0;
        }

        .scoring-explanation .scoring-method li {
            color: var(--color-text);
            font-size: 0.8125rem;
            padding: 0.25rem 0;
            display: flex;
            align-items: center;
            gap: 0.5rem;
        }

        .scoring-explanation .scoring-method li::before {
            content: "•";
            color: var(--color-text-muted);
        }

        .progress-bar {
            display: flex;
            height: 8px;
            border-radius: 4px;
            overflow: hidden;
            background-color: var(--color-bg-tertiary);
            margin: 1.5rem 0;
        }

        .progress-segment {
            height: 100%;
            transition: width 0.3s ease;
        }

        .progress-pass { background-color: var(--color-pass); }
        .progress-fail { background-color: var(--color-fail); }
        .progress-skip { background-color: var(--color-skip); }

        .progress-legend {
            display: flex;
            gap: 1.5rem;
            font-size: 0.875rem;
        }

        .legend-item {
            display: flex;
            align-items: center;
            gap: 0.5rem;
        }

        .legend-dot {
            width: 10px;
            height: 10px;
            border-radius: 50%;
        }

        .legend-dot.pass { background-color: var(--color-pass); }
        .legend-dot.fail { background-color: var(--color-fail); }
        .legend-dot.skip { background-color: var(--color-skip); }

        .section {
            margin-bottom: 2rem;
        }

        .section-header {
            display: flex;
            align-items: center;
            justify-content: space-between;
            margin-bottom: 1rem;
        }

        h2 {
            color: var(--color-text-heading);
            font-size: 1.25rem;
            font-weight: 600;
        }

        .badge {
            display: inline-flex;
            align-items: center;
            padding: 0.25rem 0.75rem;
            border-radius: 9999px;
            font-size: 0.75rem;
            font-weight: 500;
        }

        .badge-fail {
            background-color: rgba(248, 81, 73, 0.15);
            color: var(--color-fail);
        }

        .badge-pass {
            background-color: rgba(63, 185, 80, 0.15);
            color: var(--color-pass);
        }

        .badge-skip {
            background-color: rgba(139, 148, 158, 0.15);
            color: var(--color-skip);
        }

        .severity-badges {
            display: flex;
            gap: 0.5rem;
            margin-bottom: 1rem;
        }

        .severity-badge {
            display: inline-flex;
            align-items: center;
            gap: 0.25rem;
            padding: 0.25rem 0.5rem;
            border-radius: 4px;
            font-size: 0.75rem;
            font-weight: 500;
        }

        .severity-critical {
            background-color: rgba(255, 123, 114, 0.15);
            color: var(--color-critical);
        }

        .severity-high {
            background-color: rgba(255, 166, 87, 0.15);
            color: var(--color-high);
        }

        .severity-medium {
            background-color: rgba(210, 153, 34, 0.15);
            color: var(--color-medium);
        }

        .severity-low {
            background-color: rgba(88, 166, 255, 0.15);
            color: var(--color-low);
        }

        .severity-info {
            background-color: rgba(139, 148, 158, 0.15);
            color: var(--color-info);
        }

        table {
            width: 100%;
            border-collapse: collapse;
            background-color: var(--color-bg-secondary);
            border: 1px solid var(--color-border);
            border-radius: 6px;
            overflow: hidden;
        }

        th, td {
            padding: 0.75rem 1rem;
            text-align: left;
            border-bottom: 1px solid var(--color-border);
        }

        th {
            background-color: var(--color-bg-tertiary);
            color: var(--color-text-heading);
            font-weight: 600;
            font-size: 0.75rem;
            text-transform: uppercase;
            letter-spacing: 0.05em;
        }

        tr:last-child td {
            border-bottom: none;
        }

        tr:hover {
            background-color: var(--color-bg-tertiary);
        }

        .status-pass { color: var(--color-pass); }
        .status-fail { color: var(--color-fail); }
        .status-skip { color: var(--color-skip); }

        .resource-path {
            color: var(--color-text-muted);
            font-size: 0.75rem;
        }

        .check-details {
            color: var(--color-text-muted);
            font-size: 0.875rem;
            max-width: 400px;
        }

        .empty-state {
            text-align: center;
            padding: 3rem;
            color: var(--color-text-muted);
        }

        footer {
            margin-top: 3rem;
            padding: 2rem;
            border-top: 1px solid var(--color-border);
            text-align: center;
            color: var(--color-text-muted);
            background-color: var(--color-bg-secondary);
            border-radius: 6px;
        }

        .footer-brand {
            display: flex;
            align-items: center;
            justify-content: center;
            gap: 0.5rem;
            margin-bottom: 1rem;
        }

        .footer-logo {
            font-size: 1.25rem;
            font-weight: 700;
            color: var(--color-text-heading);
        }

        .footer-logo span {
            color: #6366f1;
        }

        .footer-tagline {
            font-size: 0.875rem;
            margin-bottom: 1rem;
        }

        .footer-cta {
            display: inline-flex;
            align-items: center;
            gap: 0.5rem;
            padding: 0.625rem 1.25rem;
            background: linear-gradient(135deg, #6366f1 0%, #8b5cf6 100%);
            color: white;
            text-decoration: none;
            border-radius: 6px;
            font-weight: 500;
            font-size: 0.875rem;
            transition: transform 0.2s, box-shadow 0.2s;
        }

        .footer-cta:hover {
            transform: translateY(-1px);
            box-shadow: 0 4px 12px rgba(99, 102, 241, 0.4);
        }

        .footer-links {
            margin-top: 1.5rem;
            font-size: 0.75rem;
        }

        .footer-links a {
            color: var(--color-text-muted);
            text-decoration: none;
            margin: 0 0.75rem;
        }

        .footer-links a:hover {
            color: var(--color-text);
        }

        .collapsible-header {
            cursor: pointer;
            user-select: none;
        }

        .collapsible-header:hover {
            opacity: 0.8;
        }

        .collapsible-header::before {
            content: '▼';
            display: inline-block;
            margin-right: 0.5rem;
            font-size: 0.75rem;
            transition: transform 0.2s;
        }

        .collapsible-header.collapsed::before {
            transform: rotate(-90deg);
        }

        .collapsible-content {
            overflow: hidden;
            transition: max-height 0.3s ease-out;
        }

        .collapsible-content.collapsed {
            max-height: 0 !important;
        }

        .filters {
            display: flex;
            gap: 1rem;
            margin-bottom: 1rem;
            flex-wrap: wrap;
            align-items: center;
        }

        .filter-group {
            display: flex;
            align-items: center;
            gap: 0.5rem;
        }

        .filter-label {
            font-size: 0.75rem;
            color: var(--color-text-muted);
            text-transform: uppercase;
            letter-spacing: 0.05em;
        }

        .filter-select {
            background-color: var(--color-bg-tertiary);
            border: 1px solid var(--color-border);
            border-radius: 6px;
            color: var(--color-text);
            padding: 0.5rem 0.75rem;
            font-size: 0.875rem;
            cursor: pointer;
            min-width: 150px;
        }

        .filter-select:focus {
            outline: none;
            border-color: #6366f1;
        }

        .filter-input {
            background-color: var(--color-bg-tertiary);
            border: 1px solid var(--color-border);
            border-radius: 6px;
            color: var(--color-text);
            padding: 0.5rem 0.75rem;
            font-size: 0.875rem;
            min-width: 200px;
        }

        .filter-input:focus {
            outline: none;
            border-color: #6366f1;
        }

        .filter-input::placeholder {
            color: var(--color-text-muted);
        }

        .filter-count {
            font-size: 0.75rem;
            color: var(--color-text-muted);
            margin-left: auto;
        }

        .hidden-row {
            display: none !important;
        }

        .expandable-row {
            cursor: pointer;
        }

        .expandable-row:hover {
            background-color: var(--color-bg-tertiary);
        }

        .expand-icon {
            display: inline-block;
            width: 16px;
            font-size: 0.75rem;
            color: var(--color-text-muted);
            transition: transform 0.2s;
        }

        .expandable-row.expanded .expand-icon {
            transform: rotate(90deg);
        }

        .details-row {
            display: none;
            background-color: var(--color-bg-tertiary);
        }

        .details-row.visible {
            display: table-row;
        }

        .details-row td {
            padding: 1rem 1.5rem;
            border-bottom: 1px solid var(--color-border);
        }

        .details-content {
            font-size: 0.875rem;
            color: var(--color-text);
        }

        .details-label {
            font-size: 0.75rem;
            color: var(--color-text-muted);
            text-transform: uppercase;
            letter-spacing: 0.05em;
            margin-bottom: 0.25rem;
        }

        .details-grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
            gap: 1rem;
        }

        .details-item {
            padding: 0.5rem 0;
        }

        .details-item.full-width {
            grid-column: 1 / -1;
        }

        .markdown-content {
            font-size: 0.875rem;
            line-height: 1.6;
        }

        .markdown-content p {
            margin: 0 0 0.5rem 0;
        }

        .markdown-content p:last-child {
            margin-bottom: 0;
        }

        .markdown-content code {
            background-color: var(--color-bg);
            padding: 0.125rem 0.375rem;
            border-radius: 3px;
            font-family: ui-monospace, SFMono-Regular, 'SF Mono', Menlo, Consolas, monospace;
            font-size: 0.8125rem;
        }

        .markdown-content pre {
            background-color: var(--color-bg);
            padding: 0.75rem 1rem;
            border-radius: 6px;
            overflow-x: auto;
            margin: 0.5rem 0;
        }

        .markdown-content pre code {
            padding: 0;
            background: none;
        }

        .markdown-content a {
            color: #58a6ff;
            text-decoration: none;
        }

        .markdown-content a:hover {
            text-decoration: underline;
        }

        .markdown-content strong {
            color: var(--color-text-heading);
        }

        .markdown-content em {
            font-style: italic;
        }

        .markdown-content del {
            color: var(--color-text-muted);
            text-decoration: line-through;
        }

        /* GFM Tables */
        .markdown-content table {
            border-collapse: collapse;
            width: 100%;
            margin: 0.5rem 0;
            font-size: 0.8125rem;
        }

        .markdown-content th,
        .markdown-content td {
            border: 1px solid var(--color-border);
            padding: 0.5rem 0.75rem;
            text-align: left;
        }

        .markdown-content th {
            background-color: var(--color-bg);
            font-weight: 600;
            color: var(--color-text-heading);
        }

        .markdown-content tr:nth-child(even) {
            background-color: rgba(255, 255, 255, 0.02);
        }

        /* GFM Task Lists */
        .markdown-content ul {
            list-style: disc;
            padding-left: 1.5rem;
            margin: 0.5rem 0;
        }

        .markdown-content ol {
            list-style: decimal;
            padding-left: 1.5rem;
            margin: 0.5rem 0;
        }

        .markdown-content li {
            margin: 0.25rem 0;
        }

        .markdown-content input[type="checkbox"] {
            margin-right: 0.5rem;
            vertical-align: middle;
        }

        /* Blockquotes */
        .markdown-content blockquote {
            border-left: 3px solid var(--color-border);
            padding-left: 1rem;
            margin: 0.5rem 0;
            color: var(--color-text-muted);
        }

        /* Horizontal rule */
        .markdown-content hr {
            border: none;
            border-top: 1px solid var(--color-border);
            margin: 1rem 0;
        }

        /* Headings in markdown */
        .markdown-content h1,
        .markdown-content h2,
        .markdown-content h3,
        .markdown-content h4 {
            color: var(--color-text-heading);
            margin: 0.75rem 0 0.5rem 0;
            font-weight: 600;
        }

        .markdown-content h1 { font-size: 1.25rem; }
        .markdown-content h2 { font-size: 1.125rem; }
        .markdown-content h3 { font-size: 1rem; }
        .markdown-content h4 { font-size: 0.875rem; }

        @media (max-width: 768px) {
            body {
                padding: 1rem;
            }

            .meta-info {
                flex-direction: column;
                gap: 0.5rem;
            }

            .summary-grid {
                grid-template-columns: 1fr 1fr;
            }

            table {
                font-size: 0.875rem;
            }

            th, td {
                padding: 0.5rem;
            }
        }
    </style>
</head>
<body>
    <div class="container">
        <header>
            <h1>Security Compliance Report</h1>
            <p class="subtitle">{{.Report.Metadata.ScanTarget}}</p>
            <div class="meta-info">
                <div class="meta-item">
                    <span class="meta-label">Provider:</span>
                    <span class="meta-value">{{.Report.Metadata.Provider}}</span>
                </div>
                <div class="meta-item">
                    <span class="meta-label">Generated:</span>
                    <span class="meta-value">{{.Report.Metadata.GeneratedAt.Format "2006-01-02 15:04:05 MST"}}</span>
                </div>
                <div class="meta-item">
                    <span class="meta-label">Resources:</span>
                    <span class="meta-value">{{.Summary.TotalResources}}</span>
                </div>
            </div>
        </header>

        <div class="summary-grid">
            <div class="score-card">
                <div class="score-header">
                    <div>
                        <h3>Security Score</h3>
                        <div class="score-main">
                            <span class="score-value {{gradeClass .Report.Metadata.Grade}}">{{.Report.Metadata.Score}}</span>
                            <span class="score-max">/100</span>
                        </div>
                    </div>
                    <div class="score-details" style="flex-direction: column; align-items: flex-end; gap: 0.25rem;">
                        <div class="score-detail">
                            <span class="score-label">Grade:</span>
                            <span class="{{gradeClass .Report.Metadata.Grade}}" style="font-weight: 600; font-size: 1.25rem;">{{.Report.Metadata.Grade}}</span>
                        </div>
                        <div class="score-detail">
                            <span class="score-label">Risk Level:</span>
                            <span>{{.Report.Metadata.RiskLevel}}</span>
                        </div>
                    </div>
                </div>
                <div class="score-gauge">
                    <div class="score-gauge-marker" style="left: {{.Report.Metadata.Score}}%;"></div>
                </div>
                <div class="score-legend">
                    <div class="score-legend-item">
                        <span class="score-legend-dot critical"></span>
                        <span>0-39</span>
                        <span>Critical</span>
                    </div>
                    <div class="score-legend-item">
                        <span class="score-legend-dot high"></span>
                        <span>40-59</span>
                        <span>High</span>
                    </div>
                    <div class="score-legend-item">
                        <span class="score-legend-dot medium"></span>
                        <span>60-69</span>
                        <span>Medium</span>
                    </div>
                    <div class="score-legend-item">
                        <span class="score-legend-dot low"></span>
                        <span>70-89</span>
                        <span>Low</span>
                    </div>
                    <div class="score-legend-item">
                        <span class="score-legend-dot none"></span>
                        <span>90-100</span>
                        <span>None</span>
                    </div>
                </div>
                {{if or .Report.Metadata.CriticalFindings .Report.Metadata.HighFindings .Report.Metadata.MediumFindings .Report.Metadata.LowFindings}}
                <div class="findings-badges">
                    {{if .Report.Metadata.CriticalFindings}}<span class="findings-badge critical">{{.Report.Metadata.CriticalFindings}} Critical</span>{{end}}
                    {{if .Report.Metadata.HighFindings}}<span class="findings-badge high">{{.Report.Metadata.HighFindings}} High</span>{{end}}
                    {{if .Report.Metadata.MediumFindings}}<span class="findings-badge medium">{{.Report.Metadata.MediumFindings}} Medium</span>{{end}}
                    {{if .Report.Metadata.LowFindings}}<span class="findings-badge low">{{.Report.Metadata.LowFindings}} Low</span>{{end}}
                    {{if .Report.Metadata.InfoFindings}}<span class="findings-badge info">{{.Report.Metadata.InfoFindings}} Info</span>{{end}}
                </div>
                {{end}}
                <div style="font-size: 0.75rem; color: var(--color-text-muted); margin-top: 0.5rem;">
                    Scoring System: {{.Report.Metadata.ScoringSystem}}
                </div>
            </div>
            <div class="summary-card">
                <h3>Total Checks</h3>
                <div class="summary-value">{{.Summary.TotalChecks}}</div>
            </div>
            <div class="summary-card">
                <h3>Passed</h3>
                <div class="summary-value pass">{{.Summary.PassedChecks}}</div>
            </div>
            <div class="summary-card">
                <h3>Failed</h3>
                <div class="summary-value fail">{{.Summary.FailedChecks}}</div>
            </div>
            <div class="summary-card">
                <h3>Skipped</h3>
                <div class="summary-value skip">{{.Summary.SkippedChecks}}</div>
            </div>
        </div>

        <div class="progress-bar">
            <div class="progress-segment progress-pass" style="width: {{printf "%.1f" .PassPercent}}%"></div>
            <div class="progress-segment progress-fail" style="width: {{printf "%.1f" .FailPercent}}%"></div>
            <div class="progress-segment progress-skip" style="width: {{printf "%.1f" .SkipPercent}}%"></div>
        </div>

        <div class="progress-legend">
            <div class="legend-item">
                <span class="legend-dot pass"></span>
                <span>Passed ({{printf "%.1f" .PassPercent}}%)</span>
            </div>
            <div class="legend-item">
                <span class="legend-dot fail"></span>
                <span>Failed ({{printf "%.1f" .FailPercent}}%)</span>
            </div>
            <div class="legend-item">
                <span class="legend-dot skip"></span>
                <span>Skipped ({{printf "%.1f" .SkipPercent}}%)</span>
            </div>
        </div>

        {{if .FailedRows}}
        <div class="section" id="failed-section">
            <div class="section-header">
                <h2>Failed Checks</h2>
                <span class="badge badge-fail" id="failed-count">{{len .FailedRows}} issues</span>
            </div>
            <div class="filters">
                <div class="filter-group">
                    <span class="filter-label">Severity:</span>
                    <select class="filter-select" id="severity-filter" onchange="applyFilters()">
                        <option value="">All Severities</option>
                        <option value="critical">Critical</option>
                        <option value="high">High</option>
                        <option value="medium">Medium</option>
                        <option value="low">Low</option>
                    </select>
                </div>
                <div class="filter-group">
                    <span class="filter-label">Resource:</span>
                    <input type="text" class="filter-input" id="resource-filter" placeholder="Search resources..." oninput="applyFilters()">
                </div>
                <span class="filter-count" id="visible-count"></span>
            </div>
            <table id="failed-table">
                <thead>
                    <tr>
                        <th style="width: 24px"></th>
                        <th>Severity</th>
                        <th>Check</th>
                        <th>Resource</th>
                    </tr>
                </thead>
                <tbody>
                    {{range .FailedRows}}
                    <tr class="expandable-row" data-severity="{{lower .CheckSeverity}}" data-resource="{{.ResourceName}}" data-resource-type="{{.ResourceType}}" onclick="toggleDetails(this)">
                        <td><span class="expand-icon">&#9654;</span></td>
                        <td><span class="severity-badge {{severityClass .CheckSeverity}}">{{.CheckSeverity}}</span></td>
                        <td>
                            <div>{{.CheckName}}</div>
                            <div class="resource-path">{{.CheckID}}</div>
                        </td>
                        <td>
                            <div>{{.ResourceName}}</div>
                            <div class="resource-path">{{.ResourceType}}</div>
                        </td>
                    </tr>
                    <tr class="details-row">
                        <td colspan="4">
                            <div class="details-grid">
                                <div class="details-item">
                                    <div class="details-label">Check ID</div>
                                    <div class="details-content">{{.CheckID}}</div>
                                </div>
                                <div class="details-item">
                                    <div class="details-label">Resource Path</div>
                                    <div class="details-content">{{.ResourcePath}}</div>
                                </div>
                                {{if .CheckDetails}}
                                <div class="details-item full-width">
                                    <div class="details-label">Details</div>
                                    <div class="details-content">{{.CheckDetails}}</div>
                                </div>
                                {{end}}
                                {{if .CheckRemediation}}
                                <div class="details-item full-width">
                                    <div class="details-label">Remediation</div>
                                    <div class="details-content markdown-content">{{renderMarkdown .CheckRemediation}}</div>
                                </div>
                                {{end}}
                                {{if .CheckDocs}}
                                <div class="details-item full-width">
                                    <div class="details-label">Documentation</div>
                                    <div class="details-content markdown-content">{{renderMarkdown .CheckDocs}}</div>
                                </div>
                                {{end}}
                                {{if .CheckAudit}}
                                <div class="details-item full-width">
                                    <div class="details-label">Audit</div>
                                    <div class="details-content markdown-content">{{renderMarkdown .CheckAudit}}</div>
                                </div>
                                {{end}}
                            </div>
                        </td>
                    </tr>
                    {{end}}
                </tbody>
            </table>
        </div>
        {{end}}

        {{if .PassedRows}}
        <div class="section">
            <div class="section-header collapsible-header collapsed" onclick="toggleSection(this)">
                <h2>Passed Checks</h2>
                <span class="badge badge-pass">{{len .PassedRows}} passed</span>
            </div>
            <div class="collapsible-content collapsed">
            <table>
                <thead>
                    <tr>
                        <th>Check</th>
                        <th>Resource</th>
                        <th>Status</th>
                    </tr>
                </thead>
                <tbody>
                    {{range .PassedRows}}
                    <tr>
                        <td>
                            <div>{{.CheckName}}</div>
                            <div class="resource-path">{{.CheckID}}</div>
                        </td>
                        <td>
                            <div>{{.ResourceName}}</div>
                            <div class="resource-path">{{.ResourceType}}</div>
                        </td>
                        <td><span class="status-pass">PASS</span></td>
                    </tr>
                    {{end}}
                </tbody>
            </table>
            </div>
        </div>
        {{end}}

        {{if .SkippedRows}}
        <div class="section">
            <div class="section-header collapsible-header collapsed" onclick="toggleSection(this)">
                <h2>Skipped Checks</h2>
                <span class="badge badge-skip">{{len .SkippedRows}} skipped</span>
            </div>
            <div class="collapsible-content collapsed">
            <table>
                <thead>
                    <tr>
                        <th>Check</th>
                        <th>Resource</th>
                        <th>Reason</th>
                    </tr>
                </thead>
                <tbody>
                    {{range .SkippedRows}}
                    <tr>
                        <td>
                            <div>{{.CheckName}}</div>
                            <div class="resource-path">{{.CheckID}}</div>
                        </td>
                        <td>
                            <div>{{.ResourceName}}</div>
                            <div class="resource-path">{{.ResourceType}}</div>
                        </td>
                        <td class="check-details">{{.CheckDetails}}</td>
                    </tr>
                    {{end}}
                </tbody>
            </table>
            </div>
        </div>
        {{end}}

        <div class="scoring-explanation">
            <h3>About the Security Score</h3>
            <p>
                The security score (0-100) measures your compliance posture based on check results.
                A higher score indicates better security compliance. The score is calculated using
                the <strong>{{.Report.Metadata.ScoringSystem}}</strong> scoring method.
            </p>
            {{if eq .Report.Metadata.ScoringSystem "banded"}}
            <div class="scoring-method">
                <h4>Banded Scoring</h4>
                <p style="color: var(--color-text-muted); font-size: 0.8125rem; margin-bottom: 0.75rem;">
                    Banded scoring uses severity "bands" that cap the maximum achievable score. The highest severity
                    finding determines which band applies, and additional findings of that severity reduce the score further.
                    Lower severity findings are ignored once a higher severity is present.
                </p>
                <p style="color: var(--color-text-muted); font-size: 0.8125rem; margin-bottom: 0.5rem;"><strong>Bands:</strong></p>
                <ul>
                    <li><span class="severity-badge severity-critical">Critical</span> → Max score 40, then −8 per additional (e.g., 2 critical = 32)</li>
                    <li><span class="severity-badge severity-high">High</span> → Max score 70, then −5 per additional (e.g., 2 high = 65)</li>
                    <li><span class="severity-badge severity-medium">Medium</span> → Max score 90, then −3 per additional (e.g., 3 medium = 84)</li>
                    <li><span class="severity-badge severity-low">Low</span> → Max score 99, then −1 per additional</li>
                    <li>No findings → Score 100</li>
                </ul>
                <p style="color: var(--color-text-muted); font-size: 0.75rem; margin-top: 0.75rem; font-style: italic;">
                    Example: 1 critical + 5 high + 10 medium = Score 40 (critical band dominates)
                </p>
            </div>
            {{else if eq .Report.Metadata.ScoringSystem "average"}}
            <div class="scoring-method">
                <h4>Weighted Average Scoring</h4>
                <p style="color: var(--color-text-muted); font-size: 0.8125rem; margin-bottom: 0.75rem;">
                    The score represents the ratio of "passed weight" to "total weight". Failed checks are weighted
                    by severity—a critical finding counts 10× more than a passed check. This means many passed checks
                    can offset a few failed ones, but severe findings have outsized impact.
                </p>
                <p style="color: var(--color-text-muted); font-size: 0.8125rem; margin-bottom: 0.5rem;"><strong>Weights:</strong></p>
                <ul>
                    <li><span class="severity-badge severity-critical">Critical</span> findings = weight 10</li>
                    <li><span class="severity-badge severity-high">High</span> findings = weight 7</li>
                    <li><span class="severity-badge severity-medium">Medium</span> findings = weight 4</li>
                    <li><span class="severity-badge severity-low">Low</span> findings = weight 1</li>
                    <li>Passed checks = weight 1 each</li>
                </ul>
                <p style="color: var(--color-text-muted); font-size: 0.75rem; margin-top: 0.75rem; font-style: italic;">
                    Example: 9 passed + 1 critical = 9/(9+10) × 100 = 47
                </p>
            </div>
            {{else if eq .Report.Metadata.ScoringSystem "decayed"}}
            <div class="scoring-method">
                <h4>Exponential Decay Scoring</h4>
                <p style="color: var(--color-text-muted); font-size: 0.8125rem; margin-bottom: 0.75rem;">
                    Starting from 100, each finding multiplies the score by a decay factor. The formula is:
                    <code style="background: var(--color-bg); padding: 0.125rem 0.375rem; border-radius: 3px;">score = 100 × (1−decay)^count</code>.
                    Multiple findings compound, so the score drops faster with more issues. All severities contribute.
                </p>
                <p style="color: var(--color-text-muted); font-size: 0.8125rem; margin-bottom: 0.5rem;"><strong>Decay rates:</strong></p>
                <ul>
                    <li><span class="severity-badge severity-critical">Critical</span> = 40% decay (×0.60 per finding)</li>
                    <li><span class="severity-badge severity-high">High</span> = 20% decay (×0.80 per finding)</li>
                    <li><span class="severity-badge severity-medium">Medium</span> = 10% decay (×0.90 per finding)</li>
                    <li><span class="severity-badge severity-low">Low</span> = 5% decay (×0.95 per finding)</li>
                </ul>
                <p style="color: var(--color-text-muted); font-size: 0.75rem; margin-top: 0.75rem; font-style: italic;">
                    Example: 1 critical + 1 high = 100 × 0.60 × 0.80 = 48
                </p>
            </div>
            {{else if eq .Report.Metadata.ScoringSystem "highest_impact"}}
            <div class="scoring-method">
                <h4>Highest Impact Scoring</h4>
                <p style="color: var(--color-text-muted); font-size: 0.8125rem; margin-bottom: 0.75rem;">
                    A zero-tolerance approach where only the single highest severity finding matters. The number of
                    findings is irrelevant—one critical finding has the same impact as ten. This is ideal for
                    compliance scenarios where any critical issue is unacceptable.
                </p>
                <p style="color: var(--color-text-muted); font-size: 0.8125rem; margin-bottom: 0.5rem;"><strong>Fixed scores:</strong></p>
                <ul>
                    <li>Any <span class="severity-badge severity-critical">Critical</span> finding → Score 0</li>
                    <li>Any <span class="severity-badge severity-high">High</span> finding (no critical) → Score 30</li>
                    <li>Any <span class="severity-badge severity-medium">Medium</span> finding (no high/critical) → Score 60</li>
                    <li>Any <span class="severity-badge severity-low">Low</span> finding (no medium+) → Score 80</li>
                    <li>No findings → Score 100</li>
                </ul>
                <p style="color: var(--color-text-muted); font-size: 0.75rem; margin-top: 0.75rem; font-style: italic;">
                    Example: 1 critical + 50 low = Score 0 (critical dominates everything)
                </p>
            </div>
            {{end}}
        </div>

        <footer>
            <div class="footer-brand">
                <div class="footer-logo"><span>Kopexa</span> kspec</div>
            </div>
            <p class="footer-tagline">Automate your security compliance with policy-as-code</p>
            <a href="https://kopexa.com?utm_source=kspec&utm_medium=report&utm_campaign=html_export" class="footer-cta" target="_blank" rel="noopener">
                Explore Kopexa Platform &rarr;
            </a>
            <div class="footer-links">
                <a href="https://github.com/kopexa-grc/kspec" target="_blank" rel="noopener">GitHub</a>
                <a href="https://docs.kopexa.com" target="_blank" rel="noopener">Documentation</a>
                <a href="https://kopexa.com" target="_blank" rel="noopener">kopexa.com</a>
            </div>
        </footer>
    </div>

    <script>
        function toggleSection(header) {
            header.classList.toggle('collapsed');
            const content = header.nextElementSibling;
            if (content && content.classList.contains('collapsible-content')) {
                content.classList.toggle('collapsed');
                if (!content.classList.contains('collapsed')) {
                    content.style.maxHeight = content.scrollHeight + 'px';
                } else {
                    content.style.maxHeight = null;
                }
            }
        }

        function toggleDetails(row) {
            row.classList.toggle('expanded');
            const detailsRow = row.nextElementSibling;
            if (detailsRow && detailsRow.classList.contains('details-row')) {
                detailsRow.classList.toggle('visible');
            }
        }

        function applyFilters() {
            const severityFilter = document.getElementById('severity-filter');
            const resourceFilter = document.getElementById('resource-filter');
            const table = document.getElementById('failed-table');

            if (!table) return;

            const severity = severityFilter ? severityFilter.value.toLowerCase() : '';
            const resource = resourceFilter ? resourceFilter.value.toLowerCase() : '';

            const rows = table.querySelectorAll('tbody tr.expandable-row');
            let visibleCount = 0;
            const totalCount = rows.length;

            rows.forEach(function(row) {
                const rowSeverity = row.getAttribute('data-severity') || '';
                const rowResource = row.getAttribute('data-resource') || '';
                const rowResourceType = row.getAttribute('data-resource-type') || '';

                const matchesSeverity = !severity || rowSeverity === severity;
                const matchesResource = !resource ||
                    rowResource.toLowerCase().includes(resource) ||
                    rowResourceType.toLowerCase().includes(resource);

                const detailsRow = row.nextElementSibling;

                if (matchesSeverity && matchesResource) {
                    row.classList.remove('hidden-row');
                    if (detailsRow) detailsRow.classList.remove('hidden-row');
                    visibleCount++;
                } else {
                    row.classList.add('hidden-row');
                    if (detailsRow) detailsRow.classList.add('hidden-row');
                }
            });

            // Update visible count
            const countEl = document.getElementById('visible-count');
            if (countEl) {
                if (visibleCount === totalCount) {
                    countEl.textContent = '';
                } else {
                    countEl.textContent = 'Showing ' + visibleCount + ' of ' + totalCount;
                }
            }
        }

        // Initialize expanded sections with proper max-height
        document.addEventListener('DOMContentLoaded', function() {
            document.querySelectorAll('.collapsible-content:not(.collapsed)').forEach(function(content) {
                content.style.maxHeight = content.scrollHeight + 'px';
            });
        });
    </script>
</body>
</html>
`
