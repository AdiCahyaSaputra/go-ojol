import { z } from 'zod';

const vehicleProfileSchema = z.object({
  id: z.string(),
  name: z.string(),
  license_number: z.string(),
  max_size: z.number(),
  type: z.string(),
});

const driverProfileSchema = z.object({
  id: z.string(),
  name: z.string(),
  phone_number: z.string(),
  address: z.string().optional(),
  profile_picture_url: z.string().nullable().optional(),
  vehicle: vehicleProfileSchema.nullable().optional(),
});

export const userMeSchema = z.object({
  id: z.string(),
  email: z.string(),
  role: z.string().optional(),
  driver: driverProfileSchema.nullable().optional(),
});

export type UserMe = z.infer<typeof userMeSchema>;
