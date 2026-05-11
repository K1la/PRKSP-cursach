"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { Car, LogOut } from "lucide-react";

import { useAuth } from "@/components/auth/auth-provider";
import { Button } from "@/components/ui/button";

export function Navbar() {
  const router = useRouter();
  const { user, logout } = useAuth();

  return (
    <header className="border-b border-border bg-white">
      <div className="mx-auto flex max-w-7xl items-center justify-between px-4 py-4 sm:px-6 lg:px-8">
        <Link href="/" className="flex items-center gap-3">
          <div className="flex size-10 items-center justify-center rounded bg-primary text-primary-foreground">
            <Car className="size-5" aria-hidden="true" />
          </div>
          <div>
            <p className="text-lg font-semibold leading-none">ParkEase</p>
            <p className="text-sm text-muted">Поиск и бронирование парковок</p>
          </div>
        </Link>

        <nav className="hidden items-center gap-6 text-sm font-medium text-slate-700 md:flex">
          <Link href="/parking">Парковки</Link>
          <Link href="/dashboard">Кабинет</Link>
          {user?.role === "admin" ? <span className="text-primary">admin</span> : null}
          {user?.role === "owner" ? <span className="text-primary">owner</span> : null}
        </nav>

        <div className="flex items-center gap-2">
          {user ? (
            <>
              <span className="hidden text-sm text-muted sm:inline">{user.name}</span>
              <Button
                variant="outline"
                onClick={() => {
                  logout();
                  router.push("/");
                }}
              >
                <LogOut className="mr-2 size-4" aria-hidden="true" />
                Выйти
              </Button>
            </>
          ) : (
            <>
              <Button variant="outline" onClick={() => router.push("/login")}>
                Войти
              </Button>
              <Button onClick={() => router.push("/register")}>Регистрация</Button>
            </>
          )}
        </div>
      </div>
    </header>
  );
}
