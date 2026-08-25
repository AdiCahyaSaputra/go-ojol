import {
  Camera,
  GeoJSONSource,
  Layer,
  type LngLatBounds,
  Map as MapLibreMap,
  ViewAnnotation,
} from '@maplibre/maplibre-react-native';
import { MapPin, MapPinHouse, Navigation } from 'lucide-react-native';
import { useMemo } from 'react';
import { View } from 'react-native';
import { DEFAULT_LOCATION } from '@/constants/standby';
import { OSM_RASTER_STYLE } from '@/feature/standby/constants/osm-raster-style';
import type { MockOffer } from '@/feature/standby/standby-context';

type StandbyMapProps = {
  lat: number;
  lng: number;
  offer: MockOffer | null;
  showRoute: boolean;
  padding?: { top: number; right: number; bottom: number; left: number };
};

function toLngLat(lat: number, lng: number): [number, number] {
  return [lng, lat];
}

export function StandbyMap({ lat, lng, offer, showRoute, padding }: StandbyMapProps) {
  const driverLngLat = useMemo(() => toLngLat(lat, lng), [lat, lng]);

  const routeCoordinates = useMemo(() => {
    if (!offer || !showRoute) {
      return null;
    }
    if (offer.path.length >= 2) {
      return offer.path.map(([pathLat, pathLng]) => toLngLat(pathLat, pathLng));
    }
    return [
      toLngLat(offer.pickup.lat, offer.pickup.lng),
      toLngLat(offer.destination.lat, offer.destination.lng),
    ];
  }, [offer, showRoute]);

  const routeGeoJson = useMemo(() => {
    if (!routeCoordinates) {
      return null;
    }
    return {
      type: 'Feature' as const,
      properties: {},
      geometry: {
        type: 'LineString' as const,
        coordinates: routeCoordinates,
      },
    };
  }, [routeCoordinates]);

  const bounds = useMemo(() => {
    if (!offer) {
      return undefined;
    }

    const points: [number, number][] = [
      driverLngLat,
      toLngLat(offer.pickup.lat, offer.pickup.lng),
      toLngLat(offer.destination.lat, offer.destination.lng),
    ];
    if (routeCoordinates) {
      points.push(...routeCoordinates);
    }

    let west = Number.POSITIVE_INFINITY;
    let south = Number.POSITIVE_INFINITY;
    let east = Number.NEGATIVE_INFINITY;
    let north = Number.NEGATIVE_INFINITY;

    for (const [pointLng, pointLat] of points) {
      west = Math.min(west, pointLng);
      east = Math.max(east, pointLng);
      south = Math.min(south, pointLat);
      north = Math.max(north, pointLat);
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
  }, [driverLngLat, offer, routeCoordinates]);

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
          <Camera
            center={driverLngLat}
            zoom={14}
            duration={0}
            padding={padding ?? { top: 40, right: 28, bottom: 220, left: 28 }}
          />
        )}

        {routeGeoJson ? (
          <GeoJSONSource id="standby-route" data={routeGeoJson}>
            <Layer
              id="standby-route-halo"
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
              id="standby-route-line"
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
        ) : null}

        <ViewAnnotation lngLat={driverLngLat} anchor="center">
          <View className="rounded-full border-2 border-white bg-goojol-coral p-2">
            <Navigation color="#ffffff" size={16} />
          </View>
        </ViewAnnotation>

        {offer ? (
          <>
            <ViewAnnotation lngLat={toLngLat(offer.pickup.lat, offer.pickup.lng)} anchor="bottom">
              <View className="rounded-full border-2 border-white bg-goojol-coral p-1.5">
                <MapPinHouse color="#ffffff" size={16} />
              </View>
            </ViewAnnotation>
            <ViewAnnotation
              lngLat={toLngLat(offer.destination.lat, offer.destination.lng)}
              anchor="bottom"
            >
              <View className="rounded-full border-2 border-white bg-goojol-teal p-1.5">
                <MapPin color="#ffffff" size={16} />
              </View>
            </ViewAnnotation>
          </>
        ) : null}
      </MapLibreMap>
    </View>
  );
}

export function standbyMapFallbackCenter() {
  return { lat: DEFAULT_LOCATION.lat, lng: DEFAULT_LOCATION.lng };
}
