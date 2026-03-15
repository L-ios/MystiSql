package router

import "testing"

func TestParseSQL(t *testing.T) {
	tests := []struct {
		sql       string
		wantType  SQLType
		wantInTxn bool
		wantIsTxn bool
	}{
		{"SELECT * FROM users", SQLTypeSelect, false, false},
		{"INSERT INTO t (a) VALUES (1)", SQLTypeInsert, false, false},
		{"UPDATE t SET a = 1", SQLTypeUpdate, false, false},
		{"DELETE FROM t WHERE id = 1", SQLTypeDelete, false, false},
		{"BEGIN", SQLTypeUnknown, true, true},
		{"START TRANSACTION", SQLTypeUnknown, true, true},
		{"COMMIT", SQLTypeUnknown, true, true},
		{"ROLLBACK", SQLTypeUnknown, true, true},
		{"SELECT * FROM t; BEGIN; UPDATE t SET a=2; COMMIT", SQLTypeSelect, true, false},
	}

	for _, tt := range tests {
		gotType, gotInTxn, gotIsTxn := ParseSQL(tt.sql)
		if gotType != tt.wantType || gotInTxn != tt.wantInTxn || gotIsTxn != tt.wantIsTxn {
			t.Fatalf("ParseSQL(%q) = (%v,%v,%v), want (%v,%v,%v)", tt.sql, gotType, gotInTxn, gotIsTxn, tt.wantType, tt.wantInTxn, tt.wantIsTxn)
		}
	}
}

func TestIsTransaction(t *testing.T) {
	tests := []struct {
		sql  string
		want bool
	}{
		{"BEGIN", true},
		{"START TRANSACTION", true},
		{"COMMIT", true},
		{"ROLLBACK", true},
		{"SELECT 1", false},
		{"SELECT * FROM t; BEGIN", true},
	}
	for _, tt := range tests {
		got := IsTransaction(tt.sql)
		if got != tt.want {
			t.Fatalf("IsTransaction(%q) = %v, want %v", tt.sql, got, tt.want)
		}
	}
}
