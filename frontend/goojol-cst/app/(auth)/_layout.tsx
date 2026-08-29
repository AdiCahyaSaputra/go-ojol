import { Stack } from 'expo-router';
import { View } from 'react-native';
import { ActiveTripRecovery } from '@/feature/book/active-trip-recovery';
import { BookProvider } from '@/feature/book/book-context';

const ProtectedLayout = () => {
  return (
    <BookProvider>
      <ActiveTripRecovery />
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
          <Stack.Screen
            name="book"
            options={{
              animation: 'slide_from_right',
            }}
          />
        </Stack>
      </View>
    </BookProvider>
  );
};

export default ProtectedLayout;
