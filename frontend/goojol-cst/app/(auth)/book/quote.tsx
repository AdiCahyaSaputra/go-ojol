import { useRouter } from 'expo-router';
import { Bike, Car, ChevronLeft } from 'lucide-react-native';
import { useEffect, useMemo } from 'react';
import { ActivityIndicator, Pressable, ScrollView, View } from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';
import { Button, ButtonSpinner, ButtonText } from '@/components/ui/button';
import { Heading } from '@/components/ui/heading';
import { HStack } from '@/components/ui/hstack';
import { Text } from '@/components/ui/text';
import { VStack } from '@/components/ui/vstack';
import { VEHICLE_OPTIONS } from '@/constants/book';
import { useBook } from '@/feature/book/book-context';
import { RouteMap } from '@/feature/book/components/route-map';
import { formatRupiah } from '@/feature/book/components/wizard-shell';
import { useCalculateArgoQuery } from '@/feature/book/dispatch.query';
import type { VehicleOption } from '@/feature/book/dispatch.schema';

function vehicleLabel(vehicleType: VehicleOption['vehicle_type']) {
  return VEHICLE_OPTIONS.find((option) => option.type === vehicleType)?.label ?? vehicleType;
}

function optionKey(option: VehicleOption) {
  return `${option.vehicle_type}-${option.max_size}`;
}

export default function BookQuoteScreen() {
  const router = useRouter();
  const { pickup, destination, vehicleType, vehicleMaxSize, setVehicleOption, setQuote } =
    useBook();
  const quoteQuery = useCalculateArgoQuery({ pickup, destination });
  const quote = quoteQuery.data;

  useEffect(() => {
    if (!destination) {
      router.replace('/book/location');
    }
  }, [destination, router]);

  useEffect(() => {
    if (quote) {
      setQuote(quote);
    }
  }, [quote, setQuote]);

  useEffect(() => {
    if (!quote || quote.vehicle_options.length === 0) {
      return;
    }

    const exactMatch = quote.vehicle_options.some(
      (option) => option.vehicle_type === vehicleType && option.max_size === vehicleMaxSize,
    );
    if (exactMatch) {
      return;
    }

    const sameType = quote.vehicle_options.find((option) => option.vehicle_type === vehicleType);
    const fallback = sameType ?? quote.vehicle_options[0];
    if (fallback) {
      setVehicleOption({ vehicleType: fallback.vehicle_type, maxSize: fallback.max_size });
    }
  }, [quote, setVehicleOption, vehicleMaxSize, vehicleType]);

  const selectedOption = useMemo(
    () =>
      quote?.vehicle_options.find(
        (option) => option.vehicle_type === vehicleType && option.max_size === vehicleMaxSize,
      ) ?? quote?.vehicle_options[0],
    [quote, vehicleMaxSize, vehicleType],
  );

  const onContinue = () => {
    if (!quote || !selectedOption) {
      return;
    }

    setVehicleOption({
      vehicleType: selectedOption.vehicle_type,
      maxSize: selectedOption.max_size,
    });
    router.push('/book/find-driver');
  };

  if (!destination) {
    return null;
  }

  return (
    <View className="flex-1 bg-goojol-sky">
      <RouteMap
        pickup={pickup}
        destination={destination}
        path={quote?.path ?? []}
        padding={{ top: 96, right: 28, bottom: 340, left: 28 }}
      />

      <SafeAreaView edges={['bottom']} className="absolute inset-x-0 bottom-0">
        <View className="flex-row items-center gap-0 px-4 py-2">
          <Pressable
            onPress={() => router.back()}
            className="bg-goojol-sky/90 p-2"
            accessibilityLabel="Go back"
          >
            <ChevronLeft color="#e8ecf4" size={22} />
          </Pressable>
          <Heading size="md" className="text-white">
            Choose a ride
          </Heading>
        </View>
        <View className="border-goojol-border border-t bg-goojol-sky px-5 pt-4 pb-4">
          <VStack space="md">
            <HStack className="items-start justify-between gap-3">
              <VStack className="min-w-0 flex-1">
                <Text className="text-goojol-muted text-xs">Pickup</Text>
                <Text className="text-base text-white" numberOfLines={1}>
                  {pickup.name}
                </Text>
              </VStack>
              <VStack className="min-w-0 flex-1 items-end">
                <Text className="text-goojol-muted text-xs">Drop-off</Text>
                <Text className="text-right text-base text-white" numberOfLines={1}>
                  {destination.name}
                </Text>
              </VStack>
            </HStack>

            {quote ? (
              <HStack className="justify-between">
                <Text className="text-goojol-muted text-sm">
                  {(quote.distance / 1000).toLocaleString('id-ID', { maximumFractionDigits: 1 })} km
                </Text>
                <Text className="text-goojol-muted text-sm">
                  ~{Math.ceil(quote.duration / 60)} min
                </Text>
              </HStack>
            ) : null}

            {quoteQuery.isError ? (
              <View className="rounded-xl border border-destructive/40 bg-destructive/10 px-4 py-3">
                <Text className="text-destructive text-sm">
                  {quoteQuery.error.message ?? 'Could not calculate fare.'}
                </Text>
              </View>
            ) : null}

            {quoteQuery.isPending ? (
              <HStack className="items-center gap-2 py-3">
                <ActivityIndicator color="#ff6b4a" size="small" />
                <Text className="text-goojol-muted">Calculating fares…</Text>
              </HStack>
            ) : null}

            {quote ? (
              <ScrollView
                horizontal
                showsHorizontalScrollIndicator={false}
                contentContainerStyle={{ gap: 10, paddingBottom: 10 }}
              >
                {quote.vehicle_options.map((option) => {
                  const selected =
                    selectedOption != null && optionKey(option) === optionKey(selectedOption);
                  const Icon = option.vehicle_type === 'car' ? Car : Bike;

                  return (
                    <Pressable
                      key={optionKey(option)}
                      onPress={() =>
                        setVehicleOption({
                          vehicleType: option.vehicle_type,
                          maxSize: option.max_size,
                        })
                      }
                      className={`w-40 rounded-2xl border px-4 py-3 ${
                        selected
                          ? 'border-goojol-coral bg-goojol-coral/15'
                          : 'border-goojol-border bg-goojol-surface'
                      }`}
                      accessibilityRole="button"
                      accessibilityState={{ selected }}
                      accessibilityLabel={`${vehicleLabel(option.vehicle_type)}, ${option.max_size} seats, ${formatRupiah(option.total_fare)}`}
                    >
                      <HStack className="items-center gap-2">
                        <Icon color={selected ? '#ff6b4a' : '#e8ecf4'} size={18} />
                        <Text
                          className={`font-medium ${selected ? 'text-goojol-coral' : 'text-white'}`}
                        >
                          {vehicleLabel(option.vehicle_type)}
                        </Text>
                      </HStack>
                      <Text className="mt-1 text-goojol-muted text-xs">
                        Up to {option.max_size} {option.max_size === 1 ? 'person' : 'people'}
                      </Text>
                      <Text className="mt-2 font-bold text-goojol-teal text-lg">
                        {formatRupiah(option.total_fare)}
                      </Text>
                    </Pressable>
                  );
                })}
              </ScrollView>
            ) : null}

            <Button
              className="w-full bg-goojol-coral data-[active=true]:bg-goojol-coral/90"
              onPress={onContinue}
              isDisabled={!selectedOption || quoteQuery.isPending}
            >
              {quoteQuery.isPending ? <ButtonSpinner /> : null}
              <ButtonText className="font-semibold text-white">Find driver</ButtonText>
            </Button>
          </VStack>
        </View>
      </SafeAreaView>
    </View>
  );
}
