package cmd

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(upgradeCmd)
}

var upgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "Upgrade CloudMechanic to the latest version",
	RunE:  runUpgrade,
}

const githubReleasesAPI = "https://api.github.com/repos/cloudmechanic-cli/cloudmechanic/releases/latest"

type githubRelease struct {
	TagName string `json:"tag_name"`
	Assets  []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
	} `json:"assets"`
}

func runUpgrade(cmd *cobra.Command, args []string) error {
	info := color.New(color.FgCyan)
	bold := color.New(color.FgWhite, color.Bold)
	green := color.New(color.FgGreen, color.Bold)
	warn := color.New(color.FgYellow)

	info.Println("Checking for updates...")

	release, err := fetchLatestRelease()
	if err != nil {
		return fmt.Errorf("failed to fetch latest release: %w", err)
	}

	latestVersion := strings.TrimPrefix(release.TagName, "v")
	currentVersion := strings.TrimPrefix(Version, "v")

	bold.Printf("  Current : %s\n", currentVersion)
	bold.Printf("  Latest  : %s\n", latestVersion)

	if currentVersion == latestVersion {
		green.Println("\nAlready up to date.")
		return nil
	}

	if currentVersion != "dev" && !isNewer(latestVersion, currentVersion) {
		warn.Println("\nNo newer version available.")
		return nil
	}

	assetName := buildAssetName(release.TagName)
	downloadURL := ""
	for _, a := range release.Assets {
		if a.Name == assetName {
			downloadURL = a.BrowserDownloadURL
			break
		}
	}
	if downloadURL == "" {
		return fmt.Errorf("no asset found for %s/%s (looked for %q)", runtime.GOOS, runtime.GOARCH, assetName)
	}

	info.Printf("\nDownloading %s...\n", assetName)
	tmpDir, err := os.MkdirTemp("", "cloudmechanic-upgrade-*")
	if err != nil {
		return fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	archivePath := filepath.Join(tmpDir, assetName)
	if err := downloadFile(archivePath, downloadURL); err != nil {
		return fmt.Errorf("download failed: %w", err)
	}

	info.Println("Extracting...")
	binaryPath, err := extractBinary(archivePath, tmpDir)
	if err != nil {
		return fmt.Errorf("extraction failed: %w", err)
	}

	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("could not locate current binary: %w", err)
	}
	execPath, err = filepath.EvalSymlinks(execPath)
	if err != nil {
		return fmt.Errorf("could not resolve symlink: %w", err)
	}

	info.Printf("Replacing %s...\n", execPath)
	if err := replaceBinary(binaryPath, execPath); err != nil {
		return fmt.Errorf("failed to replace binary: %w", err)
	}

	green.Printf("\nCloudMechanic upgraded to v%s\n", latestVersion)
	return nil
}

// fetchLatestRelease calls the GitHub API.
func fetchLatestRelease() (*githubRelease, error) {
	req, err := http.NewRequest(http.MethodGet, githubReleasesAPI, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "cloudmechanic-cli/"+Version)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned %d", resp.StatusCode)
	}

	var release githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, err
	}
	return &release, nil
}

// buildAssetName returns the expected archive filename for the current platform.
func buildAssetName(tag string) string {
	version := strings.TrimPrefix(tag, "v")
	goos := runtime.GOOS
	goarch := runtime.GOARCH

	// Map Go arch names to GoReleaser archive names.
	archName := goarch
	if goarch == "amd64" {
		archName = "amd64"
	} else if goarch == "arm64" {
		archName = "arm64"
	}

	if goos == "windows" {
		return fmt.Sprintf("cloudmechanic_%s_%s_%s.zip", version, goos, archName)
	}
	return fmt.Sprintf("cloudmechanic_%s_%s_%s.tar.gz", version, goos, archName)
}

// downloadFile saves the URL content to dest.
func downloadFile(dest, url string) error {
	resp, err := http.Get(url) //nolint:gosec
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status %d", resp.StatusCode)
	}

	f, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(f, resp.Body)
	return err
}

// extractBinary pulls the cloudmechanic binary out of a .tar.gz or .zip archive.
func extractBinary(archivePath, destDir string) (string, error) {
	if strings.HasSuffix(archivePath, ".tar.gz") {
		return extractTarGz(archivePath, destDir)
	}
	return extractZip(archivePath, destDir)
}

func extractTarGz(archivePath, destDir string) (string, error) {
	f, err := os.Open(archivePath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	gz, err := gzip.NewReader(f)
	if err != nil {
		return "", err
	}
	defer gz.Close()

	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}

		baseName := filepath.Base(hdr.Name)
		if baseName != "cloudmechanic" {
			continue
		}

		out := filepath.Join(destDir, "cloudmechanic")
		outF, err := os.OpenFile(out, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
		if err != nil {
			return "", err
		}

		if _, err := io.Copy(outF, tr); err != nil {
			outF.Close()
			return "", err
		}
		outF.Close()
		return out, nil
	}

	return "", fmt.Errorf("cloudmechanic binary not found in archive")
}

func extractZip(archivePath, destDir string) (string, error) {
	r, err := zip.OpenReader(archivePath)
	if err != nil {
		return "", err
	}
	defer r.Close()

	for _, f := range r.File {
		baseName := filepath.Base(f.Name)
		if baseName != "cloudmechanic.exe" && baseName != "cloudmechanic" {
			continue
		}

		out := filepath.Join(destDir, baseName)
		outF, err := os.OpenFile(out, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
		if err != nil {
			return "", err
		}

		rc, err := f.Open()
		if err != nil {
			outF.Close()
			return "", err
		}

		_, copyErr := io.Copy(outF, rc)
		rc.Close()
		outF.Close()
		if copyErr != nil {
			return "", copyErr
		}
		return out, nil
	}

	return "", fmt.Errorf("cloudmechanic binary not found in zip archive")
}

// replaceBinary atomically swaps the new binary into place.
// It writes to a temp file next to the target, then renames to avoid
// replacing an open file on Windows.
func replaceBinary(src, dest string) error {
	destDir := filepath.Dir(dest)
	tmp, err := os.CreateTemp(destDir, ".cloudmechanic-upgrade-*")
	if err != nil {
		// Fall back: try writing directly (e.g. read-only dir handled below)
		return copyAndReplace(src, dest)
	}
	tmp.Close()
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath)

	if err := copyFile(src, tmpPath, 0755); err != nil {
		return err
	}

	return os.Rename(tmpPath, dest)
}

func copyAndReplace(src, dest string) error {
	return copyFile(src, dest, 0755)
}

func copyFile(src, dest string, mode os.FileMode) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.OpenFile(dest, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}

// isNewer reports whether version a is strictly newer than b.
// Uses simple integer comparison on each semver component.
func isNewer(a, b string) bool {
	pa := parseSemver(a)
	pb := parseSemver(b)
	for i := range pa {
		if i >= len(pb) {
			return true
		}
		if pa[i] > pb[i] {
			return true
		}
		if pa[i] < pb[i] {
			return false
		}
	}
	return false
}

func parseSemver(v string) []int {
	v = strings.TrimPrefix(v, "v")
	parts := strings.Split(v, ".")
	nums := make([]int, len(parts))
	for i, p := range parts {
		fmt.Sscanf(p, "%d", &nums[i])
	}
	return nums
}
