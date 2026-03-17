import { create } from 'zustand';
export const useQueryStore = create()((set) => ({
    history: [],
    currentSql: '',
    addHistory: (item) => set((state) => ({
        history: [
            { ...item, id: crypto.randomUUID(), timestamp: Date.now() },
            ...state.history,
        ].slice(0, 100),
    })),
    clearHistory: () => set({ history: [] }),
    setCurrentSql: (sql) => set({ currentSql: sql }),
}));
