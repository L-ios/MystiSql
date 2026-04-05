package masking

import (
	"fmt"
	"strings"

	"MystiSql/pkg/types"
)

// MaskData contains fields that may be masked according to policy.
// Deprecated: Use MaskResult for dynamic column-level masking instead.
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
// Deprecated: Use MaskResult for dynamic column-level masking instead.
func (m *MaskingService) MaskForRole(role string, data MaskData) MaskData {
	rules := m.policyStore.Get(role)
	if rules == nil {
		return data
	}
	for _, rule := range rules.Rules {
		switch {
		case rule.MaskType == MaskTypePhone && matchOldField(rule, "phone"):
			data.Phone = maskPhone(data.Phone)
		case rule.MaskType == MaskTypeEmail && matchOldField(rule, "email"):
			data.Email = maskEmail(data.Email)
		case rule.MaskType == MaskTypeIDCard && matchOldField(rule, "idcard"):
			data.IDCard = maskIDCard(data.IDCard)
		case rule.MaskType == MaskTypeBankCard && matchOldField(rule, "bankcard"):
			data.BankCard = maskBankCard(data.BankCard)
		}
	}
	return data
}

// matchOldField provides backward compatibility for the legacy MaskForRole API.
func matchOldField(rule ColumnRule, field string) bool {
	return rule.Pattern == field || rule.Pattern == "*" || rule.Pattern == ""
}

// MaskResult applies masking rules to query result rows based on user role.
// It matches column names against rule patterns and applies the appropriate mask.
// Returns a new QueryResult with masked values; the original is not modified.
func (m *MaskingService) MaskResult(role string, result *types.QueryResult) *types.QueryResult {
	if result == nil || len(result.Rows) == 0 || len(result.Columns) == 0 {
		return result
	}

	rules := m.policyStore.Get(role)
	if rules == nil || len(rules.Rules) == 0 {
		return result
	}

	maskers := make([]func(string) string, len(result.Columns))
	for colIdx, col := range result.Columns {
		for _, rule := range rules.Rules {
			if rule.matchColumn(col.Name) {
				if fn := resolveMaskFunc(rule.MaskType); fn != nil {
					maskers[colIdx] = fn
					break
				}
			}
		}
	}

	hasMasking := false
	for _, fn := range maskers {
		if fn != nil {
			hasMasking = true
			break
		}
	}
	if !hasMasking {
		return result
	}

	maskedRows := make([]types.Row, len(result.Rows))
	for i, row := range result.Rows {
		newRow := make(types.Row, len(row))
		copy(newRow, row)
		for colIdx, fn := range maskers {
			if fn != nil && colIdx < len(newRow) {
				if s, ok := newRow[colIdx].(string); ok {
					newRow[colIdx] = fn(s)
				} else if newRow[colIdx] != nil {
					newRow[colIdx] = fn(fmt.Sprintf("%v", newRow[colIdx]))
				}
			}
		}
		maskedRows[i] = newRow
	}

	return &types.QueryResult{
		Columns:       result.Columns,
		Rows:          maskedRows,
		RowCount:      result.RowCount,
		Truncated:     result.Truncated,
		ExecutionTime: result.ExecutionTime,
	}
}

// resolveMaskFunc returns the masking function for the given MaskType.
func resolveMaskFunc(mt MaskType) func(string) string {
	switch mt {
	case MaskTypePhone:
		return maskPhone
	case MaskTypeEmail:
		return maskEmail
	case MaskTypeIDCard:
		return maskIDCard
	case MaskTypeBankCard:
		return maskBankCard
	default:
		return nil
	}
}

func maskPhone(s string) string {
	r := []rune(s)
	n := len(r)
	if n <= 4 {
		return strings.Repeat("*", n)
	}
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
	start := n - 4
	for i := 0; i < 4 && i+start < n; i++ {
		masked[start+i] = r[start+i]
	}
	return string(masked)
}
