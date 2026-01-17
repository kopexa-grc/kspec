// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: BUSL-1.1

package report

import (
	"fmt"
	"html/template"
	"io"
	"os"
	"strings"
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
		"lower":         strings.ToLower,
		"severityClass": severityClass,
		"statusClass":   statusClass,
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
            font-size: 0.75rem;
        }

        .collapsible {
            cursor: pointer;
        }

        .collapsible-content {
            display: none;
        }

        .collapsible-content.show {
            display: block;
        }

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
        <div class="section">
            <div class="section-header">
                <h2>Failed Checks</h2>
                <span class="badge badge-fail">{{len .FailedRows}} issues</span>
            </div>
            {{if .SeverityCounts}}
            <div class="severity-badges">
                {{range $severity, $count := .SeverityCounts}}
                <span class="severity-badge {{severityClass $severity}}">{{$severity}}: {{$count}}</span>
                {{end}}
            </div>
            {{end}}
            <table>
                <thead>
                    <tr>
                        <th>Severity</th>
                        <th>Check</th>
                        <th>Resource</th>
                        <th>Details</th>
                    </tr>
                </thead>
                <tbody>
                    {{range .FailedRows}}
                    <tr>
                        <td><span class="severity-badge {{severityClass .CheckSeverity}}">{{.CheckSeverity}}</span></td>
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
        {{end}}

        {{if .PassedRows}}
        <div class="section">
            <div class="section-header">
                <h2>Passed Checks</h2>
                <span class="badge badge-pass">{{len .PassedRows}} passed</span>
            </div>
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
        {{end}}

        {{if .SkippedRows}}
        <div class="section">
            <div class="section-header">
                <h2>Skipped Checks</h2>
                <span class="badge badge-skip">{{len .SkippedRows}} skipped</span>
            </div>
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
        {{end}}

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
                <a href="https://kopexa.com/docs" target="_blank" rel="noopener">Documentation</a>
                <a href="https://kopexa.com" target="_blank" rel="noopener">kopexa.com</a>
            </div>
        </footer>
    </div>
</body>
</html>
`
