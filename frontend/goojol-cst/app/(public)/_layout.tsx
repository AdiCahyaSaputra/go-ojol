import { Stack } from 'expo-router';
import { View } from 'react-native';

const PublicLayout = () => {
  return (
    <View className="flex-1">
      <Stack
        screenOptions={{
          headerShown: false,
          contentStyle: {
            backgroundColor: '#0F1729',
          },
          animation: 'slide_from_right',
          animationTypeForReplace: 'pop',
        }}
      />
    </View>
  );
};

export default PublicLayout;
