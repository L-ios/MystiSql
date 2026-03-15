package rest

import (
	"encoding/json"
	"net/http"
)

// OIDCLoginHandler initiates the OIDC login flow. In this minimal implementation,
// we return a 501 Not Implemented response. A full implementation would redirect
// to the IdP's authorization endpoint using a shared OIDC flow registry.
func OIDCLoginHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotImplemented)
	json.NewEncoder(w).Encode(map[string]string{
		"error": "OIDC login not implemented in this build.",
	})
}

// OIDCCallbackHandler handles the IdP callback after user authentication.
// This is a placeholder implementation that returns 501 status.
func OIDCCallbackHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotImplemented)
	json.NewEncoder(w).Encode(map[string]string{
		"error": "OIDC callback not implemented in this build.",
	})
}
