import { useMutation } from '@tanstack/react-query';
import { useRouter } from 'expo-router';
import { saveSession } from '@/lib/auth/token-storage';
import type { LoginSchema } from './login.schema';
import { loginService } from './login.service';

export const useLoginMutation = () => {
  const router = useRouter();

  return useMutation({
    mutationFn: (data: LoginSchema) => {
      // #region agent log
      fetch('http://127.0.0.1:7387/ingest/ca1b84f5-beb5-4337-ab53-4aca48bdcb69', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'X-Debug-Session-Id': '1e7040',
        },
        body: JSON.stringify({
          sessionId: '1e7040',
          location: 'login.mutation.ts:mutationFn',
          message: 'login payload shape',
          hypothesisId: 'C',
          data: {
            emailLen: data.email.length,
            passwordLen: data.password.length,
            emailHasSpace: data.email !== data.email.trim(),
            keys: Object.keys(data),
          },
          timestamp: Date.now(),
        }),
      }).catch(() => {});
      // #endregion
      return loginService(data);
    },
    onSuccess: async (data) => {
      await saveSession({
        accessToken: data.access_token,
        refreshToken: data.refresh_token,
        role: data.role,
      });
      router.replace('/(auth)/(tabs)/home');
    },
  });
};
