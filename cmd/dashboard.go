package cmd

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/cloudmechanic-cli/cloudmechanic/internal/tui"
	"github.com/spf13/cobra"
)

var (
	dashRegion     string
	dashProfile    string
	dashAllRegions bool
)

func init() {
	dashboardCmd.Flags().StringVar(&dashRegion, "region", "", "AWS region to scan (defaults to AWS_REGION or config)")
	dashboardCmd.Flags().StringVar(&dashProfile, "profile", "", "AWS named profile to use")
	dashboardCmd.Flags().BoolVar(&dashAllRegions, "all-regions", false, "Scan all available AWS regions")
	rootCmd.AddCommand(dashboardCmd)
}

var dashboardCmd = &cobra.Command{
	Use:   "dashboard",
	Short: "Interactive TUI dashboard for AWS scanning",
	RunE:  runDashboard,
}

func runDashboard(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	cfg, regions, err := LoadAWSConfig(ctx, dashRegion, dashProfile, dashAllRegions)
	if err != nil {
		return err
	}

	m := tui.NewModel(tui.ScannerBuilder{
		Cfg:     cfg,
		Regions: regions,
		Build:   BuildScanners,
		Run:     RunScanners,
	})

	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("dashboard error: %w", err)
	}

	return nil
}
