import { useQuery } from '@tanstack/react-query';
import { QUERY_KEYS } from '@/constants/query-keys';
import type { BookLocation } from './dispatch.schema';
import { calculateArgo } from './dispatch.service';

export function useCalculateArgoQuery(input: {
  pickup: BookLocation;
  destination: BookLocation | null;
  vehicleType: 'car' | 'motorcycle';
}) {
  const destination = input.destination;

  return useQuery({
    queryKey: QUERY_KEYS.calculateArgo({
      pickupLat: input.pickup.lat,
      pickupLng: input.pickup.lng,
      destinationLat: destination?.lat,
      destinationLng: destination?.lng,
      vehicleType: input.vehicleType,
    }),
    queryFn: () => {
      if (!destination) {
        throw new Error('Destination is required to calculate fare.');
      }

      return calculateArgo({
        pickup: input.pickup,
        destination,
        vehicleType: input.vehicleType,
      });
    },
    enabled: destination != null,
    staleTime: Infinity,
  });
}
