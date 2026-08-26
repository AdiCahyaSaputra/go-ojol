import { useEffect, useState } from 'react';
import { View } from 'react-native';
import { DEFAULT_LOCATION } from '@/constants/standby';
import { StandbyMap } from '@/feature/standby/components/standby-map';
import { StandbySheet } from '@/feature/standby/components/standby-sheet';
import { resolveDriverLocation } from '@/feature/standby/location';
import { useStandby } from '@/feature/standby/standby-context';

export default function StandbyPage() {
  const { phase, coords, offer } = useStandby();
  const [bootCoords, setBootCoords] = useState<{ lat: number; lng: number }>({
    lat: DEFAULT_LOCATION.lat,
    lng: DEFAULT_LOCATION.lng,
  });

  useEffect(() => {
    let cancelled = false;
    (async () => {
      const location = await resolveDriverLocation();
      if (!cancelled) {
        setBootCoords({ lat: location.lat, lng: location.lng });
      }
    })();
    return () => {
      cancelled = true;
    };
  }, []);

  const mapLat = coords?.lat ?? bootCoords.lat;
  const mapLng = coords?.lng ?? bootCoords.lng;
  const showRoute = phase === 'offer' || phase === 'accepted';
  const bottomPad = phase === 'offer' || phase === 'accepted' ? 360 : 240;

  return (
    <View className="flex-1 bg-goojol-sky">
      <StandbyMap
        lat={mapLat}
        lng={mapLng}
        offer={offer}
        showRoute={showRoute}
        padding={{ top: 48, right: 28, bottom: bottomPad, left: 28 }}
      />

      <View className="absolute inset-x-0 bottom-0" pointerEvents="box-none">
        <StandbySheet />
      </View>
    </View>
  );
}
