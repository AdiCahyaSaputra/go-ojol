import { QueryClient } from '@tanstack/react-query';
import { toast } from 'sonner-native';
import { getErrorMessage } from '@/lib/utils/error-parser';

export const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      retry: false,
    },
    mutations: {
      onError: (error) => {
        toast.error(getErrorMessage(error));
      },
    },
  },
});
