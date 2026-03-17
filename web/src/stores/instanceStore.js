import { create } from 'zustand';
export const useInstanceStore = create((set) => ({
    instances: [],
    currentInstance: null,
    loading: false,
    setInstances: (instances) => set({ instances }),
    setCurrentInstance: (name) => set({ currentInstance: name }),
    setLoading: (loading) => set({ loading }),
}));
