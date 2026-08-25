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

export type StandbyPhase = 'offline' | 'online' | 'offer' | 'accepted';

type StandbyContextValue = {
  phase: StandbyPhase;
  coords: DriverCoords | null;
  offer: TripOffer | null;
  offerSecondsLeft: number;
  isBusy: boolean;
  error: string | null;
  goOnline: () => Promise<void>;
  goOffline: () => Promise<void>;
  acceptOffer: () => Promise<void>;
  rejectOffer: () => Promise<void>;
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

export function StandbyProvider({ children }: { children: ReactNode }) {
  const [phase, setPhase] = useState<StandbyPhase>('offline');
  const [coords, setCoords] = useState<DriverCoords | null>(null);
  const [offer, setOffer] = useState<TripOffer | null>(null);
  const [offerSecondsLeft, setOfferSecondsLeft] = useState(0);
  const [isBusy, setIsBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const socketRef = useRef<StandbySocket | null>(null);
  const heartbeatRef = useRef<ReturnType<typeof setInterval> | null>(null);
  const countdownRef = useRef<ReturnType<typeof setInterval> | null>(null);
  const phaseRef = useRef(phase);
  const offerRef = useRef<TripOffer | null>(null);

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

  const clearActiveOffer = useCallback(() => {
    clearCountdown();
    setOffer(null);
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
        if (phaseRef.current === 'offline') {
          return;
        }
        const next = await resolveDriverLocation();
        setCoords(next);
        socket.sendLocation({ lat: next.lat, lng: next.lng });
      }, LOCATION_HEARTBEAT_MS);
    },
    [clearHeartbeat],
  );

  const goOffline = useCallback(async () => {
    setIsBusy(true);
    setError(null);
    clearActiveOffer();

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
  }, [clearActiveOffer, coords, disconnectSocket]);

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

      const socket = new StandbySocket();
      socketRef.current = socket;

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

        socket.connect(session.accessToken, {
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
              clearActiveOffer();
              setPhase('online');
            }
          },
          onOfferExpired: (transactionId) => {
            if (offerRef.current?.transactionId !== transactionId) {
              return;
            }
            if (phaseRef.current === 'offer') {
              clearActiveOffer();
              setPhase('online');
            }
          },
          onError: (message) => {
            if (!settled) {
              fail(message);
              return;
            }
            setError(message);
          },
          onClose: () => {
            if (phaseRef.current !== 'offline') {
              clearActiveOffer();
              setPhase('offline');
              clearHeartbeat();
              setError('Connection lost. You are offline.');
            }
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

      startHeartbeat(socket);
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
  }, [
    clearActiveOffer,
    clearHeartbeat,
    coords,
    disconnectSocket,
    startHeartbeat,
    startOfferCountdown,
  ]);

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
      clearActiveOffer();
      setPhase('online');
      setError(err instanceof Error ? err.message : 'Could not accept offer');
    } finally {
      setIsBusy(false);
    }
  }, [clearActiveOffer, clearCountdown]);

  const rejectOffer = useCallback(async () => {
    const current = offerRef.current;
    if (!current || phaseRef.current !== 'offer') {
      return;
    }

    setIsBusy(true);
    setError(null);
    try {
      await respondOffer({ transactionId: current.transactionId, action: 'reject' });
      clearActiveOffer();
      setPhase('online');
    } catch (err) {
      clearActiveOffer();
      setPhase('online');
      setError(err instanceof Error ? err.message : 'Could not reject offer');
    } finally {
      setIsBusy(false);
    }
  }, [clearActiveOffer]);

  const completeTrip = useCallback(async () => {
    clearActiveOffer();
    setPhase('online');
  }, [clearActiveOffer]);

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
      offerSecondsLeft,
      isBusy,
      error,
      goOnline,
      goOffline,
      acceptOffer,
      rejectOffer,
      completeTrip,
      clearError: () => setError(null),
    }),
    [
      phase,
      coords,
      offer,
      offerSecondsLeft,
      isBusy,
      error,
      goOnline,
      goOffline,
      acceptOffer,
      rejectOffer,
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
