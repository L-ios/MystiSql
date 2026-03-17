import { create } from 'zustand';
import { persist } from 'zustand/middleware';
export const useAuthStore = create()(persist((set, get) => ({
    token: null,
    userId: null,
    role: null,
    expiresAt: null,
    setAuth: (token, userId, role, expiresAt) => set({ token, userId, role, expiresAt }),
    clearAuth: () => set({ token: null, userId: null, role: null, expiresAt: null }),
    isAuthenticated: () => {
        const { token, expiresAt } = get();
        if (!token)
            return false;
        if (expiresAt && Date.now() > expiresAt) {
            set({ token: null, userId: null, role: null, expiresAt: null });
            return false;
        }
        return true;
    },
}), {
    name: 'mystisql-auth',
}));
