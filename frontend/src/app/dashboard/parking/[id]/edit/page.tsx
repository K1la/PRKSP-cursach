"use client";

import { FormEvent, useEffect, useState } from "react";
import { useParams, useRouter } from "next/navigation";

import { ProtectedRoute } from "@/components/auth/protected-route";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { useToast } from "@/components/ui/toast";
import { getParkingLot, updateParkingLot } from "@/lib/api";

export default function EditParkingPage() {
  const params = useParams<{ id: string }>();
  const router = useRouter();
  const [error, setError] = useState("");
  const { toast } = useToast();
  const [form, setForm] = useState({
    name: "",
    description: "",
    address: "",
    latitude: "",
    longitude: "",
    total_spots: "",
    price_per_hour: "",
    is_active: true,
  });

  useEffect(() => {
    getParkingLot(params.id)
      .then((lot) =>
        setForm({
          name: lot.name,
          description: lot.description ?? "",
          address: lot.address,
          latitude: String(lot.latitude),
          longitude: String(lot.longitude),
          total_spots: String(lot.total_spots),
          price_per_hour: String(lot.price_per_hour),
          is_active: lot.is_active,
        }),
      )
      .catch((err) => setError(err instanceof Error ? err.message : "Не удалось загрузить парковку"));
  }, [params.id]);

  async function handleSubmit(event: FormEvent) {
    event.preventDefault();
    setError("");
    try {
      await updateParkingLot(params.id, {
        name: form.name,
        description: form.description || null,
        address: form.address,
        latitude: Number(form.latitude),
        longitude: Number(form.longitude),
        total_spots: Number(form.total_spots),
        price_per_hour: Number(form.price_per_hour),
        is_active: form.is_active,
      });
      toast({ title: "Парковка обновлена" });
      router.push(`/parking/${params.id}`);
    } catch (err) {
      const message = err instanceof Error ? err.message : "Не удалось обновить парковку";
      setError(message);
      toast({ title: "Ошибка", description: message, variant: "error" });
    }
  }

  return (
    <ProtectedRoute roles={["owner", "admin"]}>
    <main className="mx-auto max-w-2xl px-4 py-8">
      <Card>
        <CardHeader>
          <CardTitle>Редактирование парковки</CardTitle>
        </CardHeader>
        <CardContent>
          <form className="grid gap-4" onSubmit={handleSubmit}>
            <Input value={form.name} onChange={(event) => setForm({ ...form, name: event.target.value })} />
            <Input value={form.address} onChange={(event) => setForm({ ...form, address: event.target.value })} />
            <Textarea value={form.description} onChange={(event) => setForm({ ...form, description: event.target.value })} />
            <div className="grid gap-3 sm:grid-cols-2">
              <Input value={form.latitude} onChange={(event) => setForm({ ...form, latitude: event.target.value })} />
              <Input value={form.longitude} onChange={(event) => setForm({ ...form, longitude: event.target.value })} />
              <Input value={form.total_spots} onChange={(event) => setForm({ ...form, total_spots: event.target.value })} />
              <Input value={form.price_per_hour} onChange={(event) => setForm({ ...form, price_per_hour: event.target.value })} />
            </div>
            <label className="flex items-center gap-2 text-sm">
              <input
                checked={form.is_active}
                onChange={(event) => setForm({ ...form, is_active: event.target.checked })}
                type="checkbox"
              />
              Активна
            </label>
            {error ? <p className="text-sm text-red-600">{error}</p> : null}
            <Button>Сохранить</Button>
          </form>
        </CardContent>
      </Card>
    </main>
    </ProtectedRoute>
  );
}
