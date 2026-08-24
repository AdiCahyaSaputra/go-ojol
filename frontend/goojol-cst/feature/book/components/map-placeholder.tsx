import { MapPin } from 'lucide-react-native';
import { View } from 'react-native';
import { Text } from '@/components/ui/text';

type MapPlaceholderProps = {
  label: string;
  lat: string;
  lng: string;
};

export function MapPlaceholder({ label, lat, lng }: MapPlaceholderProps) {
  return (
    <View className="mx-6 mt-4 flex-1 overflow-hidden rounded-2xl border border-goojol-border bg-goojol-surface">
      <View className="flex-1 items-center justify-center gap-3 px-6">
        <View className="rounded-full bg-goojol-border/40 p-4">
          <MapPin color="#ff6b4a" size={32} />
        </View>
        <Text className="text-center font-semibold text-lg text-white">{label}</Text>
        <Text className="text-center text-goojol-muted text-sm">
          Map preview — MapLibre lands here in a later phase
        </Text>
        <Text className="text-center text-goojol-muted text-xs">
          {lat}, {lng}
        </Text>
      </View>
      <View className="absolute right-0 bottom-4 left-0 items-center">
        <View className="rounded-full border-2 border-goojol-coral bg-goojol-coral/20 p-2">
          <MapPin color="#ff6b4a" size={20} />
        </View>
      </View>
    </View>
  );
}
