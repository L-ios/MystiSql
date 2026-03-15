package oidc

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
)

// FetchUserInfo calls the OIDC provider's UserInfo endpoint using the provided
// access token. It returns the raw user info as a generic map.
func (c *OIDCClient) FetchUserInfo(ctx context.Context, accessToken string) (map[string]interface{}, error) {
	if c.UserInfoURL == "" {
		// Some providers do not expose a UserInfo endpoint. Treat as not available.
		return nil, nil
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.UserInfoURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var info map[string]interface{}
	dec := json.NewDecoder(resp.Body)
	dec.UseNumber()
	if err := dec.Decode(&info); err != nil {
		return nil, err
	}
	// Normalize a timestamp to a consistent format (optional for demo)
	_ = time.Now()
	return info, nil
}
