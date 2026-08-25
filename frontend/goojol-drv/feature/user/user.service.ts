import { axiosClient } from '@/lib/api/axios-client';
import { parsedApiResponse } from '@/lib/utils/parses-api-response';
import { userMeSchema } from './user.schema';

export async function fetchCurrentUser() {
  const response = await axiosClient.get('/api/user/me');
  return parsedApiResponse(userMeSchema, response).data;
}
