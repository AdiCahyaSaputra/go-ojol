import { Link, Stack } from 'expo-router';
import { Text, View } from 'react-native';

export default function NotFoundScreen() {
  return (
    <>
      <Stack.Screen options={{ title: 'Oops!' }} />
      <View className="flex-1 items-center justify-center bg-white px-5 dark:bg-black">
        <Text className="font-bold text-black text-xl dark:text-white">
          This screen does not exist.
        </Text>
        <Link href="/" className="mt-4 py-3">
          <Text className="text-blue-600 text-sm dark:text-blue-400">Go to home screen</Text>
        </Link>
      </View>
    </>
  );
}
