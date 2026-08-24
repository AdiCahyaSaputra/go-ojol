import { useRouter } from 'expo-router';
import { ChevronLeft } from 'lucide-react-native';
import { Pressable, View } from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';
import { Heading } from '@/components/ui/heading';
import { Text } from '@/components/ui/text';
import { WizardProgress } from './wizard-progress';

type WizardShellProps = {
  title: string;
  currentStep: 1 | 2 | 3 | 4;
  children: React.ReactNode;
  footer?: React.ReactNode;
};

export function WizardShell({ title, currentStep, children, footer }: WizardShellProps) {
  const router = useRouter();

  return (
    <SafeAreaView edges={['top', 'bottom']} style={{ flex: 1, backgroundColor: '#0F1729' }}>
      <View className="flex-row items-center gap-3 px-4 py-2">
        <Pressable
          onPress={() => router.back()}
          className="rounded-full p-2 active:bg-goojol-surface"
          accessibilityLabel="Go back"
        >
          <ChevronLeft color="#8892a8" size={24} />
        </Pressable>
        <View className="flex-1">
          <Heading size="lg" className="text-white">
            {title}
          </Heading>
        </View>
      </View>

      <WizardProgress currentStep={currentStep} />

      <View className="flex-1">{children}</View>

      {footer ? (
        <View className="border-goojol-border border-t bg-goojol-sky px-6 py-4">{footer}</View>
      ) : null}
    </SafeAreaView>
  );
}

export function formatRupiah(amount: number) {
  return new Intl.NumberFormat('id-ID', {
    style: 'currency',
    currency: 'IDR',
    maximumFractionDigits: 0,
  }).format(amount);
}

export function BookError({ message }: { message: string }) {
  return (
    <View className="mx-6 rounded-xl border border-destructive/40 bg-destructive/10 px-4 py-3">
      <Text className="text-destructive text-sm">{message}</Text>
    </View>
  );
}
