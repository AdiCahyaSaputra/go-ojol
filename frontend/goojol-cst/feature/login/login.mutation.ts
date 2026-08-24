import { useMutation } from '@tanstack/react-query';
import { useRouter } from 'expo-router';
import { saveSession } from '@/lib/auth/token-storage';
import type { LoginSchema } from './login.schema';
import { loginService } from './login.service';

export const useLoginMutation = () => {
  const router = useRouter();

  return useMutation({
    mutationFn: (data: LoginSchema) => {
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
