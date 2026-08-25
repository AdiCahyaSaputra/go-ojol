import { View } from 'react-native';
import { Button, ButtonSpinner, ButtonText } from '@/components/ui/button';
import { HStack } from '@/components/ui/hstack';
import { Text } from '@/components/ui/text';
import { VStack } from '@/components/ui/vstack';
import { useStandby } from '@/feature/standby/standby-context';

function formatRupiah(amount: number) {
  return new Intl.NumberFormat('id-ID', {
    style: 'currency',
    currency: 'IDR',
    maximumFractionDigits: 0,
  }).format(amount);
}

export function StandbySheet() {
  const {
    phase,
    coords,
    offer,
    offerSecondsLeft,
    isBusy,
    error,
    goOnline,
    goOffline,
    simulateOffer,
    acceptOffer,
    rejectOffer,
    completeTrip,
  } = useStandby();

  return (
    <View className="border-goojol-border border-t bg-goojol-sky px-5 pt-4 pb-4">
      <VStack space="md">
        {error ? (
          <View className="rounded-xl border border-destructive/40 bg-destructive/10 px-4 py-3">
            <Text className="text-destructive text-sm">{error}</Text>
          </View>
        ) : null}

        {phase === 'offline' && (
          <>
            <VStack space="xs">
              <Text className="font-semibold text-lg text-white">You're offline</Text>
              <Text className="text-goojol-muted text-sm">
                Go online to appear in nearby search and start receiving trip offers.
              </Text>
              {coords && !coords.fromGps ? (
                <Text className="text-goojol-muted text-xs">
                  Using demo location (GPS unavailable).
                </Text>
              ) : null}
            </VStack>
            <Button
              className="w-full bg-goojol-coral data-[active=true]:bg-goojol-coral/90"
              onPress={goOnline}
              isDisabled={isBusy}
              accessibilityLabel="Go online"
            >
              {isBusy ? <ButtonSpinner /> : null}
              <ButtonText className="font-semibold text-white">Go online</ButtonText>
            </Button>
          </>
        )}

        {phase === 'online' && (
          <>
            <HStack className="items-center justify-between">
              <VStack space="xs" className="flex-1">
                <HStack className="items-center gap-2">
                  <View className="h-2.5 w-2.5 rounded-full bg-goojol-teal" />
                  <Text className="font-semibold text-lg text-white">Online</Text>
                </HStack>
                <Text className="text-goojol-muted text-sm">Waiting for nearby trip requests.</Text>
              </VStack>
            </HStack>
            <HStack className="gap-3">
              <Button
                variant="outline"
                className="flex-1 border-goojol-border bg-goojol-surface"
                onPress={goOffline}
                isDisabled={isBusy}
                accessibilityLabel="Go offline"
              >
                {isBusy ? <ButtonSpinner /> : null}
                <ButtonText className="font-semibold text-white">Go offline</ButtonText>
              </Button>
              <Button
                className="flex-1 bg-goojol-surface data-[active=true]:bg-goojol-border"
                onPress={simulateOffer}
                isDisabled={isBusy}
                accessibilityLabel="Simulate order"
              >
                <ButtonText className="font-semibold text-goojol-coral">Simulate order</ButtonText>
              </Button>
            </HStack>
          </>
        )}

        {phase === 'offer' && offer && (
          <>
            <HStack className="items-start justify-between gap-3">
              <VStack space="xs" className="min-w-0 flex-1">
                <Text className="font-semibold text-lg text-white">New trip offer</Text>
                <Text className="text-goojol-muted text-sm">{offer.customerName}</Text>
              </VStack>
              <View className="rounded-xl border border-goojol-border bg-goojol-surface px-3 py-2">
                <Text className="font-bold text-goojol-coral text-lg">{offerSecondsLeft}s</Text>
              </View>
            </HStack>

            <HStack className="items-start justify-between gap-3">
              <VStack className="min-w-0 flex-1">
                <Text className="text-goojol-muted text-xs">Pickup</Text>
                <Text className="text-base text-white" numberOfLines={1}>
                  {offer.pickup.name}
                </Text>
              </VStack>
              <VStack className="min-w-0 flex-1 items-end">
                <Text className="text-goojol-muted text-xs">Drop-off</Text>
                <Text className="text-right text-base text-white" numberOfLines={1}>
                  {offer.destination.name}
                </Text>
              </VStack>
            </HStack>

            <HStack className="justify-between">
              <Text className="text-goojol-muted text-sm">
                {(offer.distanceM / 1000).toLocaleString('id-ID', { maximumFractionDigits: 1 })} km
              </Text>
              <Text className="font-bold text-base text-goojol-teal">
                {formatRupiah(offer.totalFare)}
              </Text>
            </HStack>

            <HStack className="gap-3">
              <Button
                variant="outline"
                className="flex-1 border-goojol-border bg-goojol-surface"
                onPress={rejectOffer}
                accessibilityLabel="Reject offer"
              >
                <ButtonText className="font-semibold text-red-600">Reject</ButtonText>
              </Button>
              <Button
                className="flex-1 bg-goojol-coral data-[active=true]:bg-goojol-coral/90"
                onPress={acceptOffer}
                accessibilityLabel="Accept offer"
              >
                <ButtonText className="font-semibold text-white">Accept</ButtonText>
              </Button>
            </HStack>
          </>
        )}

        {phase === 'accepted' && offer && (
          <>
            <VStack space="xs">
              <Text className="font-semibold text-lg text-white">Head to pickup</Text>
              <Text className="text-goojol-muted text-sm">
                Navigate to {offer.pickup.name}. Mock trip - complete when ready.
              </Text>
            </VStack>
            <HStack className="items-start justify-between gap-3">
              <VStack className="min-w-0 flex-1">
                <Text className="text-goojol-muted text-xs">Pickup</Text>
                <Text className="text-base text-white" numberOfLines={1}>
                  {offer.pickup.name}
                </Text>
              </VStack>
              <VStack className="min-w-0 flex-1 items-end">
                <Text className="text-goojol-muted text-xs">Fare</Text>
                <Text className="font-bold text-base text-goojol-teal">
                  {formatRupiah(offer.totalFare)}
                </Text>
              </VStack>
            </HStack>
            <Button
              className="w-full bg-goojol-coral data-[active=true]:bg-goojol-coral/90"
              onPress={completeTrip}
              accessibilityLabel="Complete mock trip"
            >
              <ButtonText className="font-semibold text-white">Complete</ButtonText>
            </Button>
          </>
        )}
      </VStack>
    </View>
  );
}
