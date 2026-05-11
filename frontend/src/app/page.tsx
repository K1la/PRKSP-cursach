import dynamic from "next/dynamic";
import Link from "next/link";
import { Clock, MapPin, Search, ShieldCheck, WalletCards } from "lucide-react";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { ApiStatusButton } from "@/components/parking/api-status-button";
import { mockParkingLots } from "@/lib/mock-data";

const ParkingMap = dynamic(() => import("@/components/parking/parking-map"), {
  ssr: false,
  loading: () => <div className="h-full w-full animate-pulse bg-slate-200" />,
});

export default function Home() {
  return (
    <main className="min-h-screen bg-background">
      <section className="border-b border-border bg-white">
        <div className="mx-auto grid max-w-7xl gap-8 px-4 py-8 sm:px-6 lg:grid-cols-[minmax(0,0.95fr)_minmax(460px,1.05fr)] lg:px-8">
          <div className="flex flex-col justify-center">
            <Badge className="mb-4 w-fit">Москва · 8 парковок в демо</Badge>
            <h1 className="max-w-3xl text-4xl font-semibold tracking-normal text-slate-950 sm:text-5xl">
              Найдите свободное место рядом и забронируйте его заранее
            </h1>
            <p className="mt-4 max-w-2xl text-base leading-7 text-muted">
              ParkEase показывает парковки на карте, помогает сравнить цену,
              доступность и оформить бронь без звонков владельцу.
            </p>

            <div className="mt-7 grid gap-3 rounded border border-border bg-slate-50 p-3 sm:grid-cols-[1fr_150px_132px]">
              <div className="relative">
                <Search className="pointer-events-none absolute left-3 top-1/2 size-4 -translate-y-1/2 text-muted" />
                <Input className="pl-9" placeholder="Адрес, район или метро" />
              </div>
              <Input placeholder="До 200 ₽/ч" />
              <Button className="w-full" asChild>
                <Link href="/parking">
                <Search className="mr-2 size-4" aria-hidden="true" />
                Найти
                </Link>
              </Button>
            </div>

            <div className="mt-6 grid gap-3 text-sm text-slate-700 sm:grid-cols-3">
              <div className="flex items-center gap-2">
                <ShieldCheck className="size-4 text-primary" aria-hidden="true" />
                JWT и роли
              </div>
              <div className="flex items-center gap-2">
                <Clock className="size-4 text-primary" aria-hidden="true" />
                Бронь по времени
              </div>
              <div className="flex items-center gap-2">
                <WalletCards className="size-4 text-primary" aria-hidden="true" />
                Расчёт стоимости
              </div>
            </div>
          </div>

          <div className="h-[420px] overflow-hidden rounded border border-border bg-slate-100">
            <ParkingMap parkingLots={mockParkingLots} />
          </div>
        </div>
      </section>

      <section className="mx-auto grid max-w-7xl gap-6 px-4 py-8 sm:px-6 lg:grid-cols-[1fr_320px] lg:px-8">
        <div>
          <div className="mb-4 flex items-center justify-between gap-4">
            <div>
              <h2 className="text-2xl font-semibold">Популярные парковки</h2>
              <p className="text-sm text-muted">Стартовые данные для будущего seed-скрипта</p>
            </div>
            <Button variant="outline" asChild>
              <Link href="/parking">Все парковки</Link>
            </Button>
          </div>
          <div className="grid gap-4 md:grid-cols-2">
            {mockParkingLots.slice(0, 4).map((lot) => (
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
                  <div className="mt-4 flex items-center justify-between text-sm">
                    <span>{lot.total_spots} мест</span>
                    <Link className="font-medium text-primary" href={`/parking/${lot.id}`}>
                      Подробнее
                    </Link>
                  </div>
                </CardContent>
              </Card>
            ))}
          </div>
        </div>

        <aside className="space-y-4">
          <Card>
            <CardHeader>
              <CardTitle>Backend</CardTitle>
            </CardHeader>
            <CardContent className="space-y-3">
              <p className="text-sm leading-6 text-muted">
                Проверка обращается к `NEXT_PUBLIC_API_URL` и ожидает health endpoint.
              </p>
              <ApiStatusButton />
            </CardContent>
          </Card>
        </aside>
      </section>
    </main>
  );
}
