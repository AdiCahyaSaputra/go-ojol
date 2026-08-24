import { Camera, Map as MapLibreMap, ViewAnnotation } from '@maplibre/maplibre-react-native';
import { MapPin } from 'lucide-react-native';
import { useMemo } from 'react';
import { View } from 'react-native';
import { Text } from '@/components/ui/text';
import { DEFAULT_PICKUP } from '@/constants/book';
import { OSM_RASTER_STYLE } from '@/feature/book/constants/osm-raster-style';

type LocationMapProps = {
  label: string;
  lat: string;
  lng: string;
  compact?: boolean;
};

function parseCoordinate(value: string, fallback: string): number {
  const parsed = Number.parseFloat(value);
  const fallbackParsed = Number.parseFloat(fallback);
  return Number.isFinite(parsed) ? parsed : fallbackParsed;
}

export function LocationMap({ label, lat, lng, compact = false }: LocationMapProps) {
  const center = useMemo(
    () =>
      [parseCoordinate(lng, DEFAULT_PICKUP.lng), parseCoordinate(lat, DEFAULT_PICKUP.lat)] as [
        number,
        number,
      ],
    [lat, lng],
  );

  return (
    <View
      className={`mx-6 mt-4 overflow-hidden rounded-2xl border border-goojol-border bg-goojol-surface ${
        compact ? 'h-44' : 'flex-1'
      }`}
    >
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
        attributionPosition={{ bottom: 4, left: 4 }}
      >
        <Camera center={center} zoom={15} duration={0} />
        <ViewAnnotation lngLat={center} anchor="bottom">
          <View className="rounded-full border-2 border-goojol-coral bg-goojol-coral/60 p-1">
            <MapPin color="#ffffff" size={20} />
          </View>
        </ViewAnnotation>
      </MapLibreMap>
      {label ? (
        <View className="absolute top-3 right-3 left-3">
          <View className="self-start rounded-full border border-goojol-border bg-goojol-sky/90 px-3 py-1.5">
            <Text className="font-semibold text-sm text-white">{label}</Text>
          </View>
        </View>
      ) : null}
    </View>
  );
}
