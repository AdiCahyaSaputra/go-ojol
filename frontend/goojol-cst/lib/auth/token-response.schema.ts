import { z } from 'zod';

export const tokenResponseSchema = z.object({
  access_token: z.string(),
  refresh_token: z.string(),
  role: z.string(),
});

export type TokenResponse = z.infer<typeof tokenResponseSchema>;
