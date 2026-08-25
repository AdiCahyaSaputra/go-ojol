import { useEffect, useState } from 'react';
import { View } from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';
import { Text } from '@/components/ui/text';
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

      <SafeAreaView edges={['top']} className="absolute inset-x-0 top-0">
        <View className="px-4 pt-2">
          <View className="self-start rounded-full border border-goojol-border bg-goojol-sky/90 px-3 py-1.5">
            <Text className="font-semibold text-sm text-white">
              {phase === 'offline'
                ? 'Offline'
                : phase === 'online'
                  ? 'Standby'
                  : phase === 'offer'
                    ? 'Incoming offer'
                    : 'En route'}
            </Text>
          </View>
        </View>
      </SafeAreaView>

      <View className="absolute inset-x-0 bottom-0" pointerEvents="box-none">
        <StandbySheet />
      </View>
    </View>
  );
}
