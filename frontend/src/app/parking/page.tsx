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
import type { ParkingLot } from "@/types/parking";

const ParkingMap = dynamic(() => import("@/components/parking/parking-map"), {
  ssr: false,
  loading: () => <div className="h-full w-full animate-pulse bg-slate-200" />,
});

type ParkingFilters = {
  query: string;
  maxPrice: string;
  spotType: string;
};

const initialFilters: ParkingFilters = { query: "", maxPrice: "", spotType: "" };

export default function ParkingPage() {
  const [lots, setLots] = useState<ParkingLot[]>([]);
  const [filters, setFilters] = useState<ParkingFilters>(initialFilters);
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(true);

  const load = useCallback(async (nextFilters: ParkingFilters) => {
    setLoading(true);
    setError("");
    try {
      const params = new URLSearchParams();
      if (nextFilters.query) {
        params.set("q", nextFilters.query);
      }
      if (nextFilters.maxPrice) {
        params.set("max_price", nextFilters.maxPrice);
      }
      if (nextFilters.spotType) {
        params.set("spot_type", nextFilters.spotType);
      }
      params.set("limit", "50");
      const data = await listParkingLots(params);
      setLots(data);
    } catch (err) {
      setError(err instanceof Error ? err.message : "API недоступен");
      setLots([]);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    load(initialFilters);
  }, [load]);

  function handleSearch(event: FormEvent) {
    event.preventDefault();
    load(filters);
  }

  return (
    <main className="mx-auto max-w-7xl px-4 py-8 sm:px-6 lg:px-8">
      <div className="mb-6 flex flex-col gap-4 lg:flex-row lg:items-end lg:justify-between">
        <div>
          <h1 className="text-3xl font-semibold">Парковки</h1>
          <p className="mt-2 text-muted">Поиск по адресу, названию и карте.</p>
        </div>
        <form className="grid gap-2 sm:grid-cols-[260px_140px_150px_120px]" onSubmit={handleSearch}>
          <Input
            value={filters.query}
            onChange={(event) => setFilters({ ...filters, query: event.target.value })}
            placeholder="Адрес или название"
          />
          <Input
            min={0}
            type="number"
            value={filters.maxPrice}
            onChange={(event) => setFilters({ ...filters, maxPrice: event.target.value })}
            placeholder="До ₽/ч"
          />
          <select
            className="h-10 rounded border border-border bg-white px-3 text-sm outline-none focus:border-primary"
            value={filters.spotType}
            onChange={(event) => setFilters({ ...filters, spotType: event.target.value })}
          >
            <option value="">Любой тип</option>
            <option value="standard">standard</option>
            <option value="disabled">disabled</option>
            <option value="electric">electric</option>
            <option value="vip">vip</option>
          </select>
          <Button disabled={loading}>
            <Search className="mr-2 size-4" aria-hidden="true" />
            Найти
          </Button>
        </form>
      </div>

      <div className="grid gap-6 lg:grid-cols-[minmax(0,1fr)_460px]">
        <div className="grid gap-4 md:grid-cols-2">
          {loading
            ? Array.from({ length: 4 }).map((_, index) => (
                <div key={index} className="h-44 animate-pulse rounded border border-border bg-white" />
              ))
            : lots.map((lot) => (
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
          <ParkingMap parkingLots={lots} />
        </div>
      </div>
    </main>
  );
}
