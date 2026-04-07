package router

import (
	"testing"
)

func TestCategorizeSQL(t *testing.T) {
	tests := []struct {
		name     string
		sql      string
		expected SQLCategory
	}{
		{"SELECT", "SELECT * FROM users", SQLCategoryRead},
		{"SELECT with WHERE", "SELECT id, name FROM users WHERE id = 1", SQLCategoryRead},
		{"INSERT", "INSERT INTO users (name) VALUES ('Alice')", SQLCategoryWrite},
		{"UPDATE", "UPDATE users SET name = 'Bob' WHERE id = 1", SQLCategoryWrite},
		{"DELETE", "DELETE FROM users WHERE id = 1", SQLCategoryWrite},
		{"SHOW", "SHOW DATABASES", SQLCategoryRead},
		{"EXPLAIN", "EXPLAIN SELECT * FROM users", SQLCategoryRead},
		{"DESC", "DESC users", SQLCategoryRead},
		{"CREATE", "CREATE TABLE test (id INT)", SQLCategoryWrite},
		{"DROP", "DROP TABLE test", SQLCategoryWrite},
		{"TRUNCATE", "TRUNCATE TABLE test", SQLCategoryWrite},
		{"empty", "", SQLCategoryRead},
		{"lowercase select", "select * from users", SQLCategoryRead},
		{"lowercase insert", "insert into users values (1)", SQLCategoryWrite},
	}

	router := &ReadWriteRouter{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := router.categorizeSQL(tt.sql)
			if result != tt.expected {
				t.Errorf("categorizeSQL(%q) = %v, want %v", tt.sql, result, tt.expected)
			}
		})
	}
}

func TestIsSelectForUpdate(t *testing.T) {
	tests := []struct {
		name     string
		sql      string
		expected bool
	}{
		{"FOR UPDATE", "SELECT * FROM users WHERE id = 1 FOR UPDATE", true},
		{"for update lowercase", "select * from users for update", true},
		{"regular SELECT", "SELECT * FROM users", false},
		{"SELECT with LOCK IN SHARE MODE", "SELECT * FROM users LOCK IN SHARE MODE", false},
	}

	router := &ReadWriteRouter{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := router.isSelectForUpdate(tt.sql)
			if result != tt.expected {
				t.Errorf("isSelectForUpdate(%q) = %v, want %v", tt.sql, result, tt.expected)
			}
		})
	}
}

func TestIsSpecialFunction(t *testing.T) {
	tests := []struct {
		name     string
		sql      string
		expected bool
	}{
		{"LAST_INSERT_ID", "SELECT LAST_INSERT_ID()", true},
		{"last_insert_id lowercase", "select last_insert_id()", true},
		{"@@IDENTITY", "SELECT @@IDENTITY", true},
		{"regular SELECT", "SELECT * FROM users", false},
		{"COUNT", "SELECT COUNT(*) FROM users", false},
	}

	router := &ReadWriteRouter{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := router.isSpecialFunction(tt.sql)
			if result != tt.expected {
				t.Errorf("isSpecialFunction(%q) = %v, want %v", tt.sql, result, tt.expected)
			}
		})
	}
}
