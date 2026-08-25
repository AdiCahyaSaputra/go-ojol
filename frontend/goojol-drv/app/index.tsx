import { usePathname, useRouter } from 'expo-router';
import { useEffect, useState } from 'react';
import { Text, View } from 'react-native';
import { getSession } from '@/lib/auth/token-storage';

const EntryPoint = () => {
  const currentPath = usePathname();
  const router = useRouter();
  const [checking, setChecking] = useState(true);

  useEffect(() => {
    let cancelled = false;

    (async () => {
      const session = await getSession();
      if (cancelled) {
        return;
      }

      if (session?.refreshToken) {
        router.replace('/(auth)/home');
      } else if (currentPath !== '/login') {
        router.replace('/(public)/login');
      }
      setChecking(false);
    })();

    return () => {
      cancelled = true;
    };
  }, [currentPath, router]);

  if (checking) {
    return (
      <View className="flex-1 items-center justify-center bg-goojol-sky">
        <Text className="text-goojol-muted">Loading…</Text>
      </View>
    );
  }

  return (
    <View className="flex-1 items-center justify-center bg-goojol-sky">
      <Text className="text-goojol-muted">Loading…</Text>
    </View>
  );
};

export default EntryPoint;
