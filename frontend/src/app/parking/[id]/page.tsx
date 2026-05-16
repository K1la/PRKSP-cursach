"use client";

import Link from "next/link";
import { FormEvent, useCallback, useEffect, useState } from "react";
import { useParams } from "next/navigation";
import { MapPin, Star } from "lucide-react";
import dynamic from "next/dynamic";

import { useAuth } from "@/components/auth/auth-provider";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Textarea } from "@/components/ui/textarea";
import { useToast } from "@/components/ui/toast";
import { createReview, getParkingLot, listParkingSpots, listReviews } from "@/lib/api";
import type { ParkingLot, ParkingSpot, Review } from "@/types/parking";

const ParkingMap = dynamic(() => import("@/components/parking/parking-map"), {
  ssr: false,
  loading: () => <div className="h-full w-full animate-pulse bg-slate-200" />,
});

export default function ParkingDetailPage() {
  const params = useParams<{ id: string }>();
  const { user } = useAuth();
  const [lot, setLot] = useState<ParkingLot | null>(null);
  const [spots, setSpots] = useState<ParkingSpot[]>([]);
  const [reviews, setReviews] = useState<Review[]>([]);
  const [review, setReview] = useState({ rating: 5, comment: "" });
  const [error, setError] = useState("");
  const [tab, setTab] = useState("spots");
  const { toast } = useToast();

  const load = useCallback(async () => {
    try {
      const [lotData, spotData, reviewData] = await Promise.all([
        getParkingLot(params.id),
        listParkingSpots(params.id),
        listReviews(params.id),
      ]);
      setLot(lotData);
      setSpots(spotData);
      setReviews(reviewData);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Не удалось загрузить парковку");
    }
  }, [params.id]);

  useEffect(() => {
    load();
  }, [load]);

  async function handleReview(event: FormEvent) {
    event.preventDefault();
    setError("");
    try {
      await createReview(params.id, review);
      setReview({ rating: 5, comment: "" });
      await load();
      toast({ title: "Отзыв отправлен" });
    } catch (err) {
      const message = err instanceof Error ? err.message : "Не удалось оставить отзыв";
      setError(message);
      toast({ title: "Ошибка", description: message, variant: "error" });
    }
  }

  if (!lot) {
    return (
      <main className="mx-auto max-w-7xl px-4 py-8">
        <div className="h-72 animate-pulse rounded border border-border bg-white" />
        <p className="mt-4 text-muted">{error || "Загрузка..."}</p>
      </main>
    );
  }

  return (
    <main className="mx-auto max-w-7xl px-4 py-8 sm:px-6 lg:px-8">
      <div className="mb-6 flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between">
        <div>
          <div className="mb-3 flex items-center gap-3">
            <h1 className="text-3xl font-semibold">{lot.name}</h1>
            <Badge>{lot.price_per_hour} ₽/ч</Badge>
          </div>
          <p className="flex items-center gap-2 text-muted">
            <MapPin className="size-4" aria-hidden="true" />
            {lot.address}
          </p>
          <p className="mt-4 max-w-2xl text-slate-700">{lot.description || "Описание пока не добавлено."}</p>
        </div>
        {(user?.role === "owner" || user?.role === "admin") && (
          <Button variant="outline" asChild>
            <Link href={`/dashboard/parking/${lot.id}/edit`}>Редактировать</Link>
          </Button>
        )}
      </div>

      <div className="mb-6 h-[360px] overflow-hidden rounded border border-border bg-slate-100">
        <ParkingMap parkingLots={[lot]} center={[lot.latitude, lot.longitude]} zoom={15} />
      </div>

      <div className="grid gap-6 lg:grid-cols-[1fr_360px]">
        <section>
          <Tabs value={tab} onValueChange={setTab}>
            <TabsList>
              <TabsTrigger value="spots">Места</TabsTrigger>
              <TabsTrigger value="reviews">Отзывы</TabsTrigger>
            </TabsList>
            <TabsContent value="spots">
              <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
                {spots.map((spot) => (
                  <Card key={spot.id}>
                    <CardContent className="pt-4">
                      <div className="flex items-center justify-between">
                        <span className="font-semibold">{spot.spot_number}</span>
                        <Badge>{spot.spot_type}</Badge>
                      </div>
                      <p className="mt-2 text-sm text-muted">{spot.is_available ? "Доступно" : "Недоступно"}</p>
                      <Button className="mt-4 w-full" asChild>
                        <Link href={`/booking/new?spotId=${spot.id}`}>Забронировать</Link>
                      </Button>
                    </CardContent>
                  </Card>
                ))}
              </div>
            </TabsContent>
            <TabsContent value="reviews">
              <div className="space-y-3">
                {reviews.map((item) => (
                  <Card key={item.id}>
                    <CardContent className="pt-4">
                      <p className="flex items-center gap-1 font-medium">
                        <Star className="size-4 fill-yellow-400 text-yellow-400" aria-hidden="true" />
                        {item.rating}/5
                      </p>
                      <p className="text-sm text-muted">{item.comment || "Без комментария"}</p>
                    </CardContent>
                  </Card>
                ))}
                {reviews.length === 0 ? <p className="text-sm text-muted">Отзывов пока нет.</p> : null}
              </div>
            </TabsContent>
          </Tabs>
        </section>

        <aside className="space-y-4">
          <Card>
            <CardHeader>
              <CardTitle>Отзывы</CardTitle>
            </CardHeader>
            <CardContent className="space-y-3">
              {reviews.slice(0, 3).map((item) => (
                <div key={item.id} className="border-b border-border pb-3 last:border-0">
                  <p className="flex items-center gap-1 font-medium">
                    <Star className="size-4 fill-yellow-400 text-yellow-400" aria-hidden="true" />
                    {item.rating}/5
                  </p>
                  <p className="text-sm text-muted">{item.comment || "Без комментария"}</p>
                </div>
              ))}
              {reviews.length === 0 ? <p className="text-sm text-muted">Отзывов пока нет.</p> : null}
            </CardContent>
          </Card>

          {user ? (
            <Card>
              <CardHeader>
                <CardTitle>Оставить отзыв</CardTitle>
              </CardHeader>
              <CardContent>
                <form className="space-y-3" onSubmit={handleReview}>
                  <Input
                    min={1}
                    max={5}
                    type="number"
                    value={review.rating}
                    onChange={(event) => setReview({ ...review, rating: Number(event.target.value) })}
                  />
                  <Textarea
                    value={review.comment}
                    onChange={(event) => setReview({ ...review, comment: event.target.value })}
                    placeholder="Комментарий"
                  />
                  <Button className="w-full">Отправить</Button>
                </form>
              </CardContent>
            </Card>
          ) : null}
          {error ? <p className="text-sm text-red-600">{error}</p> : null}
        </aside>
      </div>
    </main>
  );
}
