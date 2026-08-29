import { useRouter } from 'expo-router';
import { useEffect, useRef, useState } from 'react';
import { ActivityIndicator, View } from 'react-native';
import { Button, ButtonText } from '@/components/ui/button';
import { Heading } from '@/components/ui/heading';
import { Text } from '@/components/ui/text';
import { VStack } from '@/components/ui/vstack';
import { LOCATION_HEARTBEAT_MS } from '@/constants/book';
import { useBook } from '@/feature/book/book-context';
import { TripMap } from '@/feature/book/components/trip-map';
import { formatRupiah, WizardShell } from '@/feature/book/components/wizard-shell';
import { resolveCurrentLocation } from '@/feature/book/current-location';
import { CustomerDispatchSocket } from '@/feature/book/dispatch-ws';
import { fetchActiveTransaction } from '@/feature/book/trip.service';
import { loadStoredActiveTrip, saveStoredActiveTrip } from '@/feature/book/trip-storage';
import { getSession } from '@/lib/auth/token-storage';

type TripPhase = 'accepted_offer' | 'on_the_way' | 'completed' | 'cancelled';

const statusLabel = {
  accepted_offer: 'Driver heading to pickup',
  on_the_way: 'On the way to destination',
  completed: 'Trip complete — paid',
  cancelled: 'Trip cancelled',
  default: 'Trip in progress',
};

export default function BookTripScreen() {
  const router = useRouter();
  const {
    pickup,
    destination,
    matchedDriver,
    transactionId,
    quote,
    vehicleType,
    vehicleMaxSize,
    reset,
    setMatchedDriver,
    setTransactionId,
  } = useBook();
  const [tripStatus, setTripStatus] = useState<TripPhase>('accepted_offer');
  const [driverLocation, setDriverLocation] = useState<{ lat: number; lng: number } | null>(null);
  const [customerLocation, setCustomerLocation] = useState<{ lat: number; lng: number } | null>(
    null,
  );

  const [initCancelled, setInitCancelled] = useState(false);
  const [totalFare, setTotalFare] = useState<number | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const socketRef = useRef<CustomerDispatchSocket | null>(null);
  const heartbeatRef = useRef<ReturnType<typeof setInterval> | null>(null);
  const transactionIdRef = useRef<string | null>(transactionId);

  useEffect(() => {
    transactionIdRef.current = transactionId;
  }, [transactionId]);

  const selectedFare =
    totalFare ??
    quote?.vehicle_options.find(
      (option) => option.vehicle_type === vehicleType && option.max_size === vehicleMaxSize,
    )?.total_fare ??
    null;

  const clearHeartbeat = () => {
    if (heartbeatRef.current) {
      clearInterval(heartbeatRef.current);
      heartbeatRef.current = null;
    }
  };

  const persistTrip = async () => {
    if (!transactionIdRef.current || !destination) {
      return;
    }
    await saveStoredActiveTrip({
      transactionId: transactionIdRef.current,
      pickup,
      destination,
      matchedDriver,
      totalFare: selectedFare,
    });
  };

  const startHeartbeat = (socket: CustomerDispatchSocket, txId: string) => {
    clearHeartbeat();
    heartbeatRef.current = setInterval(async () => {
      const next = await resolveCurrentLocation();
      const lat = Number.parseFloat(next.lat);
      const lng = Number.parseFloat(next.lng);
      setCustomerLocation({ lat, lng });
      socket.sendTripLocation(txId, { lat, lng });
    }, LOCATION_HEARTBEAT_MS);
  };

  const resetActiveTransaction = async (txId: string | null) => {
    const active = await fetchActiveTransaction();
    if (!txId) {
      setTransactionId(active.id);
    }

    setTotalFare(active.total_fare);

    if (active.customer_last_lat_long) {
      setCustomerLocation({
        lat: active.customer_last_lat_long[0],
        lng: active.customer_last_lat_long[1],
      });
    }

    if (active.driver) {
      setMatchedDriver(active.driver);
    }

    return active;
  };

  const init = async (cancelled: boolean) => {
    setIsLoading(true);
    setError(null);

    try {
      let txId = transactionId;
      let status: TripPhase = 'accepted_offer';
      let nextDriverLocation: { lat: number; lng: number } | null = null;

      if (!destination) {
        const stored = await loadStoredActiveTrip();
        if (stored?.destination) {
          // destination missing from context — user should re-open from recovery flow
          router.replace('/(auth)/(tabs)/home');
          return;
        }
        router.replace('/book/location');
        return;
      }

      // @1 Try to use local active trip
      if (!txId) {
        const stored = await loadStoredActiveTrip();
        if (stored?.transactionId) {
          txId = stored.transactionId;
          setTransactionId(stored.transactionId);
          if (stored.matchedDriver) {
            setMatchedDriver(stored.matchedDriver);
          }
        }
      }

      try {
        const activeTrx = await resetActiveTransaction(txId);

        // @2 If it still doesn't exist, then try to get from API
        if (!txId) {
          txId = activeTrx.id;
        }
        status = activeTrx.status as TripPhase;
        nextDriverLocation = {
          lat: activeTrx.driver_last_lat_long[0],
          lng: activeTrx.driver_last_lat_long[1],
        };
      } catch {
        router.replace('/(auth)/(tabs)/home');
        return;
      }

      transactionIdRef.current = txId;
      setTripStatus(status);
      if (nextDriverLocation) {
        setDriverLocation(nextDriverLocation);
      }

      const session = await getSession();
      if (!session?.accessToken) {
        throw new Error('Missing session. Sign in again.');
      }

      const socket = new CustomerDispatchSocket();
      socketRef.current = socket;

      const sameTransactionId = (evtTxId: string) => evtTxId === transactionIdRef.current;

      await new Promise<void>((resolve, reject) => {
        socket.connect(session.accessToken, {
          onDriverLocation: (event) => {
            if (sameTransactionId(event.transactionId)) {
              return;
            }
            setDriverLocation({ lat: event.lat, lng: event.lng });
          },
          onTripStatus: (event) => {
            if (sameTransactionId(event.transactionId)) {
              return;
            }
            setTripStatus(event.status as TripPhase);
          },
          onTripCompleted: (event) => {
            if (sameTransactionId(event.transactionId)) {
              return;
            }
            setTotalFare(event.totalFare);
            setTripStatus('completed');
            clearHeartbeat();
          },
          onError: (message) => setError(message),
          onClose: () => setError('Connection lost. Location updates paused.'),
        });
        socket.whenOpen(() => resolve());
        setTimeout(() => reject(new Error('WebSocket timed out')), 12_000);
      });

      if (cancelled) {
        return;
      }

      startHeartbeat(socket, txId);
      await persistTrip();
    } catch (err) {
      if (!cancelled) {
        setError(err instanceof Error ? err.message : 'Could not load trip');
      }
    } finally {
      if (!cancelled) {
        setIsLoading(false);
      }
    }
  };

  useEffect(() => {
    setInitCancelled(false);

    void init(initCancelled);

    return () => {
      setInitCancelled(true);
      clearHeartbeat();
      socketRef.current?.disconnect();
      socketRef.current = null;
    };
    // boome-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const onDone = async () => {
    await saveStoredActiveTrip(null);
    reset();
    router.replace('/(auth)/(tabs)/home');
  };

  if (!destination) {
    return null;
  }

  return (
    <WizardShell
      title="Your trip"
      currentStep={4}
      footer={
        tripStatus === 'completed' ? (
          <Button
            className="w-full bg-goojol-coral data-[active=true]:bg-goojol-coral/90"
            onPress={onDone}
          >
            <ButtonText className="font-semibold text-white">Done</ButtonText>
          </Button>
        ) : undefined
      }
    >
      <View className="flex-1 px-6 py-4">
        {isLoading ? (
          <View className="flex-1 items-center justify-center">
            <ActivityIndicator color="#ff6b4a" size="large" />
          </View>
        ) : (
          <VStack space="lg" className="flex-1">
            <VStack space="xs">
              <Heading size="lg" className="text-white">
                {statusLabel[tripStatus] || statusLabel.default}
              </Heading>
              {error ? <Text className="text-destructive text-sm">{error}</Text> : null}
            </VStack>

            <TripMap
              pickup={pickup}
              destination={destination}
              driverLocation={driverLocation}
              customerLocation={customerLocation}
            />

            {matchedDriver ? (
              <View className="rounded-2xl border border-goojol-border bg-goojol-surface p-4">
                <VStack space="sm">
                  <Text className="font-semibold text-lg text-white">{matchedDriver.name}</Text>
                  <Text className="text-goojol-muted">
                    {matchedDriver.vehicle_name} · {matchedDriver.license_number}
                  </Text>
                  {selectedFare != null ? (
                    <Text className="text-goojol-teal">{formatRupiah(selectedFare)}</Text>
                  ) : null}
                </VStack>
              </View>
            ) : null}
          </VStack>
        )}
      </View>
    </WizardShell>
  );
}
