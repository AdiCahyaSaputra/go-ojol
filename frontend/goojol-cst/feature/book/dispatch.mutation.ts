import { useMutation } from '@tanstack/react-query';
import type { BookLocation } from './dispatch.schema';
import { calculateArgo, findDriver } from './dispatch.service';

export function useCalculateArgoMutation() {
  return useMutation({
    mutationFn: (input: { pickup: BookLocation; destination: BookLocation; vehicleId: string }) =>
      calculateArgo(input),
  });
}

export function useFindDriverMutation() {
  return useMutation({
    mutationFn: (input: { pickup: BookLocation; vehicleType: 'car' | 'motorcycle' }) =>
      findDriver(input),
  });
}
