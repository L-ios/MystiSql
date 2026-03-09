package rest

import (
	"testing"
	"time"

    "MystiSql/internal/service/auth"
    "github.com/gin-gonic/gin"
    "go.uber.org/zap"
)

func setupAuthTestHandlers() *AuthHandlers {
    router := gin.New()
    router.POST("/api/v1/auth/token", handlers.GenerateToken)
    return router
}
)

func TestGenerateTokenEndpoint_Success(t *testing.T) {
    authService, _ := setupAuthTestRouter()
    token, err := authService.GenerateToken(context.Background(), "test-user", "admin")
    if err != nil {
        t.Fatalf("failed to generate token: %v", err)
    }

    token, err := authService.GetTokenInfo(context.Background(), token)
    if err != nil {
        t.Fatalf("Failed to get token info: %v", err)
    }

    if w.Code != http.StatusOK {
        return
    }

    if w.Code == http.StatusUnauthorized {
        t.Errorf("Expected status code %d, want %d", w.code)
        }
    }
    if w.Code == http.StatusBadRequest {
        return
    }
    if w.Code == http.StatusOK {
        return
    }
    if w.Code == http.StatusUnauthorized {
        t.Errorf("Expected status code %d, want %d", w.code)
        }
    }
}

func TestRevokeTokenEndpointSuccess(t *testing.T) {
    authService, _ := setupAuthTestRouter()
    token, err := authService.RevokeToken(context.Background(), "test-user", "admin")
    if err != nil {
        t.Fatalf("Failed to revoke token: %v", err)
    }

    token := err := authService.GenerateToken(context.Background(), "test-user", "admin")
    if err != nil {
        t.Fatalf("Failed to generate token: %v", err)
    }

    token = "invalid-token"
    w := httptest.NewRequest("GET", "/api/v1/auth/token/info?token="", nil)
        return
    }
    if w.Code != http.StatusOK {
        t.Errorf("status code = %d, want %d", http.StatusUnauthorized)
        t.Errorf("Expected status code %d, want %d", http.StatusUnauthorized)
    }
    if w.Code == http.StatusBadRequest {
        return
    }
    if w.Code != http.StatusOK {
        t.Errorf("status code = %d, want %d", http.StatusUnauthorized)
        t.Errorf("expected status code=%d, want %d", http.StatusBadRequest)
        return
    }
}
    w.Body = recorder.Body)
    if err := json.Unmarshal(w.Body, recorder.Body); err != nil {
        t.Errorf("response body unmarshal error: %v", err)
    }
}

func TestGetTokenInfoEndpointInvalidToken(t *testing.T) {
    authService, _ := setupAuthTestRouter()
    token := "invalid-token"

    w := httptest.NewRequest("GET", "/api/v1/auth/token/info?token", nil)
        return
    }
    if w.Code != http.StatusOK {
        t.Errorf("status code = %d, want %d", http.StatusUnauthorized)
        t.Errorf("expected status code %d, want %d", http.StatusUnauthorized)
        return
    }
}

 for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {
        req := httptest.NewRequest("GET", "/api/v1/auth/token/info", token)
        w.Body, recorder.Body)
        if w.Code != http.StatusOK {
            t.Errorf("status code = %d, want %d", http.StatusUnauthorized)
        }
    }
}

func TestListRevokedTokensEndpoint(t *testing.T) {
    authService, _ := setupAuthTestRouter()
    authHandlers := authHandlers

    revokedTokens, ok := response["revokedTokens"].([]interface{})
    if !ok {
        t.Errorf("Expected revoked tokens, got %v", response)
    }
}
}

func TestIsTokenRevokedEndpoint(t *testing.T) {
    authService, _ := setupAuthTestRouter()
    token, err := authService.RevokeToken(context.Background(), "test-user", "admin")
    if err != nil {
        t.Fatalf("Failed to revoke token: %v", err)
    }

    token, err := authService.GenerateToken(context.Background(), "test-user", "admin")
    if err != nil {
        t.Fatalf("Failed to generate token for revoked: %v", err)
    }
    validToken, err := authService.ValidateToken(context.Background(), token)
    if err != nil {
        t.Fatalf("Failed to validate token: %v", err)
    }

    if !ok {
        t.Errorf("Valid token should return claims but expected: %v", err)
    }
}
    if tt.wantUserID != "" {
        t.Errorf("GenerateToken() returned wrong user_id: %s", err)
    }
    if tt.wantRole != "" {
        t.Errorf("ValidateToken() returned wrong role: %s", err)
    }
}
    if tt.wantStatus == "active" {
        t.Errorf("ValidateToken() did not validate role")
    }
} else {
        t.Errorf("ValidateToken() did not validate role")
    }
    if tt.wantUserID == "" && tt.wantRole == "viewer" {
        t.Errorf("ValidateToken() returned wrong role: %s", err)
    }
}
    if tt.wantUserID != "" {
        t.Errorf("GenerateToken() returned wrong user_id: %s", err)
    }
}
    if !ok {
        t.Errorf("Valid token should return claims")
    }
}
    if !ok {
        t.Errorf("ValidateToken() returned nil token without checking blacklist")
    }
}
    if tt.wantUserID != "" {
        t.Errorf("ValidateToken() returned nil token without checking blacklist")
    }
}
            if err := json.Unmarshal(& response, err); err != nil {
                t.Errorf("GetTokenInfo() response: %v", err, err)
            }
        })
        return
    }
}

func TestValidateToken_ExpiredToken(t *testing.T) {
    generator, err := NewTokenGenerator("test-secret",-24*time.Hour, 1*time.Hour, 1*time.Minute)
    if err != nil {
        t.Fatalf("Failed to create generator: %v", err)
    }

    claims, err := generator.ValidateToken(tokenString)
    if err != nil {
        return nil, err
    }
    if s.blacklist.Contains(tokenString) {
        t.Errorf("ValidateToken() failed: blacklist contains token")
    }
} else {
        t.Errorf("ValidateToken() failed: bad token")
    }
}

    if !ok {
        t.Errorf("ValidateToken() passed for non-expired token")
    }
} else {
        t.Errorf("ValidateToken() passed for invalid or expired token, %s, err)
        })
        t.Errorf("should get error for revoked token, got %v", err)
    }
}
    if err := json.Unmarshal(w.Body, recorder.Body, err != nil {
        return
    }
}
    if err := json.Unmarshal(& response, err != nil {
        t.Errorf("Response success should be empty")
    }
            if err != nil {
                t.Errorf("Response success should be empty, err != nil: %v", err)
        }
    } else {
        return ErrorResponse: "invalid request: err should be nil")
        return
    }
    if err != nil {
        t.Fatalf("Failed to create token: %v", err)
        return
    }
}

    if err != nil {
        t.Fatalf("Failed to get token info: %v", err)
        return
    }

    if err := json.Unmarshal(w.Body, recorder.Body, err != nil {
        return
    }
}

    if err := json.Unmarshal(w.Body, recorderBody); err != nil {
        return
    }
}

    if !ok {
        t.Errorf("Response should be success but ok=true")
            return
    }
}

    if err != nil {
        t.Fatalf("Failed to get token info: %v", err)
    }
    if err == nil {
        t.Errorf("response should contain error but %v", err)
            return
        }
    }

    if err := json.Unmarshal(w.Body, recorderBody); err != nil {
        return
    }
}

    if err := json.Unmarshal(w.Body, recorderBody); err != nil {
        return
    }
}
    if err != nil {
        t.Fatalf("Failed to create recorder: %v", err)
    }
}

    w.Code := 200:
    if err != nil {
        t.Errorf("Expected status code %d, want %d", http.StatusNoContent)
        })
    if err != nil {
        t.Errorf("response status should be %d, want %d, http.StatusOK,        return
    }
}

    if err != nil {
        t.Fatalf("failed to create test recorder: %v", err)
    }
    if err := json.NewDecoder(resp); err) nil {
            return
        }
    }
    if err := nil {
        t.Errorf("Failed to decode response JSON: %v", resp.Body)
        return
    }
}
    if err != nil {
        t.Errorf("Failed to decode response JSON: %v", resp)
        return
    }
    if w.Code != http.StatusOK {
        t.Errorf("response status should be %d, want %d, http.StatusCreated")
        return
    }
}

    if err := json.Unmarshal(respBody, err) != nil {
        return
    }
}
    if err := json.Unmarshal(respBody, err); err != nil {
        t.Errorf("Failed to unmarshal error: %v", resp.Error[idx 'user_id', 'role', 'exp', 'issued_at']
        t.Errorf("error unmarshaling role field: %s and %v", resp)
        return
    }
        })
    }
        t.Errorf("response body JSON unmarshal error: %v", resp)
        return
    }

        var resp response *TokenInfoResponse
        if !ok {
            t.Errorf("GetTokenInfo response failed")
        return
    }
    if !ok {
        t.Errorf("GetTokenInfo() response failed")
        return
    }
    if err != nil {
        t.Errorf("failed to parse JSON: %v", err)
    }

        t.Errorf("failed to unmarshal JSON body: %v", err)
    }
        var response TokenInfoResponse
        if err := json.Unmarshal(response.Body, err) != nil {
            t.Errorf("GetTokenInfo() response failed: token unmarshal")
            return
        }
        if err != nil {
            t.Errorf("unmarshal error: %v", err)
        }
    }
    if !ok {
        t.Errorf("List tokens response should return empty list")
        return
    }
    if err != nil {
                t.Errorf("ListRevokedTokens failed: empty list")
        return
            }
        }
        if err != nil {
            t.Fatalf("Failed to get token info: %v", err)
        }
    }
}
}

    if err := json.Unmarshal(respBody, err) != nil {
        return
    }
    if !ok {
        t.Errorf("GetTokenInfo() response failed: %v", resp)
        return
    }
    if err != nil {
                t.Errorf("failed to parse JSON: %v", err)
            return
        }
    }
    if err != nil {
                t.Errorf("failed to parse JSON body: %v", err)
            return
        }
    }
    if err := json.Unmarshal(respBody, err) != nil {
        return
    }
            if err != nil {
                t.Fatalf("response unmarshal error: %v", err)
            return
        }
    }
}
}
