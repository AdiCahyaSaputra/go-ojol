import { Stack } from 'expo-router';
import { View } from 'react-native';
import { StandbyProvider } from '@/feature/standby/standby-context';

const ProtectedLayout = () => {
  return (
    <StandbyProvider>
      <View className="flex-1">
        <Stack
          screenOptions={{
            headerShown: false,
            contentStyle: {
              backgroundColor: '#0F1729',
            },
            animation: 'fade',
          }}
        >
          <Stack.Screen name="(tabs)" />
        </Stack>
      </View>
    </StandbyProvider>
  );
};

export default ProtectedLayout;
