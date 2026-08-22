import { useMutation } from "@tanstack/react-query";
import type { ExampleSchema } from "./example.schema";
import { exampleService } from "./example.service";

export const useExampleMutation = () => {
  return useMutation({
    mutationFn: (data: ExampleSchema) => {
      return exampleService(data);
    },
    onSuccess: () => {
      console.log("Example created successfully");
    },
  });
};