package repl

import (
	"bytes"
	"fmt"
	"io"
	"unicode"
)

type ReadLine struct {
	terminal *termWrapper
	history  *HistoryManager

	historyIndex int
	currentInput string
	cursorPos    int
}

type termWrapper struct {
	reader io.Reader
	writer io.Writer
}

func NewReadLine(reader io.Reader, writer io.Writer, history *HistoryManager) *ReadLine {
	return &ReadLine{
		terminal: &termWrapper{
			reader: reader,
			writer: writer,
		},
		history:      history,
		historyIndex: -1,
		cursorPos:    0,
	}
}

func (rl *ReadLine) ReadLine(prompt string) (string, error) {
	rl.historyIndex = -1
	rl.currentInput = ""
	rl.cursorPos = 0

	fmt.Fprint(rl.terminal.writer, prompt)

	var line []rune

	for {
		var ch [1]byte
		n, err := rl.terminal.reader.Read(ch[:])
		if err != nil {
			return "", err
		}
		if n == 0 {
			continue
		}

		switch ch[0] {
		case 1:
			rl.moveCursorToStart(line, prompt)

		case 2:
			rl.moveCursorLeft(line)

		case 5:
			rl.moveCursorToEnd(line, prompt)

		case 6:
			rl.moveCursorRight(line)

		case 3:
			fmt.Fprint(rl.terminal.writer, "^C\r\n")
			return "", nil

		case 11:
			rl.deleteToEnd(line, prompt)
			line = line[:rl.cursorPos]

		case 21:
			deleted := rl.deleteToStart(line, prompt)
			line = line[deleted:]
			rl.cursorPos = 0

		case 13, 10:
			fmt.Fprint(rl.terminal.writer, "\r\n")
			return string(line), nil

		case 27:
			escape := rl.readEscapeSequence()
			line = rl.handleEscapeSequence(escape, line, prompt)

		case 127, 8:
			if rl.cursorPos > 0 {
				line = rl.deleteChar(line, rl.cursorPos-1)
				rl.cursorPos--
				rl.refreshLine(line, prompt)
			}

		default:
			if ch[0] >= 32 && ch[0] < 127 {
				line = rl.insertChar(line, rune(ch[0]), rl.cursorPos)
				rl.cursorPos++
				rl.refreshLine(line, prompt)
			}
		}
	}
}

func (rl *ReadLine) handleEscapeSequence(escape string, line []rune, prompt string) []rune {
	switch escape {
	case "[A":
		return rl.navigateHistory(-1, line, prompt)
	case "[B":
		return rl.navigateHistory(1, line, prompt)
	case "[C":
		rl.moveCursorRight(line)
	case "[D":
		rl.moveCursorLeft(line)
	case "[H", "OH":
		rl.moveCursorToStart(line, prompt)
	case "[F", "OF":
		rl.moveCursorToEnd(line, prompt)
	case "[3~":
		if rl.cursorPos < len(line) {
			line = rl.deleteChar(line, rl.cursorPos)
			rl.refreshLine(line, prompt)
		}
	case "b":
		rl.moveCursorBackwardWord(line)
	case "f":
		rl.moveCursorForwardWord(line)
	}
	return line
}

func (rl *ReadLine) readEscapeSequence() string {
	var seq bytes.Buffer
	for i := 0; i < 4; i++ {
		var ch [1]byte
		n, err := rl.terminal.reader.Read(ch[:])
		if err != nil || n == 0 {
			break
		}
		seq.WriteByte(ch[0])
		if (ch[0] >= 'A' && ch[0] <= 'Z') || (ch[0] >= 'a' && ch[0] <= 'z') || ch[0] == '~' {
			break
		}
	}
	return seq.String()
}

func (rl *ReadLine) insertChar(line []rune, ch rune, pos int) []rune {
	if pos < 0 || pos > len(line) {
		return line
	}
	newLine := make([]rune, 0, len(line)+1)
	newLine = append(newLine, line[:pos]...)
	newLine = append(newLine, ch)
	newLine = append(newLine, line[pos:]...)
	return newLine
}

func (rl *ReadLine) deleteChar(line []rune, pos int) []rune {
	if pos < 0 || pos >= len(line) {
		return line
	}
	newLine := make([]rune, 0, len(line)-1)
	newLine = append(newLine, line[:pos]...)
	newLine = append(newLine, line[pos+1:]...)
	return newLine
}

func (rl *ReadLine) moveCursorLeft(line []rune) {
	if rl.cursorPos > 0 {
		rl.cursorPos--
		fmt.Fprint(rl.terminal.writer, "\b")
	}
}

func (rl *ReadLine) moveCursorRight(line []rune) {
	if rl.cursorPos < len(line) {
		rl.cursorPos++
		fmt.Fprint(rl.terminal.writer, "\x1b[C")
	}
}

func (rl *ReadLine) moveCursorToStart(line []rune, prompt string) {
	if rl.cursorPos > 0 {
		for rl.cursorPos > 0 {
			rl.cursorPos--
			fmt.Fprint(rl.terminal.writer, "\b")
		}
	}
}

func (rl *ReadLine) moveCursorToEnd(line []rune, prompt string) {
	if rl.cursorPos < len(line) {
		for rl.cursorPos < len(line) {
			rl.cursorPos++
			fmt.Fprint(rl.terminal.writer, "\x1b[C")
		}
	}
}

func (rl *ReadLine) moveCursorBackwardWord(line []rune) {
	if rl.cursorPos == 0 {
		return
	}

	pos := rl.cursorPos

	for pos > 0 && unicode.IsSpace(line[pos-1]) {
		pos--
	}

	for pos > 0 && !unicode.IsSpace(line[pos-1]) {
		pos--
	}

	steps := rl.cursorPos - pos
	for i := 0; i < steps; i++ {
		fmt.Fprint(rl.terminal.writer, "\b")
	}
	rl.cursorPos = pos
}

func (rl *ReadLine) moveCursorForwardWord(line []rune) {
	if rl.cursorPos >= len(line) {
		return
	}

	pos := rl.cursorPos

	for pos < len(line) && unicode.IsSpace(line[pos]) {
		pos++
	}

	for pos < len(line) && !unicode.IsSpace(line[pos]) {
		pos++
	}

	steps := pos - rl.cursorPos
	for i := 0; i < steps; i++ {
		fmt.Fprint(rl.terminal.writer, "\x1b[C")
	}
	rl.cursorPos = pos
}

func (rl *ReadLine) deleteToStart(line []rune, prompt string) int {
	if rl.cursorPos == 0 {
		return 0
	}
	deleted := rl.cursorPos
	rl.cursorPos = 0
	return deleted
}

func (rl *ReadLine) deleteToEnd(line []rune, prompt string) {
	fmt.Fprint(rl.terminal.writer, "\x1b[K")
}

func (rl *ReadLine) refreshLine(line []rune, prompt string) {
	fmt.Fprint(rl.terminal.writer, "\r")
	fmt.Fprint(rl.terminal.writer, prompt)
	fmt.Fprint(rl.terminal.writer, string(line))

	fmt.Fprint(rl.terminal.writer, "\x1b[K")

	cursorFromEnd := len(line) - rl.cursorPos
	for i := 0; i < cursorFromEnd; i++ {
		fmt.Fprint(rl.terminal.writer, "\b")
	}
}

func (rl *ReadLine) navigateHistory(direction int, currentLine []rune, prompt string) []rune {
	count := rl.history.Count()
	if count == 0 {
		return currentLine
	}

	if rl.historyIndex == -1 {
		rl.currentInput = string(currentLine)
	}

	newIndex := rl.historyIndex + direction
	if newIndex < -1 {
		newIndex = -1
	}
	if newIndex >= count {
		newIndex = count - 1
	}

	rl.historyIndex = newIndex

	var newLine string
	if newIndex == -1 {
		newLine = rl.currentInput
	} else {
		newLine = rl.history.Get(count - 1 - newIndex)
	}

	rl.cursorPos = len([]rune(newLine))
	rl.refreshLine([]rune(newLine), prompt)

	return []rune(newLine)
}

func (rl *ReadLine) Reset() {
	rl.historyIndex = -1
	rl.currentInput = ""
	rl.cursorPos = 0
}
