import { create } from 'zustand'

interface Instance {
  name: string
  type: string
  host: string
  port: number
  database?: string
  status: string
  labels?: Record<string, string>
}

interface InstanceState {
  instances: Instance[]
  currentInstance: string | null
  loading: boolean
  setInstances: (instances: Instance[]) => void
  setCurrentInstance: (name: string | null) => void
  setLoading: (loading: boolean) => void
}

export const useInstanceStore = create<InstanceState>((set) => ({
  instances: [],
  currentInstance: null,
  loading: false,
  setInstances: (instances) => set({ instances }),
  setCurrentInstance: (name) => set({ currentInstance: name }),
  setLoading: (loading) => set({ loading }),
}))
