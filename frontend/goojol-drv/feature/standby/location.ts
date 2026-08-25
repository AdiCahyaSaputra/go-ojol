import * as Location from 'expo-location';
import { DEFAULT_LOCATION } from '@/constants/standby';

export type DriverCoords = {
  lat: number;
  lng: number;
  fromGps: boolean;
};

export async function resolveDriverLocation(): Promise<DriverCoords> {
  try {
    const permission = await Location.requestForegroundPermissionsAsync();
    if (!permission.granted) {
      return { lat: DEFAULT_LOCATION.lat, lng: DEFAULT_LOCATION.lng, fromGps: false };
    }

    const position = await Location.getCurrentPositionAsync({
      accuracy: Location.Accuracy.Balanced,
    });

    return {
      lat: position.coords.latitude,
      lng: position.coords.longitude,
      fromGps: true,
    };
  } catch {
    return { lat: DEFAULT_LOCATION.lat, lng: DEFAULT_LOCATION.lng, fromGps: false };
  }
}
