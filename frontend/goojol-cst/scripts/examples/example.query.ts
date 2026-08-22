import { useQuery } from "@tanstack/react-query";

export const useExampleQuery = () => {
  return useQuery({
    queryKey: ["example"],
    queryFn: () => {}
  });
};