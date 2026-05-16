"use client";

import { FormEvent, Suspense, useState } from "react";
import { useRouter, useSearchParams } from "next/navigation";

import { ProtectedRoute } from "@/components/auth/protected-route";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { useToast } from "@/components/ui/toast";
import { createBooking } from "@/lib/api";

export default function NewBookingPage() {
  return (
    <Suspense fallback={<main className="mx-auto max-w-md px-4 py-10 text-muted">Загрузка...</main>}>
      <ProtectedRoute>
        <NewBookingForm />
      </ProtectedRoute>
    </Suspense>
  );
}

function NewBookingForm() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const [form, setForm] = useState({
    spotId: searchParams.get("spotId") ?? "",
    start: "",
    end: "",
    plate: "",
  });
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);
  const { toast } = useToast();

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
      toast({ title: "Бронь создана" });
      router.push("/dashboard");
    } catch (err) {
      const message = err instanceof Error ? err.message : "Не удалось создать бронь";
      setError(message);
      toast({ title: "Ошибка бронирования", description: message, variant: "error" });
    } finally {
      setLoading(false);
    }
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
