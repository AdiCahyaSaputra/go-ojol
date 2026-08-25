import { Box } from '@/components/ui/box';
import { HStack } from '@/components/ui/hstack';
import { Text } from '@/components/ui/text';
import { VStack } from '@/components/ui/vstack';

type RegisterStepIndicatorProps = {
  step: 1 | 2;
};

export default function RegisterStepIndicator({ step }: RegisterStepIndicatorProps) {
  return (
    <HStack space="md" className="items-center" accessibilityRole="summary">
      <VStack space="sm" className="flex-1">
        <Box className={`h-1 rounded-full ${step === 1 ? 'bg-goojol-coral' : 'bg-goojol-teal'}`} />
        <Text size="xs" className={step === 1 ? 'font-semibold text-white' : 'text-goojol-muted'}>
          Your Profile
        </Text>
      </VStack>
      <VStack space="sm" className="flex-1">
        <Box
          className={`h-1 rounded-full ${step === 2 ? 'bg-goojol-coral' : 'bg-goojol-border'}`}
        />
        <Text size="xs" className={step === 2 ? 'font-semibold text-white' : 'text-goojol-muted'}>
          Vehicle Information
        </Text>
      </VStack>
    </HStack>
  );
}
