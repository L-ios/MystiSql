// Package audit provides SQL audit logging functionality.
//
// The audit package records all SQL execution operations including:
// - Query execution (SELECT, INSERT, UPDATE, DELETE)
// - DDL operations (CREATE, ALTER, DROP)
// - Execution metadata (user, instance, time, rows affected)
//
// Audit logs are written asynchronously to JSON Lines format files
// with automatic rotation and retention management.
package audit
