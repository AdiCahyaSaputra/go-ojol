import { useQueryClient } from '@tanstack/react-query';
import { useRouter } from 'expo-router';
import { Bell, ChevronRight, HelpCircle, LogOut, User, Wallet } from 'lucide-react-native';
import { useState } from 'react';
import { Alert, Pressable, View } from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';
import { Button, ButtonSpinner, ButtonText } from '@/components/ui/button';
import { Heading } from '@/components/ui/heading';
import { Text } from '@/components/ui/text';
import { VStack } from '@/components/ui/vstack';
import { useCurrentUserQuery } from '@/feature/user/user.query';
import { axiosClient } from '@/lib/api/axios-client';
import { clearSession } from '@/lib/auth/token-storage';

type SettingsRowProps = {
  icon: React.ReactNode;
  label: string;
  onPress: () => void;
};

function SettingsRow({ icon, label, onPress }: SettingsRowProps) {
  return (
    <Pressable
      onPress={onPress}
      className="flex-row items-center justify-between border-goojol-border border-b py-4 active:opacity-70"
    >
      <View className="flex-row items-center gap-3">
        {icon}
        <Text className="text-base text-white">{label}</Text>
      </View>
      <ChevronRight color="#8892a8" size={20} />
    </Pressable>
  );
}

function showComingSoon(feature: string) {
  Alert.alert('Coming soon', `${feature} will be available in a later phase.`);
}

export default function AccountPage() {
  const router = useRouter();
  const { data: user, isLoading } = useCurrentUserQuery();
  const [signingOut, setSigningOut] = useState(false);
  const queryClient = useQueryClient();

  const displayName = user?.driver?.name ?? 'Driver';
  const email = user?.email ?? '…';
  const vehicleLabel = user?.driver?.vehicle
    ? `${user.driver.vehicle.name} · ${user.driver.vehicle.license_number}`
    : null;

  const onSignOut = async () => {
    setSigningOut(true);
    try {
      await axiosClient.post('/api/auth/logout');
    } catch {
      // Clear local session even if the network call fails.
    } finally {
      await clearSession();
      queryClient.clear();
      setSigningOut(false);
      router.replace('/(public)/login');
    }
  };

  return (
    <SafeAreaView edges={['top']} style={{ flex: 1, backgroundColor: '#0F1729' }}>
      <View className="flex-1 px-6 pt-6">
        <Heading size="2xl" className="mb-6 text-white">
          Account
        </Heading>

        <View className="mb-8 flex-row items-center gap-4 rounded-2xl border border-goojol-border bg-goojol-surface p-4">
          <View className="rounded-full bg-goojol-border/50 p-3">
            <User color="#ff6b4a" size={28} />
          </View>
          <VStack space="xs" className="flex-1">
            <Text className="font-semibold text-lg text-white">
              {isLoading ? 'Loading…' : displayName}
            </Text>
            <Text className="text-goojol-muted text-sm">{email}</Text>
            {vehicleLabel ? (
              <Text className="text-goojol-muted text-xs">{vehicleLabel}</Text>
            ) : null}
          </VStack>
        </View>

        <View className="rounded-2xl border border-goojol-border bg-goojol-surface px-4">
          <SettingsRow
            icon={<Bell color="#8892a8" size={20} />}
            label="Notifications"
            onPress={() => showComingSoon('Notifications')}
          />
          <SettingsRow
            icon={<Wallet color="#8892a8" size={20} />}
            label="Payouts"
            onPress={() => showComingSoon('Payouts')}
          />
          <SettingsRow
            icon={<HelpCircle color="#8892a8" size={20} />}
            label="Help"
            onPress={() => showComingSoon('Help')}
          />
        </View>

        <Button
          className="mt-8 w-full border border-goojol-border bg-transparent data-[active=true]:bg-goojol-surface"
          onPress={onSignOut}
          isDisabled={signingOut}
          accessibilityLabel="Sign out"
        >
          {signingOut ? <ButtonSpinner /> : <LogOut color="#ff6b4a" size={18} />}
          <ButtonText className="font-semibold text-goojol-coral">Sign out</ButtonText>
        </Button>
      </View>
    </SafeAreaView>
  );
}
