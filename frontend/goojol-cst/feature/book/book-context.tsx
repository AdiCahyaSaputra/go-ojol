import { createContext, type ReactNode, useCallback, useContext, useMemo, useState } from 'react';
import { DEFAULT_PICKUP } from '@/constants/book';
import type { BookLocation, CalculateArgoResponse, MatchedDriver } from './dispatch.schema';

type VehicleType = 'car' | 'motorcycle';

export type BookContextValue = {
  pickup: BookLocation;
  destination: BookLocation | null;
  vehicleType: VehicleType;
  vehicleMaxSize: number;
  quote: CalculateArgoResponse | null;
  matchedDriver: MatchedDriver | null;
  transactionId: string | null;
  setPickup: (location: BookLocation) => void;
  setDestination: (location: BookLocation) => void;
  setVehicleOption: (option: { vehicleType: VehicleType; maxSize: number }) => void;
  setQuote: (quote: CalculateArgoResponse | null) => void;
  setMatchedDriver: (driver: MatchedDriver | null) => void;
  setTransactionId: (transactionId: string | null) => void;
  reset: () => void;
};

const BookContext = createContext<BookContextValue | null>(null);

export function BookProvider({ children }: { children: ReactNode }) {
  const [pickup, setPickupState] = useState<BookLocation>(DEFAULT_PICKUP);
  const [destination, setDestinationState] = useState<BookLocation | null>(null);
  const [vehicleType, setVehicleType] = useState<VehicleType>('motorcycle');
  const [vehicleMaxSize, setVehicleMaxSize] = useState(1);
  const [quote, setQuoteState] = useState<CalculateArgoResponse | null>(null);
  const [matchedDriver, setMatchedDriverState] = useState<MatchedDriver | null>(null);
  const [transactionId, setTransactionIdState] = useState<string | null>(null);

  const setPickup = useCallback((location: BookLocation) => {
    setPickupState(location);
  }, []);

  const setDestination = useCallback((location: BookLocation) => {
    setDestinationState(location);
  }, []);

  const setVehicleOption = useCallback((option: { vehicleType: VehicleType; maxSize: number }) => {
    setVehicleType(option.vehicleType);
    setVehicleMaxSize(option.maxSize);
  }, []);

  const reset = useCallback(() => {
    setPickupState(DEFAULT_PICKUP);
    setDestinationState(null);
    setVehicleType('motorcycle');
    setVehicleMaxSize(1);
    setQuoteState(null);
    setMatchedDriverState(null);
    setTransactionIdState(null);
  }, []);

  const value = useMemo<BookContextValue>(
    () => ({
      pickup,
      destination,
      vehicleType,
      vehicleMaxSize,
      quote,
      matchedDriver,
      transactionId,
      setPickup,
      setDestination,
      setVehicleOption,
      setQuote: setQuoteState,
      setMatchedDriver: setMatchedDriverState,
      setTransactionId: setTransactionIdState,
      reset,
    }),
    [
      pickup,
      destination,
      vehicleType,
      vehicleMaxSize,
      quote,
      matchedDriver,
      transactionId,
      setPickup,
      setDestination,
      setVehicleOption,
      reset,
    ],
  );

  return <BookContext.Provider value={value}>{children}</BookContext.Provider>;
}

export function useBook() {
  const context = useContext(BookContext);
  if (!context) {
    throw new Error('useBook must be used within BookProvider');
  }
  return context;
}

export function useOptionalBook() {
  return useContext(BookContext);
}
