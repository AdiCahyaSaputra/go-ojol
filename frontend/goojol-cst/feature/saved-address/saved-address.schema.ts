import { z } from 'zod';
import type { BookLocation } from '@/feature/book/dispatch.schema';

export const savedAddressSchema = z.object({
  id: z.string(),
  name: z.string().min(1),
  lat_long: z.tuple([z.string(), z.string()]),
  is_default_pickup: z.boolean(),
  created_at: z.string(),
  updated_at: z.string(),
});

export type SavedAddress = z.infer<typeof savedAddressSchema>;

export const savedAddressListSchema = z.array(savedAddressSchema);

export const createSavedAddressRequestSchema = z.object({
  name: z.string().min(1).max(255),
  lat_long: z.tuple([z.string(), z.string()]),
  is_default_pickup: z.boolean(),
});

export type CreateSavedAddressRequest = z.infer<typeof createSavedAddressRequestSchema>;

export function savedAddressToBookLocation(address: SavedAddress): BookLocation {
  return {
    name: address.name,
    lat: address.lat_long[0],
    lng: address.lat_long[1],
  };
}
