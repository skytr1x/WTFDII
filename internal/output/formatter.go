package output

import (
	"fmt"
	"strings"

	"github.com/skytr1x/wtfdii/internal/models"
)

const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorGray   = "\033[90m"
	colorBold   = "\033[1m"
	symbolError   = "🔴"
	symbolWarning = "🟡"
	symbolInfo    = "🟢"
)

type Formatter struct {
	colorEnabled bool
}

func New(colorEnabled bool) *Formatter {
	return &Formatter{
		colorEnabled: colorEnabled,
	}
}

func (f *Formatter) color(color, text string) string {
	if !f.colorEnabled {
		return text
	}
	return color + text + colorReset
}

func (f *Formatter) FormatOverview(tree *models.DependencyTree) string {
	var sb strings.Builder

	sb.WriteString(f.color(colorBold, "WTF Did I Install?\n\n"))
	sb.WriteString(fmt.Sprintf("Project: %s\n", f.color(colorBlue, tree.ProjectName)))
	sb.WriteString(fmt.Sprintf("Package manager: %s\n\n", f.color(colorBlue, tree.PackageManager)))
	sb.WriteString(f.color(colorBold, "Dependencies\n"))
	sb.WriteString(strings.Repeat("─", 40) + "\n")
	sb.WriteString(fmt.Sprintf("%-25s %d\n", "Direct dependencies", tree.DirectCount))
	sb.WriteString(fmt.Sprintf("%-25s %d\n", "Transitive dependencies", tree.TransitiveCount))
	sb.WriteString(fmt.Sprintf("%-25s %d\n\n", "Total packages", tree.TotalCount))
	sb.WriteString(f.color(colorBold, "Disk usage\n"))
	sb.WriteString(strings.Repeat("─", 40) + "\n")
	sb.WriteString(fmt.Sprintf("%-25s %s\n\n", "node_modules", f.formatSize(tree.TotalSize)))
	sb.WriteString(f.color(colorBold, "⚠ Things you might want to know\n\n"))

	if tree.SecurityIssuesCount.Total > 0 {
		sb.WriteString(fmt.Sprintf("  %s %d packages have security advisories\n",
			symbolError, tree.SecurityIssuesCount.Total))
	}
	sb.WriteString(f.color(colorGray, "  Run 'wtfdii security' for security audit\n"))
	sb.WriteString(f.color(colorGray, "  Run 'wtfdii stale' for outdated packages\n"))
	sb.WriteString(f.color(colorGray, "  Run 'wtfdii unused' for unused dependencies\n"))

	return sb.String()
}

func (f *Formatter) FormatDependencyChain(packageName string, chain []string) string {
	var sb strings.Builder

	sb.WriteString(f.color(colorBold, packageName) + "\n\n")
	sb.WriteString("Installed because:\n\n")

	if len(chain) == 0 {
		sb.WriteString(f.color(colorGray, "  (direct dependency)\n"))
		return sb.String()
	}

	indent := ""
	for i, pkg := range chain {
		if i == len(chain)-1 {
			sb.WriteString(fmt.Sprintf("%s└── %s\n", indent, f.color(colorBlue, pkg)))
		} else {
			sb.WriteString(fmt.Sprintf("%s└── %s\n", indent, pkg))
			indent += "    "
		}
	}

	return sb.String()
}

func (f *Formatter) FormatPackagesBySize(packages []*models.Package, limit int) string {
	var sb strings.Builder

	sb.WriteString(f.color(colorBold, "Largest packages\n"))
	sb.WriteString(strings.Repeat("─", 60) + "\n")
	sb.WriteString(fmt.Sprintf("%-40s %15s\n", "Package", "Size"))
	sb.WriteString(strings.Repeat("─", 60) + "\n")

	count := limit
	if count == 0 || count > len(packages) {
		count = len(packages)
	}

	for i := 0; i < count && i < len(packages); i++ {
		pkg := packages[i]
		sb.WriteString(fmt.Sprintf("%-40s %15s\n",
			f.color(colorBlue, pkg.Name),
			f.formatSize(pkg.Size)))
	}

	return sb.String()
}

func (f *Formatter) FormatSecurityIssues(issues []models.SecurityIssue) string {
	var sb strings.Builder

	sb.WriteString(f.color(colorBold, "Security\n\n"))

	if len(issues) == 0 {
		sb.WriteString(f.color(colorGreen, symbolInfo+" No known vulnerabilities\n"))
		return sb.String()
	}

	severityCounts := make(map[string]int)
	for _, issue := range issues {
		severityCounts[issue.Severity]++
	}

	if count := severityCounts["high"] + severityCounts["critical"]; count > 0 {
		sb.WriteString(fmt.Sprintf("%s %d high severity vulnerabilities\n", symbolError, count))
	}
	if count := severityCounts["moderate"]; count > 0 {
		sb.WriteString(fmt.Sprintf("%s %d moderate vulnerabilities\n", symbolWarning, count))
	}
	if count := severityCounts["low"]; count > 0 {
		sb.WriteString(fmt.Sprintf("%s %d low severity vulnerabilities\n", symbolInfo, count))
	}

	sb.WriteString("\nAffected packages:\n\n")

	for _, issue := range issues {
		color := colorGray
		switch issue.Severity {
		case "critical", "high":
			color = colorRed
		case "moderate":
			color = colorYellow
		case "low":
			color = colorGreen
		}

		sb.WriteString(fmt.Sprintf("%s %s\n",
			f.color(color, issue.Severity),
			f.color(colorBold, issue.Title)))

		if issue.Description != "" {
			sb.WriteString(fmt.Sprintf("  %s\n", issue.Description))
		}
		if issue.URL != "" {
			sb.WriteString(fmt.Sprintf("  %s\n", f.color(colorBlue, issue.URL)))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

func (f *Formatter) FormatStalePackages(packages []*models.StalePackage) string {
	var sb strings.Builder

	sb.WriteString(f.color(colorBold, "Potentially stale dependencies\n\n"))

	if len(packages) == 0 {
		sb.WriteString(f.color(colorGreen, symbolInfo+" All packages are up to date\n"))
		return sb.String()
	}

	sb.WriteString(fmt.Sprintf("%-40s %-15s %-15s\n", "package", "current", "latest"))
	sb.WriteString(strings.Repeat("─", 75) + "\n")

	for _, pkg := range packages {
		pkgType := ""
		if pkg.Type == "devDependencies" {
			pkgType = f.color(colorGray, " (dev)")
		}

		sb.WriteString(fmt.Sprintf("%-40s %-15s %-15s%s\n",
			f.color(colorBlue, pkg.Name),
			pkg.Current,
			f.color(colorYellow, pkg.Latest),
			pkgType))
	}

	sb.WriteString(fmt.Sprintf("\n%s %d outdated packages found\n",
		symbolWarning, len(packages)))

	return sb.String()
}

func (f *Formatter) FormatError(err error) string {
	return f.color(colorRed, "Error: "+err.Error())
}

func (f *Formatter) formatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}

	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
