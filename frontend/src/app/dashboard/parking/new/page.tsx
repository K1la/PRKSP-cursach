"use client";

import { FormEvent, useState } from "react";
import { useRouter } from "next/navigation";

import { ProtectedRoute } from "@/components/auth/protected-route";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { useToast } from "@/components/ui/toast";
import { createParkingLot } from "@/lib/api";

export default function NewParkingPage() {
  const router = useRouter();
  const [error, setError] = useState("");
  const { toast } = useToast();
  const [form, setForm] = useState({
    name: "",
    description: "",
    address: "",
    latitude: "55.7558",
    longitude: "37.6173",
    total_spots: "20",
    price_per_hour: "150",
  });

  async function handleSubmit(event: FormEvent) {
    event.preventDefault();
    setError("");
    try {
      const lot = await createParkingLot({
        name: form.name,
        description: form.description || null,
        address: form.address,
        latitude: Number(form.latitude),
        longitude: Number(form.longitude),
        total_spots: Number(form.total_spots),
        price_per_hour: Number(form.price_per_hour),
      });
      toast({ title: "Парковка создана" });
      router.push(`/parking/${lot.id}`);
    } catch (err) {
      const message = err instanceof Error ? err.message : "Не удалось создать парковку";
      setError(message);
      toast({ title: "Ошибка", description: message, variant: "error" });
    }
  }

  return (
    <ProtectedRoute roles={["owner", "admin"]}>
    <main className="mx-auto max-w-2xl px-4 py-8">
      <Card>
        <CardHeader>
          <CardTitle>Создание парковки</CardTitle>
        </CardHeader>
        <CardContent>
          <form className="grid gap-4" onSubmit={handleSubmit}>
            <Input value={form.name} onChange={(event) => setForm({ ...form, name: event.target.value })} placeholder="Название" />
            <Input value={form.address} onChange={(event) => setForm({ ...form, address: event.target.value })} placeholder="Адрес" />
            <Textarea
              value={form.description}
              onChange={(event) => setForm({ ...form, description: event.target.value })}
              placeholder="Описание"
            />
            <div className="grid gap-3 sm:grid-cols-2">
              <Input value={form.latitude} onChange={(event) => setForm({ ...form, latitude: event.target.value })} placeholder="Широта" />
              <Input value={form.longitude} onChange={(event) => setForm({ ...form, longitude: event.target.value })} placeholder="Долгота" />
              <Input value={form.total_spots} onChange={(event) => setForm({ ...form, total_spots: event.target.value })} placeholder="Мест" />
              <Input value={form.price_per_hour} onChange={(event) => setForm({ ...form, price_per_hour: event.target.value })} placeholder="Цена/час" />
            </div>
            {error ? <p className="text-sm text-red-600">{error}</p> : null}
            <Button>Создать</Button>
          </form>
        </CardContent>
      </Card>
    </main>
    </ProtectedRoute>
  );
}
