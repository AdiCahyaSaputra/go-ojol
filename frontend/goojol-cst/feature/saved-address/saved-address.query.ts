import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { QUERY_KEYS } from '@/constants/query-keys';
import type { SavedAddress } from './saved-address.schema';
import { listSavedAddresses, upsertSavedAddress } from './saved-address.storage';

export function useSavedAddressesQuery() {
  return useQuery({
    queryKey: QUERY_KEYS.savedAddresses,
    queryFn: listSavedAddresses,
  });
}

export function useUpsertSavedAddressMutation() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (input: Omit<SavedAddress, 'id'> & { id?: string }) => upsertSavedAddress(input),
    onSuccess: (data) => {
      queryClient.setQueryData(QUERY_KEYS.savedAddresses, data);
    },
  });
}
