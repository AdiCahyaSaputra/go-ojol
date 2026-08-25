import { useMutation } from '@tanstack/react-query';
import type { BookLocation } from './dispatch.schema';
import { findDriver } from './dispatch.service';

export function useFindDriverMutation() {
  return useMutation({
    mutationFn: (input: {
      pickup: BookLocation;
      destination: BookLocation;
      vehicleType: 'car' | 'motorcycle';
      maxSize: number;
    }) => findDriver(input),
  });
}
