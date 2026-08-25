import { z } from 'zod';
import { axiosClient } from '@/lib/api/axios-client';
import { parsedApiResponse } from '@/lib/utils/parses-api-response';

export type DriverMode = 'online' | 'offline';

type SetDriverModeInput = {
  mode: DriverMode;
  lat: number;
  lng: number;
};

export type OfferAction = 'accept' | 'reject';

type RespondOfferInput = {
  transactionId: string;
  action: OfferAction;
};

const respondOfferResponseSchema = z.object({
  transaction_id: z.string(),
  status: z.string(),
});

export async function setDriverMode(input: SetDriverModeInput) {
  const response = await axiosClient.post('/api/trip/dispatch/driver/mode', {
    mode: input.mode,
    current_lat_long: [String(input.lat), String(input.lng)],
  });
  return parsedApiResponse(z.unknown(), response).data;
}

export async function respondOffer(input: RespondOfferInput) {
  const response = await axiosClient.post(
    `/api/trip/dispatch/driver/offers/${input.transactionId}/respond`,
    { action: input.action },
  );
  return parsedApiResponse(respondOfferResponseSchema, response).data;
}
