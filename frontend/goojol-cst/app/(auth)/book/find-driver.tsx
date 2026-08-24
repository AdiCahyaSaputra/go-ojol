import { useRouter } from 'expo-router';
import { useEffect, useRef, useState } from 'react';
import { ActivityIndicator, View } from 'react-native';
import { Button, ButtonText } from '@/components/ui/button';
import { Text } from '@/components/ui/text';
import { VStack } from '@/components/ui/vstack';
import { useBook } from '@/feature/book/book-context';
import { BookError, WizardShell } from '@/feature/book/components/wizard-shell';
import { useFindDriverMutation } from '@/feature/book/dispatch.mutation';

export default function BookFindDriverScreen() {
  const router = useRouter();
  const { pickup, vehicleType, setMatchedDriver } = useBook();
  const mutation = useFindDriverMutation();
  const [noDrivers, setNoDrivers] = useState(false);
  const hasRequested = useRef(false);

  useEffect(() => {
    if (hasRequested.current) {
      return;
    }
    hasRequested.current = true;

    mutation.mutate(
      { pickup, vehicleType },
      {
        onSuccess: (data) => {
          const driver = data.drivers[0];
          if (driver) {
            setMatchedDriver(driver);
            router.replace('/book/success');
            return;
          }
          setNoDrivers(true);
        },
      },
    );
  }, [pickup, vehicleType, setMatchedDriver, router, mutation]);

  return (
    <WizardShell
      title="Finding driver"
      currentStep={4}
      footer={
        noDrivers || mutation.isError ? (
          <Button
            className="w-full border border-goojol-border bg-goojol-surface data-[active=true]:bg-goojol-surface/80"
            onPress={() => router.back()}
          >
            <ButtonText className="font-semibold text-white">Go back</ButtonText>
          </Button>
        ) : undefined
      }
    >
      <View className="flex-1 items-center justify-center px-6">
        {!noDrivers && !mutation.isError ? (
          <VStack space="lg" className="items-center">
            <ActivityIndicator color="#ff6b4a" size="large" />
            <Text className="font-medium text-lg text-white">Looking for nearby drivers…</Text>
            <Text className="text-center text-goojol-muted">
              Searching within 3 km of your pickup point.
            </Text>
          </VStack>
        ) : null}

        {noDrivers ? (
          <BookError message="No drivers nearby. Ask a driver to go online in the demo area, then try again." />
        ) : null}

        {mutation.isError ? (
          <BookError
            message={mutation.error.message ?? 'Could not find a driver. Try again later.'}
          />
        ) : null}
      </View>
    </WizardShell>
  );
}
