import { z } from 'zod';
import { matchedDriverSchema } from './dispatch.schema';

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
  driver: matchedDriverSchema.optional(),
});

export type ActiveTransaction = z.infer<typeof activeTransactionSchema>;

export const ACTIVE_TRIP_STORAGE_KEY = 'goojol-cst-active-trip';

export type StoredActiveTrip = {
  transactionId: string;
  pickup: { name: string; lat: string; lng: string };
  destination: { name: string; lat: string; lng: string };
  matchedDriver: z.infer<typeof matchedDriverSchema> | null;
  totalFare: number | null;
};
