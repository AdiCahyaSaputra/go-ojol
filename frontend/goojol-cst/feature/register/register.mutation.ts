import { useMutation } from '@tanstack/react-query';
import { useRouter } from 'expo-router';
import type { RegisterRequestSchema } from './register.schema';
import { registerService } from './register.service';

export const useRegisterMutation = () => {
  const router = useRouter();

  return useMutation({
    mutationFn: (data: RegisterRequestSchema) => {
      return registerService(data);
    },
    onSuccess: () => {
      router.replace('/(public)/login');
    },
  });
};
