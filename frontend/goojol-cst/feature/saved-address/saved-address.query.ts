import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { QUERY_KEYS } from '@/constants/query-keys';
import type { CreateSavedAddressRequest } from './saved-address.schema';
import { createSavedAddress, listSavedAddresses } from './saved-address.service';

export function useSavedAddressesQuery() {
  return useQuery({
    queryKey: QUERY_KEYS.savedAddresses,
    queryFn: listSavedAddresses,
  });
}

export function useCreateSavedAddressMutation() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (input: CreateSavedAddressRequest) => createSavedAddress(input),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: QUERY_KEYS.savedAddresses });
    },
  });
}
