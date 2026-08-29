import { z } from 'zod';
import { axiosClient } from '@/lib/api/axios-client';
import { parsedApiResponse } from '@/lib/utils/parses-api-response';

const latLongPairSchema = z.tuple([z.number(), z.number()]);

export const activeTransactionSchema = z.object({
  id: z.string(),
  status: z.enum(['accepted_offer', 'on_the_way', 'completed', 'cancelled']),
  pickup_lat_long: latLongPairSchema,
  destination_lat_long: latLongPairSchema,
  driver_last_lat_long: latLongPairSchema,
  customer_last_lat_long: latLongPairSchema.nullish(),
  distance: z.number(),
  total_fare: z.number(),
  paid_at: z.string().optional().nullable(),
});

export type ActiveTransaction = z.infer<typeof activeTransactionSchema>;

const startTripResponseSchema = z.object({
  transaction_id: z.string(),
  status: z.string(),
});

const completeTripResponseSchema = z.object({
  transaction_id: z.string(),
  status: z.string(),
  total_fare: z.number(),
  paid_at: z.string(),
});

export async function fetchActiveTransaction() {
  const response = await axiosClient.get('/api/trip/transactions/active');
  return parsedApiResponse(activeTransactionSchema, response).data;
}

export async function startTrip(transactionId: string) {
  const response = await axiosClient.post(`/api/trip/transactions/${transactionId}/start`);
  return parsedApiResponse(startTripResponseSchema, response).data;
}

export async function completeTrip(transactionId: string) {
  const response = await axiosClient.post(`/api/trip/transactions/${transactionId}/complete`);
  return parsedApiResponse(completeTripResponseSchema, response).data;
}

export async function cancelTrip(transactionId: string) {
  const response = await axiosClient.post(`/api/trip/transactions/${transactionId}/cancel`);
  return parsedApiResponse(z.object({ transaction_id: z.string(), status: z.string() }), response).data;
}
