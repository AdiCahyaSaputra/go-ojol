import { z } from "zod";

export const loginSchema = z.object({
});

export type LoginSchema = z.infer<typeof loginSchema>;

export const loginResponse = z.object({
  id: z.number(),
  name: z.string(),
});

export type LoginResponse = z.infer<typeof loginResponse>;