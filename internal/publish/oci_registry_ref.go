package publish

import (
	"fmt"
	"strings"
)

type ociReference struct {
	Registry   string
	Repository string
	Reference  string
}

func parseDigestReference(ref string) (*ociReference, error) {
	parts := strings.SplitN(ref, "@", 2)
	if len(parts) != 2 || strings.TrimSpace(parts[1]) == "" {
		return nil, fmt.Errorf("invalid source digest reference: %s", ref)
	}

	registry, repository, err := splitRegistryRepository(parts[0])
	if err != nil {
		return nil, err
	}

	return &ociReference{Registry: registry, Repository: repository, Reference: strings.TrimSpace(parts[1])}, nil
}

func parseTagReference(ref string) (*ociReference, error) {
	slash := strings.Index(ref, "/")
	colon := strings.LastIndex(ref, ":")
	if slash < 0 || colon <= slash || colon == len(ref)-1 {
		return nil, fmt.Errorf("invalid target tag reference: %s", ref)
	}

	registry, repository, err := splitRegistryRepository(ref[:colon])
	if err != nil {
		return nil, err
	}

	return &ociReference{Registry: registry, Repository: repository, Reference: strings.TrimSpace(ref[colon+1:])}, nil
}

func splitRegistryRepository(value string) (string, string, error) {
	parts := strings.SplitN(value, "/", 2)
	if len(parts) != 2 || strings.TrimSpace(parts[0]) == "" || strings.TrimSpace(parts[1]) == "" {
		return "", "", fmt.Errorf("invalid OCI reference: %s", value)
	}

	return strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1]), nil
}
