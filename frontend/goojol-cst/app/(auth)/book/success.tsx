import { useRouter } from 'expo-router';
import { CheckCircle2 } from 'lucide-react-native';
import { View } from 'react-native';
import { Button, ButtonText } from '@/components/ui/button';
import { Heading } from '@/components/ui/heading';
import { Text } from '@/components/ui/text';
import { VStack } from '@/components/ui/vstack';
import { useBook } from '@/feature/book/book-context';
import { formatRupiah, WizardShell } from '@/feature/book/components/wizard-shell';

export default function BookSuccessScreen() {
  const router = useRouter();
  const { matchedDriver, quote, vehicleType, vehicleMaxSize, reset } = useBook();
  const selectedFare = quote?.vehicle_options.find(
    (option) => option.vehicle_type === vehicleType && option.max_size === vehicleMaxSize,
  )?.total_fare;

  const onDone = () => {
    reset();
    router.replace('/(auth)/(tabs)/home');
  };

  return (
    <WizardShell
      title="Driver found"
      currentStep={4}
      footer={
        <Button
          className="w-full bg-goojol-coral data-[active=true]:bg-goojol-coral/90"
          onPress={onDone}
        >
          <ButtonText className="font-semibold text-white">Done</ButtonText>
        </Button>
      }
    >
      <View className="flex-1 px-6 py-6">
        <VStack space="xl" className="items-center pt-8">
          <CheckCircle2 color="#3ddba8" size={64} />
          <VStack space="xs" className="items-center">
            <Heading size="xl" className="text-center text-white">
              Driver found!
            </Heading>
            <Text className="text-center text-goojol-muted">
              Your driver accepted the trip and is heading to the pickup.
            </Text>
          </VStack>
        </VStack>

        {matchedDriver ? (
          <View className="mt-8 rounded-2xl border border-goojol-border bg-goojol-surface p-4">
            <VStack space="sm">
              <Text className="font-semibold text-lg text-white">{matchedDriver.name}</Text>
              <Text className="text-goojol-muted">
                {matchedDriver.vehicle_name} · {matchedDriver.license_number}
              </Text>
              <Text className="text-goojol-muted">
                {matchedDriver.phone_number} · {matchedDriver.vehicle_type}
              </Text>
              {selectedFare != null ? (
                <Text className="mt-2 text-goojol-teal">
                  {formatRupiah(selectedFare)} estimated
                </Text>
              ) : null}
            </VStack>
          </View>
        ) : null}
      </View>
    </WizardShell>
  );
}
