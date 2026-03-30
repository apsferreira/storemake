import { createContext, useContext, useState, useEffect, type ReactNode } from 'react'
import { isAuthenticated, clearToken, setToken, getUserFromToken } from '../api/client'

interface AuthState {
  authenticated: boolean
  userId: string | null
  storeId: string | null
}

interface AuthContextType extends AuthState {
  login: (token: string) => void
  logout: () => void
}

const AuthContext = createContext<AuthContextType | null>(null)

export function AuthProvider({ children }: { children: ReactNode }) {
  const [state, setState] = useState<AuthState>(() => {
    const authed = isAuthenticated()
    const user = authed ? getUserFromToken() : null
    return {
      authenticated: authed,
      userId: user?.sub ?? null,
      storeId: (user as Record<string, string>)?.store_id ?? null,
    }
  })

  useEffect(() => {
    const interval = setInterval(() => {
      if (!isAuthenticated() && state.authenticated) {
        setState({ authenticated: false, userId: null, storeId: null })
      }
    }, 60_000)
    return () => clearInterval(interval)
  }, [state.authenticated])

  const login = (token: string) => {
    setToken(token)
    const user = getUserFromToken()
    setState({
      authenticated: true,
      userId: user?.sub ?? null,
      storeId: (user as Record<string, string>)?.store_id ?? null,
    })
  }

  const logout = () => {
    clearToken()
    setState({ authenticated: false, userId: null, storeId: null })
  }

  return (
    <AuthContext.Provider value={{ ...state, login, logout }}>
      {children}
    </AuthContext.Provider>
  )
}

export function useAuth() {
  const ctx = useContext(AuthContext)
  if (!ctx) throw new Error('useAuth deve ser usado dentro de AuthProvider')
  return ctx
}
