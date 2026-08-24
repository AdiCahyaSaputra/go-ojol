import { useRouter } from "expo-router";
import { useEffect } from "react";
import { ActivityIndicator, Pressable, View } from "react-native";
import { Button, ButtonSpinner, ButtonText } from "@/components/ui/button";
import { Text } from "@/components/ui/text";
import { VStack } from "@/components/ui/vstack";
import { HStack } from "@/components/ui/hstack";
import { VEHICLE_OPTIONS } from "@/constants/book";
import { useBook } from "@/feature/book/book-context";
import {
  BookError,
  formatRupiah,
  WizardShell,
} from "@/feature/book/components/wizard-shell";
import { useCalculateArgoMutation } from "@/feature/book/dispatch.mutation";
import { truncate } from "@/lib/utils/string";

export default function BookQuoteScreen() {
  const router = useRouter();
  const { pickup, destination, vehicleId, setVehicleId, quote, setQuote } =
    useBook();
  const mutation = useCalculateArgoMutation();

  useEffect(() => {
    if (!destination) {
      router.replace("/book/destination");
    }
  }, [destination, router]);

  useEffect(() => {
    if (!destination) {
      return;
    }

    mutation.mutate(
      { pickup, destination, vehicleId },
      {
        onSuccess: (data) => setQuote(data),
      }
    );
  }, [destination, pickup, vehicleId, setQuote, mutation.mutate]);

  const onContinue = () => {
    if (quote) {
      router.push("/book/find-driver");
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
          isDisabled={!quote || mutation.isPending}
        >
          {mutation.isPending ? <ButtonSpinner /> : null}
          <ButtonText className="font-semibold text-white">
            Find driver
          </ButtonText>
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
                      ? "border-goojol-coral bg-goojol-coral/10"
                      : "border-goojol-border bg-goojol-surface"
                  }`}
                >
                  <Text
                    className={`text-center font-medium ${
                      selected ? "text-goojol-coral" : "text-white"
                    }`}
                  >
                    {option.label}
                  </Text>
                </Pressable>
              );
            })}
          </View>
        </VStack>

        <View className="rounded-2xl border border-goojol-border bg-goojol-surface p-4">
          <HStack space="lg" className="justify-between items-center">
            <View>
              <Text className="text-goojol-muted text-xs">Pickup</Text>
              <Text className="text-base text-white">
                {truncate(pickup.name, 20)}
              </Text>
            </View>

            <View className="flex-1 h-px border-goojol-coral border-t-4 border-dotted"/>

            <View>
              <Text className="text-goojol-muted text-xs">Destination</Text>
              <Text className="text-base text-white">
                {truncate(destination?.name ?? "—", 20)}
              </Text>
            </View>
          </HStack>
        </View>

        {mutation.isError ? (
          <BookError
            message={mutation.error.message ?? "Could not calculate fare."}
          />
        ) : null}

        <View className="rounded-2xl border border-goojol-teal/30 bg-goojol-teal/10 p-4">
          <Text className="text-goojol-muted text-sm">Estimated fare</Text>
          {mutation.isPending ? (
            <HStack className="items-center gap-2 py-2">
              <ActivityIndicator color="#ff6b4a" size="small" />
              <Text className="text-goojol-muted text-3xl font-bold">
                Calculating fare…
              </Text>
            </HStack>
          ) : quote ? (
            <>
              <Text className="font-bold text-3xl text-goojol-teal py-2">
                {formatRupiah(quote.total_fare)}
              </Text>

              <HStack className="justify-between">
                <Text className="mt-2 text-goojol-muted text-sm">
                  {quote.distance.toLocaleString("id-ID")} m
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
