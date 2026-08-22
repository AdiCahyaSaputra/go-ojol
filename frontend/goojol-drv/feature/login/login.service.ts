import type { LoginSchema } from "./login.schema";
import { axiosClient } from "@/lib/api/axios-client";
import { parsedApiResponse } from "@/lib/utils/parses-api-response";
import { loginResponse } from "./login.schema";

export const loginService = async (data: LoginSchema) => {
  const response = await axiosClient.post('/login', data);
  const parsedResponse = parsedApiResponse(loginResponse, response);

  return parsedResponse.data;
};