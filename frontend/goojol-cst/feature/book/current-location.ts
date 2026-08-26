import * as Location from 'expo-location';
import { DEFAULT_PICKUP } from '@/constants/book';
import type { BookLocation } from '@/feature/book/dispatch.schema';

export type CurrentLocationResult = BookLocation & {
  fromGps: boolean;
};

export async function resolveCurrentLocation(): Promise<CurrentLocationResult> {
  try {
    const permission = await Location.requestForegroundPermissionsAsync();
    if (!permission.granted) {
      return { ...DEFAULT_PICKUP, fromGps: false };
    }

    const position = await Location.getCurrentPositionAsync({
      accuracy: Location.Accuracy.Balanced,
    });

    return {
      name: DEFAULT_PICKUP.name,
      lat: String(position.coords.latitude),
      lng: String(position.coords.longitude),
      fromGps: true,
    };
  } catch {
    return { ...DEFAULT_PICKUP, fromGps: false };
  }
}
