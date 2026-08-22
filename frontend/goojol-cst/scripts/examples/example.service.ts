import type { ExampleSchema } from "./example.schema";
import { axiosClient } from "@/lib/api/axios-client";
import { parsedApiResponse } from "@/lib/utils/parses-api-response";
import { exampleResponse } from "./example.schema";

export const exampleService = async (data: ExampleSchema) => {
  const response = await axiosClient.post('/examples', data);
  const parsedResponse = parsedApiResponse(exampleResponse, response);

  return parsedResponse.data;
};