package publish

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

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
