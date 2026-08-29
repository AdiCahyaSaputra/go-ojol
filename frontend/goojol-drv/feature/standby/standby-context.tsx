import {
  createContext,
  type ReactNode,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useRef,
  useState,
} from 'react';
import { LOCATION_HEARTBEAT_MS } from '@/constants/standby';
import { getSession } from '@/lib/auth/token-storage';
import { type DriverCoords, resolveDriverLocation } from './location';
import { respondOffer, setDriverMode } from './standby.service';
import { StandbySocket, type TripOffer } from './standby-ws';
import {
  completeTrip as completeTripRequest,
  fetchActiveTransaction,
  startTrip as startTripRequest,
} from './trip.service';
import { set } from 'zod';

export type StandbyPhase = 'offline' | 'online' | 'offer' | 'accepted' | 'in_trip';

type StandbyContextValue = {
  phase: StandbyPhase;
  coords: DriverCoords | null;
  offer: TripOffer | null;
  customerLocation: { lat: number; lng: number } | null;
  offerSecondsLeft: number;
  isBusy: boolean;
  error: string | null;
  goOnline: () => Promise<void>;
  goOffline: () => Promise<void>;
  acceptOffer: () => Promise<void>;
  rejectOffer: () => Promise<void>;
  startTrip: () => Promise<void>;
  completeTrip: () => Promise<void>;
  clearError: () => void;
};

const StandbyContext = createContext<StandbyContextValue | null>(null);

function formatCoordPair(coords: [number, number]) {
  return `${coords[0].toFixed(5)}, ${coords[1].toFixed(5)}`;
}

export function formatOfferCoord(coords: [number, number]) {
  return formatCoordPair(coords);
}

function isInTripPhase(phase: StandbyPhase) {
  return phase === 'accepted' || phase === 'in_trip';
}

export function StandbyProvider({ children }: { children: ReactNode }) {
  const [phase, setPhase] = useState<StandbyPhase>('offline');
  const [coords, setCoords] = useState<DriverCoords | null>(null);
  const [offer, setOffer] = useState<TripOffer | null>(null);
  const [customerLocation, setCustomerLocation] = useState<{ lat: number; lng: number } | null>(
    null,
  );
  const [rehydrateCancelled, setRehydrateCancelled] = useState(false);
  const [offerSecondsLeft, setOfferSecondsLeft] = useState(0);
  const [isBusy, setIsBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const socketRef = useRef<StandbySocket | null>(null);
  const heartbeatRef = useRef<ReturnType<typeof setInterval> | null>(null);
  const countdownRef = useRef<ReturnType<typeof setInterval> | null>(null);
  const phaseRef = useRef(phase);
  const offerRef = useRef<TripOffer | null>(null);
  const rehydratedRef = useRef(false);

  useEffect(() => {
    phaseRef.current = phase;
  }, [phase]);

  useEffect(() => {
    offerRef.current = offer;
  }, [offer]);

  const clearHeartbeat = useCallback(() => {
    if (heartbeatRef.current) {
      clearInterval(heartbeatRef.current);
      heartbeatRef.current = null;
    }
  }, []);

  const clearCountdown = useCallback(() => {
    if (countdownRef.current) {
      clearInterval(countdownRef.current);
      countdownRef.current = null;
    }
    setOfferSecondsLeft(0);
  }, []);

  const clearActiveTrip = useCallback(() => {
    clearCountdown();
    setOffer(null);
    setCustomerLocation(null);
  }, [clearCountdown]);

  const startOfferCountdown = useCallback(
    (expiresInSec: number) => {
      clearCountdown();
      setOfferSecondsLeft(expiresInSec);
      countdownRef.current = setInterval(() => {
        setOfferSecondsLeft((prev) => {
          if (prev <= 1) {
            clearCountdown();
            if (phaseRef.current === 'offer') {
              setOffer(null);
              setPhase('online');
            }
            return 0;
          }
          return prev - 1;
        });
      }, 1000);
    },
    [clearCountdown],
  );

  const disconnectSocket = useCallback(() => {
    clearHeartbeat();
    socketRef.current?.disconnect();
    socketRef.current = null;
  }, [clearHeartbeat]);

  const startHeartbeat = useCallback(
    (socket: StandbySocket) => {
      clearHeartbeat();
      heartbeatRef.current = setInterval(async () => {
        const currentPhase = phaseRef.current;
        if (currentPhase === 'offline') {
          return;
        }
        const next = await resolveDriverLocation();
        setCoords(next);
        if (isInTripPhase(currentPhase)) {
          const transactionId = offerRef.current?.transactionId;
          if (transactionId) {
            socket.sendTripLocation(transactionId, { lat: next.lat, lng: next.lng });
          }
          return;
        }
        socket.sendLocation({ lat: next.lat, lng: next.lng });
      }, LOCATION_HEARTBEAT_MS);
    },
    [clearHeartbeat],
  );

  const connectSocket = useCallback(
    async (accessToken: string, location: DriverCoords, waitForStandby: boolean) => {
      const socket = new StandbySocket();
      socketRef.current = socket;

      if (waitForStandby) {
        await new Promise<void>((resolve, reject) => {
          let settled = false;
          const fail = (message: string) => {
            if (settled) {
              return;
            }
            settled = true;
            reject(new Error(message));
          };
          const succeed = () => {
            if (settled) {
              return;
            }
            settled = true;
            resolve();
          };

          socket.connect(accessToken, {
            onStandbyOk: succeed,
            onTripOffer: (nextOffer) => {
              if (phaseRef.current !== 'online') {
                return;
              }
              setOffer(nextOffer);
              setPhase('offer');
              startOfferCountdown(nextOffer.expiresInSec);
            },
            onOfferTaken: (transactionId) => {
              if (offerRef.current?.transactionId !== transactionId) {
                return;
              }
              if (phaseRef.current === 'offer') {
                clearActiveTrip();
                setPhase('online');
              }
            },
            onOfferExpired: (transactionId) => {
              if (offerRef.current?.transactionId !== transactionId) {
                return;
              }
              if (phaseRef.current === 'offer') {
                clearActiveTrip();
                setPhase('online');
              }
            },
            onCustomerLocation: (event) => {
              if (offerRef.current?.transactionId !== event.transactionId) {
                return;
              }
              setCustomerLocation({ lat: event.lat, lng: event.lng });
            },
            onTripStatus: (event) => {
              if (offerRef.current?.transactionId !== event.transactionId) {
                return;
              }
              if (event.status === 'on_the_way') {
                setPhase('in_trip');
              }
              if (event.status === 'cancelled') {
                clearActiveTrip();
                setPhase('online');
              }
            },
            onTripCompleted: () => {
              clearActiveTrip();
              setPhase('online');
            },
            onError: (message) => {
              if (!settled) {
                fail(message);
                return;
              }
              setError(message);
            },
            onClose: () => {
              if (phaseRef.current === 'offline') {
                return;
              }
              if (isInTripPhase(phaseRef.current)) {
                clearHeartbeat();
                setError('Connection lost. Reconnect to keep sharing location.');
                return;
              }
              clearActiveTrip();
              setPhase('offline');
              clearHeartbeat();
              setError('Connection lost. You are offline.');
            },
          });

          socket.whenOpen(() => {
            const sent = socket.sendStandby({ lat: location.lat, lng: location.lng });
            if (!sent) {
              fail('Could not send standby');
            }
          });

          setTimeout(() => fail('Standby timed out'), 12_000);
        });
      } else {
        await new Promise<void>((resolve, reject) => {
          socket.connect(accessToken, {
            onCustomerLocation: (event) => {
              if (offerRef.current?.transactionId !== event.transactionId) {
                return;
              }
              setCustomerLocation({ lat: event.lat, lng: event.lng });
            },
            onTripStatus: (event) => {
              if (offerRef.current?.transactionId !== event.transactionId) {
                return;
              }
              if (event.status === 'on_the_way') {
                setPhase('in_trip');
              }
              if (event.status === 'cancelled') {
                clearActiveTrip();
                setPhase('online');
              }
            },
            onTripCompleted: () => {
              clearActiveTrip();
              setPhase('online');
            },
            onError: (message) => setError(message),
            onClose: () => {
              if (isInTripPhase(phaseRef.current)) {
                clearHeartbeat();
                setError('Connection lost. Reconnect to keep sharing location.');
              }
            },
          });
          socket.whenOpen(() => resolve());
          setTimeout(() => reject(new Error('WebSocket timed out')), 12_000);
        });
      }

      startHeartbeat(socket);
      return socket;
    },
    [clearActiveTrip, clearHeartbeat, startHeartbeat, startOfferCountdown],
  );

  const goOffline = useCallback(async () => {
    setIsBusy(true);
    setError(null);
    clearActiveTrip();

    const current = coords ?? (await resolveDriverLocation());
    disconnectSocket();

    try {
      await setDriverMode({ mode: 'offline', lat: current.lat, lng: current.lng });
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Could not go offline');
    } finally {
      setPhase('offline');
      setIsBusy(false);
    }
  }, [clearActiveTrip, coords, disconnectSocket]);

  const goOnline = useCallback(async () => {
    setIsBusy(true);
    setError(null);

    try {
      const location = await resolveDriverLocation();
      setCoords(location);

      await setDriverMode({ mode: 'online', lat: location.lat, lng: location.lng });

      const session = await getSession();
      if (!session?.accessToken) {
        throw new Error('Missing session. Sign in again.');
      }

      await connectSocket(session.accessToken, location, true);
      setPhase('online');
    } catch (err) {
      disconnectSocket();
      try {
        const location = coords ?? (await resolveDriverLocation());
        await setDriverMode({ mode: 'offline', lat: location.lat, lng: location.lng });
      } catch {
        // Best-effort rollback.
      }
      setPhase('offline');
      setError(err instanceof Error ? err.message : 'Could not go online');
    } finally {
      setIsBusy(false);
    }
  }, [connectSocket, coords, disconnectSocket]);

  const acceptOffer = useCallback(async () => {
    const current = offerRef.current;
    if (!current || phaseRef.current !== 'offer') {
      return;
    }

    setIsBusy(true);
    setError(null);
    try {
      await respondOffer({ transactionId: current.transactionId, action: 'accept' });
      clearCountdown();
      setPhase('accepted');
    } catch (err) {
      clearActiveTrip();
      setPhase('online');
      setError(err instanceof Error ? err.message : 'Could not accept offer');
    } finally {
      setIsBusy(false);
    }
  }, [clearActiveTrip, clearCountdown]);

  const rejectOffer = useCallback(async () => {
    const current = offerRef.current;
    if (!current || phaseRef.current !== 'offer') {
      return;
    }

    setIsBusy(true);
    setError(null);
    try {
      await respondOffer({ transactionId: current.transactionId, action: 'reject' });
      clearActiveTrip();
      setPhase('online');
    } catch (err) {
      clearActiveTrip();
      setPhase('online');
      setError(err instanceof Error ? err.message : 'Could not reject offer');
    } finally {
      setIsBusy(false);
    }
  }, [clearActiveTrip]);

  const startTrip = useCallback(async () => {
    const current = offerRef.current;
    if (!current || phaseRef.current !== 'accepted') {
      return;
    }

    setIsBusy(true);
    setError(null);
    try {
      await startTripRequest(current.transactionId);
      setPhase('in_trip');
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Could not start trip');
    } finally {
      setIsBusy(false);
    }
  }, []);

  const completeTrip = useCallback(async () => {
    const current = offerRef.current;
    if (!current || phaseRef.current !== 'in_trip') {
      return;
    }

    setIsBusy(true);
    setError(null);
    try {
      await completeTripRequest(current.transactionId);
      clearActiveTrip();
      setPhase('online');
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Could not complete trip');
    } finally {
      setIsBusy(false);
    }
  }, [clearActiveTrip]);

  const rehydrate = async () => {
    try {
      const txn = await fetchActiveTransaction();
      if (rehydrateCancelled) {
        return;
      }

      const restoredOffer: TripOffer = {
        transactionId: txn.id,
        customerName: 'Customer',
        pickup: txn.pickup_lat_long,
        destination: txn.destination_lat_long,
        distanceM: txn.distance,
        totalFare: txn.total_fare,
        expiresInSec: 0,
      };

      if (txn.customer_last_lat_long) {
        setCustomerLocation({
          lat: txn.customer_last_lat_long[0],
          lng: txn.customer_last_lat_long[1],
        });
      }

      setOffer(restoredOffer);
      setPhase(txn.status === 'on_the_way' ? 'in_trip' : 'accepted');

      const location = await resolveDriverLocation();
      setCoords(location);

      const session = await getSession();
      if (!session?.accessToken) {
        return;
      }

      await connectSocket(session.accessToken, location, false);
    } catch {
      // No active trip or network error — stay offline until user goes online.
    }
  };

  useEffect(() => {
    if (rehydratedRef.current) {
      return;
    }
    rehydratedRef.current = true;

    setRehydrateCancelled(false);

    void rehydrate();

    return () => {
      setRehydrateCancelled(true);
    };
  }, [connectSocket]);

  useEffect(() => {
    return () => {
      clearCountdown();
      disconnectSocket();
    };
  }, [clearCountdown, disconnectSocket]);

  const value = useMemo<StandbyContextValue>(
    () => ({
      phase,
      coords,
      offer,
      customerLocation,
      offerSecondsLeft,
      isBusy,
      error,
      goOnline,
      goOffline,
      acceptOffer,
      rejectOffer,
      startTrip,
      completeTrip,
      clearError: () => setError(null),
    }),
    [
      phase,
      coords,
      offer,
      customerLocation,
      offerSecondsLeft,
      isBusy,
      error,
      goOnline,
      goOffline,
      acceptOffer,
      rejectOffer,
      startTrip,
      completeTrip,
    ],
  );

  return <StandbyContext.Provider value={value}>{children}</StandbyContext.Provider>;
}

export function useStandby() {
  const ctx = useContext(StandbyContext);
  if (!ctx) {
    throw new Error('useStandby must be used within StandbyProvider');
  }
  return ctx;
}
