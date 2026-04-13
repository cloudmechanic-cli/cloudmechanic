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
		return "" // not yet sized
	}

	var sections []string

	// Header
	sections = append(sections, m.viewHeader())

	if m.state == stateLoading {
		sections = append(sections, m.viewLoading())
	} else {
		sections = append(sections, m.viewBody())
	}

	// Status bar
	sections = append(sections, m.viewStatusBar())

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

func (m Model) viewHeader() string {
	art := headerStyle.Render(logo)
	subtitle := lipgloss.NewStyle().
		Foreground(colorDim).
		Padding(0, 2).
		Render("An OBD scanner for your AWS environment")
	return lipgloss.JoinVertical(lipgloss.Left, art, subtitle)
}

func (m Model) viewLoading() string {
	s := fmt.Sprintf("\n   %s Scanning AWS resources...\n\n", m.spinner.View())
	regions := ""
	for _, r := range m.builder.Regions {
		regions += fmt.Sprintf("     %s %s\n", regionActiveStyle.Render("*"), r)
	}
	if regions != "" {
		s += lipgloss.NewStyle().Foreground(colorDim).Render("   Regions:\n") + regions
	}
	return s
}

func (m Model) viewBody() string {
	sidebarWidth := 24
	mainWidth := m.width - sidebarWidth - 4
	if mainWidth < 30 {
		mainWidth = 30
	}

	sidebar := m.viewSidebar(sidebarWidth)
	main := m.viewMain(mainWidth)

	return lipgloss.JoinHorizontal(lipgloss.Top, sidebar, main)
}

func (m Model) viewSidebar(width int) string {
	var b strings.Builder
	b.WriteString(sidebarTitleStyle.Render("  REGIONS"))
	b.WriteString("\n")

	for _, r := range m.builder.Regions {
		line := fmt.Sprintf("  %s %s", regionActiveStyle.Render("*"), r)
		b.WriteString(line + "\n")
	}

	b.WriteString("\n")
	b.WriteString(sidebarTitleStyle.Render("  SUMMARY"))
	b.WriteString("\n")

	critCount := countSeverity(m.issues, scanner.SeverityCritical)
	warnCount := countSeverity(m.issues, scanner.SeverityWarning)

	b.WriteString(fmt.Sprintf("  %s %s\n",
		countCriticalStyle.Render(fmt.Sprintf("%d", critCount)),
		lipgloss.NewStyle().Foreground(colorDim).Render("Critical")))
	b.WriteString(fmt.Sprintf("  %s %s\n",
		countWarningStyle.Render(fmt.Sprintf("%d", warnCount)),
		lipgloss.NewStyle().Foreground(colorDim).Render("Warnings")))
	b.WriteString(fmt.Sprintf("  %s %s\n",
		countGreenStyle.Render(fmt.Sprintf("%d", len(m.issues))),
		lipgloss.NewStyle().Foreground(colorDim).Render("Total")))

	if len(m.errors) > 0 {
		b.WriteString(fmt.Sprintf("  %s %s\n",
			errorStyle.Render(fmt.Sprintf("%d", len(m.errors))),
			lipgloss.NewStyle().Foreground(colorDim).Render("Errors")))
	}

	return sidebarStyle.Width(width).Render(b.String())
}

func (m Model) viewMain(width int) string {
	var b strings.Builder

	if len(m.issues) == 0 && len(m.errors) == 0 {
		b.WriteString(countGreenStyle.Render("\n  No issues found. Your account looks clean!\n"))
		return mainPanelStyle.Width(width).Render(b.String())
	}

	// Calculate visible area: header ~8 lines, statusbar 2, sidebar overhead ~6
	maxVisible := m.height - 14
	if maxVisible < 5 {
		maxVisible = 5
	}

	// Issue list
	start := 0
	if m.cursor >= maxVisible {
		start = m.cursor - maxVisible + 1
	}

	for i := start; i < len(m.issues) && i < start+maxVisible; i++ {
		issue := m.issues[i]
		badge := severityBadge(issue.Severity)
		line := fmt.Sprintf(" %s  %s", badge, truncate(issue.Description, width-20))

		if i == m.cursor {
			line = issueSelectedStyle.Width(width - 4).Render(line)
			if m.expanded {
				line += "\n" + m.viewDetail(issue, width-4)
			}
		} else {
			line = issueNormalStyle.Width(width - 4).Render(line)
		}

		b.WriteString(line + "\n")
	}

	// Scroll indicator
	if len(m.issues) > maxVisible {
		b.WriteString(lipgloss.NewStyle().Foreground(colorDim).Render(
			fmt.Sprintf("\n  Showing %d-%d of %d issues", start+1, min(start+maxVisible, len(m.issues)), len(m.issues))))
	}

	// Errors at the bottom
	if len(m.errors) > 0 {
		b.WriteString("\n")
		b.WriteString(errorStyle.Render(fmt.Sprintf("  Scan Errors (%d):", len(m.errors))))
		b.WriteString("\n")
		for _, err := range m.errors {
			b.WriteString(fmt.Sprintf("    %s\n", lipgloss.NewStyle().Foreground(colorDim).Render(err.Error())))
		}
	}

	return mainPanelStyle.Width(width).Render(b.String())
}

func (m Model) viewDetail(issue scanner.Issue, width int) string {
	var b strings.Builder
	indent := "       "

	b.WriteString(indent + detailLabelStyle.Render("Scanner:  ") + detailValueStyle.Render(issue.Scanner) + "\n")
	b.WriteString(indent + detailLabelStyle.Render("Resource: ") + detailValueStyle.Render(issue.ResourceID) + "\n")
	b.WriteString(indent + detailLabelStyle.Render("Fix:      ") + suggestionStyle.Render(issue.Suggestion) + "\n")

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorCyan).
		Padding(0, 1).
		Width(width - 10).
		Render(b.String())

	return "     " + box
}

func (m Model) viewStatusBar() string {
	left := ""
	if m.state == stateReady {
		left = fmt.Sprintf(" Scanned in %s", m.elapsed.Round(time.Millisecond))
	}

	right := " [j/k] Navigate  [Enter] Expand  [R] Re-scan  [Q] Quit "

	gap := m.width - lipgloss.Width(left) - lipgloss.Width(right)
	if gap < 0 {
		gap = 0
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
		return warningBadge.Render("WARNING")
	default:
		return "UNKNOWN"
	}
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

func truncate(s string, max int) string {
	if max < 4 {
		max = 4
	}
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
