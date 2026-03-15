import { create } from 'zustand'

interface HistoryItem {
  id: string
  sql: string
  instance: string
  timestamp: number
  success: boolean
}

interface QueryState {
  history: HistoryItem[]
  currentSql: string
  addHistory: (item: Omit<HistoryItem, 'id' | 'timestamp'>) => void
  clearHistory: () => void
  setCurrentSql: (sql: string) => void
}

export const useQueryStore = create<QueryState>()(
  (set) => ({
    history: [],
    currentSql: '',
    addHistory: (item) =>
      set((state) => ({
        history: [
          { ...item, id: crypto.randomUUID(), timestamp: Date.now() },
          ...state.history,
        ].slice(0, 100),
      })),
    clearHistory: () => set({ history: [] }),
    setCurrentSql: (sql) => set({ currentSql: sql }),
  })
)
