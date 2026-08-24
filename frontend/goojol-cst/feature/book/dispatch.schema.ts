import { z } from 'zod';

export const bookLocationSchema = z.object({
  name: z.string(),
  lat: z.string(),
  lng: z.string(),
});

export type BookLocation = z.infer<typeof bookLocationSchema>;

export const vehicleOptionSchema = z.object({
  vehicle_type: z.enum(['car', 'motorcycle']),
  max_size: z.number(),
  total_fare: z.number(),
});

export type VehicleOption = z.infer<typeof vehicleOptionSchema>;

export const calculateArgoResponseSchema = z.object({
  distance: z.number(),
  duration: z.number(),
  path: z.array(z.tuple([z.number(), z.number()])),
  platform_percentage: z.number(),
  vehicle_options: z.array(vehicleOptionSchema),
});

export type CalculateArgoResponse = z.infer<typeof calculateArgoResponseSchema>;

const nearbyDriverProfileSchema = z.object({
  user_id: z.string(),
  driver_id: z.string(),
  name: z.string(),
  phone_number: z.string(),
  vehicle_name: z.string(),
  license_number: z.string(),
  type: z.enum(['car', 'motorcycle']),
});

export const nearbyDriverSchema = z.object({
  user_id: z.string(),
  distance_m: z.number(),
  location: z.tuple([z.number(), z.number()]),
  profile: nearbyDriverProfileSchema,
});

export type NearbyDriver = z.infer<typeof nearbyDriverSchema>;

export const findDriverResponseSchema = z.object({
  drivers: z.array(nearbyDriverSchema),
});

export type FindDriverResponse = z.infer<typeof findDriverResponseSchema>;
