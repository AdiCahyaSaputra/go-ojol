import { useRouter } from 'expo-router';
import { View } from 'react-native';
import { Button, ButtonText } from '@/components/ui/button';
import { useBook } from '@/feature/book/book-context';
import { LocationMap } from '@/feature/book/components/location-map';
import { WizardShell } from '@/feature/book/components/wizard-shell';
import { SavedAddressChips } from '@/feature/saved-address/components/saved-address-chips';

export default function BookPickupScreen() {
  const router = useRouter();
  const { pickup, setPickup } = useBook();

  return (
    <WizardShell
      title="Confirm pickup"
      currentStep={1}
      footer={
        <Button
          className="w-full bg-goojol-coral data-[active=true]:bg-goojol-coral/90"
          onPress={() => router.push('/book/destination')}
        >
          <ButtonText className="font-semibold text-white">Confirm pickup</ButtonText>
        </Button>
      }
    >
      <LocationMap label={pickup.name} lat={pickup.lat} lng={pickup.lng} />
      <View className="px-6 py-4">
        <SavedAddressChips selected={pickup} onSelect={setPickup} />
      </View>
    </WizardShell>
  );
}
