package publish

import (
	"fmt"
	"net/http"
	"strings"
	"time"
)

func ociCopyByDigest(sourceRef, targetRef, username, password string, insecure bool) error {
	source, err := parseDigestReference(sourceRef)
	if err != nil {
		return err
	}

	target, err := parseTagReference(targetRef)
	if err != nil {
		return err
	}

	scheme := "https"
	if insecure {
		scheme = "http"
	}

	client := &http.Client{Timeout: 30 * time.Second}
	accept := strings.Join([]string{
		"application/vnd.oci.image.manifest.v1+json",
		"application/vnd.oci.artifact.manifest.v1+json",
		"application/vnd.docker.distribution.manifest.v2+json",
	}, ", ")

	getURL := fmt.Sprintf("%s://%s/v2/%s/manifests/%s", scheme, source.Registry, source.Repository, source.Reference)
	getResp, err := doRegistryRequest(client, "GET", getURL, nil, map[string]string{"Accept": accept}, username, password)
	if err != nil {
		return fmt.Errorf("failed to fetch source manifest: %w", err)
	}
	if getResp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to fetch source manifest %s: status %d, body: %s", sourceRef, getResp.StatusCode, strings.TrimSpace(string(getResp.Body)))
	}

	contentType := getResp.Headers.Get("Content-Type")
	if idx := strings.Index(contentType, ";"); idx >= 0 {
		contentType = strings.TrimSpace(contentType[:idx])
	}
	if contentType == "" {
		return fmt.Errorf("registry response missing Content-Type for %s", sourceRef)
	}

	putURL := fmt.Sprintf("%s://%s/v2/%s/manifests/%s", scheme, target.Registry, target.Repository, target.Reference)
	putResp, err := doRegistryRequest(client, "PUT", putURL, getResp.Body, map[string]string{"Content-Type": contentType}, username, password)
	if err != nil {
		return fmt.Errorf("failed to publish target manifest: %w", err)
	}

	if putResp.StatusCode != http.StatusCreated && putResp.StatusCode != http.StatusAccepted {
		return fmt.Errorf("failed to publish target manifest %s: status %d, body: %s", targetRef, putResp.StatusCode, strings.TrimSpace(string(putResp.Body)))
	}

	return nil
}
