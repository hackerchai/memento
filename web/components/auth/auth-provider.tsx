"use client";

import { createContext, useContext, useEffect, useState } from "react";
import { authAPI, userAPI } from "@/lib/api-client";

// User type definition
interface User {
  id: string;
  name: string;
  email: string;
  created_at?: string;
  role?: string;
}

// Authentication context type
interface AuthContextType {
  user: User | null;
  isLoading: boolean;
  isAuthenticated: boolean;
  login: (email: string, password: string) => Promise<void>;
  logout: () => void;
  register: (name: string, email: string, password: string) => Promise<void>;
}

// Create authentication context
const AuthContext = createContext<AuthContextType | undefined>(undefined);

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [user, setUser] = useState<User | null>(null);
  const [isLoading, setIsLoading] = useState<boolean>(true);
  const [isAuthenticated, setIsAuthenticated] = useState<boolean>(false);

  // Check token from local storage
  useEffect(() => {
    const checkAuth = async () => {
      setIsLoading(true);
      try {
        const token = localStorage.getItem("mementoToken");
        
        if (!token) {
          setUser(null);
          setIsAuthenticated(false);
          return;
        }
        
        // Validate token validity
        const profileResponse = await userAPI.getProfile();
        if (profileResponse?.data) {
          setUser(profileResponse.data.data);
          setIsAuthenticated(true);
        } else {
          // Invalid token, clear it
          localStorage.removeItem("mementoToken");
          setUser(null);
          setIsAuthenticated(false);
        }
      } catch (error) {
        console.error("Authentication check failed:", error);
        localStorage.removeItem("mementoToken");
        setUser(null);
        setIsAuthenticated(false);
      } finally {
        setIsLoading(false);
      }
    };

    checkAuth();
  }, []);

  // Login function
  const login = async (email: string, password: string) => {
    setIsLoading(true);
    try {
      const response = await authAPI.login(email, password);
      
      if (response?.data?.token) {
        // Save token
        localStorage.setItem("mementoToken", response.data.token);
        setUser(response.data.user);
        setIsAuthenticated(true);
      } else {
        throw new Error("Login failed: Invalid response");
      }
    } catch (error) {
      console.error("Login failed:", error);
      throw error;
    } finally {
      setIsLoading(false);
    }
  };

  // Logout function
  const logout = () => {
    localStorage.removeItem("mementoToken");
    setUser(null);
    setIsAuthenticated(false);
  };

  // Register function
  const register = async (name: string, email: string, password: string) => {
    setIsLoading(true);
    try {
      const response = await authAPI.register(name, email, password);
      
      if (response?.data) {
        // Auto login after successful registration
        await login(email, password);
      } else {
        throw new Error("Registration failed: Invalid response");
      }
    } catch (error) {
      console.error("Registration failed:", error);
      throw error;
    } finally {
      setIsLoading(false);
    }
  };

  const value = {
    user,
    isLoading,
    isAuthenticated,
    login,
    logout,
    register
  };

  return (
    <AuthContext.Provider value={value}>
      {children}
    </AuthContext.Provider>
  );
}

// Provide useAuth hook for components
export function useAuth() {
  const context = useContext(AuthContext);
  if (context === undefined) {
    throw new Error("useAuth must be used within an AuthProvider");
  }
  return context;
}