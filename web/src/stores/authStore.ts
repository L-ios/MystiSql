import { create } from 'zustand'
import { persist } from 'zustand/middleware'

interface AuthState {
  token: string | null
  userId: string | null
  role: string | null
  expiresAt: number | null
  setAuth: (token: string, userId: string, role: string, expiresAt: number) => void
  clearAuth: () => void
  isAuthenticated: () => boolean
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set, get) => ({
      token: null,
      userId: null,
      role: null,
      expiresAt: null,
      setAuth: (token, userId, role, expiresAt) =>
        set({ token, userId, role, expiresAt }),
      clearAuth: () =>
        set({ token: null, userId: null, role: null, expiresAt: null }),
      isAuthenticated: () => {
        const { token, expiresAt } = get()
        if (!token) return false
        if (expiresAt && Date.now() > expiresAt) {
          set({ token: null, userId: null, role: null, expiresAt: null })
          return false
        }
        return true
      },
    }),
    {
      name: 'mystisql-auth',
    }
  )
)
