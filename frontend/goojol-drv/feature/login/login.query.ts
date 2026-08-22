import { useQuery } from "@tanstack/react-query";

export const useLoginQuery = () => {
  return useQuery({
    queryKey: ["login"],
    queryFn: () => {}
  });
};