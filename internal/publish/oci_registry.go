package publish

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type ociReference struct {
	Registry   string
	Repository string
	Reference  string
}

type registryResponse struct {
	StatusCode int
	Headers    http.Header
	Body       []byte
}

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

func doRegistryRequest(client *http.Client, method, endpoint string, body []byte, headers map[string]string, username, password string) (*registryResponse, error) {
	authHeader := ""
	basicHeader := ""
	if username != "" || password != "" {
		basicHeader = "Basic " + base64.StdEncoding.EncodeToString([]byte(username+":"+password))
	}

	for attempt := 0; attempt < 3; attempt++ {
		resp, err := doRawHTTP(client, method, endpoint, body, headers, authHeader, basicHeader)
		if err != nil {
			return nil, err
		}

		if resp.StatusCode != http.StatusUnauthorized {
			return resp, nil
		}

		challenge := resp.Headers.Get("WWW-Authenticate")
		realm, service, scope, ok := parseBearerChallenge(challenge)
		if !ok {
			if basicHeader == "" || authHeader == basicHeader {
				return resp, nil
			}
			authHeader = basicHeader
			continue
		}

		token, tokenErr := fetchBearerToken(client, realm, service, scope, username, password)
		if tokenErr != nil {
			return nil, tokenErr
		}
		authHeader = "Bearer " + token
	}

	return nil, fmt.Errorf("request %s %s failed after retries", method, endpoint)
}

func doRawHTTP(client *http.Client, method, endpoint string, body []byte, headers map[string]string, authHeader, basicHeader string) (*registryResponse, error) {
	var reader io.Reader
	if body != nil {
		reader = bytes.NewReader(body)
	}

	req, err := http.NewRequest(method, endpoint, reader)
	if err != nil {
		return nil, err
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	if authHeader != "" {
		req.Header.Set("Authorization", authHeader)
	} else if basicHeader != "" {
		req.Header.Set("Authorization", basicHeader)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return &registryResponse{StatusCode: resp.StatusCode, Headers: resp.Header.Clone(), Body: respBody}, nil
}

func fetchBearerToken(client *http.Client, realm, service, scope, username, password string) (string, error) {
	parsed, err := url.Parse(realm)
	if err != nil {
		return "", fmt.Errorf("invalid bearer realm %q: %w", realm, err)
	}

	query := parsed.Query()
	if service != "" {
		query.Set("service", service)
	}
	if scope != "" {
		query.Set("scope", scope)
	}
	parsed.RawQuery = query.Encode()

	req, err := http.NewRequest("GET", parsed.String(), nil)
	if err != nil {
		return "", err
	}
	if username != "" || password != "" {
		req.SetBasicAuth(username, password)
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("token request failed: status %d, body: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var tokenResp struct {
		Token       string `json:"token"`
		AccessToken string `json:"access_token"`
	}
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return "", fmt.Errorf("failed to parse token response: %w", err)
	}

	if tokenResp.Token != "" {
		return tokenResp.Token, nil
	}
	if tokenResp.AccessToken != "" {
		return tokenResp.AccessToken, nil
	}

	return "", fmt.Errorf("token response missing token field")
}

func parseBearerChallenge(challenge string) (string, string, string, bool) {
	trimmed := strings.TrimSpace(challenge)
	if !strings.HasPrefix(strings.ToLower(trimmed), "bearer ") {
		return "", "", "", false
	}

	params := parseChallengeParams(trimmed[len("Bearer "):])
	realm := params["realm"]
	if realm == "" {
		return "", "", "", false
	}

	return realm, params["service"], params["scope"], true
}

func parseChallengeParams(raw string) map[string]string {
	params := map[string]string{}
	for _, part := range splitAuthParams(raw) {
		kv := strings.SplitN(strings.TrimSpace(part), "=", 2)
		if len(kv) != 2 {
			continue
		}
		key := strings.ToLower(strings.TrimSpace(kv[0]))
		value := strings.Trim(strings.TrimSpace(kv[1]), `"`)
		params[key] = value
	}
	return params
}

func splitAuthParams(raw string) []string {
	parts := make([]string, 0)
	start := 0
	inQuotes := false

	for i, ch := range raw {
		switch ch {
		case '"':
			inQuotes = !inQuotes
		case ',':
			if inQuotes {
				continue
			}
			parts = append(parts, strings.TrimSpace(raw[start:i]))
			start = i + 1
		}
	}

	if start < len(raw) {
		parts = append(parts, strings.TrimSpace(raw[start:]))
	}

	return parts
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
