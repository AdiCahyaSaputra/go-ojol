import { KeyboardAvoidingView, Platform, ScrollView, Text, View } from 'react-native';
import { Heading } from '@/components/ui/heading';
import { VStack } from '@/components/ui/vstack';
import RegisterForm from './register-form';
import RegisterHero from './register-hero';

export default function RegisterPage() {
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
          <RegisterHero />
        </View>

        <VStack space="xl" className="px-6 pt-8">
          <VStack space="xs">
            <Heading size="2xl" className="text-center text-white">
              Join the ride
            </Heading>
            <Text className="text-center text-goojol-muted">Tell us about who are you</Text>
          </VStack>

          <RegisterForm />
        </VStack>
      </ScrollView>
    </KeyboardAvoidingView>
  );
}
