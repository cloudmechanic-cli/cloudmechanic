package tui

import (
	"context"
	"sort"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/cloudmechanic-cli/cloudmechanic/internal/scanner"
)

// --- Messages ---

type scanResultMsg struct {
	issues  []scanner.Issue
	errors  []error
	elapsed time.Duration
}

type scanStartMsg struct{}

// --- Types ---

type pane int

const (
	paneSidebar pane = iota
	paneMain
)

type state int

const (
	stateLoading state = iota
	stateReady
)

type sevFilter int

const (
	filterAll sevFilter = iota
	filterCritical
	filterWarning
)

func (f sevFilter) String() string {
	switch f {
	case filterCritical:
		return "Critical Only"
	case filterWarning:
		return "Warnings Only"
	default:
		return "All"
	}
}

func (f sevFilter) next() sevFilter {
	return (f + 1) % 3
}

// ScannerBuilder abstracts scanner creation for the TUI.
type ScannerBuilder struct {
	Cfg        aws.Config
	Regions    []string
	AllRegions bool
	Build      func(cfg aws.Config, regions []string) []scanner.Scanner
	Run        func(ctx context.Context, scanners []scanner.Scanner) ([]scanner.Issue, []error)
}

// Model is the top-level Bubble Tea model.
type Model struct {
	state   state
	builder ScannerBuilder
	spinner spinner.Model

	// Data.
	issues  []scanner.Issue
	errors  []error
	elapsed time.Duration

	// Two-pane navigation.
	focus        pane
	regionCursor int // 0 = "Global", 1..N = individual regions
	issueCursor  int
	expanded     bool

	// Filtering and search.
	sevFilter  sevFilter
	searchMode bool
	searchText string

	// Derived (recomputed on data/filter change).
	filtered []scanner.Issue // issues after region + severity + search filters
	groups   []serviceGroup // filtered issues grouped by service

	// Terminal size.
	width  int
	height int
}

// serviceGroup holds issues for a single service category.
type serviceGroup struct {
	service string
	issues  []scanner.Issue
}

// NewModel creates a Model ready to be run.
func NewModel(b ScannerBuilder) Model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = spinnerStyle

	return Model{
		state:   stateLoading,
		builder: b,
		spinner: s,
		focus:   paneSidebar,
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, startScan())
}

func startScan() tea.Cmd {
	return func() tea.Msg { return scanStartMsg{} }
}

func (m Model) runScanCmd() tea.Cmd {
	b := m.builder
	return func() tea.Msg {
		ctx := context.Background()
		start := time.Now()

		scanners := b.Build(b.Cfg, b.Regions)
		issues, errs := b.Run(ctx, scanners)

		sort.Slice(issues, func(i, j int) bool {
			return issues[i].Severity > issues[j].Severity
		})

		return scanResultMsg{
			issues:  issues,
			errors:  errs,
			elapsed: time.Since(start),
		}
	}
}

// Update handles messages.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		if m.searchMode {
			return m.handleSearchKey(msg)
		}
		return m.handleKey(msg)

	case scanStartMsg:
		m.state = stateLoading
		m.issues = nil
		m.errors = nil
		m.issueCursor = 0
		m.expanded = false
		m.filtered = nil
		m.groups = nil
		return m, tea.Batch(m.spinner.Tick, m.runScanCmd())

	case scanResultMsg:
		m.state = stateReady
		m.issues = msg.issues
		m.errors = msg.errors
		m.elapsed = msg.elapsed
		m.refilter()
		return m, nil

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m *Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit

	case "tab":
		if m.state == stateReady {
			if m.focus == paneSidebar {
				m.focus = paneMain
			} else {
				m.focus = paneSidebar
			}
			m.expanded = false
		}

	case "r":
		if m.state == stateReady {
			return m, startScan()
		}

	case "f":
		if m.state == stateReady {
			m.sevFilter = m.sevFilter.next()
			m.issueCursor = 0
			m.expanded = false
			m.refilter()
		}

	case "/":
		if m.state == stateReady {
			m.searchMode = true
			m.searchText = ""
		}

	case "up", "k":
		if m.state != stateReady {
			break
		}
		if m.focus == paneSidebar {
			if m.regionCursor > 0 {
				m.regionCursor--
				m.issueCursor = 0
				m.expanded = false
				m.refilter()
			}
		} else {
			if m.issueCursor > 0 {
				m.issueCursor--
				m.expanded = false
			}
		}

	case "down", "j":
		if m.state != stateReady {
			break
		}
		if m.focus == paneSidebar {
			maxRegion := len(m.builder.Regions) // 0=Global + N regions
			if m.regionCursor < maxRegion {
				m.regionCursor++
				m.issueCursor = 0
				m.expanded = false
				m.refilter()
			}
		} else {
			if m.issueCursor < len(m.filtered)-1 {
				m.issueCursor++
				m.expanded = false
			}
		}

	case "enter":
		if m.state == stateReady && m.focus == paneMain && len(m.filtered) > 0 {
			m.expanded = !m.expanded
		}

	case "esc":
		m.expanded = false
		if m.searchText != "" {
			m.searchText = ""
			m.issueCursor = 0
			m.refilter()
		}
	}

	return m, nil
}

func (m *Model) handleSearchKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter", "esc":
		m.searchMode = false
		m.focus = paneMain
	case "backspace":
		if len(m.searchText) > 0 {
			m.searchText = m.searchText[:len(m.searchText)-1]
			m.issueCursor = 0
			m.refilter()
		}
	case "ctrl+c":
		return m, tea.Quit
	default:
		if len(msg.String()) == 1 {
			m.searchText += msg.String()
			m.issueCursor = 0
			m.refilter()
		}
	}
	return m, nil
}

// selectedRegion returns the region string for the current sidebar cursor.
// Empty string means "Global" (show all).
func (m Model) selectedRegion() string {
	if m.regionCursor == 0 {
		return ""
	}
	idx := m.regionCursor - 1
	if idx < len(m.builder.Regions) {
		return m.builder.Regions[idx]
	}
	return ""
}

// refilter rebuilds the filtered issue list and service groups.
func (m *Model) refilter() {
	region := m.selectedRegion()
	search := strings.ToLower(m.searchText)

	var filtered []scanner.Issue
	for _, issue := range m.issues {
		// Region filter.
		if region != "" && issue.Region != region && issue.Region != "global" {
			continue
		}
		// Severity filter.
		if m.sevFilter == filterCritical && issue.Severity != scanner.SeverityCritical {
			continue
		}
		if m.sevFilter == filterWarning && issue.Severity != scanner.SeverityWarning {
			continue
		}
		// Search filter.
		if search != "" {
			hay := strings.ToLower(issue.Description + " " + issue.ResourceID + " " + issue.Scanner)
			if !strings.Contains(hay, search) {
				continue
			}
		}
		filtered = append(filtered, issue)
	}

	m.filtered = filtered
	if m.issueCursor >= len(m.filtered) {
		m.issueCursor = max(0, len(m.filtered)-1)
	}

	// Group by service.
	groupMap := make(map[string][]scanner.Issue)
	var order []string
	for _, issue := range filtered {
		svc := serviceFromScanner(issue.Scanner)
		if _, exists := groupMap[svc]; !exists {
			order = append(order, svc)
		}
		groupMap[svc] = append(groupMap[svc], issue)
	}

	m.groups = make([]serviceGroup, 0, len(order))
	for _, svc := range order {
		m.groups = append(m.groups, serviceGroup{service: svc, issues: groupMap[svc]})
	}
}

// serviceFromScanner maps scanner names to service categories.
func serviceFromScanner(name string) string {
	n := strings.ToLower(name)
	switch {
	case strings.Contains(n, "ebs"), strings.Contains(n, "security group"),
		strings.Contains(n, "elastic ip"), strings.Contains(n, "eip"),
		strings.Contains(n, "snapshot"):
		return "EC2"
	case strings.Contains(n, "s3"):
		return "S3"
	case strings.Contains(n, "rds"):
		return "RDS"
	case strings.Contains(n, "iam"), strings.Contains(n, "mfa"):
		return "IAM"
	case strings.Contains(n, "dynamo"):
		return "DynamoDB"
	case strings.Contains(n, "nat"), strings.Contains(n, "vpc"), strings.Contains(n, "flow"):
		return "VPC"
	case strings.Contains(n, "lambda"):
		return "Lambda"
	default:
		return "Other"
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
