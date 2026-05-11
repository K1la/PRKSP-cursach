"use client";

import { FormEvent, useState } from "react";
import { Suspense } from "react";
import { useRouter, useSearchParams } from "next/navigation";

import { useAuth } from "@/components/auth/auth-provider";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { createBooking } from "@/lib/api";

export default function NewBookingPage() {
  return (
    <Suspense fallback={<main className="mx-auto max-w-md px-4 py-10 text-muted">Загрузка...</main>}>
      <NewBookingForm />
    </Suspense>
  );
}

function NewBookingForm() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const { user } = useAuth();
  const [form, setForm] = useState({
    spotId: searchParams.get("spotId") ?? "",
    start: "",
    end: "",
    plate: "",
  });
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);

  async function handleSubmit(event: FormEvent) {
    event.preventDefault();
    setError("");
    setLoading(true);
    try {
      await createBooking({
        parking_spot_id: form.spotId,
        start_time: new Date(form.start).toISOString(),
        end_time: new Date(form.end).toISOString(),
        vehicle_plate: form.plate,
      });
      router.push("/dashboard");
    } catch (err) {
      setError(err instanceof Error ? err.message : "Не удалось создать бронь");
    } finally {
      setLoading(false);
    }
  }

  if (!user) {
    return (
      <main className="mx-auto max-w-md px-4 py-10">
        <Card>
          <CardContent className="pt-4">
            <p className="mb-4 text-muted">Для бронирования нужно войти.</p>
            <Button onClick={() => router.push("/login")}>Войти</Button>
          </CardContent>
        </Card>
      </main>
    );
  }

  return (
    <main className="mx-auto max-w-md px-4 py-10">
      <Card>
        <CardHeader>
          <CardTitle>Новая бронь</CardTitle>
        </CardHeader>
        <CardContent>
          <form className="space-y-4" onSubmit={handleSubmit}>
            <Input
              value={form.spotId}
              onChange={(event) => setForm({ ...form, spotId: event.target.value })}
              placeholder="ID места"
            />
            <Input
              type="datetime-local"
              value={form.start}
              onChange={(event) => setForm({ ...form, start: event.target.value })}
            />
            <Input
              type="datetime-local"
              value={form.end}
              onChange={(event) => setForm({ ...form, end: event.target.value })}
            />
            <Input
              value={form.plate}
              onChange={(event) => setForm({ ...form, plate: event.target.value })}
              placeholder="Номер авто"
            />
            {error ? <p className="text-sm text-red-600">{error}</p> : null}
            <Button className="w-full" disabled={loading}>
              {loading ? "Создание..." : "Забронировать"}
            </Button>
          </form>
        </CardContent>
      </Card>
    </main>
  );
}
