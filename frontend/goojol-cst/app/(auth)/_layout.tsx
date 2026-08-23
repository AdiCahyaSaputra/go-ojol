import { Stack } from 'expo-router';
import { View } from 'react-native';

const ProtectedLayout = () => {
  return (
    <View className="flex-1">
      <Stack
        screenOptions={{
          headerShown: false,
          contentStyle: {
            backgroundColor: '#0F1729',
          },
          animation: 'fade',
        }}
      />
    </View>
  );
};

export default ProtectedLayout;
