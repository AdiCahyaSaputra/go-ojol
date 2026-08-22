import { useMutation } from "@tanstack/react-query";
import type { LoginSchema } from "./login.schema";
import { loginService } from "./login.service";

export const useLoginMutation = () => {
  return useMutation({
    mutationFn: (data: LoginSchema) => {
      return loginService(data);
    },
    onSuccess: () => {
      console.log("Login created successfully");
    },
  });
};