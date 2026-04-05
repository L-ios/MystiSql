package masking

import (
	"testing"

	"MystiSql/pkg/types"
)

func TestMaskPhone(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"13812345678", "138******78"},
		{"1234", "****"},
		{"12", "**"},
		{"", ""},
		{"138123456", "138****56"},
		{"12345", "12345"},
	}
	for _, tt := range tests {
		got := maskPhone(tt.input)
		if got != tt.want {
			t.Errorf("maskPhone(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestMaskEmail(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"user@example.com", "u***@example.com"},
		{"a@b.com", "*@b.com"},
		{"@onlydomain.com", "@onlydomain.com"},
		{"nouser", "******"},
		{"ab@example.com", "a*@example.com"},
	}
	for _, tt := range tests {
		got := maskEmail(tt.input)
		if got != tt.want {
			t.Errorf("maskEmail(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestMaskIDCard(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"110101199001011234", "11**************34"},
		{"1234", "****"},
		{"12345", "12*45"},
		{"123456", "12**56"},
		{"", ""},
	}
	for _, tt := range tests {
		got := maskIDCard(tt.input)
		if got != tt.want {
			t.Errorf("maskIDCard(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestMaskBankCard(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"6222021234567890123", "***************0123"},
		{"1234", "****"},
		{"12345", "*2345"},
		{"", ""},
	}
	for _, tt := range tests {
		got := maskBankCard(tt.input)
		if got != tt.want {
			t.Errorf("maskBankCard(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestPolicyStore_Get_Nonexistent(t *testing.T) {
	ps := NewPolicyStore()
	policy := ps.Get("nonexistent")
	if policy != nil {
		t.Error("expected nil for nonexistent role")
	}
}

func TestPolicyStore_SetAndGet(t *testing.T) {
	ps := NewPolicyStore()
	ps.Set("viewer", &RoleRules{
		Rules: []ColumnRule{
			{Pattern: "phone", MaskType: MaskTypePhone},
			{Pattern: "email", MaskType: MaskTypeEmail},
		},
	})

	got := ps.Get("viewer")
	if got == nil {
		t.Fatal("expected non-nil rules for viewer")
	}
	if len(got.Rules) != 2 {
		t.Fatalf("expected 2 rules, got %d", len(got.Rules))
	}
	if got.Rules[0].MaskType != MaskTypePhone {
		t.Error("first rule should be phone masking")
	}
	if got.Rules[1].MaskType != MaskTypeEmail {
		t.Error("second rule should be email masking")
	}
}

func TestPolicyStore_Nil(t *testing.T) {
	var ps *PolicyStore
	policy := ps.Get("any")
	if policy != nil {
		t.Error("nil PolicyStore should return nil")
	}
}

func TestPolicyStore_AddRule(t *testing.T) {
	ps := NewPolicyStore()
	ps.AddRule("viewer", ColumnRule{Pattern: "phone", MaskType: MaskTypePhone})
	ps.AddRule("viewer", ColumnRule{Pattern: "email", MaskType: MaskTypeEmail})

	got := ps.Get("viewer")
	if got == nil {
		t.Fatal("expected non-nil rules for viewer")
	}
	if len(got.Rules) != 2 {
		t.Fatalf("expected 2 rules, got %d", len(got.Rules))
	}
}

func TestPolicyStore_RemoveRule(t *testing.T) {
	ps := NewPolicyStore()
	ps.AddRule("viewer", ColumnRule{Pattern: "phone", MaskType: MaskTypePhone})
	ps.AddRule("viewer", ColumnRule{Pattern: "email", MaskType: MaskTypeEmail})

	ps.RemoveRule("viewer", "phone")

	got := ps.Get("viewer")
	if got == nil {
		t.Fatal("expected non-nil rules for viewer")
	}
	if len(got.Rules) != 1 {
		t.Fatalf("expected 1 rule after removal, got %d", len(got.Rules))
	}
	if got.Rules[0].Pattern != "email" {
		t.Error("remaining rule should be email")
	}
}

func TestNewPolicyStore(t *testing.T) {
	ps := NewPolicyStore()
	if ps == nil {
		t.Fatal("expected non-nil PolicyStore")
	}
}

func TestMaskingService_MaskForRole(t *testing.T) {
	ps := NewPolicyStore()
	ps.AddRule("viewer", ColumnRule{Pattern: "phone", MaskType: MaskTypePhone})
	ps.AddRule("viewer", ColumnRule{Pattern: "email", MaskType: MaskTypeEmail})

	svc := NewMaskingService(ps)

	data := MaskData{
		Phone:    "13812345678",
		Email:    "user@example.com",
		IDCard:   "110101199001011234",
		BankCard: "6222021234567890123",
	}

	result := svc.MaskForRole("viewer", data)
	if result.Phone == "13812345678" {
		t.Error("phone should be masked for viewer")
	}
	if result.Email == "user@example.com" {
		t.Error("email should be masked for viewer")
	}
	if result.IDCard != "110101199001011234" {
		t.Error("IDCard should NOT be masked for viewer")
	}
	if result.BankCard != "6222021234567890123" {
		t.Error("BankCard should NOT be masked for viewer")
	}

	result = svc.MaskForRole("admin", data)
	if result.Phone != "13812345678" {
		t.Error("phone should NOT be masked for admin")
	}
	if result.Email != "user@example.com" {
		t.Error("email should NOT be masked for admin")
	}
}

func TestMaskingService_NoPolicy(t *testing.T) {
	ps := NewPolicyStore()
	svc := NewMaskingService(ps)

	data := MaskData{Phone: "13812345678", Email: "user@example.com"}
	result := svc.MaskForRole("unknown", data)

	if result.Phone != "13812345678" {
		t.Error("should not mask when no policy exists")
	}
}

func TestMaskResult_BasicColumnMatching(t *testing.T) {
	ps := NewPolicyStore()
	ps.AddRule("viewer", ColumnRule{Pattern: "phone", MaskType: MaskTypePhone})
	ps.AddRule("viewer", ColumnRule{Pattern: "email", MaskType: MaskTypeEmail})
	ps.AddRule("viewer", ColumnRule{Pattern: "id_card", MaskType: MaskTypeIDCard})

	svc := NewMaskingService(ps)

	result := &types.QueryResult{
		Columns: []types.ColumnInfo{
			{Name: "name", Type: "varchar"},
			{Name: "phone", Type: "varchar"},
			{Name: "email", Type: "varchar"},
			{Name: "id_card", Type: "varchar"},
		},
		Rows: []types.Row{
			{"Alice", "13812345678", "user@example.com", "110101199001011234"},
			{"Bob", "13987654321", "bob@test.com", "220202199002022345"},
		},
		RowCount: 2,
	}

	masked := svc.MaskResult("viewer", result)

	if masked.Rows[0][0] != "Alice" {
		t.Error("name should not be masked")
	}
	if masked.Rows[0][1] == "13812345678" {
		t.Error("phone should be masked")
	}
	if masked.Rows[0][2] == "user@example.com" {
		t.Error("email should be masked")
	}
	if masked.Rows[0][3] == "110101199001011234" {
		t.Error("id_card should be masked")
	}

	if result.Rows[0][1] != "13812345678" {
		t.Error("original result should not be modified")
	}
}

func TestMaskResult_PatternWithQualifier(t *testing.T) {
	ps := NewPolicyStore()
	ps.AddRule("viewer", ColumnRule{Pattern: "*.users.phone", MaskType: MaskTypePhone})

	svc := NewMaskingService(ps)

	result := &types.QueryResult{
		Columns: []types.ColumnInfo{
			{Name: "phone", Type: "varchar"},
			{Name: "address", Type: "varchar"},
		},
		Rows: []types.Row{
			{"13812345678", "Beijing"},
		},
	}

	masked := svc.MaskResult("viewer", result)
	if masked.Rows[0][0] == "13812345678" {
		t.Error("phone column should be masked via qualified pattern")
	}
	if masked.Rows[0][1] != "Beijing" {
		t.Error("address should not be masked")
	}
}

func TestMaskResult_NoRulesForRole(t *testing.T) {
	ps := NewPolicyStore()
	svc := NewMaskingService(ps)

	result := &types.QueryResult{
		Columns: []types.ColumnInfo{{Name: "phone", Type: "varchar"}},
		Rows:    []types.Row{{"13812345678"}},
	}

	masked := svc.MaskResult("unknown", result)
	if masked.Rows[0][0] != "13812345678" {
		t.Error("should not mask when no rules for role")
	}
}

func TestMaskResult_NilResult(t *testing.T) {
	ps := NewPolicyStore()
	ps.AddRule("viewer", ColumnRule{Pattern: "phone", MaskType: MaskTypePhone})
	svc := NewMaskingService(ps)

	masked := svc.MaskResult("viewer", nil)
	if masked != nil {
		t.Error("nil input should return nil")
	}
}

func TestMaskResult_EmptyRows(t *testing.T) {
	ps := NewPolicyStore()
	ps.AddRule("viewer", ColumnRule{Pattern: "phone", MaskType: MaskTypePhone})
	svc := NewMaskingService(ps)

	result := &types.QueryResult{
		Columns: []types.ColumnInfo{{Name: "phone", Type: "varchar"}},
		Rows:    []types.Row{},
	}

	masked := svc.MaskResult("viewer", result)
	if masked != result {
		t.Error("empty rows should return same result")
	}
}

func TestMaskResult_NonStringValues(t *testing.T) {
	ps := NewPolicyStore()
	ps.AddRule("viewer", ColumnRule{Pattern: "phone", MaskType: MaskTypePhone})
	svc := NewMaskingService(ps)

	result := &types.QueryResult{
		Columns: []types.ColumnInfo{
			{Name: "phone", Type: "varchar"},
			{Name: "age", Type: "int"},
		},
		Rows: []types.Row{
			{13812345678, 25},
			{nil, 30},
		},
	}

	masked := svc.MaskResult("viewer", result)

	phoneVal, ok := masked.Rows[0][0].(string)
	if !ok || phoneVal == "13812345678" {
		t.Error("non-string phone should be converted and masked")
	}

	if masked.Rows[1][0] != nil {
		t.Error("nil phone should remain nil")
	}

	if masked.Rows[0][1] != 25 {
		t.Error("age should not be modified")
	}
}

func TestColumnRule_MatchColumn(t *testing.T) {
	tests := []struct {
		pattern     string
		columnName  string
		shouldMatch bool
	}{
		{"phone", "phone", true},
		{"phone", "Phone", true},
		{"phone", "email", false},
		{"*.users.phone", "phone", true},
		{"*.users.phone", "email", false},
		{"instance.users.phone", "phone", true},
		{"", "phone", false},
		{"*", "anything", true},
	}

	for _, tt := range tests {
		rule := ColumnRule{Pattern: tt.pattern}
		got := rule.matchColumn(tt.columnName)
		if got != tt.shouldMatch {
			t.Errorf("matchColumn(%q vs %q) = %v, want %v", tt.pattern, tt.columnName, got, tt.shouldMatch)
		}
	}
}

func TestResolveMaskFunc(t *testing.T) {
	if resolveMaskFunc(MaskTypePhone) == nil {
		t.Error("phone mask func should not be nil")
	}
	if resolveMaskFunc(MaskTypeEmail) == nil {
		t.Error("email mask func should not be nil")
	}
	if resolveMaskFunc(MaskTypeIDCard) == nil {
		t.Error("idcard mask func should not be nil")
	}
	if resolveMaskFunc(MaskTypeBankCard) == nil {
		t.Error("bankcard mask func should not be nil")
	}
	if resolveMaskFunc(MaskTypeCustom) != nil {
		t.Error("custom mask func should be nil (not implemented)")
	}
	if resolveMaskFunc("unknown") != nil {
		t.Error("unknown mask type should return nil")
	}
}
