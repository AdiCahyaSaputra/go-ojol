import { useQuery } from '@tanstack/react-query';
import { QUERY_KEYS } from '@/constants/query-keys';
import { fetchCurrentUser } from './user.service';

export function useCurrentUserQuery() {
  return useQuery({
    queryKey: QUERY_KEYS.userMe,
    queryFn: fetchCurrentUser,
  });
}
