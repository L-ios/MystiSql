package validator

import (
	"sync"
	"testing"
)

func TestWhitelistManager_Add(t *testing.T) {
	tests := []struct {
		name        string
		patterns    []string
		wantErr     bool
		wantCount   int
		description string
	}{
		{
			name:        "添加有效的正则表达式模式",
			patterns:    []string{`^SELECT.*`, `^SHOW.*`},
			wantErr:     false,
			wantCount:   2,
			description: "应该成功添加两个有效模式",
		},
		{
			name:        "添加单个模式",
			patterns:    []string{`^SELECT\s+\*\s+FROM\s+\w+`},
			wantErr:     false,
			wantCount:   1,
			description: "应该成功添加单个模式",
		},
		{
			name:        "添加无效的正则表达式模式",
			patterns:    []string{`[INVALID(`},
			wantErr:     true,
			wantCount:   0,
			description: "应该返回错误且不添加到列表",
		},
		{
			name:        "添加空模式",
			patterns:    []string{""},
			wantErr:     false,
			wantCount:   1,
			description: "空字符串也是有效的正则表达式",
		},
		{
			name:        "添加多个模式包含一个无效的",
			patterns:    []string{`^SELECT.*`, `[INVALID(`, `^SHOW.*`},
			wantErr:     true,
			wantCount:   1,
			description: "遇到无效模式应返回错误，但之前的有效模式已添加",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wm := NewWhitelistManager()
			var err error
			var addedCount int

			for _, pattern := range tt.patterns {
				err = wm.Add(pattern)
				if err != nil {
					break
				}
				addedCount++
			}

			if (err != nil) != tt.wantErr {
				t.Errorf("WhitelistManager.Add() error = %v, wantErr %v", err, tt.wantErr)
			}

			if got := len(wm.GetAll()); got != tt.wantCount {
				t.Errorf("WhitelistManager.GetAll() count = %v, wantCount %v", got, tt.wantCount)
			}
		})
	}
}

func TestWhitelistManager_Remove(t *testing.T) {
	tests := []struct {
		name        string
		initial     []string
		remove      string
		wantCount   int
		wantMatch   string
		wantMatched bool
		description string
	}{
		{
			name:        "移除存在的模式",
			initial:     []string{`^SELECT.*`, `^SHOW.*`, `^DESCRIBE.*`},
			remove:      `^SHOW.*`,
			wantCount:   2,
			wantMatch:   "SHOW TABLES",
			wantMatched: false,
			description: "移除后应该只有2个模式，且不应匹配已移除的模式",
		},
		{
			name:        "移除不存在的模式",
			initial:     []string{`^SELECT.*`},
			remove:      `^DELETE.*`,
			wantCount:   1,
			wantMatch:   "SELECT * FROM users",
			wantMatched: true,
			description: "移除不存在的模式不应影响现有模式",
		},
		{
			name:        "从空列表移除",
			initial:     []string{},
			remove:      `^SELECT.*`,
			wantCount:   0,
			wantMatch:   "SELECT * FROM users",
			wantMatched: false,
			description: "从空列表移除应该安全处理",
		},
		{
			name:        "移除第一个模式",
			initial:     []string{`^SELECT.*`, `^SHOW.*`},
			remove:      `^SELECT.*`,
			wantCount:   1,
			wantMatch:   "SELECT * FROM users",
			wantMatched: false,
			description: "移除第一个模式后应正确更新",
		},
		{
			name:        "移除最后一个模式",
			initial:     []string{`^SELECT.*`, `^SHOW.*`},
			remove:      `^SHOW.*`,
			wantCount:   1,
			wantMatch:   "SHOW TABLES",
			wantMatched: false,
			description: "移除最后一个模式后应正确更新",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wm := NewWhitelistManager()
			for _, pattern := range tt.initial {
				if err := wm.Add(pattern); err != nil {
					t.Fatalf("添加初始模式失败: %v", err)
				}
			}

			wm.Remove(tt.remove)

			if got := len(wm.GetAll()); got != tt.wantCount {
				t.Errorf("移除后 GetAll() count = %v, wantCount %v", got, tt.wantCount)
			}

			if matched := wm.Match(tt.wantMatch); matched != tt.wantMatched {
				t.Errorf("移除后 Match(%q) = %v, wantMatched %v", tt.wantMatch, matched, tt.wantMatched)
			}
		})
	}
}

func TestWhitelistManager_Match(t *testing.T) {
	tests := []struct {
		name        string
		patterns    []string
		query       string
		wantMatch   bool
		description string
	}{
		{
			name:        "匹配简单的SELECT查询",
			patterns:    []string{`^SELECT.*`},
			query:       "SELECT * FROM users",
			wantMatch:   true,
			description: "应该匹配以SELECT开头的查询",
		},
		{
			name:        "匹配带参数的查询",
			patterns:    []string{`^SELECT\s+\*\s+FROM\s+\w+\s+WHERE\s+\w+\s*=\s*\?`},
			query:       "SELECT * FROM users WHERE id = ?",
			wantMatch:   true,
			description: "应该匹配带参数的SELECT查询",
		},
		{
			name:        "不匹配不同类型的查询",
			patterns:    []string{`^SELECT.*`},
			query:       "DELETE FROM users",
			wantMatch:   false,
			description: "不应该匹配DELETE查询",
		},
		{
			name:        "匹配多个模式之一",
			patterns:    []string{`^SELECT.*`, `^SHOW.*`, `^DESCRIBE.*`},
			query:       "SHOW TABLES",
			wantMatch:   true,
			description: "应该匹配多个模式中的一个",
		},
		{
			name:        "空模式列表不匹配任何查询",
			patterns:    []string{},
			query:       "SELECT * FROM users",
			wantMatch:   false,
			description: "空列表不应匹配任何查询",
		},
		{
			name:        "大小写不敏感匹配",
			patterns:    []string{`(?i)^select.*`},
			query:       "select * from users",
			wantMatch:   true,
			description: "使用(?i)标志应该大小写不敏感匹配",
		},
		{
			name:        "匹配特定表名",
			patterns:    []string{`^SELECT.*FROM\s+(users|orders|products)`},
			query:       "SELECT * FROM users WHERE id = 1",
			wantMatch:   true,
			description: "应该匹配特定表名",
		},
		{
			name:        "不匹配不在列表中的表",
			patterns:    []string{`^SELECT.*FROM\s+(users|orders|products)`},
			query:       "SELECT * FROM admin WHERE id = 1",
			wantMatch:   false,
			description: "不应该匹配不在白名单中的表",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wm := NewWhitelistManager()
			for _, pattern := range tt.patterns {
				if err := wm.Add(pattern); err != nil {
					t.Fatalf("添加模式失败: %v", err)
				}
			}

			if matched := wm.Match(tt.query); matched != tt.wantMatch {
				t.Errorf("WhitelistManager.Match(%q) = %v, wantMatch %v", tt.query, matched, tt.wantMatch)
			}
		})
	}
}

func TestWhitelistManager_GetAll(t *testing.T) {
	tests := []struct {
		name        string
		patterns    []string
		wantCount   int
		description string
	}{
		{
			name:        "获取所有添加的模式",
			patterns:    []string{`^SELECT.*`, `^SHOW.*`, `^DESCRIBE.*`},
			wantCount:   3,
			description: "应该返回所有添加的模式",
		},
		{
			name:        "获取空列表",
			patterns:    []string{},
			wantCount:   0,
			description: "空列表应返回0个元素",
		},
		{
			name:        "添加重复模式",
			patterns:    []string{`^SELECT.*`, `^SELECT.*`},
			wantCount:   2,
			description: "应该保留重复的模式",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wm := NewWhitelistManager()
			for _, pattern := range tt.patterns {
				if err := wm.Add(pattern); err != nil {
					t.Fatalf("添加模式失败: %v", err)
				}
			}

			got := wm.GetAll()
			if len(got) != tt.wantCount {
				t.Errorf("WhitelistManager.GetAll() count = %v, wantCount %v", len(got), tt.wantCount)
			}

			for i, pattern := range tt.patterns {
				if i >= len(got) || got[i] != pattern {
					if i < len(got) {
						t.Errorf("GetAll()[%d] = %v, want %v", i, got[i], pattern)
					} else {
						t.Errorf("GetAll() missing pattern[%d] = %v", i, pattern)
					}
				}
			}
		})
	}
}

func TestBlacklistManager_Add(t *testing.T) {
	tests := []struct {
		name        string
		patterns    []string
		wantErr     bool
		wantCount   int
		description string
	}{
		{
			name:        "添加危险操作模式",
			patterns:    []string{`(?i)^\s*DROP\s+`, `(?i)^\s*TRUNCATE\s+`},
			wantErr:     false,
			wantCount:   2,
			description: "应该成功添加危险操作模式",
		},
		{
			name:        "添加DELETE不带WHERE的模式",
			patterns:    []string{`(?i)^\s*DELETE\s+FROM\s+\w+\s*;?\s*$`},
			wantErr:     false,
			wantCount:   1,
			description: "应该成功添加DELETE模式",
		},
		{
			name:        "添加无效的正则表达式",
			patterns:    []string{`[INVALID(`},
			wantErr:     true,
			wantCount:   0,
			description: "应该返回错误",
		},
		{
			name:        "添加SQL注入模式",
			patterns:    []string{`(?i)(\bOR\b|\bAND\b).*?=\s*.*?--`, `(?i)\bUNION\s+SELECT\b`},
			wantErr:     false,
			wantCount:   2,
			description: "应该成功添加SQL注入模式",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bm := NewBlacklistManager()
			var err error
			var addedCount int

			for _, pattern := range tt.patterns {
				err = bm.Add(pattern)
				if err != nil {
					break
				}
				addedCount++
			}

			if (err != nil) != tt.wantErr {
				t.Errorf("BlacklistManager.Add() error = %v, wantErr %v", err, tt.wantErr)
			}

			if got := len(bm.GetAll()); got != tt.wantCount {
				t.Errorf("BlacklistManager.GetAll() count = %v, wantCount %v", got, tt.wantCount)
			}
		})
	}
}

func TestBlacklistManager_Remove(t *testing.T) {
	tests := []struct {
		name        string
		initial     []string
		remove      string
		wantCount   int
		query       string
		wantMatched bool
		description string
	}{
		{
			name:        "移除黑名单模式",
			initial:     []string{`(?i)^\s*DROP\s+`, `(?i)^\s*TRUNCATE\s+`},
			remove:      `(?i)^\s*DROP\s+`,
			wantCount:   1,
			query:       "DROP TABLE users",
			wantMatched: false,
			description: "移除后不应匹配DROP语句",
		},
		{
			name:        "移除不存在的模式",
			initial:     []string{`(?i)^\s*DROP\s+`},
			remove:      `(?i)^\s*DELETE\s+`,
			wantCount:   1,
			query:       "DROP TABLE users",
			wantMatched: true,
			description: "移除不存在的模式不应影响现有模式",
		},
		{
			name:        "从空黑名单移除",
			initial:     []string{},
			remove:      `(?i)^\s*DROP\s+`,
			wantCount:   0,
			query:       "DROP TABLE users",
			wantMatched: false,
			description: "从空黑名单移除应该安全处理",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bm := NewBlacklistManager()
			for _, pattern := range tt.initial {
				if err := bm.Add(pattern); err != nil {
					t.Fatalf("添加初始模式失败: %v", err)
				}
			}

			bm.Remove(tt.remove)

			if got := len(bm.GetAll()); got != tt.wantCount {
				t.Errorf("移除后 GetAll() count = %v, wantCount %v", got, tt.wantCount)
			}

			if matched := bm.Match(tt.query); matched != tt.wantMatched {
				t.Errorf("移除后 Match(%q) = %v, wantMatched %v", tt.query, matched, tt.wantMatched)
			}
		})
	}
}

func TestBlacklistManager_Match(t *testing.T) {
	tests := []struct {
		name        string
		patterns    []string
		query       string
		wantMatch   bool
		description string
	}{
		{
			name:        "匹配DROP语句",
			patterns:    []string{`(?i)^\s*DROP\s+`},
			query:       "DROP TABLE users",
			wantMatch:   true,
			description: "应该匹配DROP语句",
		},
		{
			name:        "匹配TRUNCATE语句",
			patterns:    []string{`(?i)^\s*TRUNCATE\s+`},
			query:       "TRUNCATE TABLE logs",
			wantMatch:   true,
			description: "应该匹配TRUNCATE语句",
		},
		{
			name:        "不匹配安全的SELECT查询",
			patterns:    []string{`(?i)^\s*DROP\s+`, `(?i)^\s*TRUNCATE\s+`},
			query:       "SELECT * FROM users",
			wantMatch:   false,
			description: "不应该匹配SELECT查询",
		},
		{
			name:        "匹配SQL注入模式",
			patterns:    []string{`(?i)\bOR\b.*?=.*?--`},
			query:       "SELECT * FROM users WHERE id = 1 OR 1=1--",
			wantMatch:   true,
			description: "应该匹配SQL注入模式",
		},
		{
			name:        "匹配UNION注入",
			patterns:    []string{`(?i)\bUNION\s+SELECT\b`},
			query:       "SELECT * FROM users UNION SELECT * FROM admin",
			wantMatch:   true,
			description: "应该匹配UNION SELECT注入",
		},
		{
			name:        "空黑名单不匹配任何查询",
			patterns:    []string{},
			query:       "DROP TABLE users",
			wantMatch:   false,
			description: "空黑名单不应匹配任何查询",
		},
		{
			name:        "大小写不敏感匹配DROP",
			patterns:    []string{`(?i)^\s*DROP\s+`},
			query:       "drop table users",
			wantMatch:   true,
			description: "应该大小写不敏感匹配",
		},
		{
			name:        "匹配DELETE无WHERE",
			patterns:    []string{`(?i)^\s*DELETE\s+FROM\s+\w+\s*;?\s*$`},
			query:       "DELETE FROM users;",
			wantMatch:   true,
			description: "应该匹配不带WHERE的DELETE",
		},
		{
			name:        "不匹配带WHERE的DELETE",
			patterns:    []string{`(?i)^\s*DELETE\s+FROM\s+\w+\s*;?\s*$`},
			query:       "DELETE FROM users WHERE id = 1",
			wantMatch:   false,
			description: "不应该匹配带WHERE的DELETE",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bm := NewBlacklistManager()
			for _, pattern := range tt.patterns {
				if err := bm.Add(pattern); err != nil {
					t.Fatalf("添加模式失败: %v", err)
				}
			}

			if matched := bm.Match(tt.query); matched != tt.wantMatch {
				t.Errorf("BlacklistManager.Match(%q) = %v, wantMatch %v", tt.query, matched, tt.wantMatch)
			}
		})
	}
}

func TestBlacklistManager_GetAll(t *testing.T) {
	tests := []struct {
		name        string
		patterns    []string
		wantCount   int
		description string
	}{
		{
			name:        "获取所有黑名单模式",
			patterns:    []string{`(?i)^\s*DROP\s+`, `(?i)^\s*TRUNCATE\s+`, `(?i)^\s*DELETE\s+FROM\s+\w+\s*;?\s*$`},
			wantCount:   3,
			description: "应该返回所有添加的黑名单模式",
		},
		{
			name:        "获取空黑名单",
			patterns:    []string{},
			wantCount:   0,
			description: "空黑名单应返回0个元素",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bm := NewBlacklistManager()
			for _, pattern := range tt.patterns {
				if err := bm.Add(pattern); err != nil {
					t.Fatalf("添加模式失败: %v", err)
				}
			}

			got := bm.GetAll()
			if len(got) != tt.wantCount {
				t.Errorf("BlacklistManager.GetAll() count = %v, wantCount %v", len(got), tt.wantCount)
			}

			for i, pattern := range tt.patterns {
				if i >= len(got) || got[i] != pattern {
					if i < len(got) {
						t.Errorf("GetAll()[%d] = %v, want %v", i, got[i], pattern)
					} else {
						t.Errorf("GetAll() missing pattern[%d] = %v", i, pattern)
					}
				}
			}
		})
	}
}

func TestWhitelistManager_ConcurrentAccess(t *testing.T) {
	wm := NewWhitelistManager()
	var wg sync.WaitGroup
	var mu sync.Mutex
	const goroutines = 50
	const operations = 100

	patterns := []string{`^SELECT.*`, `^SHOW.*`, `^DESCRIBE.*`, `^INSERT.*`, `^UPDATE.*`}

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < operations; j++ {
				pattern := patterns[(id+j)%len(patterns)]

				switch j % 4 {
				case 0:
					mu.Lock()
					wm.Add(pattern)
					mu.Unlock()
				case 1:
					mu.Lock()
					wm.Remove(pattern)
					mu.Unlock()
				case 2:
					wm.Match("SELECT * FROM users")
				case 3:
					wm.GetAll()
				}
			}
		}(i)
	}

	wg.Wait()
}

func TestBlacklistManager_ConcurrentAccess(t *testing.T) {
	bm := NewBlacklistManager()
	var wg sync.WaitGroup
	var mu sync.Mutex
	const goroutines = 50
	const operations = 100

	patterns := []string{
		`(?i)^\s*DROP\s+`,
		`(?i)^\s*TRUNCATE\s+`,
		`(?i)^\s*DELETE\s+FROM\s+\w+\s*;?\s*$`,
		`(?i)\bUNION\s+SELECT\b`,
		`(?i)\bOR\b.*?=.*?--`,
	}

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < operations; j++ {
				pattern := patterns[(id+j)%len(patterns)]

				switch j % 4 {
				case 0:
					mu.Lock()
					bm.Add(pattern)
					mu.Unlock()
				case 1:
					mu.Lock()
					bm.Remove(pattern)
					mu.Unlock()
				case 2:
					bm.Match("DROP TABLE users")
				case 3:
					bm.GetAll()
				}
			}
		}(i)
	}

	wg.Wait()
}

func TestWhitelistManager_ComplexPatterns(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
		queries []struct {
			query string
			match bool
		}
		description string
	}{
		{
			name:    "复杂的表访问控制模式",
			pattern: `^SELECT\s+[\w\*,\s]+\s+FROM\s+(users|orders|products)(?:\s+WHERE\s+.*)?$`,
			queries: []struct {
				query string
				match bool
			}{
				{"SELECT * FROM users", true},
				{"SELECT id, name FROM users WHERE id = 1", true},
				{"SELECT * FROM admin", false},
				{"SELECT * FROM users ORDER BY id", false},
			},
			description: "测试复杂的表访问控制",
		},
		{
			name:    "带分组和量词的模式",
			pattern: `^SELECT\s+(?:DISTINCT\s+)?[\w\*,\s]+\s+FROM\s+\w+(?:\s+(?:WHERE|ORDER|GROUP|LIMIT)\s+.*)?`,
			queries: []struct {
				query string
				match bool
			}{
				{"SELECT * FROM users", true},
				{"SELECT DISTINCT name FROM users WHERE id = 1", true},
				{"SELECT * FROM users ORDER BY created_at DESC", true},
				{"DELETE FROM users", false},
			},
			description: "测试带分组和量词的模式",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wm := NewWhitelistManager()
			if err := wm.Add(tt.pattern); err != nil {
				t.Fatalf("添加模式失败: %v", err)
			}

			for _, q := range tt.queries {
				if matched := wm.Match(q.query); matched != q.match {
					t.Errorf("Match(%q) = %v, want %v", q.query, matched, q.match)
				}
			}
		})
	}
}

func TestBlacklistManager_ComplexPatterns(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
		queries []struct {
			query string
			match bool
		}
		description string
	}{
		{
			name:    "检测多种SQL注入模式",
			pattern: `(?i)(\bOR\b|\bAND\b)\s+[\w']+\s*=\s*[\w']+.*?(?:--|;|\s*$)`,
			queries: []struct {
				query string
				match bool
			}{
				{"SELECT * FROM users WHERE id = 1 OR '1'='1'--", true},
				{"SELECT * FROM users WHERE id = 1 AND '1'='1';", true},
				{"SELECT * FROM users WHERE id = 1", false},
				{"SELECT * FROM users WHERE name = 'john'", false},
			},
			description: "测试多种SQL注入检测",
		},
		{
			name:    "检测危险操作组合",
			pattern: `(?i)^\s*(?:DROP|TRUNCATE|ALTER|CREATE)\s+(?:TABLE|DATABASE|SCHEMA)`,
			queries: []struct {
				query string
				match bool
			}{
				{"DROP TABLE users", true},
				{"TRUNCATE TABLE logs", true},
				{"ALTER DATABASE test", true},
				{"CREATE SCHEMA app", true},
				{"SELECT * FROM users", false},
			},
			description: "测试危险操作组合检测",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bm := NewBlacklistManager()
			if err := bm.Add(tt.pattern); err != nil {
				t.Fatalf("添加模式失败: %v", err)
			}

			for _, q := range tt.queries {
				if matched := bm.Match(q.query); matched != q.match {
					t.Errorf("Match(%q) = %v, want %v", q.query, matched, q.match)
				}
			}
		})
	}
}

func TestWhitelistAndBlacklist_Integration(t *testing.T) {
	wm := NewWhitelistManager()
	bm := NewBlacklistManager()

	whitelistPatterns := []string{
		`^SELECT\s+[\w\*,\s]+\s+FROM\s+\w+`,
		`^SHOW\s+`,
	}

	blacklistPatterns := []string{
		`(?i)^\s*DROP\s+`,
		`(?i)\bUNION\s+SELECT\b`,
	}

	for _, p := range whitelistPatterns {
		if err := wm.Add(p); err != nil {
			t.Fatalf("添加白名单模式失败: %v", err)
		}
	}

	for _, p := range blacklistPatterns {
		if err := bm.Add(p); err != nil {
			t.Fatalf("添加黑名单模式失败: %v", err)
		}
	}

	tests := []struct {
		name          string
		query         string
		allowByWhite  bool
		blockByBlack  bool
		shouldProcess bool
		description   string
	}{
		{
			name:          "安全的SELECT查询",
			query:         "SELECT * FROM users",
			allowByWhite:  true,
			blockByBlack:  false,
			shouldProcess: true,
			description:   "在白名单中且不在黑名单中，应允许处理",
		},
		{
			name:          "SQL注入攻击",
			query:         "SELECT * FROM users UNION SELECT * FROM admin",
			allowByWhite:  true,
			blockByBlack:  true,
			shouldProcess: false,
			description:   "虽然符合白名单，但在黑名单中，应拒绝",
		},
		{
			name:          "不在白名单的查询",
			query:         "DELETE FROM users",
			allowByWhite:  false,
			blockByBlack:  false,
			shouldProcess: false,
			description:   "不在白名单中，应拒绝",
		},
		{
			name:          "DROP语句",
			query:         "DROP TABLE users",
			allowByWhite:  false,
			blockByBlack:  true,
			shouldProcess: false,
			description:   "在黑名单中，应拒绝",
		},
		{
			name:          "SHOW查询",
			query:         "SHOW TABLES",
			allowByWhite:  true,
			blockByBlack:  false,
			shouldProcess: true,
			description:   "SHOW查询在白名单中，应允许",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			allowByWhite := wm.Match(tt.query)
			blockByBlack := bm.Match(tt.query)
			shouldProcess := allowByWhite && !blockByBlack

			if allowByWhite != tt.allowByWhite {
				t.Errorf("白名单匹配结果 = %v, want %v", allowByWhite, tt.allowByWhite)
			}

			if blockByBlack != tt.blockByBlack {
				t.Errorf("黑名单匹配结果 = %v, want %v", blockByBlack, tt.blockByBlack)
			}

			if shouldProcess != tt.shouldProcess {
				t.Errorf("最终处理决策 = %v, want %v", shouldProcess, tt.shouldProcess)
			}
		})
	}
}
