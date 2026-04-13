package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/cloudmechanic-cli/cloudmechanic/internal/scanner"
)

const logo = `
  ______  __                    __  __  ___              __                   _
 / ____/ / /____   __  __ ____/ / /  |/  /___   _____  / /_   ____ _ ____   (_)_____
/ /     / // __ \ / / / // __  / / /|_/ // _ \ / ___/ / __ \ / __ '// __ \ / // ___/
/ /___ / // /_/ // /_/ // /_/ / / /  / //  __// /__  / / / // /_/ // / / // // /__
\____//_/ \____/ \__,_/ \__,_/ /_/  /_/ \___/ \___/ /_/ /_/ \__,_//_/ /_//_/ \___/`

// View renders the entire UI.
func (m Model) View() string {
	if m.width == 0 {
		return ""
	}

	var sections []string

	sections = append(sections, m.viewHeader())

	if m.state == stateLoading {
		sections = append(sections, m.viewLoading())
	} else {
		sections = append(sections, m.viewBody())
	}

	sections = append(sections, m.viewFooter())

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

func (m Model) viewHeader() string {
	art := headerStyle.Render(logo)
	subtitle := dimStyle.Padding(0, 2).Render("An OBD scanner for your AWS environment")
	return lipgloss.JoinVertical(lipgloss.Left, art, subtitle)
}

func (m Model) viewLoading() string {
	s := fmt.Sprintf("\n   %s Scanning AWS resources...\n\n", m.spinner.View())
	for _, r := range m.builder.Regions {
		s += fmt.Sprintf("     %s %s\n",
			lipgloss.NewStyle().Foreground(colorGreen).Bold(true).Render("*"), r)
	}
	return s
}

// viewBody renders the two-pane explorer layout.
func (m Model) viewBody() string {
	sidebarWidth := 30
	// Account for border (2 chars each side) and gap
	mainWidth := m.width - sidebarWidth - 6
	if mainWidth < 40 {
		mainWidth = 40
	}

	// Height for panes: total - header(~8) - footer(1) - padding
	paneHeight := m.height - 11
	if paneHeight < 8 {
		paneHeight = 8
	}

	sidebar := m.viewSidebar(sidebarWidth, paneHeight)
	main := m.viewMain(mainWidth, paneHeight)

	// Apply focused/unfocused borders.
	var sidebarBox, mainBox string
	if m.focus == paneSidebar {
		sidebarBox = focusedBorder.Width(sidebarWidth).Height(paneHeight).Render(sidebar)
		mainBox = unfocusedBorder.Width(mainWidth).Height(paneHeight).Render(main)
	} else {
		sidebarBox = unfocusedBorder.Width(sidebarWidth).Height(paneHeight).Render(sidebar)
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
			b.WriteString(groupHeaderStyle.Render(fmt.Sprintf("  %s", item.header)))
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

			if m.expanded {
				detail := m.viewDetail(issue, width-4)
				b.WriteString(detail)
				b.WriteString("\n")
				linesWritten += 4
			}
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

func (m Model) viewDetail(issue scanner.Issue, width int) string {
	var b strings.Builder
	indent := "      "

	b.WriteString(indent + detailLabelStyle.Render("Scanner:  ") + detailValueStyle.Render(issue.Scanner) + "\n")
	b.WriteString(indent + detailLabelStyle.Render("Resource: ") + detailValueStyle.Render(issue.ResourceID) + "\n")
	if issue.Region != "" {
		b.WriteString(indent + detailLabelStyle.Render("Region:   ") + detailValueStyle.Render(issue.Region) + "\n")
	}
	b.WriteString(indent + detailLabelStyle.Render("Fix:      ") + suggestionStyle.Render(issue.Suggestion) + "\n")

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorCyan).
		Padding(0, 1).
		Width(width - 8).
		Render(b.String())

	return "    " + box
}

// --- Footer ---

func (m Model) viewFooter() string {
	left := ""
	if m.state == stateReady {
		region := m.selectedRegion()
		if region == "" {
			region = "All Regions"
		}
		left = fmt.Sprintf(" %s | %s | Filter: %s",
			region,
			m.elapsed.Round(time.Millisecond),
			m.sevFilter)
	}

	right := " [Tab] Switch Pane  [j/k] Navigate  [Enter] Expand  [F] Filter  [/] Search  [R] Re-scan  [Q] Quit "

	gap := m.width - lipgloss.Width(left) - lipgloss.Width(right)
	if gap < 0 {
		gap = 0
		// Truncate right hints if terminal is narrow.
		right = " [Tab] Pane [F] Filter [/] Search [R] Scan [Q] Quit "
		gap = m.width - lipgloss.Width(left) - lipgloss.Width(right)
		if gap < 0 {
			gap = 0
		}
	}

	bar := left + strings.Repeat(" ", gap) + right
	return statusBarStyle.Width(m.width).Render(bar)
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
