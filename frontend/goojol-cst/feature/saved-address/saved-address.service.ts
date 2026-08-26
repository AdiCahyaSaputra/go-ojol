import { axiosClient } from '@/lib/api/axios-client';
import { parsedApiResponse } from '@/lib/utils/parses-api-response';
import {
  type CreateSavedAddressRequest,
  savedAddressListSchema,
  savedAddressSchema,
} from './saved-address.schema';

export async function listSavedAddresses() {
  const response = await axiosClient.get('/api/trip/saved-addresses');

	console.log(response);

  return parsedApiResponse(savedAddressListSchema, response).data;
}

export async function createSavedAddress(input: CreateSavedAddressRequest) {
  const response = await axiosClient.post('/api/trip/saved-addresses', input);
  return parsedApiResponse(savedAddressSchema, response).data;
}
