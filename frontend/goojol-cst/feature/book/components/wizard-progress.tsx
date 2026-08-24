import { View } from 'react-native';
import { Text } from '@/components/ui/text';

type WizardProgressProps = {
  currentStep: 1 | 2 | 3 | 4;
};

const STEPS = ['Pickup', 'Destination', 'Quote', 'Driver'] as const;

export function WizardProgress({ currentStep }: WizardProgressProps) {
  return (
    <View className="px-6 pt-2 pb-6">
      <View className="mb-2 flex-row items-center gap-2">
        {STEPS.map((label, index) => {
          const stepNumber = index + 1;
          const isActive = stepNumber === currentStep;
          const isComplete = stepNumber < currentStep;

          return (
            <View
              key={label}
              className={`h-1.5 flex-1 rounded-full ${isComplete || isActive ? 'bg-goojol-coral' : 'bg-goojol-border'}`}
            />
          );
        })}
      </View>
      <Text className="text-goojol-muted text-xs">
        Step {currentStep} of 4 · {STEPS[currentStep - 1]}
      </Text>
    </View>
  );
}
