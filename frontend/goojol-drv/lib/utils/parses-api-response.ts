import type { AxiosResponse } from 'axios';
import { z } from 'zod';

const apiEnvelopeSchema = z.object({
  status: z.boolean(),
  message: z.string(),
  error: z.unknown().optional(),
  data: z.unknown().optional(),
  meta: z.unknown().optional(),
});

export type ApiEnvelope<TData> = {
  status: boolean;
  message: string;
  data: TData;
  error?: unknown;
  meta?: unknown;
};

export function parsedApiResponse<TSchema extends z.ZodType>(
  schema: TSchema,
  response: AxiosResponse<unknown>,
): ApiEnvelope<z.infer<TSchema>> {
  const envelope = apiEnvelopeSchema.parse(response.data);

  if (!envelope.status) {
    const errorMessage =
      typeof envelope.error === 'string' ? envelope.error : envelope.message || 'Request failed';
    throw new Error(errorMessage);
  }

  return {
    status: envelope.status,
    message: envelope.message,
    data: schema.parse(envelope.data),
    error: envelope.error,
    meta: envelope.meta,
  };
}
