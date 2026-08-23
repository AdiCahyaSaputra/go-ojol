import { useQuery } from '@tanstack/react-query';

export const useRegisterQuery = () => {
  return useQuery({
    queryKey: ['register'],
    queryFn: () => {},
  });
};
