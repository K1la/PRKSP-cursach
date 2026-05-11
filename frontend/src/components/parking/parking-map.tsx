"use client";

import L from "leaflet";
import { MapContainer, Marker, Popup, TileLayer } from "react-leaflet";

import type { ParkingLot } from "@/types/parking";

const markerIcon = L.icon({
  iconUrl: "https://unpkg.com/leaflet@1.9.4/dist/images/marker-icon.png",
  iconRetinaUrl: "https://unpkg.com/leaflet@1.9.4/dist/images/marker-icon-2x.png",
  shadowUrl: "https://unpkg.com/leaflet@1.9.4/dist/images/marker-shadow.png",
  iconSize: [25, 41],
  iconAnchor: [12, 41],
  popupAnchor: [1, -34],
  shadowSize: [41, 41],
});

type ParkingMapProps = {
  parkingLots: ParkingLot[];
};

export default function ParkingMap({ parkingLots }: ParkingMapProps) {
  return (
    <MapContainer center={[55.7558, 37.6173]} zoom={11} scrollWheelZoom={false}>
      <TileLayer
        attribution='&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a>'
        url="https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png"
      />
      {parkingLots.map((lot) => (
        <Marker key={lot.id} position={[lot.latitude, lot.longitude]} icon={markerIcon}>
          <Popup>
            <strong>{lot.name}</strong>
            <br />
            {lot.total_spots} мест · {lot.price_per_hour} ₽/ч
          </Popup>
        </Marker>
      ))}
    </MapContainer>
  );
}
