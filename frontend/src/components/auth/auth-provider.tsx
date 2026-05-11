"use client";

import { createContext, useContext, useEffect, useMemo, useState } from "react";

import {
  clearTokens,
  login as loginRequest,
  me,
  register as registerRequest,
  storeTokens,
} from "@/lib/api";
import type { User } from "@/types/parking";

type AuthContextValue = {
  user: User | null;
  loading: boolean;
  login: (email: string, password: string) => Promise<void>;
  register: (payload: { email: string; password: string; name: string; phone?: string }) => Promise<void>;
  logout: () => void;
  refreshUser: () => Promise<void>;
};

const AuthContext = createContext<AuthContextValue | null>(null);

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState(true);

  async function refreshUser() {
    try {
      const currentUser = await me();
      setUser(currentUser);
    } catch {
      clearTokens();
      setUser(null);
    }
  }

  useEffect(() => {
    refreshUser().finally(() => setLoading(false));
  }, []);

  const value = useMemo<AuthContextValue>(
    () => ({
      user,
      loading,
      async login(email, password) {
        const response = await loginRequest({ email, password });
        storeTokens(response);
        setUser(response.user);
      },
      async register(payload) {
        const response = await registerRequest(payload);
        storeTokens(response);
        setUser(response.user);
      },
      logout() {
        clearTokens();
        setUser(null);
      },
      refreshUser,
    }),
    [loading, user],
  );

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth() {
  const value = useContext(AuthContext);
  if (!value) {
    throw new Error("useAuth must be used inside AuthProvider");
  }
  return value;
}
