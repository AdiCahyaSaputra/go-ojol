import { useEffect, useState } from 'react';
import { Modal, Pressable, View } from 'react-native';
import { Button, ButtonText } from '@/components/ui/button';
import { FormControl, FormControlLabel, FormControlLabelText } from '@/components/ui/form-control';
import { Heading } from '@/components/ui/heading';
import { Input, InputField } from '@/components/ui/input';
import { Text } from '@/components/ui/text';
import { VStack } from '@/components/ui/vstack';
import type { BookLocation } from '@/feature/book/dispatch.schema';
import { useCreateSavedAddressMutation } from '../saved-address.query';
import { savedAddressToBookLocation } from '../saved-address.schema';

type AddSavedAddressModalProps = {
  open: boolean;
  onClose: () => void;
  initialName?: string;
  initialLat?: string;
  initialLng?: string;
  onSaved?: (location: BookLocation) => void;
};

export function AddSavedAddressModal({
  open,
  onClose,
  initialName = '',
  initialLat = '-6.2088',
  initialLng = '106.8456',
  onSaved,
}: AddSavedAddressModalProps) {
  const [name, setName] = useState(initialName);
  const [lat, setLat] = useState(initialLat);
  const [lng, setLng] = useState(initialLng);
  const mutation = useCreateSavedAddressMutation();

  useEffect(() => {
    if (!open) {
      return;
    }

    setName(initialName);
    setLat(initialLat);
    setLng(initialLng);
  }, [open, initialName, initialLat, initialLng]);

  const onSave = async () => {
    if (!name.trim()) {
      return;
    }

    try {
      const created = await mutation.mutateAsync({
        name: name.trim(),
        lat_long: [lat, lng],
        is_default_pickup: false,
      });

      onSaved?.(savedAddressToBookLocation(created));
      onClose();
    } catch {
      // mutation.isError surfaces the failure message
    }
  };

  return (
    <Modal visible={open} animationType="slide" transparent onRequestClose={onClose}>
      <View className="flex-1 justify-end bg-black/50">
        <View className="rounded-t-3xl bg-goojol-sky px-6 pt-6 pb-10">
          <Heading size="lg" className="mb-4 text-white">
            Custom address
          </Heading>

          <VStack space="md">
            <FormControl>
              <FormControlLabel>
                <FormControlLabelText className="text-goojol-muted">Name</FormControlLabelText>
              </FormControlLabel>
              <Input className="border-goojol-border bg-goojol-surface">
                <InputField
                  value={name}
                  onChangeText={setName}
                  placeholder="e.g. Gym"
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
                  placeholder="-6.2088"
                  placeholderTextColor="#8892a8"
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
                  placeholder="106.8456"
                  placeholderTextColor="#8892a8"
                  keyboardType="numeric"
                  className="text-white"
                />
              </Input>
            </FormControl>

            {mutation.isError ? (
              <Text className="text-destructive text-sm">Could not save address. Try again.</Text>
            ) : null}

            <View className="flex-row gap-3 pt-2">
              <Pressable
                onPress={onClose}
                className="flex-1 items-center rounded-xl border border-goojol-border py-3"
              >
                <Text className="font-medium text-goojol-muted">Cancel</Text>
              </Pressable>
              <Button
                className="flex-1 bg-goojol-coral data-[active=true]:bg-goojol-coral/90"
                onPress={onSave}
                isDisabled={mutation.isPending || !name.trim()}
              >
                <ButtonText className="font-semibold text-white">Save</ButtonText>
              </Button>
            </View>
          </VStack>
        </View>
      </View>
    </Modal>
  );
}
