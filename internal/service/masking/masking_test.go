package masking

import "testing"

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

func TestPolicyStore(t *testing.T) {
	ps := NewPolicyStore()

	policy := ps.Get("nonexistent")
	if policy.MaskPhone || policy.MaskEmail || policy.MaskIDCard || policy.MaskBankCard {
		t.Error("default policy should have all masking disabled")
	}

	ps.Set("viewer", MaskPolicy{MaskPhone: true, MaskEmail: true})
	got := ps.Get("viewer")
	if !got.MaskPhone || !got.MaskEmail {
		t.Error("viewer policy should have phone and email masking enabled")
	}
	if got.MaskIDCard || got.MaskBankCard {
		t.Error("viewer policy should NOT have IDCard and BankCard masking")
	}
}

func TestPolicyStore_Nil(t *testing.T) {
	var ps *PolicyStore
	policy := ps.Get("any")
	if policy.MaskPhone || policy.MaskEmail {
		t.Error("nil PolicyStore should return zero policy")
	}
}

func TestMaskingService_MaskForRole(t *testing.T) {
	ps := NewPolicyStore()
	ps.Set("viewer", MaskPolicy{MaskPhone: true, MaskEmail: true})
	ps.Set("admin", MaskPolicy{})

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

func TestNewPolicyStore(t *testing.T) {
	ps := NewPolicyStore()
	if ps.Policies == nil {
		t.Error("Policies map should be initialized")
	}
}
