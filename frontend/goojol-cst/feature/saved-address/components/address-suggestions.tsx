import { ArrowUpLeft, LocateFixed, MapPin, Plus } from 'lucide-react-native';
import { useMemo } from 'react';
import { ActivityIndicator, Pressable, View } from 'react-native';
import { Text } from '@/components/ui/text';
import { VStack } from '@/components/ui/vstack';
import { DEFAULT_PICKUP, DESTINATION_PRESETS } from '@/constants/book';
import type { BookLocation } from '@/feature/book/dispatch.schema';
import { useSavedAddressesQuery } from '../saved-address.query';
import { savedAddressToBookLocation } from '../saved-address.schema';

type AddressSuggestionsProps = {
  query: string;
  activeField: 'pickup' | 'destination';
  onSelect: (location: BookLocation) => void;
  onSelectCurrentLocation: () => void;
  onCustomAddress: () => void;
  currentLocationLoading?: boolean;
};

function isSameLocation(a: BookLocation, b: BookLocation) {
  return a.name === b.name && a.lat === b.lat && a.lng === b.lng;
}

function matchesCurrentLocationQuery(query: string) {
  const normalized = query.trim().toLowerCase();
  if (!normalized) {
    return true;
  }

  return (
    DEFAULT_PICKUP.name.toLowerCase().includes(normalized) ||
    normalized.includes('current') ||
    normalized.includes('location')
  );
}

export function AddressSuggestions({
  query,
  activeField,
  onSelect,
  onSelectCurrentLocation,
  onCustomAddress,
  currentLocationLoading = false,
}: AddressSuggestionsProps) {
  const { data, isLoading, isError } = useSavedAddressesQuery();
  const trimmedQuery = query.trim();
  const normalizedQuery = trimmedQuery.toLowerCase();

  const suggestions = useMemo(() => {
    const fromApi = (data ?? []).map(savedAddressToBookLocation);
    const presets = DESTINATION_PRESETS.map((preset) => ({
      name: preset.name,
      lat: preset.lat,
      lng: preset.lng,
    }));
    const combined = [...fromApi, ...presets].filter(
      (location, index, list) => list.findIndex((item) => isSameLocation(item, location)) === index,
    );

    if (!normalizedQuery) {
      return combined;
    }

    return combined.filter((location) => location.name.toLowerCase().includes(normalizedQuery));
  }, [data, normalizedQuery]);

  const hasExactNameMatch = suggestions.some(
    (location) => location.name.toLowerCase() === normalizedQuery,
  );
  const showCurrentLocation = matchesCurrentLocationQuery(trimmedQuery);
  const showCustomAddress = trimmedQuery.length > 0 && !hasExactNameMatch;

  return (
    <VStack space="sm">
      <Text className="font-medium text-goojol-muted text-sm">
        {activeField === 'pickup' ? 'Pickup suggestions' : 'Destination suggestions'}
      </Text>
      <View className="overflow-hidden rounded-2xl border border-goojol-border bg-goojol-surface">
        {showCurrentLocation ? (
          <Pressable
            onPress={onSelectCurrentLocation}
            disabled={currentLocationLoading}
            className="flex-row items-center justify-between px-4 py-4 active:bg-goojol-sky"
            accessibilityRole="button"
            accessibilityLabel="Use current location"
          >
            <View className="flex-row items-center gap-3">
              <LocateFixed color="#ff6b4a" size={18} />
              <VStack space="xs">
                <Text className="text-base text-white">{DEFAULT_PICKUP.name}</Text>
                <Text className="text-goojol-muted text-xs">
                  {currentLocationLoading ? 'Getting GPS…' : 'Use your live position'}
                </Text>
              </VStack>
            </View>
            {currentLocationLoading ? (
              <ActivityIndicator color="#ff6b4a" />
            ) : (
              <ArrowUpLeft color="#8892a8" size={18} />
            )}
          </Pressable>
        ) : null}

        {isLoading && (
          <View className="items-center px-4 py-4">
            <ActivityIndicator color="#ff6b4a" />
          </View>
        )}

        {isError && (
          <View className="px-4 py-4">
            <Text className="text-destructive text-sm">Could not load saved addresses.</Text>
          </View>
        )}

        {(!isLoading && !isError)
          && suggestions.map((location, index) => (
              <View key={`${location.name}-${location.lat}-${location.lng}`}>
                {showCurrentLocation || index > 0 ? (
                  <View className="ml-11 h-px bg-goojol-border" />
                ) : null}
                <Pressable
                  onPress={() => onSelect(location)}
                  className="flex-row items-center justify-between px-4 py-4 active:bg-goojol-sky"
                  accessibilityRole="button"
                  accessibilityLabel={`Use ${location.name} for ${activeField}`}
                >
                  <View className="flex-row items-center gap-3">
                    <MapPin color="#8892a8" size={18} />
                    <VStack space="xs">
                      <Text className="text-base text-white">{location.name}</Text>
                      <Text className="text-goojol-muted text-xs">
                        {activeField === 'pickup' ? 'Set as pickup' : 'Set as destination'}
                      </Text>
                    </VStack>
                  </View>
                  <ArrowUpLeft color="#8892a8" size={18} />
                </Pressable>
              </View>
            ))
          }

        {showCustomAddress && (
          <>
            {(showCurrentLocation || suggestions.length > 0) && !isLoading ? (
              <View className="ml-11 h-px bg-goojol-border" />
            ) : null}
            <Pressable
              onPress={onCustomAddress}
              className="flex-row items-center gap-3 px-4 py-4 active:bg-goojol-sky"
              accessibilityRole="button"
              accessibilityLabel={`Save custom address ${trimmedQuery}`}
            >
              <Plus color="#ff6b4a" size={18} />
              <VStack space="xs" className="flex-1">
                <Text className="font-medium text-base text-goojol-coral">Custom address</Text>
                <Text className="text-goojol-muted text-xs" numberOfLines={1}>
                  {trimmedQuery}
                </Text>
              </VStack>
            </Pressable>
          </>
        )}

        {!isLoading &&
        !isError &&
        !showCurrentLocation &&
        suggestions.length === 0 &&
        !showCustomAddress ? (
          <View className="px-4 py-4">
            <Text className="text-goojol-muted text-sm">No matching places.</Text>
          </View>
        ) : null}
      </View>
    </VStack>
  );
}
