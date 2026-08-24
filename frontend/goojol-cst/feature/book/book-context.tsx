import { createContext, type ReactNode, useCallback, useContext, useMemo, useState } from 'react';
import { DEFAULT_PICKUP, VEHICLE_OPTIONS } from '@/constants/book';
import type { BookLocation, CalculateArgoResponse, NearbyDriver } from './dispatch.schema';

type BookContextValue = {
  pickup: BookLocation;
  destination: BookLocation | null;
  vehicleId: string;
  vehicleType: 'car' | 'motorcycle';
  quote: CalculateArgoResponse | null;
  matchedDriver: NearbyDriver | null;
  setPickup: (location: BookLocation) => void;
  setDestination: (location: BookLocation) => void;
  setVehicleId: (vehicleId: string) => void;
  setQuote: (quote: CalculateArgoResponse | null) => void;
  setMatchedDriver: (driver: NearbyDriver | null) => void;
  reset: () => void;
};

const defaultVehicle = VEHICLE_OPTIONS[0];

const BookContext = createContext<BookContextValue | null>(null);

export function BookProvider({ children }: { children: ReactNode }) {
  const [pickup, setPickupState] = useState<BookLocation>(DEFAULT_PICKUP);
  const [destination, setDestinationState] = useState<BookLocation | null>(null);
  const [vehicleId, setVehicleIdState] = useState<string>(defaultVehicle.id);
  const [quote, setQuoteState] = useState<CalculateArgoResponse | null>(null);
  const [matchedDriver, setMatchedDriverState] = useState<NearbyDriver | null>(null);

  const setPickup = useCallback((location: BookLocation) => {
    setPickupState(location);
  }, []);

  const setDestination = useCallback((location: BookLocation) => {
    setDestinationState(location);
  }, []);

  const setVehicleId = useCallback((id: string) => {
    setVehicleIdState(id);
    const option = VEHICLE_OPTIONS.find((item) => item.id === id);
    if (option) {
      // vehicleType is derived when vehicleId changes
    }
  }, []);

  const vehicleType = useMemo(() => {
    return VEHICLE_OPTIONS.find((item) => item.id === vehicleId)?.type ?? 'motorcycle';
  }, [vehicleId]);

  const reset = useCallback(() => {
    setPickupState(DEFAULT_PICKUP);
    setDestinationState(null);
    setVehicleIdState(defaultVehicle.id);
    setQuoteState(null);
    setMatchedDriverState(null);
  }, []);

  const value = useMemo<BookContextValue>(
    () => ({
      pickup,
      destination,
      vehicleId,
      vehicleType,
      quote,
      matchedDriver,
      setPickup,
      setDestination,
      setVehicleId,
      setQuote: setQuoteState,
      setMatchedDriver: setMatchedDriverState,
      reset,
    }),
    [
      pickup,
      destination,
      vehicleId,
      vehicleType,
      quote,
      matchedDriver,
      setPickup,
      setDestination,
      setVehicleId,
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
