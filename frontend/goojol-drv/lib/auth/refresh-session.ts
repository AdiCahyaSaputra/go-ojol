import axios from 'axios';
import { parsedApiResponse } from '@/lib/utils/parses-api-response';
import { tokenResponseSchema } from './token-response.schema';
import { clearSession, getSession, saveSession } from './token-storage';

const refreshClient = axios.create({
  baseURL: process.env.EXPO_PUBLIC_API_URL,
  headers: {
    Accept: 'application/json',
    'Content-Type': 'application/json',
  },
});

export async function refreshSession() {
  const session = await getSession();
  if (!session?.refreshToken) {
    await clearSession();
    throw new Error('No refresh token');
  }

  const response = await refreshClient.post('/api/auth/refresh', {
    refresh_token: session.refreshToken,
  });
  const parsed = parsedApiResponse(tokenResponseSchema, response);

  await saveSession({
    accessToken: parsed.data.access_token,
    refreshToken: parsed.data.refresh_token,
    role: parsed.data.role,
  });

  return parsed.data;
}
