"use client";

import { createContext, useContext, useEffect, useState } from "react";
import { authAPI, userAPI } from "@/lib/api-client";
import { useRouter, usePathname } from "next/navigation";

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
  error: string | null;
  login: (email: string, password: string) => Promise<void>;
  logout: () => void;
  register: (name: string, email: string, password: string) => Promise<void>;
  clearError: () => void;
}

// Public routes that don't require authentication
const publicRoutes = ["/", "/login", "/register"];

// Create authentication context
const AuthContext = createContext<AuthContextType | undefined>(undefined);

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [user, setUser] = useState<User | null>(null);
  const [isLoading, setIsLoading] = useState<boolean>(true);
  const [isAuthenticated, setIsAuthenticated] = useState<boolean>(false);
  const [error, setError] = useState<string | null>(null);
  const router = useRouter();
  const pathname = usePathname();

  // Clear error function
  const clearError = () => setError(null);

  // Check token from local storage
  useEffect(() => {
    const checkAuth = async () => {
      setIsLoading(true);
      try {
        const token = localStorage.getItem("mementoToken");
        
        if (!token) {
          setUser(null);
          setIsAuthenticated(false);
          // Redirect to login if not on a public route
          if (!publicRoutes.includes(pathname as string) && pathname !== null) {
            router.push("/login");
          }
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
          // Redirect to login if not on a public route
          if (!publicRoutes.includes(pathname as string) && pathname !== null) {
            router.push("/login");
          }
        }
      } catch (error) {
        console.error("Authentication check failed:", error);
        localStorage.removeItem("mementoToken");
        setUser(null);
        setIsAuthenticated(false);
        // Redirect to login if not on a public route
        if (!publicRoutes.includes(pathname as string) && pathname !== null) {
          router.push("/login");
        }
      } finally {
        setIsLoading(false);
      }
    };

    checkAuth();
  }, [pathname, router]);

  // Login function
  const login = async (email: string, password: string) => {
    setIsLoading(true);
    clearError();
    try {
      const response = await authAPI.login(email, password);
      
      if (response?.data?.token) {
        // Save token
        localStorage.setItem("mementoToken", response.data.token);
        setUser(response.data.user);
        setIsAuthenticated(true);
        // Redirect to articles page after successful login
        router.push("/articles");
      } else {
        throw new Error("Login failed: Invalid response");
      }
    } catch (error: any) {
      console.error("Login failed:", error);
      setError(error.message || "Login failed, please check your username and password");
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
    router.push("/login");
  };

  // Register function
  const register = async (name: string, email: string, password: string) => {
    setIsLoading(true);
    clearError();
    try {
      const response = await authAPI.register(name, email, password);
      
      if (response?.data) {
        // Auto login after successful registration
        await login(email, password);
      } else {
        throw new Error("Registration failed: Invalid response");
      }
    } catch (error: any) {
      console.error("Registration failed:", error);
      setError(error.message || "Registration failed, please try again later");
      throw error;
    } finally {
      setIsLoading(false);
    }
  };

  const value = {
    user,
    isLoading,
    isAuthenticated,
    error,
    login,
    logout,
    register,
    clearError
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