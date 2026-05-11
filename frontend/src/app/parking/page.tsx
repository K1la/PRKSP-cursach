"use client";

import dynamic from "next/dynamic";
import Link from "next/link";
import { FormEvent, useCallback, useEffect, useState } from "react";
import { MapPin, Search } from "lucide-react";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { listParkingLots } from "@/lib/api";
import { mockParkingLots } from "@/lib/mock-data";
import type { ParkingLot } from "@/types/parking";

const ParkingMap = dynamic(() => import("@/components/parking/parking-map"), {
  ssr: false,
  loading: () => <div className="h-full w-full animate-pulse bg-slate-200" />,
});

export default function ParkingPage() {
  const [lots, setLots] = useState<ParkingLot[]>([]);
  const [query, setQuery] = useState("");
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(true);

  const load = useCallback(async (search = "") => {
    setLoading(true);
    setError("");
    try {
      const params = new URLSearchParams();
      if (search) {
        params.set("q", search);
      }
      params.set("limit", "50");
      const data = await listParkingLots(params);
      setLots(data);
    } catch (err) {
      setError(err instanceof Error ? err.message : "API недоступен, показаны demo-данные");
      setLots(mockParkingLots);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    load("");
  }, [load]);

  function handleSearch(event: FormEvent) {
    event.preventDefault();
    load(query);
  }

  return (
    <main className="mx-auto max-w-7xl px-4 py-8 sm:px-6 lg:px-8">
      <div className="mb-6 flex flex-col gap-4 lg:flex-row lg:items-end lg:justify-between">
        <div>
          <h1 className="text-3xl font-semibold">Парковки</h1>
          <p className="mt-2 text-muted">Поиск по адресу, названию и карте.</p>
        </div>
        <form className="grid gap-2 sm:grid-cols-[320px_120px]" onSubmit={handleSearch}>
          <Input value={query} onChange={(event) => setQuery(event.target.value)} placeholder="Адрес или название" />
          <Button disabled={loading}>
            <Search className="mr-2 size-4" aria-hidden="true" />
            Найти
          </Button>
        </form>
      </div>

      <div className="grid gap-6 lg:grid-cols-[minmax(0,1fr)_460px]">
        <div className="grid gap-4 md:grid-cols-2">
          {lots.map((lot) => (
            <Card key={lot.id}>
              <CardHeader>
                <div className="flex items-start justify-between gap-3">
                  <CardTitle>{lot.name}</CardTitle>
                  <Badge>{lot.price_per_hour} ₽/ч</Badge>
                </div>
              </CardHeader>
              <CardContent>
                <p className="flex items-center gap-2 text-sm text-muted">
                  <MapPin className="size-4" aria-hidden="true" />
                  {lot.address}
                </p>
                <p className="mt-3 text-sm text-slate-700">{lot.total_spots} мест</p>
                <Button className="mt-4 w-full" asChild>
                  <Link href={`/parking/${lot.id}`}>Открыть</Link>
                </Button>
              </CardContent>
            </Card>
          ))}
          {!loading && lots.length === 0 ? <p className="text-muted">Парковки не найдены.</p> : null}
          {error ? <p className="text-sm text-amber-700">{error}</p> : null}
        </div>
        <div className="h-[520px] overflow-hidden rounded border border-border bg-slate-100">
          <ParkingMap parkingLots={lots.length ? lots : mockParkingLots} />
        </div>
      </div>
    </main>
  );
}
