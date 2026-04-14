package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/cloudmechanic-cli/cloudmechanic/internal/scanner"
)

// View renders the entire UI.
func (m Model) View() string {
	if m.width == 0 {
		return ""
	}

	// Remediation view takes over the whole screen.
	if m.state == stateRemediation {
		return m.viewRemediation()
	}

	var sections []string
	sections = append(sections, m.viewHeader())
	sections = append(sections, m.viewTabBar())

	if m.state == stateLoading {
		sections = append(sections, m.viewLoading())
	} else {
		sections = append(sections, m.viewBody())
	}

	sections = append(sections, m.viewFooter())

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

// viewHeader renders a minimal single-line title bar.
func (m Model) viewHeader() string {
	title := headerTitleStyle.Render("☁  CloudMechanic")
	dot := headerDotStyle.Render("  ·  ")
	sub := headerSubStyle.Render("AWS Security Scanner")

	line := title + dot + sub
	return headerBarStyle.Width(m.width).Render(line)
}

// viewTabBar renders two button-like tabs above the panes showing which is active.
func (m Model) viewTabBar() string {
	var regTab, issTab string
	if m.focus == paneSidebar {
		regTab = tabActiveStyle.Render("  ⬡  Regions  ")
		issTab = tabInactiveStyle.Render("  ≡  Issues   ")
	} else {
		regTab = tabInactiveStyle.Render("  ⬡  Regions  ")
		issTab = tabActiveStyle.Render("  ≡  Issues   ")
	}
	return "\n " + regTab + "  " + issTab + "\n"
}

// viewLoading renders the scanning progress screen.
func (m Model) viewLoading() string {
	heading := lipgloss.NewStyle().Bold(true).Foreground(catBlue).
		Render("Scanning AWS resources...")
	s := fmt.Sprintf("\n   %s  %s\n\n", m.spinner.View(), heading)
	for _, r := range m.builder.Regions {
		s += fmt.Sprintf("     %s  %s\n",
			lipgloss.NewStyle().Foreground(catTeal).Render("◌"),
			lipgloss.NewStyle().Foreground(catSubtext1).Render(r))
	}
	return s
}

// viewBody renders the two-pane explorer layout.
func (m Model) viewBody() string {
	sidebarWidth := 30
	// Account for border (2 chars each side) and gap.
	mainWidth := m.width - sidebarWidth - 6
	if mainWidth < 40 {
		mainWidth = 40
	}

	// Overhead: header(1) + tabbar(3) + footer(1) + padding(3) = 8
	paneHeight := m.height - 8
	if paneHeight < 8 {
		paneHeight = 8
	}

	sidebar := m.viewSidebar(sidebarWidth, paneHeight)
	main := m.viewMain(mainWidth, paneHeight)

	// Active pane: bright border + full colour content.
	// Inactive pane: dim border + faint content for instant focus clarity.
	var sidebarBox, mainBox string
	if m.focus == paneSidebar {
		sidebarBox = focusedBorder.Width(sidebarWidth).Height(paneHeight).Render(sidebar)
		mainBox = unfocusedBorder.Width(mainWidth).Height(paneHeight).Render(
			faintStyle.Render(main))
	} else {
		sidebarBox = unfocusedBorder.Width(sidebarWidth).Height(paneHeight).Render(
			faintStyle.Render(sidebar))
		mainBox = focusedBorder.Width(mainWidth).Height(paneHeight).Render(main)
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, sidebarBox, mainBox)
}

// --- Sidebar ---

func (m Model) viewSidebar(width, height int) string {
	var b strings.Builder

	b.WriteString(sidebarTitleStyle.Render(" REGIONS"))
	b.WriteString("\n\n")

	// "Global View" entry at index 0.
	globalCrit := countSeverity(m.issues, scanner.SeverityCritical)
	globalWarn := countSeverity(m.issues, scanner.SeverityWarning)
	globalLabel := fmt.Sprintf("  All Regions  %s",
		regionBadges(globalCrit, globalWarn))

	if m.regionCursor == 0 {
		b.WriteString(regionSelectedStyle.Width(width - 2).Render(globalLabel))
	} else {
		b.WriteString(regionNormalStyle.Render(globalLabel))
	}
	b.WriteString("\n")

	// Individual regions.
	for i, region := range m.builder.Regions {
		crit := countSeverityInRegion(m.issues, region, scanner.SeverityCritical)
		warn := countSeverityInRegion(m.issues, region, scanner.SeverityWarning)
		label := fmt.Sprintf("  %s  %s", region, regionBadges(crit, warn))

		idx := i + 1
		if m.regionCursor == idx {
			b.WriteString(regionSelectedStyle.Width(width - 2).Render(label))
		} else if crit+warn > 0 {
			b.WriteString(regionNormalStyle.Render(label))
		} else {
			b.WriteString(regionDimStyle.Render(label))
		}
		b.WriteString("\n")
	}

	// Global services (IAM) shown under a separator when in global view.
	if m.regionCursor == 0 {
		globalIssues := countRegion(m.issues, "global")
		if globalIssues > 0 {
			crit := countSeverityInRegion(m.issues, "global", scanner.SeverityCritical)
			warn := countSeverityInRegion(m.issues, "global", scanner.SeverityWarning)
			b.WriteString(regionDimStyle.Render(
				fmt.Sprintf("  global (IAM)  %s", regionBadges(crit, warn))))
			b.WriteString("\n")
		}
	}

	// Summary section.
	b.WriteString("\n")
	b.WriteString(sidebarTitleStyle.Render(" SUMMARY"))
	b.WriteString("\n\n")

	fc := len(m.filtered)
	total := len(m.issues)

	filteredCrit := countSeverity(m.filtered, scanner.SeverityCritical)
	filteredWarn := countSeverity(m.filtered, scanner.SeverityWarning)

	b.WriteString(fmt.Sprintf("  %s %s\n",
		countCriticalStyle.Render(fmt.Sprintf("%d", filteredCrit)),
		dimStyle.Render("Critical")))
	b.WriteString(fmt.Sprintf("  %s %s\n",
		countWarningStyle.Render(fmt.Sprintf("%d", filteredWarn)),
		dimStyle.Render("Warnings")))

	if fc != total {
		b.WriteString(fmt.Sprintf("  %s %s\n",
			countGreenStyle.Render(fmt.Sprintf("%d/%d", fc, total)),
			dimStyle.Render("Showing")))
	} else {
		b.WriteString(fmt.Sprintf("  %s %s\n",
			countGreenStyle.Render(fmt.Sprintf("%d", total)),
			dimStyle.Render("Total")))
	}

	if len(m.errors) > 0 {
		b.WriteString(fmt.Sprintf("  %s %s\n",
			errorStyle.Render(fmt.Sprintf("%d", len(m.errors))),
			dimStyle.Render("Errors")))
	}

	return b.String()
}

// --- Main pane ---

func (m Model) viewMain(width, height int) string {
	var b strings.Builder

	// Search bar.
	if m.searchMode {
		b.WriteString(searchStyle.Render(" / ") +
			lipgloss.NewStyle().Foreground(colorWhite).Render(m.searchText) +
			searchStyle.Render("_"))
		b.WriteString("\n\n")
	} else if m.searchText != "" {
		b.WriteString(dimStyle.Render(fmt.Sprintf(" Search: \"%s\"  [Esc] clear", m.searchText)))
		b.WriteString("\n\n")
	}

	// Filter indicator.
	if m.sevFilter != filterAll {
		b.WriteString(filterActiveStyle.Render(fmt.Sprintf(" Filter: %s", m.sevFilter)))
		b.WriteString("\n\n")
	}

	if len(m.filtered) == 0 {
		if len(m.issues) == 0 && len(m.errors) == 0 {
			b.WriteString(countGreenStyle.Render("\n  No issues found. Your account looks clean!\n"))
		} else {
			b.WriteString(dimStyle.Render("\n  No issues match current filters.\n"))
		}
		return b.String()
	}

	// Table header.
	colSev := 12
	colSvc := 10
	colDesc := width - colSev - colSvc - 8
	if colDesc < 20 {
		colDesc = 20
	}

	hdr := fmt.Sprintf(" %-*s %-*s %s",
		colSev, "SEVERITY",
		colSvc, "SERVICE",
		"DESCRIPTION")
	b.WriteString(tableHeaderStyle.Width(width - 2).Render(hdr))
	b.WriteString("\n")

	// Grouped issue list with scrolling.
	maxVisible := height - 6
	if m.searchMode || m.searchText != "" {
		maxVisible -= 2
	}
	if m.sevFilter != filterAll {
		maxVisible -= 2
	}
	if maxVisible < 4 {
		maxVisible = 4
	}

	// Build flat display lines: interleave group headers + issues.
	type displayItem struct {
		isHeader bool
		header   string
		issue    scanner.Issue
		issueIdx int // index into m.filtered
	}

	var items []displayItem
	flatIdx := 0
	for _, g := range m.groups {
		items = append(items, displayItem{isHeader: true, header: g.service})
		for _, issue := range g.issues {
			items = append(items, displayItem{issue: issue, issueIdx: flatIdx})
			flatIdx++
		}
	}

	// Find which display item corresponds to the cursor.
	cursorDisplayIdx := 0
	for i, item := range items {
		if !item.isHeader && item.issueIdx == m.issueCursor {
			cursorDisplayIdx = i
			break
		}
	}

	// Calculate scroll window.
	start := 0
	if cursorDisplayIdx >= maxVisible {
		start = cursorDisplayIdx - maxVisible + 1
	}

	linesWritten := 0
	for i := start; i < len(items) && linesWritten < maxVisible; i++ {
		item := items[i]

		if item.isHeader {
			icon := serviceIcon(item.header)
			b.WriteString(groupHeaderStyle.Render(fmt.Sprintf("  %s%s", icon, item.header)))
			b.WriteString("\n")
			linesWritten++
			continue
		}

		issue := item.issue
		svc := serviceFromScanner(issue.Scanner)
		badge := severityBadge(issue.Severity)
		desc := truncate(issue.Description, colDesc)

		line := fmt.Sprintf(" %s %-*s %s", badge, colSvc, svc, desc)

		if item.issueIdx == m.issueCursor && m.focus == paneMain {
			b.WriteString(issueSelectedStyle.Width(width - 2).Render(line))
			b.WriteString("\n")
			linesWritten++
		} else {
			b.WriteString(issueNormalStyle.Width(width - 2).Render(line))
			b.WriteString("\n")
			linesWritten++
		}
	}

	// Scroll indicator.
	if len(m.filtered) > maxVisible {
		b.WriteString(dimStyle.Render(
			fmt.Sprintf("\n  %d of %d issues", minInt(m.issueCursor+1, len(m.filtered)), len(m.filtered))))
	}

	return b.String()
}

// --- Footer ---

func (m Model) viewFooter() string {
	// ── Left: context info ───────────────────────────────────────────────────
	var left string
	if m.state == stateReady {
		region := m.selectedRegion()
		if region == "" {
			region = "All Regions"
		}
		left = statusRegionStyle.Render(" ☁  "+region+" ") +
			dimStyle.Render("  "+m.elapsed.Round(time.Millisecond).String())
		if m.sevFilter != filterAll {
			left += "  " + filterActiveStyle.Render("⚡ "+m.sevFilter.String())
		}
		if m.searchText != "" {
			left += "  " + searchStyle.Render("/ "+m.searchText)
		}
	}

	// ── Right: pill-shaped key hints ─────────────────────────────────────────
	pills := renderPill("Tab", "Switch") +
		renderPill("↑↓", "Nav") +
		renderPill("↵", "Fix") +
		renderPill("F", "Filter") +
		renderPill("/", "Search") +
		renderPill("R", "Rescan") +
		renderPill("Q", "Quit")

	gap := m.width - lipgloss.Width(left) - lipgloss.Width(pills)
	if gap < 0 {
		// Condensed version for narrow terminals.
		pills = renderPill("↵", "Fix") + renderPill("⇥", "Switch") + renderPill("Q", "Quit")
		gap = m.width - lipgloss.Width(left) - lipgloss.Width(pills)
		if gap < 0 {
			gap = 0
		}
	}

	bar := left + strings.Repeat(" ", gap) + pills
	return statusBarStyle.Width(m.width).Render(bar)
}

// renderPill returns a styled key + label pair for the status bar.
func renderPill(key, action string) string {
	k := statusKeyStyle.Render(key)
	a := statusActionStyle.Render(action)
	return k + a + " "
}

// serviceIcon returns an emoji prefix for a service category.
func serviceIcon(svc string) string {
	switch svc {
	case "EC2":
		return "🖥  "
	case "S3":
		return "🪣  "
	case "RDS":
		return "🗄  "
	case "IAM":
		return "🔑  "
	case "DynamoDB":
		return "⚡  "
	case "VPC":
		return "☁  "
	case "Lambda":
		return "λ  "
	default:
		return "◈  "
	}
}

// --- Helpers ---

func severityBadge(s scanner.Severity) string {
	switch s {
	case scanner.SeverityCritical:
		return criticalBadge.Render("CRITICAL")
	case scanner.SeverityWarning:
		return warningBadge.Render("WARNING ")
	default:
		return "UNKNOWN"
	}
}

func regionBadges(crit, warn int) string {
	parts := ""
	if crit > 0 {
		parts += countCriticalStyle.Render(fmt.Sprintf("%d\xf0\x9f\x94\xb4", crit))
	}
	if warn > 0 {
		if parts != "" {
			parts += " "
		}
		parts += countWarningStyle.Render(fmt.Sprintf("%d\xf0\x9f\x9f\xa1", warn))
	}
	return parts
}

func countSeverity(issues []scanner.Issue, sev scanner.Severity) int {
	n := 0
	for _, i := range issues {
		if i.Severity == sev {
			n++
		}
	}
	return n
}

func countSeverityInRegion(issues []scanner.Issue, region string, sev scanner.Severity) int {
	n := 0
	for _, i := range issues {
		if i.Region == region && i.Severity == sev {
			n++
		}
	}
	return n
}

func countRegion(issues []scanner.Issue, region string) int {
	n := 0
	for _, i := range issues {
		if i.Region == region {
			n++
		}
	}
	return n
}

func truncate(s string, maxLen int) string {
	if maxLen < 4 {
		maxLen = 4
	}
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// ── Remediation view ─────────────────────────────────────────────────────────

// viewRemediation renders the full-screen Terraform remediation editor.
func (m Model) viewRemediation() string {
	if len(m.filtered) == 0 || m.issueCursor >= len(m.filtered) {
		return ""
	}
	issue := m.filtered[m.issueCursor]
	rem := GetRemediation(issue)

	var sections []string

	// ── Header bar ──────────────────────────────────────────────────────────
	title := remHeaderTitleStyle.Render("  TERRAFORM REMEDIATION  ")
	sections = append(sections, title)
	sections = append(sections, remSepStyle.Render(strings.Repeat("─", m.width-1)))
	sections = append(sections, "")

	// ── Issue context ────────────────────────────────────────────────────────
	badge := severityBadge(issue.Severity)
	svc := serviceFromScanner(issue.Scanner)
	sections = append(sections, fmt.Sprintf("  %s  %s  %s",
		badge,
		lipgloss.NewStyle().Bold(true).Foreground(colorCyan).Render(svc),
		remIssueValueStyle.Render(issue.Scanner),
	))
	sections = append(sections, fmt.Sprintf("  %s %s",
		remIssueLabelStyle.Render("Resource:"),
		remIssueValueStyle.Render(issue.ResourceID),
	))
	if issue.Region != "" {
		regionLabel := issue.Region
		if regionLabel == "global" {
			regionLabel = "global (IAM)"
		}
		sections = append(sections, fmt.Sprintf("  %s %s",
			remIssueLabelStyle.Render("Region:  "),
			remIssueValueStyle.Render(regionLabel),
		))
	}
	sections = append(sections, "")

	// ── Fix title and description ────────────────────────────────────────────
	sections = append(sections, remFixTitleStyle.Render(rem.Title))
	sections = append(sections, remFixDescStyle.Width(m.width-4).Render(rem.Description))
	sections = append(sections, "")

	// ── Code block ──────────────────────────────────────────────────────────
	// Calculate available height for the code block.
	// header(1) + sep(1) + blank(1) + issue lines(2-3) + blank(1) +
	// fix title(1) + fix desc(~2) + blank(1) + footer(2) = ~13 lines overhead
	const overheadLines = 14
	codeViewHeight := m.height - overheadLines
	if codeViewHeight < 5 {
		codeViewHeight = 5
	}

	codeLines := strings.Split(rem.TerraformCode, "\n")
	totalLines := len(codeLines)

	// Clamp scroll so we can never scroll past the last line.
	maxScroll := totalLines - codeViewHeight
	if maxScroll < 0 {
		maxScroll = 0
	}
	scroll := m.remScroll
	if scroll > maxScroll {
		scroll = maxScroll
	}

	visibleLines := codeLines[scroll:]
	if len(visibleLines) > codeViewHeight {
		visibleLines = visibleLines[:codeViewHeight]
	}

	// Render each visible line with HCL syntax highlighting.
	var codeContent strings.Builder
	for _, line := range visibleLines {
		codeContent.WriteString(syntaxHighlightHCL(line) + "\n")
	}

	// Code block width: terminal width minus outer padding and border.
	codeWidth := m.width - 6
	if codeWidth < 40 {
		codeWidth = 40
	}

	// File name header row inside the code block.
	scrollInfo := ""
	if maxScroll > 0 {
		scrollInfo = remScrollHintStyle.Render(
			fmt.Sprintf("  [%d–%d / %d lines]  j/k to scroll",
				scroll+1, minInt(scroll+codeViewHeight, totalLines), totalLines))
	}
	fileHeader := remCodeFileStyle.Render(" main.tf ") + scrollInfo
	codeSep := remCodeSepStyle.Render(strings.Repeat("─", codeWidth))

	innerContent := fileHeader + "\n" + codeSep + "\n" + codeContent.String()

	codeBlock := remCodeBorderStyle.
		Width(codeWidth).
		Render(innerContent)

	sections = append(sections, "  "+codeBlock)

	// ── Footer ───────────────────────────────────────────────────────────────
	footerHints := remFooterStyle.Render(
		"  [j/k] Scroll   [Esc / q] Back to Issues   [Ctrl+C] Quit",
	)
	sections = append(sections, "")
	sections = append(sections, remSepStyle.Render(strings.Repeat("─", m.width-1)))
	sections = append(sections, footerHints)

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

// syntaxHighlightHCL applies basic HCL/Terraform syntax colouring to one line.
func syntaxHighlightHCL(line string) string {
	trimmed := strings.TrimSpace(line)
	if trimmed == "" {
		return ""
	}

	// Compute leading whitespace.
	indentLen := len(line) - len(strings.TrimLeft(line, " \t"))
	indent := line[:indentLen]

	// Comments
	if strings.HasPrefix(trimmed, "#") {
		return remCommentStyle.Render(line)
	}

	// Lone closing braces / brackets
	switch trimmed {
	case "}", "})", "}]", "],", "},", "]":
		return remHCLBraceStyle.Render(line)
	}

	// Top-level HCL block keywords: resource, module, data, output, …
	hclKeywords := []string{
		"resource", "module", "data", "output",
		"variable", "locals", "provider", "terraform",
	}
	for _, kw := range hclKeywords {
		if trimmed == kw || strings.HasPrefix(trimmed, kw+" ") {
			rest := trimmed[len(kw):]
			return indent + remHCLKeywordStyle.Render(kw) + remHCLBlockArgStyle.Render(rest)
		}
	}

	// Assignment: attr_name = <value>
	if eqIdx := strings.Index(trimmed, " = "); eqIdx != -1 {
		attr := trimmed[:eqIdx]
		val := trimmed[eqIdx+3:]
		return indent +
			remHCLAttrStyle.Render(attr) +
			remHCLEqualsStyle.Render(" = ") +
			styleHCLValue(val)
	}

	// Nested block openers: schedule {, create_rule {, etc.
	if strings.HasSuffix(trimmed, " {") || trimmed == "{" {
		return indent + remHCLInnerBlockStyle.Render(trimmed)
	}

	return indent + remDefaultLineStyle.Render(trimmed)
}

// styleHCLValue colours the right-hand side of an HCL assignment.
func styleHCLValue(v string) string {
	// String literals
	if strings.HasPrefix(v, `"`) {
		return remHCLStringStyle.Render(v)
	}
	// Heredoc / function calls (jsonencode, file, etc.)
	if strings.HasPrefix(v, "jsonencode(") || strings.HasPrefix(v, "<<") ||
		strings.HasPrefix(v, "file(") || strings.HasPrefix(v, "filebase64") {
		return remHCLRefStyle.Render(v)
	}
	// Booleans and null
	switch v {
	case "true", "false", "null":
		return remHCLBoolStyle.Render(v)
	}
	// Numbers
	isNum := len(v) > 0
	for _, c := range v {
		if (c < '0' || c > '9') && c != '.' && c != '-' {
			isNum = false
			break
		}
	}
	if isNum {
		return remHCLBoolStyle.Render(v)
	}
	// Resource references (e.g., aws_iam_role.flow_logs.arn)
	if strings.Contains(v, ".") && !strings.Contains(v, " ") {
		return remHCLRefStyle.Render(v)
	}
	return remDefaultLineStyle.Render(v)
}
