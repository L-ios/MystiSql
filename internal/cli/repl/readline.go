package repl

import (
	"bytes"
	"fmt"
	"io"
)

type ReadLine struct {
	terminal *termWrapper
	history  *HistoryManager

	historyIndex int
	currentInput string
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
	}
}

func (rl *ReadLine) ReadLine(prompt string) (string, error) {
	rl.historyIndex = -1
	rl.currentInput = ""

	fmt.Fprint(rl.terminal.writer, prompt)

	var buf bytes.Buffer
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
		case 3:
			fmt.Fprintln(rl.terminal.writer, "^C")
			return "", nil

		case 13, 10:
			fmt.Fprintln(rl.terminal.writer)
			return string(line), nil

		case 127, 8:
			if len(line) > 0 {
				line = line[:len(line)-1]
				fmt.Fprint(rl.terminal.writer, "\b \b")
			}

		case 27:
			escape := rl.readEscapeSequence()
			if escape == "[A" {
				line = rl.navigateHistory(-1, line, prompt)
			} else if escape == "[B" {
				line = rl.navigateHistory(1, line, prompt)
			}

		default:
			if ch[0] >= 32 && ch[0] < 127 {
				line = append(line, rune(ch[0]))
				fmt.Fprintf(rl.terminal.writer, "%c", ch[0])
			}
		}

		buf.Reset()
	}

}

func (rl *ReadLine) readEscapeSequence() string {
	var seq bytes.Buffer
	for i := 0; i < 3; i++ {
		var ch [1]byte
		n, err := rl.terminal.reader.Read(ch[:])
		if err != nil || n == 0 {
			break
		}
		seq.WriteByte(ch[0])
		if (ch[0] >= 'A' && ch[0] <= 'Z') || (ch[0] >= 'a' && ch[0] <= 'z') {
			break
		}
	}
	return seq.String()
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

	rl.clearLine(len(prompt), len(currentLine))
	fmt.Fprint(rl.terminal.writer, prompt)
	fmt.Fprint(rl.terminal.writer, newLine)

	return []rune(newLine)
}

func (rl *ReadLine) clearLine(promptLen, lineLen int) {
	totalLen := promptLen + lineLen
	for i := 0; i < totalLen; i++ {
		fmt.Fprint(rl.terminal.writer, "\b")
	}
	fmt.Fprint(rl.terminal.writer, "\r")
}

func (rl *ReadLine) Reset() {
	rl.historyIndex = -1
	rl.currentInput = ""
}
