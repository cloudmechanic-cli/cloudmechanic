package tui

import "github.com/charmbracelet/lipgloss"

// Color palette.
var (
	colorCritical = lipgloss.Color("#FF4444")
	colorWarning  = lipgloss.Color("#FFAA00")
	colorGreen    = lipgloss.Color("#00CC66")
	colorCyan     = lipgloss.Color("#00BFFF")
	colorDim      = lipgloss.Color("#666666")
	colorWhite    = lipgloss.Color("#FFFFFF")
	colorBg       = lipgloss.Color("#1A1A2E")
	colorSidebar  = lipgloss.Color("#16213E")
	colorSelected = lipgloss.Color("#0F3460")
)

// Layout styles.
var (
	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorCyan).
			Padding(0, 1)

	sidebarStyle = lipgloss.NewStyle().
			Background(colorSidebar).
			Padding(1, 2).
			MarginRight(1)

	sidebarTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(colorCyan).
				MarginBottom(1)

	regionActiveStyle = lipgloss.NewStyle().
				Foreground(colorGreen).
				Bold(true)

	regionIdleStyle = lipgloss.NewStyle().
			Foreground(colorDim)

	mainPanelStyle = lipgloss.NewStyle().
			Padding(1, 2)

	statusBarStyle = lipgloss.NewStyle().
			Foreground(colorDim).
			Padding(0, 1)
)

// Issue severity styles.
var (
	criticalBadge = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorWhite).
			Background(colorCritical).
			Padding(0, 1)

	warningBadge = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#000000")).
			Background(colorWarning).
			Padding(0, 1)

	issueNormalStyle = lipgloss.NewStyle().
				Padding(0, 1)

	issueSelectedStyle = lipgloss.NewStyle().
				Background(colorSelected).
				Foreground(colorWhite).
				Bold(true).
				Padding(0, 1)

	detailLabelStyle = lipgloss.NewStyle().
				Foreground(colorCyan).
				Bold(true)

	detailValueStyle = lipgloss.NewStyle().
				Foreground(colorWhite)

	suggestionStyle = lipgloss.NewStyle().
			Foreground(colorGreen).
			Italic(true)

	spinnerStyle = lipgloss.NewStyle().
			Foreground(colorCyan)

	errorStyle = lipgloss.NewStyle().
			Foreground(colorCritical)

	countCriticalStyle = lipgloss.NewStyle().
				Foreground(colorCritical).
				Bold(true)

	countWarningStyle = lipgloss.NewStyle().
				Foreground(colorWarning).
				Bold(true)

	countGreenStyle = lipgloss.NewStyle().
			Foreground(colorGreen).
			Bold(true)
)
