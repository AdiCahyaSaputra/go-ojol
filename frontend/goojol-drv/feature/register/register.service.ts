import { axiosClient } from '@/lib/api/axios-client';
import { parsedApiResponse } from '@/lib/utils/parses-api-response';
import type { RegisterRequestSchema } from './register.schema';
import { registerResponse } from './register.schema';

export const registerService = async (data: RegisterRequestSchema) => {
  const response = await axiosClient.post('/api/auth/register', data);
  const parsedResponse = parsedApiResponse(registerResponse, response);

  return parsedResponse.data;
};
