import type {
  AdminStats,
  AuthResponse,
  Booking,
  HealthResponse,
  ParkingLot,
  ParkingSpot,
  Review,
  User,
} from "@/types/parking";

const apiBaseURL = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080/api/v1";

type RequestOptions = RequestInit & {
  token?: string | null;
};

type ParkingLotPayload = {
  name: string;
  description?: string | null;
  address: string;
  latitude: number;
  longitude: number;
  total_spots: number;
  price_per_hour: number;
  is_active?: boolean;
};

export function getStoredAccessToken() {
  if (typeof window === "undefined") {
    return null;
  }
  return localStorage.getItem("access_token");
}

export function storeTokens(tokens: Pick<AuthResponse, "access_token" | "refresh_token">) {
  localStorage.setItem("access_token", tokens.access_token);
  localStorage.setItem("refresh_token", tokens.refresh_token);
}

export function clearTokens() {
  localStorage.removeItem("access_token");
  localStorage.removeItem("refresh_token");
}

async function apiFetch<T>(path: string, options: RequestOptions = {}): Promise<T> {
  const headers = new Headers(options.headers);
  headers.set("Accept", "application/json");

  if (options.body && !headers.has("Content-Type")) {
    headers.set("Content-Type", "application/json");
  }

  const token = options.token === undefined ? getStoredAccessToken() : options.token;
  if (token) {
    headers.set("Authorization", `Bearer ${token}`);
  }

  const response = await fetch(`${apiBaseURL}${path}`, {
    ...options,
    headers,
    cache: "no-store",
  });

  if (!response.ok) {
    let message = `Request failed with status ${response.status}`;
    try {
      const payload = (await response.json()) as { error?: { message?: string } };
      message = payload.error?.message ?? message;
    } catch {
      // Keep fallback message.
    }
    throw new Error(message);
  }

  if (response.status === 204) {
    return undefined as T;
  }

  return response.json() as Promise<T>;
}

export function getHealth(): Promise<HealthResponse> {
  return apiFetch<HealthResponse>("/health", { token: null });
}

export function register(payload: {
  email: string;
  password: string;
  name: string;
  phone?: string;
}): Promise<AuthResponse> {
  return apiFetch<AuthResponse>("/auth/register", {
    method: "POST",
    body: JSON.stringify(payload),
    token: null,
  });
}

export function login(payload: { email: string; password: string }): Promise<AuthResponse> {
  return apiFetch<AuthResponse>("/auth/login", {
    method: "POST",
    body: JSON.stringify(payload),
    token: null,
  });
}

export function me(): Promise<User> {
  return apiFetch<User>("/users/me");
}

export function listUsers(): Promise<User[]> {
  return apiFetch<User[]>("/users");
}

export function updateMe(payload: { name: string; phone?: string | null }): Promise<User> {
  return apiFetch<User>("/users/me", {
    method: "PUT",
    body: JSON.stringify(payload),
  });
}

export function listParkingLots(params: URLSearchParams = new URLSearchParams()): Promise<ParkingLot[]> {
  const query = params.toString();
  return apiFetch<ParkingLot[]>(`/parking-lots${query ? `?${query}` : ""}`, { token: null });
}

export function getParkingLot(id: string): Promise<ParkingLot> {
  return apiFetch<ParkingLot>(`/parking-lots/${id}`, { token: null });
}

export function listParkingSpots(id: string): Promise<ParkingSpot[]> {
  return apiFetch<ParkingSpot[]>(`/parking-lots/${id}/spots`, { token: null });
}

export function createParkingLot(payload: ParkingLotPayload): Promise<ParkingLot> {
  return apiFetch<ParkingLot>("/parking-lots", {
    method: "POST",
    body: JSON.stringify(payload),
  });
}

export function updateParkingLot(id: string, payload: ParkingLotPayload): Promise<ParkingLot> {
  return apiFetch<ParkingLot>(`/parking-lots/${id}`, {
    method: "PUT",
    body: JSON.stringify(payload),
  });
}

export function createBooking(payload: {
  parking_spot_id: string;
  start_time: string;
  end_time: string;
  vehicle_plate: string;
}): Promise<Booking> {
  return apiFetch<Booking>("/bookings", {
    method: "POST",
    body: JSON.stringify(payload),
  });
}

export function listMyBookings(): Promise<Booking[]> {
  return apiFetch<Booking[]>("/bookings");
}

export function cancelBooking(id: string): Promise<Booking> {
  return apiFetch<Booking>(`/bookings/${id}/cancel`, { method: "PUT" });
}

export function listReviews(lotId: string): Promise<Review[]> {
  return apiFetch<Review[]>(`/parking-lots/${lotId}/reviews`, { token: null });
}

export function createReview(lotId: string, payload: { rating: number; comment?: string }): Promise<Review> {
  return apiFetch<Review>(`/parking-lots/${lotId}/reviews`, {
    method: "POST",
    body: JSON.stringify(payload),
  });
}

export function adminStats(): Promise<AdminStats> {
  return apiFetch<AdminStats>("/admin/stats");
}
