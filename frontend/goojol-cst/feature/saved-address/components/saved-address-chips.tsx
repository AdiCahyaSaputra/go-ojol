import { Plus } from 'lucide-react-native';
import { useState } from 'react';
import { ActivityIndicator, Pressable, ScrollView, View } from 'react-native';
import { Text } from '@/components/ui/text';
import { DEFAULT_PICKUP } from '@/constants/book';
import type { BookLocation } from '@/feature/book/dispatch.schema';
import { useSavedAddressesQuery } from '../saved-address.query';
import { AddSavedAddressModal } from './add-saved-address-modal';

type SavedAddressChipsProps = {
  selected: BookLocation;
  onSelect: (location: BookLocation) => void;
};

function isSameLocation(a: BookLocation, b: BookLocation) {
  return a.name === b.name && a.lat === b.lat && a.lng === b.lng;
}

function PickupChip({
  label,
  selected,
  onPress,
  accessibilityLabel,
  dashed,
}: {
  label: string;
  selected: boolean;
  onPress: () => void;
  accessibilityLabel: string;
  dashed?: boolean;
}) {
  return (
    <Pressable
      onPress={onPress}
      className={`rounded-full border px-4 py-2 ${
        dashed
          ? 'flex-row items-center gap-1 border-dashed active:border-goojol-coral'
          : 'active:border-goojol-coral'
      } ${
        selected
          ? 'border-goojol-coral bg-goojol-coral/10'
          : 'border-goojol-border bg-goojol-surface'
      }`}
      accessibilityLabel={accessibilityLabel}
      accessibilityState={{ selected }}
    >
      {dashed ? <Plus color={selected ? '#ff6b4a' : '#8892a8'} size={16} /> : null}
      <Text className={`font-medium ${selected ? 'text-goojol-coral' : dashed ? 'text-goojol-muted' : 'text-white'}`}>
        {label}
      </Text>
    </Pressable>
  );
}

export function SavedAddressChips({ selected, onSelect }: SavedAddressChipsProps) {
  const { data, isLoading } = useSavedAddressesQuery();
  const [addOpen, setAddOpen] = useState(false);

  return (
    <>
      <View>
        <Text className="mb-3 font-medium text-goojol-muted text-sm">Pickup point</Text>
        <ScrollView horizontal showsHorizontalScrollIndicator={false}>
          <View className="flex-row gap-2">
            <PickupChip
              label={DEFAULT_PICKUP.name}
              selected={isSameLocation(selected, DEFAULT_PICKUP)}
              onPress={() => onSelect(DEFAULT_PICKUP)}
              accessibilityLabel="Use current location as pickup"
            />
            {isLoading ? (
              <View className="justify-center px-2">
                <ActivityIndicator color="#ff6b4a" />
              </View>
            ) : (
              data?.map((address) => (
                <PickupChip
                  key={address.id}
                  label={address.name}
                  selected={isSameLocation(selected, address)}
                  onPress={() => onSelect({ name: address.name, lat: address.lat, lng: address.lng })}
                  accessibilityLabel={`Use ${address.name} as pickup`}
                />
              ))
            )}
            <PickupChip
              label="Add"
              selected={false}
              onPress={() => setAddOpen(true)}
              accessibilityLabel="Add pickup place"
              dashed
            />
          </View>
        </ScrollView>
      </View>

      <AddSavedAddressModal open={addOpen} onClose={() => setAddOpen(false)} />
    </>
  );
}
