package oidc

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"golang.org/x/oauth2"
)

// OIDCFlow coordinates the Authorization Code flow for configured OIDC clients.
type OIDCFlow struct {
	mu       sync.RWMutex
	clients  map[string]*OIDCClient
	stateMap map[string]string // state -> provider name
}

func NewOIDCFlow(clients map[string]*OIDCClient) *OIDCFlow {
	return &OIDCFlow{
		clients:  clients,
		stateMap: make(map[string]string),
	}
}

// Login redirects the user to the Identity Provider to begin the flow.
func (f *OIDCFlow) Login(providerName string, w http.ResponseWriter, r *http.Request) {
	f.mu.Lock()
	defer f.mu.Unlock()
	client, ok := f.clients[providerName]
	if !ok {
		http.Error(w, "Unknown OIDC provider", http.StatusBadRequest)
		return
	}
	state := randomState()
	f.stateMap[state] = providerName
	url := client.OAuthConfig.AuthCodeURL(state, oauth2.AccessTypeOffline)
	http.Redirect(w, r, url, http.StatusFound)
}

// Callback handles the IdP callback after user authentication.
func (f *OIDCFlow) Callback(w http.ResponseWriter, r *http.Request) {
	f.mu.Lock()
	defer f.mu.Unlock()
	state := r.URL.Query().Get("state")
	code := r.URL.Query().Get("code")
	providerName, ok := f.stateMap[state]
	if !ok {
		http.Error(w, "Invalid state", http.StatusBadRequest)
		return
	}
	client, ok := f.clients[providerName]
	if !ok {
		http.Error(w, "Unknown provider", http.StatusBadRequest)
		return
	}
	// Exchange code for tokens
	token, err := client.OAuthConfig.Exchange(context.Background(), code)
	if err != nil {
		http.Error(w, "Token exchange failed", http.StatusInternalServerError)
		return
	}
	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		http.Error(w, "No id_token in token response", http.StatusInternalServerError)
		return
	}
	// Verify the ID Token
	idToken, err := client.Verifier.Verify(r.Context(), rawIDToken)
	if err != nil {
		http.Error(w, "Invalid ID Token", http.StatusUnauthorized)
		return
	}
	// Extract claims as a generic map to support role mapping
	claims := map[string]interface{}{}
	if err := idToken.Claims(&claims); err != nil {
		http.Error(w, "Failed to parse claims", http.StatusInternalServerError)
		return
	}

	// Basic user identity for demonstration purposes
	user := map[string]interface{}{
		"provider": providerName,
		"claims":   claims,
		"token":    rawIDToken,
		"expires":  time.Now().Add(5 * time.Minute).Unix(),
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

// randomState generates a secure random string for the OAuth2 state parameter.
func randomState() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		// Fallback to time-based value on failure
		return base64.RawURLEncoding.EncodeToString([]byte(time.Now().String()))
	}
	return base64.RawURLEncoding.EncodeToString(b)
}
