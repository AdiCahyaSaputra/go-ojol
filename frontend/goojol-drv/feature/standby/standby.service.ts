import { z } from 'zod';
import { axiosClient } from '@/lib/api/axios-client';
import { parsedApiResponse } from '@/lib/utils/parses-api-response';

export type DriverMode = 'online' | 'offline';

type SetDriverModeInput = {
  mode: DriverMode;
  lat: number;
  lng: number;
};

export async function setDriverMode(input: SetDriverModeInput) {
  const response = await axiosClient.post('/api/trip/dispatch/driver/mode', {
    mode: input.mode,
    current_lat_long: [String(input.lat), String(input.lng)],
  });
  return parsedApiResponse(z.unknown(), response).data;
}
