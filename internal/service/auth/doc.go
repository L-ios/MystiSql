// Package auth provides JWT token-based authentication services.
//
// This package implements token generation, validation, and revocation
// using JWT (JSON Web Tokens) with HS256 signing algorithm.
//
// Key Features:
// - Token generation with user_id, role, and expiration
// - Token validation and revocation
// - In-memory token blacklist management
// - Thread-safe operations
package auth
