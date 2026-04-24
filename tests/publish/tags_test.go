package publish_test

import (
	"reflect"
	"strings"
	"testing"

	publish "github.com/ben-wangz/forgekit/internal/publish"
)

func TestResolvePublishTagsStableSemverMultiTag(t *testing.T) {
	tags, warnings, err := publish.ResolvePublishTags("1.6.0", true, true, true)
	if err != nil {
		t.Fatalf("ResolvePublishTags returned error: %v", err)
	}

	expectedTags := []string{"latest", "1", "1.6", "1.6.0"}
	if !reflect.DeepEqual(tags, expectedTags) {
		t.Fatalf("unexpected tags: want %v, got %v", expectedTags, tags)
	}

	if len(warnings) != 0 {
		t.Fatalf("expected no warnings, got %v", warnings)
	}
}

func TestResolvePublishTagsPrereleaseDowngradesToSingleTag(t *testing.T) {
	tags, warnings, err := publish.ResolvePublishTags("1.6.0-rc.1", true, true, true)
	if err != nil {
		t.Fatalf("ResolvePublishTags returned error: %v", err)
	}

	expectedTags := []string{"1.6.0-rc.1"}
	if !reflect.DeepEqual(tags, expectedTags) {
		t.Fatalf("unexpected tags: want %v, got %v", expectedTags, tags)
	}

	if len(warnings) != 1 || !strings.Contains(warnings[0], "prerelease") {
		t.Fatalf("expected prerelease warning, got %v", warnings)
	}
}

func TestResolvePublishTagsRejectsBuildMetadata(t *testing.T) {
	_, _, err := publish.ResolvePublishTags("1.6.0+build.1", true, true, true)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !strings.Contains(err.Error(), "OCI tag") || !strings.Contains(err.Error(), "+") {
		t.Fatalf("unexpected error message: %v", err)
	}
}
