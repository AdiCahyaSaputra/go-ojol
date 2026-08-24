import { z } from 'zod';

const customerProfileSchema = z.object({
  id: z.string(),
  name: z.string(),
  phone_number: z.string(),
  profile_picture_url: z.string().nullable().optional(),
});

export const userMeSchema = z.object({
  id: z.string(),
  email: z.string(),
  role: z.string().optional(),
  customer: customerProfileSchema.nullable().optional(),
});

export type UserMe = z.infer<typeof userMeSchema>;
