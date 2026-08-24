import { useRouter } from 'expo-router';
import { LucideChevronsLeftRightEllipsis } from 'lucide-react-native';
import { useEffect } from 'react';
import { ActivityIndicator, Pressable, View } from 'react-native';
import { Button, ButtonSpinner, ButtonText } from '@/components/ui/button';
import { HStack } from '@/components/ui/hstack';
import { Text } from '@/components/ui/text';
import { VStack } from '@/components/ui/vstack';
import { VEHICLE_OPTIONS } from '@/constants/book';
import { useBook } from '@/feature/book/book-context';
import { BookError, formatRupiah, WizardShell } from '@/feature/book/components/wizard-shell';
import { useCalculateArgoQuery } from '@/feature/book/dispatch.query';
import { truncate } from '@/lib/utils/string';

export default function BookQuoteScreen() {
  const router = useRouter();
  const { pickup, destination, vehicleId, vehicleType, setVehicleId, setQuote } = useBook();
  const quoteQuery = useCalculateArgoQuery({ pickup, destination, vehicleType });
  const quote = quoteQuery.data;

  useEffect(() => {
    if (!destination) {
      router.replace('/book/destination');
    }
  }, [destination, router]);

  useEffect(() => {
    if (quote) {
      setQuote(quote);
    }
  }, [quote, setQuote]);

  const onContinue = () => {
    if (quote) {
      router.push('/book/find-driver');
    }
  };

  return (
    <WizardShell
      title="Review fare"
      currentStep={3}
      footer={
        <Button
          className="w-full bg-goojol-coral data-[active=true]:bg-goojol-coral/90"
          onPress={onContinue}
          isDisabled={!quote || quoteQuery.isPending}
        >
          {quoteQuery.isPending ? <ButtonSpinner /> : null}
          <ButtonText className="font-semibold text-white">Find driver</ButtonText>
        </Button>
      }
    >
      <VStack space="lg" className="flex-1 px-6 py-4">
        <VStack space="sm">
          <Text className="text-goojol-muted text-sm">Vehicle</Text>
          <View className="flex-row gap-2">
            {VEHICLE_OPTIONS.map((option) => {
              const selected = option.id === vehicleId;
              return (
                <Pressable
                  key={option.id}
                  onPress={() => setVehicleId(option.id)}
                  className={`flex-1 rounded-xl border px-4 py-3 ${
                    selected
                      ? 'border-goojol-coral bg-goojol-coral/10'
                      : 'border-goojol-border bg-goojol-surface'
                  }`}
                >
                  <Text
                    className={`text-center font-medium ${
                      selected ? 'text-goojol-coral' : 'text-white'
                    }`}
                  >
                    {option.label}
                  </Text>
                </Pressable>
              );
            })}
          </View>
        </VStack>

        <View className="rounded-2xl border border-goojol-border bg-goojol-surface">
          <HStack space="lg" className="justify-between">
            <VStack className="w-1/3 p-4">
              <Text className="text-goojol-muted text-xs">Pickup</Text>
              <Text className="line-clamp-1 break-all text-base text-white">
                {truncate(pickup.name, 20)}
              </Text>
            </VStack>

            <VStack className="relative items-center justify-center">
              <View className="absolute inset-y-0 w-px bg-goojol-border" />
              <View className="rounded-full bg-goojol-surface p-1">
                <LucideChevronsLeftRightEllipsis color="#ff6b4a" size={16} />
              </View>
            </VStack>

            <VStack className="w-1/3 items-end p-4">
              <Text className="text-goojol-muted text-xs">Destination</Text>
              <Text className="line-clamp-1 break-all text-base text-white">
                {truncate(destination?.name ?? '—', 20)}
              </Text>
            </VStack>
          </HStack>
        </View>

        {quoteQuery.isError ? (
          <BookError message={quoteQuery.error.message ?? 'Could not calculate fare.'} />
        ) : null}

        <View className="rounded-2xl border border-goojol-teal/30 bg-goojol-teal/10 p-4">
          <Text className="text-goojol-muted text-sm">Estimated fare</Text>
          {quoteQuery.isPending ? (
            <HStack className="items-center gap-2 py-2">
              <ActivityIndicator color="#ff6b4a" size="small" />
              <Text className="font-bold text-base text-goojol-muted">Calculating fare…</Text>
            </HStack>
          ) : quote ? (
            <>
              <Text className="py-2 font-bold text-3xl text-goojol-teal">
                {formatRupiah(quote.total_fare)}
              </Text>

              <HStack className="justify-between">
                <Text className="mt-2 text-goojol-muted text-sm">
                  {quote.distance.toLocaleString('id-ID')} m
                </Text>

                <Text className="mt-2 text-goojol-muted text-sm">
                  ~{Math.ceil(quote.duration / 60)} min
                </Text>
              </HStack>
            </>
          ) : null}
        </View>
      </VStack>
    </WizardShell>
  );
}
