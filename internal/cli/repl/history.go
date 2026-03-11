package repl

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type HistoryManager struct {
	entries []string
	maxSize int
	file    string
}

func NewHistoryManager() *HistoryManager {
	homeDir, _ := os.UserHomeDir()
	historyDir := filepath.Join(homeDir, ".mystisql")
	historyFile := filepath.Join(historyDir, "history")

	return &HistoryManager{
		entries: make([]string, 0),
		maxSize: 1000,
		file:    historyFile,
	}
}

func (h *HistoryManager) Add(entry string) {
	entry = strings.TrimSpace(entry)
	if entry == "" {
		return
	}

	if len(h.entries) > 0 && h.entries[len(h.entries)-1] == entry {
		return
	}

	for i, e := range h.entries {
		if e == entry {
			h.entries = append(h.entries[:i], h.entries[i+1:]...)
			break
		}
	}

	h.entries = append(h.entries, entry)

	if len(h.entries) > h.maxSize {
		h.entries = h.entries[len(h.entries)-h.maxSize:]
	}
}

func (h *HistoryManager) Get(index int) string {
	if index < 0 || index >= len(h.entries) {
		return ""
	}
	return h.entries[index]
}

func (h *HistoryManager) Count() int {
	return len(h.entries)
}

func (h *HistoryManager) All() []string {
	return h.entries
}

func (h *HistoryManager) Clear() {
	h.entries = make([]string, 0)
}

func (h *HistoryManager) Load() error {
	dir := filepath.Dir(h.file)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create history directory: %w", err)
	}

	file, err := os.Open(h.file)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("failed to open history file: %w", err)
	}
	defer file.Close()

	h.entries = make([]string, 0)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			h.entries = append(h.entries, line)
		}
	}

	return scanner.Err()
}

func (h *HistoryManager) Save() error {
	dir := filepath.Dir(h.file)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create history directory: %w", err)
	}

	file, err := os.Create(h.file)
	if err != nil {
		return fmt.Errorf("failed to create history file: %w", err)
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	for _, entry := range h.entries {
		if _, err := writer.WriteString(entry + "\n"); err != nil {
			return fmt.Errorf("failed to write history: %w", err)
		}
	}

	return writer.Flush()
}
