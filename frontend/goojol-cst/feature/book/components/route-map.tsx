import {
  Camera,
  GeoJSONSource,
  Layer,
  type LngLatBounds,
  Map as MapLibreMap,
  ViewAnnotation,
} from '@maplibre/maplibre-react-native';
import { MapPin, MapPinHouse } from 'lucide-react-native';
import { useMemo } from 'react';
import { View } from 'react-native';
import { DEFAULT_PICKUP } from '@/constants/book';
import { OSM_RASTER_STYLE } from '@/feature/book/constants/osm-raster-style';
import type { BookLocation } from '@/feature/book/dispatch.schema';

type RouteMapProps = {
  pickup: BookLocation;
  destination: BookLocation;
  path: [number, number][];
  padding?: { top: number; right: number; bottom: number; left: number };
};

function parseCoordinate(value: string, fallback: string): number {
  const parsed = Number.parseFloat(value);
  const fallbackParsed = Number.parseFloat(fallback);
  return Number.isFinite(parsed) ? parsed : fallbackParsed;
}

function toLngLat(lat: number, lng: number): [number, number] {
  return [lng, lat];
}

export function RouteMap({ pickup, destination, path, padding }: RouteMapProps) {
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

  const routeCoordinates = useMemo(() => {
    if (path.length >= 2) {
      return path.map(([lat, lng]) => toLngLat(lat, lng));
    }
    return [pickupLngLat, destinationLngLat];
  }, [destinationLngLat, path, pickupLngLat]);

  const routeGeoJson = useMemo(
    () => ({
      type: 'Feature' as const,
      properties: {},
      geometry: {
        type: 'LineString' as const,
        coordinates: routeCoordinates,
      },
    }),
    [routeCoordinates],
  );

  const bounds = useMemo(() => {
    let west = Number.POSITIVE_INFINITY;
    let south = Number.POSITIVE_INFINITY;
    let east = Number.NEGATIVE_INFINITY;
    let north = Number.NEGATIVE_INFINITY;

    for (const [lng, lat] of routeCoordinates) {
      west = Math.min(west, lng);
      east = Math.max(east, lng);
      south = Math.min(south, lat);
      north = Math.max(north, lat);
    }

    if (!Number.isFinite(west)) {
      return undefined;
    }

    const padLng = Math.max((east - west) * 0.08, 0.002);
    const padLat = Math.max((north - south) * 0.08, 0.002);

    const fittedBounds: LngLatBounds = [
      west - padLng,
      south - padLat,
      east + padLng,
      north + padLat,
    ];
    return fittedBounds;
  }, [routeCoordinates]);

  return (
    <View className="flex-1 bg-goojol-sky">
      <MapLibreMap
        style={{ flex: 1 }}
        mapStyle={OSM_RASTER_STYLE}
        dragPan
        touchZoom
        doubleTapZoom
        doubleTapHoldZoom
        touchRotate={false}
        touchPitch={false}
        logo={false}
        attribution
        attributionPosition={{ bottom: 8, left: 8 }}
      >
        {bounds ? (
          <Camera bounds={bounds} padding={padding} duration={0} />
        ) : (
          <Camera center={pickupLngLat} zoom={14} duration={0} />
        )}
        <GeoJSONSource id="quote-route" data={routeGeoJson}>
          <Layer
            id="quote-route-halo"
            type="line"
            paint={{
              'line-color': '#0F1729',
              'line-width': 8,
              'line-opacity': 0.35,
            }}
            layout={{
              'line-cap': 'round',
              'line-join': 'round',
            }}
          />
          <Layer
            id="quote-route-line"
            type="line"
            paint={{
              'line-color': '#ff6b4a',
              'line-width': 4,
              'line-opacity': 0.95,
            }}
            layout={{
              'line-cap': 'round',
              'line-join': 'round',
            }}
          />
        </GeoJSONSource>
        <ViewAnnotation lngLat={pickupLngLat} anchor="bottom">
          <View className="rounded-full border-2 border-white bg-goojol-coral p-1.5">
            <MapPinHouse color="#ffffff" size={16} />
          </View>
        </ViewAnnotation>
        <ViewAnnotation lngLat={destinationLngLat} anchor="bottom">
          <View className="rounded-full border-2 border-white bg-goojol-teal p-1.5">
            <MapPin color="#ffffff" size={16} />
          </View>
        </ViewAnnotation>
      </MapLibreMap>
    </View>
  );
}
