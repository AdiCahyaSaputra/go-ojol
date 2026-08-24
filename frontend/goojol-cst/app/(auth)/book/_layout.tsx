import { Stack } from 'expo-router';

export default function BookLayout() {
  return (
    <Stack
      screenOptions={{
        headerShown: false,
        contentStyle: { backgroundColor: '#0F1729' },
        animation: 'slide_from_right',
      }}
    />
  );
}
