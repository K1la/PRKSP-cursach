"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { useCallback, useEffect, useState } from "react";

import { useAuth } from "@/components/auth/auth-provider";
import { ProtectedRoute } from "@/components/auth/protected-route";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Dialog } from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { useToast } from "@/components/ui/toast";
import { adminStats, cancelBooking, listMyBookings, listParkingLots, listUsers, updateMe } from "@/lib/api";
import type { AdminStats, Booking, ParkingLot, User } from "@/types/parking";

export default function DashboardPage() {
  const router = useRouter();
  const { user, loading, refreshUser } = useAuth();
  const [bookings, setBookings] = useState<Booking[]>([]);
  const [lots, setLots] = useState<ParkingLot[]>([]);
  const [users, setUsers] = useState<User[]>([]);
  const [stats, setStats] = useState<AdminStats | null>(null);
  const [profile, setProfile] = useState({ name: "", phone: "" });
  const [error, setError] = useState("");
  const [bookingToCancel, setBookingToCancel] = useState<string | null>(null);
  const { toast } = useToast();

  const load = useCallback(async () => {
    if (!user) {
      return;
    }
    setError("");
    try {
      setProfile({ name: user.name, phone: user.phone ?? "" });
      const [bookingData, parkingData] = await Promise.all([listMyBookings(), listParkingLots()]);
      setBookings(bookingData);
      setLots(parkingData);

      if (user.role === "admin") {
        const [statsData, usersData] = await Promise.all([adminStats(), listUsers()]);
        setStats(statsData);
        setUsers(usersData);
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : "Не удалось загрузить кабинет");
    }
  }, [user]);

  useEffect(() => {
    if (!loading && !user) {
      router.push("/login");
      return;
    }
    load();
  }, [load, loading, router, user]);

  async function saveProfile() {
    setError("");
    try {
      await updateMe({ name: profile.name, phone: profile.phone || null });
      await refreshUser();
      toast({ title: "Профиль обновлен" });
    } catch (err) {
      const message = err instanceof Error ? err.message : "Не удалось обновить профиль";
      setError(message);
      toast({ title: "Ошибка", description: message, variant: "error" });
    }
  }

  async function handleCancel(id: string) {
    try {
      await cancelBooking(id);
      await load();
      toast({ title: "Бронь отменена" });
    } catch (err) {
      toast({ title: "Ошибка", description: err instanceof Error ? err.message : "Не удалось отменить бронь", variant: "error" });
    } finally {
      setBookingToCancel(null);
    }
  }

  if (loading || !user) {
    return <main className="mx-auto max-w-7xl px-4 py-8 text-muted">Загрузка...</main>;
  }

  const ownedLots = user.role === "admin" ? lots : lots.filter((lot) => lot.owner_id === user.id);

  return (
    <ProtectedRoute>
    <main className="mx-auto max-w-7xl px-4 py-8 sm:px-6 lg:px-8">
      <div className="mb-6 flex items-start justify-between gap-4">
        <div>
          <h1 className="text-3xl font-semibold">Личный кабинет</h1>
          <p className="mt-2 text-muted">{user.email}</p>
        </div>
        <Badge>{user.role}</Badge>
      </div>

      {error ? <p className="mb-4 text-sm text-red-600">{error}</p> : null}

      {stats ? (
        <section className="mb-6 grid gap-4 md:grid-cols-4">
          <StatCard title="Пользователи" value={stats.total_users} />
          <StatCard title="Брони" value={stats.total_bookings} />
          <StatCard title="Парковки" value={stats.total_parking_lots} />
          <StatCard title="Выручка" value={`${stats.revenue.toFixed(0)} ₽`} />
        </section>
      ) : null}

      <div className="grid gap-6 lg:grid-cols-[360px_1fr]">
        <aside className="space-y-4">
          <Card>
            <CardHeader>
              <CardTitle>Профиль</CardTitle>
            </CardHeader>
            <CardContent className="space-y-3">
              <Input value={profile.name} onChange={(event) => setProfile({ ...profile, name: event.target.value })} />
              <Input
                value={profile.phone}
                onChange={(event) => setProfile({ ...profile, phone: event.target.value })}
                placeholder="Телефон"
              />
              <Button className="w-full" onClick={saveProfile}>
                Сохранить
              </Button>
            </CardContent>
          </Card>

          {(user.role === "owner" || user.role === "admin") && (
            <Card>
              <CardHeader>
                <CardTitle>Управление парковками</CardTitle>
              </CardHeader>
              <CardContent>
                <Button className="w-full" asChild>
                  <Link href="/dashboard/parking/new">Создать парковку</Link>
                </Button>
              </CardContent>
            </Card>
          )}
        </aside>

        <section className="space-y-6">
          <Card>
            <CardHeader>
              <CardTitle>Мои бронирования</CardTitle>
            </CardHeader>
            <CardContent className="space-y-3">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Авто</TableHead>
                    <TableHead>Период</TableHead>
                    <TableHead>Статус</TableHead>
                    <TableHead />
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {bookings.map((booking) => (
                    <TableRow key={booking.id}>
                      <TableCell className="font-medium">{booking.vehicle_plate}</TableCell>
                      <TableCell>
                        {new Date(booking.start_time).toLocaleString()} - {new Date(booking.end_time).toLocaleString()}
                      </TableCell>
                      <TableCell>
                        <Badge>{booking.status}</Badge>
                      </TableCell>
                      <TableCell className="text-right">
                        <Button variant="outline" onClick={() => setBookingToCancel(booking.id)}>
                          Отменить
                        </Button>
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
              {bookings.length === 0 ? <p className="text-sm text-muted">Бронирований пока нет.</p> : null}
            </CardContent>
          </Card>

          {(user.role === "owner" || user.role === "admin") && (
            <Card>
              <CardHeader>
                <CardTitle>{user.role === "admin" ? "Все парковки" : "Мои парковки"}</CardTitle>
              </CardHeader>
              <CardContent className="space-y-3">
                {ownedLots.map((lot) => (
                  <div key={lot.id} className="flex items-center justify-between gap-3 border-b border-border pb-3">
                    <div>
                      <p className="font-medium">{lot.name}</p>
                      <p className="text-sm text-muted">{lot.address}</p>
                    </div>
                    <Button variant="outline" asChild>
                      <Link href={`/dashboard/parking/${lot.id}/edit`}>Изменить</Link>
                    </Button>
                  </div>
                ))}
              </CardContent>
            </Card>
          )}

          {user.role === "admin" && (
            <Card>
              <CardHeader>
                <CardTitle>Пользователи</CardTitle>
              </CardHeader>
              <CardContent>
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>Имя</TableHead>
                      <TableHead>Email</TableHead>
                      <TableHead>Роль</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {users.map((item) => (
                      <TableRow key={item.id}>
                        <TableCell className="font-medium">{item.name}</TableCell>
                        <TableCell>{item.email}</TableCell>
                        <TableCell>
                          <Badge>{item.role}</Badge>
                        </TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              </CardContent>
            </Card>
          )}
        </section>
      </div>
      <Dialog
        destructive
        open={bookingToCancel !== null}
        title="Отменить бронирование?"
        description="Действие изменит статус брони на cancelled, если она еще pending или active."
        confirmText="Отменить бронь"
        onClose={() => setBookingToCancel(null)}
        onConfirm={() => bookingToCancel && handleCancel(bookingToCancel)}
      />
    </main>
    </ProtectedRoute>
  );
}

function StatCard({ title, value }: { title: string; value: number | string }) {
  return (
    <Card>
      <CardContent className="pt-4">
        <p className="text-sm text-muted">{title}</p>
        <p className="mt-2 text-2xl font-semibold">{value}</p>
      </CardContent>
    </Card>
  );
}
