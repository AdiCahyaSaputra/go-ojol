import { Image, KeyboardAvoidingView, Platform, ScrollView, Text, View } from 'react-native';
import { Heading } from '@/components/ui/heading';
import { VStack } from '@/components/ui/vstack';
import LoginForm from './login-form';

export default function LoginPage() {
  return (
    <KeyboardAvoidingView
      style={{ flex: 1, backgroundColor: '#0F1729' }}
      behavior={Platform.OS === 'ios' ? 'padding' : undefined}
    >
      <ScrollView
        style={{ flex: 1 }}
        contentContainerStyle={{ flexGrow: 1, paddingBottom: 32 }}
        keyboardShouldPersistTaps="handled"
        showsVerticalScrollIndicator={false}
      >
        <View className="overflow-hidden rounded-b-3xl bg-goojol-sky">
          <Image
            source={require('@/assets/images/login/pixel-ojek-hero.png')}
            style={{ height: 208, width: '100%' }}
            resizeMode="contain"
            accessibilityLabel="Pixel art street scene with a scooter and a street sign"
          />
        </View>

        <VStack space="xl" className="px-6 pt-8">
          <VStack space="xs">
            <Heading size="2xl" className="text-center text-white">
              Your ride, Two taps away
            </Heading>
            <Text className="text-center text-goojol-muted">Get where you're going, faster.</Text>
          </VStack>

          <LoginForm />
        </VStack>
      </ScrollView>
    </KeyboardAvoidingView>
  );
}
