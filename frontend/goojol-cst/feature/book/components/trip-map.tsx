import { Camera, Map as MapLibreMap, ViewAnnotation } from '@maplibre/maplibre-react-native';
import { MapPin, MapPinHouse, Navigation } from 'lucide-react-native';
import { useMemo } from 'react';
import { View } from 'react-native';
import { DEFAULT_PICKUP } from '@/constants/book';
import { OSM_RASTER_STYLE } from '@/feature/book/constants/osm-raster-style';
import type { BookLocation } from '@/feature/book/dispatch.schema';

type TripMapProps = {
  pickup: BookLocation;
  destination: BookLocation;
  driverLocation?: { lat: number; lng: number } | null;
  customerLocation?: { lat: number; lng: number } | null;
};

function parseCoordinate(value: string, fallback: string): number {
  const parsed = Number.parseFloat(value);
  const fallbackParsed = Number.parseFloat(fallback);
  return Number.isFinite(parsed) ? parsed : fallbackParsed;
}

function toLngLat(lat: number, lng: number): [number, number] {
  return [lng, lat];
}

export function TripMap({ pickup, destination, driverLocation, customerLocation }: TripMapProps) {
  const pickupLngLat = useMemo(
    () =>
      toLngLat(
        parseCoordinate(pickup.lat, DEFAULT_PICKUP.lat),
        parseCoordinate(pickup.lng, DEFAULT_PICKUP.lng),
      ),
    [pickup.lat, pickup.lng],
  );

  const destinationLngLat = useMemo(
    () =>
      toLngLat(
        parseCoordinate(destination.lat, DEFAULT_PICKUP.lat),
        parseCoordinate(destination.lng, DEFAULT_PICKUP.lng),
      ),
    [destination.lat, destination.lng],
  );

  const driverLngLat = driverLocation
    ? toLngLat(driverLocation.lat, driverLocation.lng)
    : pickupLngLat;

  const customerLngLat = customerLocation
    ? toLngLat(customerLocation.lat, customerLocation.lng)
    : pickupLngLat;

  const center = driverLocation ? driverLngLat : pickupLngLat;

  return (
    <View className="min-h-[280px] flex-1 overflow-hidden rounded-2xl border border-goojol-border">
      <MapLibreMap
        style={{ flex: 1, minHeight: 280 }}
        mapStyle={OSM_RASTER_STYLE}
        logoEnabled={false}
        attributionEnabled={false}
      >
        <Camera centerCoordinate={center} zoomLevel={14} animationDuration={500} />
        <ViewAnnotation id="pickup" coordinate={pickupLngLat}>
          <View className="rounded-full bg-goojol-coral p-2">
            <MapPinHouse color="#ffffff" size={16} />
          </View>
        </ViewAnnotation>
        <ViewAnnotation id="destination" coordinate={destinationLngLat}>
          <View className="rounded-full bg-goojol-teal p-2">
            <MapPin color="#ffffff" size={16} />
          </View>
        </ViewAnnotation>
        {driverLocation ? (
          <ViewAnnotation id="driver" coordinate={driverLngLat}>
            <View className="rounded-full bg-blue-500 p-2">
              <Navigation color="#ffffff" size={16} />
            </View>
          </ViewAnnotation>
        ) : null}
        {customerLocation ? (
          <ViewAnnotation id="customer" coordinate={customerLngLat}>
            <View className="rounded-full bg-amber-400 p-2">
              <MapPin color="#ffffff" size={14} />
            </View>
          </ViewAnnotation>
        ) : null}
      </MapLibreMap>
    </View>
  );
}
