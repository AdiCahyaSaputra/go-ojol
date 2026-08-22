import { z } from "zod";

export const exampleSchema = z.object({
});

export type ExampleSchema = z.infer<typeof exampleSchema>;

export const exampleResponse = z.object({
  id: z.number(),
  name: z.string(),
});

export type ExampleResponse = z.infer<typeof exampleResponse>;