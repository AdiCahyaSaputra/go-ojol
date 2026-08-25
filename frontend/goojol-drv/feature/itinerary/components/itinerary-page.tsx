import { Route } from 'lucide-react-native';
import { View } from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';
import { Heading } from '@/components/ui/heading';
import { Text } from '@/components/ui/text';
import { VStack } from '@/components/ui/vstack';

export default function ItineraryPage() {
  return (
    <SafeAreaView edges={['top']} style={{ flex: 1, backgroundColor: '#0F1729' }}>
      <View className="flex-1 px-6 pt-6">
        <Heading size="2xl" className="mb-6 text-white">
          Itinerary
        </Heading>

        <View className="flex-1 items-center justify-center pb-24">
          <VStack space="lg" className="items-center">
            <View className="rounded-full bg-goojol-surface p-6">
              <Route color="#ff6b4a" size={40} />
            </View>

            <VStack space="xs" className="items-center">
              <Heading size="lg" className="text-center text-white">
                No trips yet
              </Heading>
              <Text className="max-w-xs text-center text-goojol-muted">
                Completed and cancelled trips will show up here once trip history is connected.
              </Text>
            </VStack>
          </VStack>
        </View>
      </View>
    </SafeAreaView>
  );
}
