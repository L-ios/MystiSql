package query

import (
	"context"
)

type contextKey string

const (
	userIDKey   contextKey = "user_id"
	clientIPKey contextKey = "client_ip"
	roleKey     contextKey = "role"
)

func getUserIDFromContext(ctx context.Context) string {
	if userID, ok := ctx.Value(userIDKey).(string); ok {
		return userID
	}
	return "unknown"
}

func getClientIPFromContext(ctx context.Context) string {
	if clientIP, ok := ctx.Value(clientIPKey).(string); ok {
		return clientIP
	}
	return "unknown"
}

func ContextWithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}

func ContextWithClientIP(ctx context.Context, clientIP string) context.Context {
	return context.WithValue(ctx, clientIPKey, clientIP)
}

func getRoleFromContext(ctx context.Context) string {
	if role, ok := ctx.Value(roleKey).(string); ok {
		return role
	}
	return ""
}

func ContextWithRole(ctx context.Context, role string) context.Context {
	return context.WithValue(ctx, roleKey, role)
}
