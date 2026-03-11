package repl

import (
	"strings"
)

type InputBuffer struct {
	lines      []string
	inString   bool
	stringChar rune
	inComment  bool
	commentEnd string
}

func NewInputBuffer() *InputBuffer {
	return &InputBuffer{
		lines: make([]string, 0),
	}
}

func (b *InputBuffer) IsEmpty() bool {
	return len(b.lines) == 0 && !b.inString && !b.inComment
}

func (b *InputBuffer) Append(line string) {
	b.lines = append(b.lines, line)
	b.parseState(line)
}

func (b *InputBuffer) parseState(line string) {
	i := 0
	for i < len(line) {
		ch := rune(line[i])

		if b.inComment {
			if len(b.commentEnd) == 2 && i+1 < len(line) && line[i:i+2] == b.commentEnd {
				b.inComment = false
				b.commentEnd = ""
				i += 2
				continue
			}
			i++
			continue
		}

		if b.inString {
			if ch == '\\' && i+1 < len(line) {
				i += 2
				continue
			}
			if ch == b.stringChar {
				b.inString = false
				b.stringChar = 0
			}
			i++
			continue
		}

		if i+1 < len(line) && line[i:i+2] == "/*" {
			b.inComment = true
			b.commentEnd = "*/"
			i += 2
			continue
		}

		if i+1 < len(line) && (line[i:i+2] == "--" || line[i:i+2] == "#") {
			break
		}

		if ch == '\'' || ch == '"' || ch == '`' {
			b.inString = true
			b.stringChar = ch
		}

		i++
	}
}

func (b *InputBuffer) IsComplete() bool {
	if b.inString || b.inComment {
		return false
	}

	sql := b.GetSQL()
	sql = strings.TrimSpace(sql)

	if sql == "" {
		return false
	}

	if strings.HasSuffix(sql, ";") || strings.HasSuffix(sql, "\\g") {
		return true
	}

	return false
}

func (b *InputBuffer) GetSQL() string {
	return strings.Join(b.lines, "\n")
}

func (b *InputBuffer) GetContinuePrompt() string {
	if b.inComment {
		return "    */> "
	}
	if b.inString {
		switch b.stringChar {
		case '\'':
			return "    '> "
		case '"':
			return "    \"> "
		case '`':
			return "    `> "
		}
	}
	return "    -> "
}

func (b *InputBuffer) Reset() {
	b.lines = make([]string, 0)
	b.inString = false
	b.stringChar = 0
	b.inComment = false
	b.commentEnd = ""
}
