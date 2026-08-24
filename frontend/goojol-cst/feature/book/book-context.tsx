import { createContext, type ReactNode, useCallback, useContext, useMemo, useState } from 'react';
import { DEFAULT_PICKUP } from '@/constants/book';
import type { BookLocation, CalculateArgoResponse, NearbyDriver } from './dispatch.schema';

type VehicleType = 'car' | 'motorcycle';

type BookContextValue = {
  pickup: BookLocation;
  destination: BookLocation | null;
  vehicleType: VehicleType;
  vehicleMaxSize: number;
  quote: CalculateArgoResponse | null;
  matchedDriver: NearbyDriver | null;
  setPickup: (location: BookLocation) => void;
  setDestination: (location: BookLocation) => void;
  setVehicleOption: (option: { vehicleType: VehicleType; maxSize: number }) => void;
  setQuote: (quote: CalculateArgoResponse | null) => void;
  setMatchedDriver: (driver: NearbyDriver | null) => void;
  reset: () => void;
};

const BookContext = createContext<BookContextValue | null>(null);

export function BookProvider({ children }: { children: ReactNode }) {
  const [pickup, setPickupState] = useState<BookLocation>(DEFAULT_PICKUP);
  const [destination, setDestinationState] = useState<BookLocation | null>(null);
  const [vehicleType, setVehicleType] = useState<VehicleType>('motorcycle');
  const [vehicleMaxSize, setVehicleMaxSize] = useState(1);
  const [quote, setQuoteState] = useState<CalculateArgoResponse | null>(null);
  const [matchedDriver, setMatchedDriverState] = useState<NearbyDriver | null>(null);

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
  }, []);

  const value = useMemo<BookContextValue>(
    () => ({
      pickup,
      destination,
      vehicleType,
      vehicleMaxSize,
      quote,
      matchedDriver,
      setPickup,
      setDestination,
      setVehicleOption,
      setQuote: setQuoteState,
      setMatchedDriver: setMatchedDriverState,
      reset,
    }),
    [
      pickup,
      destination,
      vehicleType,
      vehicleMaxSize,
      quote,
      matchedDriver,
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
