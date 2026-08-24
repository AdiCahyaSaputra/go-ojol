import { axiosClient } from '@/lib/api/axios-client';
import { parsedApiResponse } from '@/lib/utils/parses-api-response';
import {
  type BookLocation,
  calculateArgoResponseSchema,
  findDriverResponseSchema,
} from './dispatch.schema';

type CalculateArgoInput = {
  pickup: BookLocation;
  destination: BookLocation;
};

type FindDriverInput = {
  pickup: BookLocation;
  vehicleType: 'car' | 'motorcycle';
};

export async function calculateArgo(input: CalculateArgoInput) {
  const params = new URLSearchParams();
  params.append('pickup_loc', input.pickup.lat);
  params.append('pickup_loc', input.pickup.lng);
  params.append('destination', input.destination.lat);
  params.append('destination', input.destination.lng);

  const response = await axiosClient.get(
    `/api/trip/dispatch/customer/calculate-argo?${params.toString()}`,
  );

  return parsedApiResponse(calculateArgoResponseSchema, response).data;
}

export async function findDriver(input: FindDriverInput) {
  const response = await axiosClient.post('/api/trip/dispatch/customer/find-driver', {
    current_lat_long: [input.pickup.lat, input.pickup.lng],
    vehicle_type: input.vehicleType,
  });

  return parsedApiResponse(findDriverResponseSchema, response).data;
}
