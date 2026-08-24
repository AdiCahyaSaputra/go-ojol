import { z } from 'zod';

export const savedAddressSchema = z.object({
  id: z.string(),
  name: z.string().min(1),
  lat: z.string(),
  lng: z.string(),
});

export type SavedAddress = z.infer<typeof savedAddressSchema>;

export const savedAddressListSchema = z.array(savedAddressSchema);
