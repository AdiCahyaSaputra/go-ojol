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
  destination: BookLocation;
  vehicleType: 'car' | 'motorcycle';
  maxSize: number;
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
    pickup_lat_long: [input.pickup.lat, input.pickup.lng],
    destination_lat_long: [input.destination.lat, input.destination.lng],
    vehicle_type: input.vehicleType,
    max_size: input.maxSize,
  });

  return parsedApiResponse(findDriverResponseSchema, response).data;
}
