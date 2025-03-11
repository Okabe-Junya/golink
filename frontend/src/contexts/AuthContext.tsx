import type React from "react"
import { createContext, useState, useEffect, useContext, useCallback } from "react"
import axios from "axios"

// Get API base URL from environment variable
const API_BASE_URL = import.meta.env.VITE_API_URL || "http://localhost:8080/api"

// Define User type
interface User {
  id: string
  email: string
  name: string
  picture: string
}

// Define AuthContext type
interface AuthContextType {
  user: User | null
  loading: boolean
  error: string | null
  login: () => void
  logout: () => void
  isAuthenticated: boolean
}

// Create auth context
const AuthContext = createContext<AuthContextType>({
  user: null,
  loading: false,
  error: null,
  login: () => {},
  logout: () => {},
  isAuthenticated: false,
})

// Auth provider component
export const AuthProvider: React.FC<{ children: React.ReactNode }> = ({
  children,
}) => {
  const [user, setUser] = useState<User | null>(null)
  const [loading, setLoading] = useState<boolean>(true)
  const [error, setError] = useState<string | null>(null)

  // Function to fetch user information
  const fetchUserInfo = useCallback(async () => {
    try {
      setLoading(true)
      setError(null)
      const response = await axios.get(`${API_BASE_URL}/auth/user`, {
        withCredentials: true, // Include cookies
      })
      setUser(response.data)
      return true
    } catch (err) {
      // For 401 errors, user is just not logged in, don't treat as error
      if (axios.isAxiosError(err) && err.response?.status === 401) {
        setUser(null)
        return false
      }
      setError("Failed to fetch authentication information")
      console.error("Failed to fetch user info:", err)
      return false
    } finally {
      setLoading(false)
    }
  }, []) // Empty dependency array since all dependencies are stable

  // Fetch user information on initial load
  useEffect(() => {
    fetchUserInfo()
  }, [fetchUserInfo])

  // Login handling
  const login = () => {
    // Redirect to backend login endpoint
    window.location.href = `${API_BASE_URL}/auth/login`
  }

  // Logout handling
  const logout = async () => {
    // Just remove session on client side for now (should have backend logout API in production)
    document.cookie =
      "session_token=; expires=Thu, 01 Jan 1970 00:00:00 UTC; path=/;"
    setUser(null)
  }

  // Calculate authentication state
  const isAuthenticated = !!user

  // Create context value
  const contextValue: AuthContextType = {
    user,
    loading,
    error,
    login,
    logout,
    isAuthenticated,
  }

  return (
    <AuthContext.Provider value={contextValue}>{children}</AuthContext.Provider>
  )
}

// Custom hook
export const useAuth = () => useContext(AuthContext)
export default AuthContext
