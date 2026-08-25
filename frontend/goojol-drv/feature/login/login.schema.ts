import { z } from 'zod';
import { tokenResponseSchema } from '@/lib/auth/token-response.schema';

export const loginSchema = z.object({
  email: z.email('Enter a valid email'),
  password: z.string().min(1, 'Password is required'),
});

export type LoginSchema = z.infer<typeof loginSchema>;

export const loginResponse = tokenResponseSchema;

export type LoginResponse = z.infer<typeof loginResponse>;
