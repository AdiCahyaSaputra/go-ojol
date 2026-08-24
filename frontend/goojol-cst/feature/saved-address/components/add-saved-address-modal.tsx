import { useState } from 'react';
import { Modal, Pressable, View } from 'react-native';
import { Button, ButtonText } from '@/components/ui/button';
import { FormControl, FormControlLabel, FormControlLabelText } from '@/components/ui/form-control';
import { Heading } from '@/components/ui/heading';
import { Input, InputField } from '@/components/ui/input';
import { Text } from '@/components/ui/text';
import { VStack } from '@/components/ui/vstack';
import { useUpsertSavedAddressMutation } from '../saved-address.query';

type AddSavedAddressModalProps = {
  open: boolean;
  onClose: () => void;
};

export function AddSavedAddressModal({ open, onClose }: AddSavedAddressModalProps) {
  const [name, setName] = useState('');
  const [lat, setLat] = useState('-6.2088');
  const [lng, setLng] = useState('106.8456');
  const mutation = useUpsertSavedAddressMutation();

  const onSave = async () => {
    if (!name.trim()) {
      return;
    }

    await mutation.mutateAsync({ name: name.trim(), lat, lng });
    setName('');
    onClose();
  };

  return (
    <Modal visible={open} animationType="slide" transparent onRequestClose={onClose}>
      <View className="flex-1 justify-end bg-black/50">
        <View className="rounded-t-3xl bg-goojol-sky px-6 pt-6 pb-10">
          <Heading size="lg" className="mb-4 text-white">
            Add pickup place
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
