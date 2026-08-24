import { useRouter } from 'expo-router';
import { ChevronRight } from 'lucide-react-native';
import { Image, Pressable, ScrollView, View } from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';
import { Heading } from '@/components/ui/heading';
import { Text } from '@/components/ui/text';
import { VStack } from '@/components/ui/vstack';
import { useCurrentUserQuery } from '@/feature/user/user.query';

export default function HomePage() {
  const router = useRouter();
  const { data: user } = useCurrentUserQuery();

  const displayName = user?.customer?.name ?? 'there';

  const startBook = () => {
    router.push('/book/location');
  };

  return (
    <SafeAreaView edges={['top']} style={{ flex: 1, backgroundColor: '#0F1729' }}>
      <ScrollView
        className="flex-1"
        contentContainerStyle={{ flexGrow: 1, paddingBottom: 24 }}
        showsVerticalScrollIndicator={false}
      >
        <VStack space="lg" className="flex-1 px-6 pt-6">
          <VStack space="xs">
            <Heading size="2xl" className="text-white">
              Hi, {displayName}
            </Heading>
            <Text className="text-goojol-muted">Ready when you are.</Text>
          </VStack>

          <Pressable
            onPress={startBook}
            className="flex-row items-center justify-between rounded-2xl border border-goojol-border bg-goojol-surface px-5 py-5 active:border-goojol-coral"
            accessibilityLabel="Where to?"
          >
            <VStack space="xs">
              <Text className="text-goojol-muted text-sm">Book a ride</Text>
              <Text className="font-semibold text-white text-xl">Where to?</Text>
            </VStack>
            <ChevronRight color="#ff6b4a" size={24} />
          </Pressable>

          <View className="mt-auto overflow-hidden rounded-2xl bg-goojol-surface">
            <Image
              source={require('@/assets/images/login/pixel-ojek-hero.png')}
              style={{ height: 120, width: '100%' }}
              resizeMode="cover"
              accessibilityLabel="Pixel art street scene"
            />
            <View className="px-4 py-3">
              <Text className="font-medium text-sm text-white">Your ride, two taps away</Text>
              <Text className="text-goojol-muted text-xs">
                Fast pickup · Fair fares · Local drivers
              </Text>
            </View>
          </View>
        </VStack>
      </ScrollView>
    </SafeAreaView>
  );
}
