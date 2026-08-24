import { useRouter } from 'expo-router';
import { Car } from 'lucide-react-native';
import { View } from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';
import { Button, ButtonText } from '@/components/ui/button';
import { Heading } from '@/components/ui/heading';
import { Text } from '@/components/ui/text';
import { VStack } from '@/components/ui/vstack';

export default function RidesPage() {
  const router = useRouter();

  return (
    <SafeAreaView edges={['top']} style={{ flex: 1, backgroundColor: '#0F1729' }}>
      <View className="flex-1 px-6 pt-6">
        <Heading size="2xl" className="mb-6 text-white">
          Rides
        </Heading>

        <View className="flex-1 items-center justify-center pb-24">
          <VStack space="lg" className="items-center">
            <View className="rounded-full bg-goojol-surface p-6">
              <Car color="#ff6b4a" size={40} />
            </View>

            <VStack space="xs" className="items-center">
              <Heading size="lg" className="text-center text-white">
                No rides yet
              </Heading>
              <Text className="max-w-xs text-center text-goojol-muted">
                Your past and active rides will show up here once trip history is connected.
              </Text>
            </VStack>

            <Button
              className="w-full max-w-xs bg-goojol-coral data-[active=true]:bg-goojol-coral/90"
              onPress={() => router.push('/book/pickup')}
              accessibilityLabel="Book a ride"
            >
              <ButtonText className="font-semibold text-white">Book a ride</ButtonText>
            </Button>
          </VStack>
        </View>
      </View>
    </SafeAreaView>
  );
}
