package publish_test

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	publish "github.com/ben-wangz/forgekit/internal/publish"
)

func TestPushChartWithDepsPublishesAliasesFromDigest(t *testing.T) {
	tags, warnings, err := publish.ResolvePublishTags("1.6.0", true, true, true)
	if err != nil {
		t.Fatalf("ResolvePublishTags returned error: %v", err)
	}
	if len(warnings) != 0 {
		t.Fatalf("unexpected warnings: %v", warnings)
	}

	cfg := &publish.ChartConfig{ChartRegistry: "registry.example/charts", ChartRegistryHost: "registry.example"}

	pushCalls := 0
	copySources := make([]string, 0)
	copyTargets := make([]string, 0)

	deps := publish.ChartPushDeps{
		HelmPush: func(chartPackage, registry string, insecure bool) (string, error) {
			pushCalls++
			if chartPackage != "/tmp/demo-1.6.0.tgz" {
				t.Fatalf("unexpected chart package: %s", chartPackage)
			}
			if registry != cfg.ChartRegistry {
				t.Fatalf("unexpected registry: %s", registry)
			}
			return "sha256:abc123", nil
		},
		OCICopyByDigest: func(sourceRef, targetRef, username, password string, insecure bool) error {
			copySources = append(copySources, sourceRef)
			copyTargets = append(copyTargets, targetRef)
			return nil
		},
		HelmRegistryLogin: func(registry, username, password string, insecure bool) error {
			t.Fatalf("login should not be called without credentials")
			return nil
		},
		HelmRegistryLogout: func(registry string) error {
			t.Fatalf("logout should not be called without credentials")
			return nil
		},
	}

	err = publish.PushChartWithDeps(cfg, "/tmp/demo-1.6.0.tgz", "demo", "1.6.0", tags, deps)
	if err != nil {
		t.Fatalf("PushChartWithDeps returned error: %v", err)
	}

	if pushCalls != 1 {
		t.Fatalf("expected helm push once, got %d", pushCalls)
	}

	expectedSource := "registry.example/charts/demo@sha256:abc123"
	for _, source := range copySources {
		if source != expectedSource {
			t.Fatalf("unexpected alias source ref: %s", source)
		}
	}

	expectedTargets := []string{
		"registry.example/charts/demo:latest",
		"registry.example/charts/demo:1",
		"registry.example/charts/demo:1.6",
	}
	if !reflect.DeepEqual(copyTargets, expectedTargets) {
		t.Fatalf("unexpected alias targets: want %v, got %v", expectedTargets, copyTargets)
	}
}

func TestPushChartWithDepsFailFastReportsPublishedTags(t *testing.T) {
	tags, warnings, err := publish.ResolvePublishTags("1.6.0", true, true, true)
	if err != nil {
		t.Fatalf("ResolvePublishTags returned error: %v", err)
	}
	if len(warnings) != 0 {
		t.Fatalf("unexpected warnings: %v", warnings)
	}

	cfg := &publish.ChartConfig{ChartRegistry: "registry.example/charts", ChartRegistryHost: "registry.example"}

	pushCalls := 0
	copyCalls := 0
	deps := publish.ChartPushDeps{
		HelmPush: func(chartPackage, registry string, insecure bool) (string, error) {
			pushCalls++
			return "sha256:abc123", nil
		},
		OCICopyByDigest: func(sourceRef, targetRef, username, password string, insecure bool) error {
			copyCalls++
			if strings.HasSuffix(targetRef, ":latest") {
				return nil
			}
			return errors.New("copy failed")
		},
	}

	err = publish.PushChartWithDeps(cfg, "/tmp/demo-1.6.0.tgz", "demo", "1.6.0", tags, deps)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if pushCalls != 1 {
		t.Fatalf("expected helm push once, got %d", pushCalls)
	}

	if copyCalls != 2 {
		t.Fatalf("expected fail-fast on second alias publish, got %d alias calls", copyCalls)
	}

	if !strings.Contains(err.Error(), "already published tags: 1.6.0, latest") {
		t.Fatalf("unexpected error message: %v", err)
	}
}
