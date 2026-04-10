package publish

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	versiontool "github.com/ben-wangz/forgekit/internal/version"
	"gopkg.in/yaml.v3"
)

const (
	deprecatedRegistryWarning   = "[WARN] DEPRECATED: chart registry is resolved via compatibility fallback path. Please set CHART_REGISTRY explicitly. This compatibility may be removed in the next release."
	deprecatedCredentialWarning = "[WARN] DEPRECATED: chart credentials are resolved via compatibility fallback path. Please set CHART_REGISTRY_USERNAME and CHART_REGISTRY_PASSWORD explicitly. This compatibility may be removed in the next release."
)

type ChartPushDeps struct {
	HelmRegistryLogin  func(registry, username, password string, insecure bool) error
	HelmPush           func(chartPackage, registry string, insecure bool) (string, error)
	HelmRegistryLogout func(registry string) error
	OCICopyByDigest    func(sourceRef, targetRef, username, password string, insecure bool) error
}

func DefaultChartPushDeps() ChartPushDeps {
	return ChartPushDeps{
		HelmRegistryLogin:  helmRegistryLogin,
		HelmPush:           helmPush,
		HelmRegistryLogout: helmRegistryLogout,
		OCICopyByDigest:    ociCopyByDigest,
	}
}

func (d ChartPushDeps) withDefaults() ChartPushDeps {
	defaults := DefaultChartPushDeps()
	if d.HelmRegistryLogin == nil {
		d.HelmRegistryLogin = defaults.HelmRegistryLogin
	}
	if d.HelmPush == nil {
		d.HelmPush = defaults.HelmPush
	}
	if d.HelmRegistryLogout == nil {
		d.HelmRegistryLogout = defaults.HelmRegistryLogout
	}
	if d.OCICopyByDigest == nil {
		d.OCICopyByDigest = defaults.OCICopyByDigest
	}
	return d
}

type ChartYAML struct {
	Name    string `yaml:"name"`
	Version string `yaml:"version"`
}

func cmdChartBuild(args []string, projectRoot string) error {
	cfg, err := loadChartConfig(args, projectRoot)
	if err != nil {
		return err
	}

	if cfg.RegistryFallbackDeprecated {
		fmt.Fprintln(os.Stderr, deprecatedRegistryWarning)
	}
	if cfg.CredentialFallbackDeprecated {
		fmt.Fprintln(os.Stderr, deprecatedCredentialWarning)
	}

	chartYAML, err := readChartYAML(cfg.ChartDir)
	if err != nil {
		return fmt.Errorf("failed to read Chart.yaml: %w", err)
	}

	versionManager, err := versiontool.NewManager(cfg.ProjectRoot)
	if err != nil {
		return err
	}

	publishVersion, err := resolveChartPublishVersion(versionManager, chartYAML.Name, cfg.Semver)
	if err != nil {
		return fmt.Errorf("failed to resolve chart publish version: %w", err)
	}

	chartVersion := publishVersion.Value
	tags, warnings, err := ResolvePublishTags(chartVersion, cfg.Semver, cfg.Push, cfg.MultiTag)
	if err != nil {
		return err
	}

	for _, warning := range warnings {
		fmt.Fprintln(os.Stderr, warning)
	}

	printChartBuildInfo(cfg, chartYAML.Name, publishVersion.Mode, chartYAML.Version, chartVersion, tags)

	fmt.Println()
	fmt.Println("Packaging chart...")
	fmt.Println()

	chartPackage, err := helmPackage(cfg.ChartDir, chartVersion)
	if err != nil {
		return fmt.Errorf("failed to package chart: %w", err)
	}
	defer os.Remove(chartPackage)

	fmt.Println()
	fmt.Println("===========================================")
	fmt.Println("Package Complete")
	fmt.Println("===========================================")
	fmt.Printf("Chart package: %s\n", chartPackage)
	fmt.Println()

	if cfg.Push {
		if err := PushChartWithDeps(cfg, chartPackage, chartYAML.Name, chartVersion, tags, DefaultChartPushDeps()); err != nil {
			return err
		}
	}

	return nil
}

func readChartYAML(chartDir string) (*ChartYAML, error) {
	chartFile := filepath.Join(chartDir, "Chart.yaml")
	data, err := os.ReadFile(chartFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read Chart.yaml: %w", err)
	}

	var chart ChartYAML
	if err := yaml.Unmarshal(data, &chart); err != nil {
		return nil, fmt.Errorf("failed to parse Chart.yaml: %w", err)
	}

	return &chart, nil
}

func printChartBuildInfo(cfg *ChartConfig, chartName, versionMode, baseVersion, fullVersion string, tags []string) {
	fmt.Println("===========================================")
	fmt.Println("Building Helm Chart")
	fmt.Println("===========================================")
	fmt.Printf("Project root: %s\n", cfg.ProjectRoot)
	fmt.Printf("Chart dir: %s\n", cfg.ChartDir)
	fmt.Printf("Chart name: %s\n", chartName)
	fmt.Printf("Version mode: %s\n", versionMode)
	fmt.Printf("Base version: %s\n", baseVersion)
	fmt.Printf("Package version: %s\n", fullVersion)
	if cfg.Push {
		fmt.Println("Push: enabled")
		fmt.Printf("Registry: %s\n", cfg.ChartRegistry)
		fmt.Printf("Push tags: %s\n", strings.Join(tags, ", "))
	}
}

func PushChartWithDeps(cfg *ChartConfig, chartPackage, chartName, chartVersion string, tags []string, deps ChartPushDeps) error {
	deps = deps.withDefaults()

	fmt.Println("===========================================")
	fmt.Println("Pushing Helm Chart")
	fmt.Println("===========================================")

	registryPath := cfg.ChartRegistry
	fullTagRef := fmt.Sprintf("%s/%s:%s", cfg.ChartRegistry, chartName, chartVersion)
	fmt.Printf("Chart reference: %s\n", fullTagRef)
	fmt.Println()
	publishedTags := make([]string, 0, len(tags))

	loggedIn := false
	if cfg.Username != "" && cfg.Password != "" {
		fmt.Println("Logging in to registry...")
		if err := deps.HelmRegistryLogin(cfg.ChartRegistryHost, cfg.Username, cfg.Password, cfg.RegistryPlainHTTP); err != nil {
			return fmt.Errorf("failed to login to registry: %w", err)
		}
		fmt.Println()
		loggedIn = true
	}

	fmt.Println("Pushing chart...")
	digest, err := deps.HelmPush(chartPackage, registryPath, cfg.RegistryPlainHTTP)
	if err != nil {
		if loggedIn {
			_ = deps.HelmRegistryLogout(cfg.ChartRegistryHost)
		}
		return fmt.Errorf("failed to push chart: %w", err)
	}
	publishedTags = append(publishedTags, chartVersion)

	extraTags := chartExtraTags(tags, chartVersion)
	if len(extraTags) > 0 {
		sourceDigestRef := fmt.Sprintf("%s/%s@%s", cfg.ChartRegistry, chartName, digest)
		fmt.Println()
		fmt.Printf("Pushed digest: %s\n", digest)
		for _, tag := range extraTags {
			targetRef := fmt.Sprintf("%s/%s:%s", cfg.ChartRegistry, chartName, tag)
			fmt.Printf("Publishing alias tag: %s\n", targetRef)
			if err := deps.OCICopyByDigest(sourceDigestRef, targetRef, cfg.Username, cfg.Password, cfg.RegistryPlainHTTP); err != nil {
				if loggedIn {
					_ = deps.HelmRegistryLogout(cfg.ChartRegistryHost)
				}
				return fmt.Errorf("failed to publish alias tag %s: %w (already published tags: %s)", tag, err, strings.Join(publishedTags, ", "))
			}
			publishedTags = append(publishedTags, tag)
		}
	}

	if loggedIn {
		fmt.Println()
		fmt.Println("Logging out from registry...")
		if err := deps.HelmRegistryLogout(cfg.ChartRegistryHost); err != nil {
			return fmt.Errorf("failed to logout from registry: %w", err)
		}
	}

	fmt.Println()
	fmt.Println("===========================================")
	fmt.Println("Push Complete")
	fmt.Println("===========================================")
	fmt.Printf("Registry: %s\n", cfg.ChartRegistry)
	fmt.Printf("Chart: %s\n", chartName)
	fmt.Printf("Digest: %s\n", digest)
	fmt.Printf("Tags: %s\n", strings.Join(tags, ", "))
	fmt.Println()

	return nil
}

func chartExtraTags(tags []string, fullTag string) []string {
	extra := make([]string, 0, len(tags))
	for _, tag := range tags {
		if tag == fullTag {
			continue
		}
		extra = append(extra, tag)
	}
	return extra
}
