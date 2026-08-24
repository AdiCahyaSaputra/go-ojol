import { useRouter } from 'expo-router';
import { ArrowUpLeft, Clock3, LucideMapPinHouse, LucideMapPinned } from 'lucide-react-native';
import { useMemo, useState } from 'react';
import { Pressable, ScrollView, View } from 'react-native';
import { Button, ButtonText } from '@/components/ui/button';
import { Input, InputField } from '@/components/ui/input';
import { Text } from '@/components/ui/text';
import { VStack } from '@/components/ui/vstack';
import { DEFAULT_PICKUP, DESTINATION_PRESETS } from '@/constants/book';
import { useBook } from '@/feature/book/book-context';
import { LocationMap } from '@/feature/book/components/location-map';
import { WizardShell } from '@/feature/book/components/wizard-shell';
import type { BookLocation } from '@/feature/book/dispatch.schema';
import { HStack } from '@/components/ui/hstack';

type ActiveField = 'pickup' | 'destination';

function isSameLocation(a: BookLocation, b: BookLocation) {
  return a.name === b.name && a.lat === b.lat && a.lng === b.lng;
}

export default function BookLocationScreen() {
  const router = useRouter();
  const { pickup, destination, setPickup, setDestination } = useBook();
  const [pickupDraft, setPickupDraft] = useState<BookLocation>(pickup);
  const [destinationDraft, setDestinationDraft] = useState<BookLocation>(
    destination ?? DESTINATION_PRESETS[0],
  );
  const [activeField, setActiveField] = useState<ActiveField>(
    destination ? 'destination' : 'pickup',
  );

  const activeLocation = activeField === 'pickup' ? pickupDraft : destinationDraft;
  const canContinue = pickupDraft.name.trim().length > 0 && destinationDraft.name.trim().length > 0;

  const recentLocations = useMemo(() => {
    const locations = [destinationDraft, pickupDraft, DEFAULT_PICKUP, ...DESTINATION_PRESETS];
    return locations.filter(
      (location, index) => locations.findIndex((item) => isSameLocation(item, location)) === index,
    );
  }, [destinationDraft, pickupDraft]);

  const updateLocationName = (field: ActiveField, name: string) => {
    if (field === 'pickup') {
      setPickupDraft((current) => ({ ...current, name }));
      return;
    }

    setDestinationDraft((current) => ({ ...current, name }));
  };

  const applyRecentLocation = (location: BookLocation) => {
    if (activeField === 'pickup') {
      setPickupDraft(location);
      if (!destinationDraft.name.trim()) {
        setActiveField('destination');
      }
      return;
    }

    setDestinationDraft(location);
    if (!pickupDraft.name.trim()) {
      setActiveField('pickup');
    }
  };

  const onContinue = () => {
    if (!canContinue) {
      return;
    }

    setPickup({ ...pickupDraft, name: pickupDraft.name.trim() });
    setDestination({ ...destinationDraft, name: destinationDraft.name.trim() });
    router.push('/book/quote');
  };

  return (
    <WizardShell
      title="Set route"
      currentStep={1}
      footer={
        <Button
          className="w-full bg-goojol-coral data-[active=true]:bg-goojol-coral/90"
          onPress={onContinue}
          isDisabled={!canContinue}
        >
          <ButtonText className="font-semibold text-white">Continue</ButtonText>
        </Button>
      }
    >
      <LocationMap
        label={activeLocation.name || (activeField === 'pickup' ? 'Pickup' : 'Destination')}
        lat={activeLocation.lat}
        lng={activeLocation.lng}
        compact
      />

      <ScrollView className="flex-1" keyboardShouldPersistTaps="handled">
        <VStack space="lg" className="px-6 py-4">
          <View className="rounded-2xl border border-goojol-border bg-goojol-surface">
            <Pressable
              onPress={() => setActiveField('pickup')}
              className={`flex-row gap-3 p-4`}
              accessibilityRole="button"
              accessibilityLabel="Edit pickup location"
            >
              <VStack space="xs" className="flex-1">
                <HStack className="items-center gap-2">
                  <LucideMapPinHouse color="#ff6b4a" width={14} height={14} />
                  <Text className="text-goojol-muted text-xs">Pickup</Text>
                </HStack>
                <Input className="border-0 bg-transparent px-0">
                  <InputField
                    value={pickupDraft.name}
                    onChangeText={(value) => updateLocationName('pickup', value)}
                    onFocus={() => setActiveField('pickup')}
                    placeholder="Enter pickup point"
                    placeholderTextColor="#8892a8"
                    className="px-0 font-medium text-base text-white"
                  />
                </Input>
              </VStack>
            </Pressable>

            <View className="border-t border-goojol-border" />

            <Pressable
              onPress={() => setActiveField('destination')}
              className={`flex-row gap-3 p-4`}
              accessibilityRole="button"
              accessibilityLabel="Edit destination location"
            >
              <VStack space="xs" className="flex-1">
                <HStack className="items-center gap-2">
                  <LucideMapPinned color="#3ddba8" width={14} height={14} />
                  <Text className="text-goojol-muted text-xs">Destination</Text>
                </HStack>
                <Input className="border-0 bg-transparent px-0">
                  <InputField
                    value={destinationDraft.name}
                    onChangeText={(value) => updateLocationName('destination', value)}
                    onFocus={() => setActiveField('destination')}
                    placeholder="Where are you going?"
                    placeholderTextColor="#8892a8"
                    className="px-0 font-medium text-base text-white"
                  />
                </Input>
              </VStack>
            </Pressable>
          </View>

          <VStack space="sm">
            <Text className="font-medium text-goojol-muted text-sm">Recent</Text>
            <View className="overflow-hidden rounded-2xl border border-goojol-border bg-goojol-surface">
              {recentLocations.map((location, index) => (
                <View key={`${location.name}-${location.lat}-${location.lng}`}>
                  <Pressable
                    onPress={() => applyRecentLocation(location)}
                    className="flex-row items-center justify-between px-4 py-4 active:bg-goojol-sky"
                    accessibilityRole="button"
                    accessibilityLabel={`Use ${location.name} for ${activeField}`}
                  >
                    <View className="flex-row items-center gap-3">
                      <Clock3 color="#8892a8" size={18} />
                      <VStack space="xs">
                        <Text className="text-base text-white">{location.name}</Text>
                        <Text className="text-goojol-muted text-xs">
                          {activeField === 'pickup' ? 'Set as pickup' : 'Set as destination'}
                        </Text>
                      </VStack>
                    </View>
                    <ArrowUpLeft color="#8892a8" size={18} />
                  </Pressable>
                  {index < recentLocations.length - 1 ? (
                    <View className="ml-11 h-px bg-goojol-border" />
                  ) : null}
                </View>
              ))}
            </View>
          </VStack>
        </VStack>
      </ScrollView>
    </WizardShell>
  );
}
