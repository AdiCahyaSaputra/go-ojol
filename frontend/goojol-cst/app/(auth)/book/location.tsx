import { useRouter } from 'expo-router';
import { LucideMapPinHouse, LucideMapPinned } from 'lucide-react-native';
import { useState } from 'react';
import { Pressable, ScrollView, View } from 'react-native';
import { Button, ButtonText } from '@/components/ui/button';
import { HStack } from '@/components/ui/hstack';
import { Input, InputField } from '@/components/ui/input';
import { Text } from '@/components/ui/text';
import { VStack } from '@/components/ui/vstack';
import { DESTINATION_PRESETS } from '@/constants/book';
import { useBook } from '@/feature/book/book-context';
import { LocationMap } from '@/feature/book/components/location-map';
import { WizardShell } from '@/feature/book/components/wizard-shell';
import { resolveCurrentLocation } from '@/feature/book/current-location';
import type { BookLocation } from '@/feature/book/dispatch.schema';
import { AddSavedAddressModal } from '@/feature/saved-address/components/add-saved-address-modal';
import { AddressSuggestions } from '@/feature/saved-address/components/address-suggestions';

type ActiveField = 'pickup' | 'destination';

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
  const [customOpen, setCustomOpen] = useState(false);
  const [currentLocationLoading, setCurrentLocationLoading] = useState(false);

  const activeLocation = activeField === 'pickup' ? pickupDraft : destinationDraft;
  const canContinue = pickupDraft.name.trim().length > 0 && destinationDraft.name.trim().length > 0;

  const updateLocationName = (field: ActiveField, name: string) => {
    if (field === 'pickup') {
      setPickupDraft((current) => ({ ...current, name }));
      return;
    }

    setDestinationDraft((current) => ({ ...current, name }));
  };

  const applyLocation = (location: BookLocation) => {
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

  const onSelectCurrentLocation = async () => {
    setCurrentLocationLoading(true);
    try {
      const location = await resolveCurrentLocation();
      applyLocation({ name: location.name, lat: location.lat, lng: location.lng });
    } finally {
      setCurrentLocationLoading(false);
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
              className="flex-row gap-3 p-4"
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

            <View className="border-goojol-border border-t" />

            <Pressable
              onPress={() => setActiveField('destination')}
              className="flex-row gap-3 p-4"
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

          <AddressSuggestions
            query={activeLocation.name}
            activeField={activeField}
            onSelect={applyLocation}
            onSelectCurrentLocation={onSelectCurrentLocation}
            onCustomAddress={() => setCustomOpen(true)}
            currentLocationLoading={currentLocationLoading}
          />
        </VStack>
      </ScrollView>

      <AddSavedAddressModal
        open={customOpen}
        onClose={() => setCustomOpen(false)}
        initialName={activeLocation.name.trim()}
        initialLat={activeLocation.lat}
        initialLng={activeLocation.lng}
        onSaved={applyLocation}
      />
    </WizardShell>
  );
}
