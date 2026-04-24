package publish

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
)

type registryResponse struct {
	StatusCode int
	Headers    http.Header
	Body       []byte
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
