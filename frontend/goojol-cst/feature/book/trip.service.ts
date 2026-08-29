import { z } from 'zod';
import { axiosClient } from '@/lib/api/axios-client';
import { parsedApiResponse } from '@/lib/utils/parses-api-response';
import { activeTransactionSchema } from './trip.schema';

export async function fetchActiveTransaction() {
  const response = await axiosClient.get('/api/trip/transactions/active');
  return parsedApiResponse(activeTransactionSchema, response).data;
}

export async function cancelActiveTrip(transactionId: string) {
  const response = await axiosClient.post(`/api/trip/transactions/${transactionId}/cancel`);
  return parsedApiResponse(
    z.object({
      transaction_id: z.string(),
      status: z.string(),
    }),
    response,
  ).data;
}
