import { useRouter } from 'expo-router';
import { useState } from 'react';
import { Pressable, ScrollView, View } from 'react-native';
import { Button, ButtonText } from '@/components/ui/button';
import { FormControl, FormControlLabel, FormControlLabelText } from '@/components/ui/form-control';
import { Input, InputField } from '@/components/ui/input';
import { Text } from '@/components/ui/text';
import { VStack } from '@/components/ui/vstack';
import { DESTINATION_PRESETS } from '@/constants/book';
import { useBook } from '@/feature/book/book-context';
import { LocationMap } from '@/feature/book/components/location-map';
import { WizardShell } from '@/feature/book/components/wizard-shell';

export default function BookDestinationScreen() {
  const router = useRouter();
  const { destination, setDestination } = useBook();
  const [name, setName] = useState(destination?.name ?? '');
  const [lat, setLat] = useState(destination?.lat ?? DESTINATION_PRESETS[0].lat);
  const [lng, setLng] = useState(destination?.lng ?? DESTINATION_PRESETS[0].lng);

  const applyPreset = (preset: (typeof DESTINATION_PRESETS)[number]) => {
    setName(preset.name);
    setLat(preset.lat);
    setLng(preset.lng);
  };

  const onContinue = () => {
    if (!name.trim()) {
      return;
    }
    setDestination({ name: name.trim(), lat, lng });
    router.push('/book/quote');
  };

  return (
    <WizardShell
      title="Set destination"
      currentStep={2}
      footer={
        <Button
          className="w-full bg-goojol-coral data-[active=true]:bg-goojol-coral/90"
          onPress={onContinue}
          isDisabled={!name.trim()}
        >
          <ButtonText className="font-semibold text-white">Continue</ButtonText>
        </Button>
      }
    >
      <LocationMap label={name || 'Destination'} lat={lat} lng={lng} />

      <ScrollView className="flex-1 px-6 py-4" keyboardShouldPersistTaps="handled">
        <VStack space="md">
          <Text className="text-goojol-muted text-sm">Quick picks</Text>
          <View className="flex-row flex-wrap gap-2">
            {DESTINATION_PRESETS.map((preset) => (
              <Pressable
                key={preset.name}
                onPress={() => applyPreset(preset)}
                className="rounded-full border border-goojol-border bg-goojol-surface px-4 py-2 active:border-goojol-coral"
              >
                <Text className="text-white">{preset.name}</Text>
              </Pressable>
            ))}
          </View>

          <FormControl>
            <FormControlLabel>
              <FormControlLabelText className="text-goojol-muted">Place name</FormControlLabelText>
            </FormControlLabel>
            <Input className="border-goojol-border bg-goojol-surface">
              <InputField
                value={name}
                onChangeText={setName}
                placeholder="Where are you going?"
                placeholderTextColor="#8892a8"
                className="text-white"
              />
            </Input>
          </FormControl>

          <FormControl>
            <FormControlLabel>
              <FormControlLabelText className="text-goojol-muted">Latitude</FormControlLabelText>
            </FormControlLabel>
            <Input className="border-goojol-border bg-goojol-surface">
              <InputField
                value={lat}
                onChangeText={setLat}
                keyboardType="numeric"
                className="text-white"
              />
            </Input>
          </FormControl>

          <FormControl>
            <FormControlLabel>
              <FormControlLabelText className="text-goojol-muted">Longitude</FormControlLabelText>
            </FormControlLabel>
            <Input className="border-goojol-border bg-goojol-surface">
              <InputField
                value={lng}
                onChangeText={setLng}
                keyboardType="numeric"
                className="text-white"
              />
            </Input>
          </FormControl>
        </VStack>
      </ScrollView>
    </WizardShell>
  );
}
