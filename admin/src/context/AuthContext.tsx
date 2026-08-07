import { createContext, useContext, useState, useEffect, type ReactNode } from 'react';
import { authService, type LoginCredentials, type RegisterCredentials } from '../services/auth';

interface User {
  id: string;
  email: string;
  name?: string;
  status: string;
}

interface AuthContextType {
  user: User | null;
  accessToken: string | null;
  refreshToken: string | null;
  loading: boolean;
  login: (credentials: LoginCredentials) => Promise<void>;
  register: (credentials: RegisterCredentials) => Promise<void>;
  logout: () => Promise<void>;
  isAuthenticated: boolean;
}

const AuthContext = createContext<AuthContextType | null>(null);

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null);
  const [accessToken, setAccessToken] = useState<string | null>(null);
  const [refreshToken, setRefreshToken] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    // Check for existing session on mount
    const storedToken = localStorage.getItem('access_token');
    const storedRefreshToken = localStorage.getItem('refresh_token');
    
    if (storedToken && storedRefreshToken) {
      setAccessToken(storedToken);
      setRefreshToken(storedRefreshToken);
      // TODO: Fetch user info with token
      setUser({ id: '1', email: 'user@example.com', status: 'active' });
    }
    setLoading(false);
  }, []);

  const login = async (credentials: LoginCredentials) => {
    const response = await authService.login(credentials);
    setAccessToken(response.access_token);
    setRefreshToken(response.refresh_token);
    localStorage.setItem('access_token', response.access_token);
    localStorage.setItem('refresh_token', response.refresh_token);
    // TODO: Fetch user info
    setUser({ id: '1', email: credentials.email, status: 'active' });
  };

  const register = async (credentials: RegisterCredentials) => {
    await authService.register(credentials);
    // Register returns user_id, not access_token
    // Need to login after registration
    await login({ email: credentials.email, password: credentials.password });
  };

  const logout = async () => {
    if (accessToken && refreshToken) {
      try {
        await authService.logout(accessToken, refreshToken);
      } catch (error) {
        console.error('Logout error:', error);
      }
    }
    setAccessToken(null);
    setRefreshToken(null);
    setUser(null);
    localStorage.removeItem('access_token');
    localStorage.removeItem('refresh_token');
  };

  const value: AuthContextType = {
    user,
    accessToken,
    refreshToken,
    loading,
    login,
    register,
    logout,
    isAuthenticated: !!accessToken,
  };

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth() {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error('useAuth must be used within AuthProvider');
  }
  return context;
}
