export type UserRole = "user" | "owner" | "admin";

export type User = {
  id: string;
  email: string;
  name: string;
  phone?: string | null;
  role: UserRole;
  created_at: string;
  updated_at: string;
};

export type ParkingLot = {
  id: string;
  owner_id: string;
  name: string;
  description?: string | null;
  address: string;
  latitude: number;
  longitude: number;
  total_spots: number;
  price_per_hour: number;
  is_active: boolean;
  created_at: string;
  updated_at: string;
};

export type ParkingSpot = {
  id: string;
  parking_lot_id: string;
  spot_number: string;
  spot_type: "standard" | "disabled" | "electric" | "vip";
  is_available: boolean;
  floor?: number | null;
  created_at: string;
};

export type Booking = {
  id: string;
  user_id: string;
  parking_spot_id: string;
  start_time: string;
  end_time: string;
  status: "pending" | "active" | "completed" | "cancelled";
  total_price: number;
  vehicle_plate: string;
  created_at: string;
  updated_at: string;
};

export type Review = {
  id: string;
  user_id: string;
  parking_lot_id: string;
  rating: number;
  comment?: string | null;
  created_at: string;
};

export type AuthResponse = {
  user: User;
  access_token: string;
  refresh_token: string;
};

export type HealthResponse = {
  status: string;
  db: string;
  version: string;
};

export type AdminStats = {
  total_users: number;
  total_bookings: number;
  total_parking_lots: number;
  revenue: number;
};
