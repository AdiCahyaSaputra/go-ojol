import { useRouter } from 'expo-router';
import { useEffect, useRef, useState } from 'react';
import { ActivityIndicator, View } from 'react-native';
import { Button, ButtonSpinner, ButtonText } from '@/components/ui/button';
import { Text } from '@/components/ui/text';
import { VStack } from '@/components/ui/vstack';
import { useBook } from '@/feature/book/book-context';
import { BookError, WizardShell } from '@/feature/book/components/wizard-shell';
import { useFindDriverMutation } from '@/feature/book/dispatch.mutation';
import { CustomerDispatchSocket } from '@/feature/book/dispatch-ws';
import { getSession } from '@/lib/auth/token-storage';

type FindPhase =
  | 'connecting'
  | 'searching'
  | 'waiting'
  | 'no_drivers'
  | 'expired'
  | 'rejected'
  | 'error';

export default function BookFindDriverScreen() {
  const router = useRouter();
  const {
    pickup,
    destination,
    vehicleType,
    vehicleMaxSize,
    setMatchedDriver,
    setTransactionId,
  } = useBook();
  const mutation = useFindDriverMutation();
  const [phase, setPhase] = useState<FindPhase>('connecting');
  const [errorMessage, setErrorMessage] = useState<string | null>(null);
  const [secondsLeft, setSecondsLeft] = useState(0);
  const [isRetrying, setIsRetrying] = useState(false);

  const socketRef = useRef<CustomerDispatchSocket | null>(null);
  const countdownRef = useRef<ReturnType<typeof setInterval> | null>(null);
  const startedRef = useRef(false);

  const clearCountdown = () => {
    if (countdownRef.current) {
      clearInterval(countdownRef.current);
      countdownRef.current = null;
    }
    setSecondsLeft(0);
  };

  const startCountdown = (expiresInSec: number) => {
    clearCountdown();
    setSecondsLeft(expiresInSec);
    countdownRef.current = setInterval(() => {
      setSecondsLeft((prev) => {
        if (prev <= 1) {
          if (countdownRef.current) {
            clearInterval(countdownRef.current);
            countdownRef.current = null;
          }
          return 0;
        }
        return prev - 1;
      });
    }, 1000);
  };

  useEffect(() => {
    if (!destination) {
      router.replace('/book/location');
      return;
    }

    if (startedRef.current) {
      return;
    }
    startedRef.current = true;

    let cancelled = false;

    const start = async () => {
      setPhase('connecting');
      setErrorMessage(null);

      const session = await getSession();
      if (!session?.accessToken) {
        setPhase('error');
        setErrorMessage('Missing session. Sign in again.');
        return;
      }

      const socket = new CustomerDispatchSocket();
      socketRef.current = socket;

      const requestFindDriver = () => {
        setPhase('searching');
        mutation.mutate(
          {
            pickup,
            destination,
            vehicleType,
            maxSize: vehicleMaxSize,
          },
          {
            onSuccess: (data) => {
              if (cancelled) {
                return;
              }
              if (data.transaction_id) {
                setTransactionId(data.transaction_id);
              }
              if (data.drivers.length === 0) {
                setPhase('no_drivers');
              }
            },
            onError: (err) => {
              if (cancelled) {
                return;
              }
              setPhase('error');
              setErrorMessage(err.message ?? 'Could not find a driver. Try again later.');
            },
          },
        );
      };

      socket.connect(session.accessToken, {
        onWaiting: (event) => {
          if (cancelled) {
            return;
          }
          setIsRetrying(false);
          setTransactionId(event.transactionId || null);
          setPhase('waiting');
          startCountdown(event.expiresInSec);
        },
        onDriverMatched: (event) => {
          if (cancelled) {
            return;
          }
          clearCountdown();
          setMatchedDriver(event.matchedDriver);
          setTransactionId(event.transactionId || null);
          socket.disconnect();
          router.replace('/book/trip');
        },
        onOfferExpired: () => {
          if (cancelled) {
            return;
          }
          clearCountdown();
          setIsRetrying(false);
          setPhase('expired');
        },
        onOfferRejected: () => {
          if (cancelled) {
            return;
          }
          clearCountdown();
          setIsRetrying(false);
          setPhase('rejected');
        },
        onNoDrivers: () => {
          if (cancelled) {
            return;
          }
          clearCountdown();
          setIsRetrying(false);
          setPhase('no_drivers');
        },
        onError: (message) => {
          if (cancelled) {
            return;
          }
          setIsRetrying(false);
          setPhase('error');
          setErrorMessage(message);
        },
        onClose: () => {
          if (cancelled) {
            return;
          }
          setPhase((current) => {
            if (
              current === 'waiting' ||
              current === 'searching' ||
              current === 'connecting'
            ) {
              setErrorMessage('Connection lost. Try again.');
              return 'error';
            }
            return current;
          });
        },
      });

      socket.whenOpen(() => {
        if (cancelled) {
          return;
        }
        requestFindDriver();
      });
    };

    void start();

    return () => {
      cancelled = true;
      clearCountdown();
      socketRef.current?.disconnect();
      socketRef.current = null;
    };

    // Intentionally run once on mount for this booking attempt.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const onRetry = () => {
    const socket = socketRef.current;
    if (!socket?.isConnected) {
      setPhase('error');
      setErrorMessage('Connection lost. Go back and try again.');
      return;
    }
    setIsRetrying(true);
    setErrorMessage(null);
    setPhase('searching');
    const sent = socket.sendRetry();
    if (!sent) {
      setIsRetrying(false);
      setPhase('error');
      setErrorMessage('Could not retry. Go back and try again.');
    }
  };

  const showRetry = phase === 'expired' || phase === 'rejected';
  const showBack =
    phase === 'no_drivers' || phase === 'error' || phase === 'expired' || phase === 'rejected';

  return (
    <WizardShell
      title="Finding driver"
      currentStep={4}
      footer={
        showBack ? (
          <VStack space="sm" className="w-full">
            {showRetry ? (
              <Button
                className="w-full bg-goojol-coral data-[active=true]:bg-goojol-coral/90"
                onPress={onRetry}
                isDisabled={isRetrying}
              >
                {isRetrying ? <ButtonSpinner /> : null}
                <ButtonText className="font-semibold text-white">Retry</ButtonText>
              </Button>
            ) : null}
            <Button
              className="w-full border border-goojol-border bg-goojol-surface data-[active=true]:bg-goojol-surface/80"
              onPress={() => router.back()}
            >
              <ButtonText className="font-semibold text-white">Go back</ButtonText>
            </Button>
          </VStack>
        ) : undefined
      }
    >
      <View className="flex-1 items-center justify-center px-6">
        {phase === 'connecting' || phase === 'searching' || phase === 'waiting' ? (
          <VStack space="lg" className="items-center">
            <ActivityIndicator color="#ff6b4a" size="large" />
            <Text className="font-medium text-lg text-white">
              {phase === 'waiting' ? 'Waiting for a driver…' : 'Looking for nearby drivers…'}
            </Text>
            <Text className="text-center text-goojol-muted">
              {phase === 'waiting'
                ? secondsLeft > 0
                  ? `Offer expires in ${secondsLeft}s`
                  : 'Drivers have been notified. Hang tight.'
                : 'Searching within 3 km of your pickup point.'}
            </Text>
          </VStack>
        ) : null}

        {phase === 'no_drivers' ? (
          <BookError message="No drivers nearby. Ask a driver to go online in the demo area, then try again." />
        ) : null}

        {phase === 'expired' ? (
          <BookError message="No driver accepted in time. Retry to search again." />
        ) : null}

        {phase === 'rejected' ? (
          <BookError message="Drivers declined this trip. Retry to search again." />
        ) : null}

        {phase === 'error' ? (
          <BookError message={errorMessage ?? 'Could not find a driver. Try again later.'} />
        ) : null}
      </View>
    </WizardShell>
  );
}
