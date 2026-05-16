"use client";

import { useEffect } from "react";
import { useRouter } from "next/navigation";

import { useAuth } from "@/components/auth/auth-provider";
import type { UserRole } from "@/types/parking";

export function ProtectedRoute({
  roles,
  children,
}: {
  roles?: UserRole[];
  children: React.ReactNode;
}) {
  const router = useRouter();
  const { user, loading } = useAuth();
  const allowed = user && (!roles || roles.includes(user.role));

  useEffect(() => {
    if (!loading && !user) {
      router.push("/login");
    }
  }, [loading, router, user]);

  if (loading) {
    return <main className="mx-auto max-w-7xl px-4 py-8 text-muted">Загрузка...</main>;
  }
  if (!user) {
    return <main className="mx-auto max-w-7xl px-4 py-8 text-muted">Нужен вход в аккаунт.</main>;
  }
  if (!allowed) {
    return <main className="mx-auto max-w-7xl px-4 py-8 text-muted">Недостаточно прав для этого раздела.</main>;
  }
  return children;
}
