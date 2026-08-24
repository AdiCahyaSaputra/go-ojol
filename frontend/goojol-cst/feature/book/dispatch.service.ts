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
  vehicleId: string;
};

type FindDriverInput = {
  pickup: BookLocation;
  vehicleType: 'car' | 'motorcycle';
};

export async function calculateArgo(input: CalculateArgoInput) {
  const response = await axiosClient.post('/api/trip/dispatch/customer/calculate-argo', {
    pickup_loc: [input.pickup.lat, input.pickup.lng],
    destination: [input.destination.lat, input.destination.lng],
    vehicle_id: input.vehicleId,
  });

  return parsedApiResponse(calculateArgoResponseSchema, response).data;
}

export async function findDriver(input: FindDriverInput) {
  const response = await axiosClient.post('/api/trip/dispatch/customer/find-driver', {
    current_lat_long: [input.pickup.lat, input.pickup.lng],
    vehicle_type: input.vehicleType,
  });

  return parsedApiResponse(findDriverResponseSchema, response).data;
}
