import { z } from 'zod';

export const registerFormSchema = z
  .object({
    name: z.string().min(1, 'Name is required'),
    phone_number: z.string().min(1, 'Phone number is required'),
    email: z.email('Enter a valid email'),
    password: z.string().min(8, 'Password must be at least 8 characters'),
    confirmPassword: z.string().min(1, 'Confirm your password'),
  })
  .refine((data) => data.password === data.confirmPassword, {
    message: 'Passwords do not match',
    path: ['confirmPassword'],
  });

export type RegisterFormSchema = z.infer<typeof registerFormSchema>;

export const registerRequestSchema = z.object({
  email: z.email(),
  password: z.string().min(8),
  role: z.literal('customer'),
  name: z.string().min(1),
  phone_number: z.string().min(1),
});

export type RegisterRequestSchema = z.infer<typeof registerRequestSchema>;

export const registerResponse = z.object({
  id: z.string(),
  email: z.string(),
  role: z.string(),
  customer: z
    .object({
      id: z.string(),
      name: z.string(),
      phone_number: z.string(),
      profile_picture_url: z.string().nullable().optional(),
    })
    .optional(),
  created_at: z.string(),
  updated_at: z.string(),
});

export type RegisterResponse = z.infer<typeof registerResponse>;
