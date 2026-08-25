import { useRouter } from 'expo-router';
import { SafeAreaView } from 'react-native-safe-area-context';
import { Box } from '@/components/ui/box';
import { Button, ButtonText } from '@/components/ui/button';
import { Heading } from '@/components/ui/heading';
import { Text } from '@/components/ui/text';
import { VStack } from '@/components/ui/vstack';
import { clearSession } from '@/lib/auth/token-storage';

export default function HomePage() {
  const router = useRouter();

  const signOut = async () => {
    await clearSession();
    router.replace('/(public)/login');
  };

  return (
    <SafeAreaView edges={['top']} style={{ flex: 1, backgroundColor: '#0F1729' }}>
      <VStack className="flex-1 justify-between px-6 pt-8 pb-10">
        <VStack space="md">
          <Heading size="2xl" className="text-white" style={{ fontFamily: 'SpaceMono' }}>
            You're online soon
          </Heading>
          <Text className="text-base text-goojol-muted">
            Driver home is a placeholder for now. Auth is live — trip tools come next.
          </Text>
        </VStack>

        <Box className="gap-4 rounded-2xl border border-goojol-border bg-goojol-surface p-5">
          <Text className="font-semibold text-lg text-white">Dummy home</Text>
          <Text size="sm" className="text-goojol-muted">
            Sign-in and registration land here until the real driver dashboard ships.
          </Text>
          <Button
            variant="outline"
            className="border-goojol-border data-[active=true]:bg-goojol-surface"
            onPress={signOut}
            accessibilityLabel="Sign out"
          >
            <ButtonText className="font-semibold text-goojol-coral text-sm">Sign out</ButtonText>
          </Button>
        </Box>
      </VStack>
    </SafeAreaView>
  );
}
