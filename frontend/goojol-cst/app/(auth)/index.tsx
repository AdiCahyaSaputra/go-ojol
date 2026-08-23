import { useRouter } from 'expo-router';
import { useEffect, useState } from 'react';
import { View } from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';
import { Button, ButtonSpinner, ButtonText } from '@/components/ui/button';
import { Heading } from '@/components/ui/heading';
import { Text } from '@/components/ui/text';
import { VStack } from '@/components/ui/vstack';
import { axiosClient } from '@/lib/api/axios-client';
import { type AuthSession, clearSession, getSession } from '@/lib/auth/token-storage';

const AppHome = () => {
  const router = useRouter();
  const [session, setSession] = useState<AuthSession | null>(null);
  const [signingOut, setSigningOut] = useState(false);

  useEffect(() => {
    getSession().then(setSession);
  }, []);

  const onSignOut = async () => {
    setSigningOut(true);
    try {
      await axiosClient.post('/api/auth/logout');
    } catch {
      // Clear local session even if the network call fails.
    } finally {
      await clearSession();
      setSigningOut(false);
      router.replace('/(public)/login');
    }
  };

  return (
    <SafeAreaView edges={['top']} style={{ flex: 1, backgroundColor: '#0F1729' }}>
      <View className="flex-1 px-6 pt-10">
        <VStack space="xl">
          <VStack space="xs">
            <Heading size="2xl" className="text-white">
              You're signed in
            </Heading>
            <Text className="text-goojol-muted">Role: {session?.role ?? '…'}</Text>
          </VStack>

          <Button
            className="w-full bg-goojol-coral data-[active=true]:bg-goojol-coral/90 data-[hover=true]:bg-goojol-coral/90"
            onPress={onSignOut}
            isDisabled={signingOut}
            accessibilityLabel="Sign out"
          >
            {signingOut ? <ButtonSpinner /> : null}
            <ButtonText className="font-semibold text-white">Sign out</ButtonText>
          </Button>
        </VStack>
      </View>
    </SafeAreaView>
  );
};

export default AppHome;
