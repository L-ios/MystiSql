package masking

import (
	"strings"
)

// MaskData contains fields that may be masked according to policy.
type MaskData struct {
	Phone    string
	Email    string
	IDCard   string
	BankCard string
}

// MaskingService applies masking rules based on a role-based policy store.
type MaskingService struct {
	policyStore *PolicyStore
}

// NewMaskingService creates a masking service with the given policy store.
func NewMaskingService(ps *PolicyStore) *MaskingService {
	return &MaskingService{policyStore: ps}
}

// MaskForRole applies masking rules for the provided role to the data.
func (m *MaskingService) MaskForRole(role string, data MaskData) MaskData {
	policy := m.policyStore.Get(role)
	if policy.MaskPhone {
		data.Phone = maskPhone(data.Phone)
	}
	if policy.MaskEmail {
		data.Email = maskEmail(data.Email)
	}
	if policy.MaskIDCard {
		data.IDCard = maskIDCard(data.IDCard)
	}
	if policy.MaskBankCard {
		data.BankCard = maskBankCard(data.BankCard)
	}
	return data
}

// Internal masking helpers
func maskPhone(s string) string {
	r := []rune(s)
	n := len(r)
	if n <= 4 {
		return strings.Repeat("*", n)
	}
	// reveal first 3 and last 2 characters, mask the middle
	res := make([]rune, 0, n)
	for i := 0; i < 3 && i < n; i++ {
		res = append(res, r[i])
	}
	mid := n - 5
	if mid < 0 {
		mid = 0
	}
	for i := 0; i < mid; i++ {
		res = append(res, '*')
	}
	for i := n - 2; i < n && i >= 0; i++ {
		res = append(res, r[i])
	}
	return string(res)
}

func maskEmail(s string) string {
	parts := strings.SplitN(s, "@", 2)
	if len(parts) != 2 {
		return strings.Repeat("*", len(s))
	}
	local, domain := parts[0], parts[1]
	if len(local) <= 1 {
		localMasked := strings.Repeat("*", len(local))
		return localMasked + "@" + domain
	}
	maskedLocal := local[:1] + strings.Repeat("*", len(local)-1)
	return maskedLocal + "@" + domain
}

func maskIDCard(s string) string {
	r := []rune(s)
	n := len(r)
	if n <= 4 {
		return strings.Repeat("*", n)
	}
	res := make([]rune, 0, n)
	res = append(res, r[0], r[1])
	for i := 2; i < n-2; i++ {
		res = append(res, '*')
	}
	res = append(res, r[n-2], r[n-1])
	return string(res)
}

func maskBankCard(s string) string {
	r := []rune(s)
	n := len(r)
	if n <= 4 {
		return strings.Repeat("*", n)
	}
	masked := make([]rune, n)
	for i := range masked {
		masked[i] = '*'
	}
	// reveal last 4 digits if available
	start := n - 4
	for i := 0; i < 4 && i+start < n; i++ {
		masked[start+i] = r[start+i]
	}
	return string(masked)
}
