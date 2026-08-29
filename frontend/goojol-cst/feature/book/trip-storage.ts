import { getJson, setJson } from '@/lib/storage/kv-storage';
import type { StoredActiveTrip } from './trip.schema';
import { ACTIVE_TRIP_STORAGE_KEY } from './trip.schema';

export async function loadStoredActiveTrip(): Promise<StoredActiveTrip | null> {
  return getJson<StoredActiveTrip>(ACTIVE_TRIP_STORAGE_KEY);
}

export async function saveStoredActiveTrip(trip: StoredActiveTrip | null): Promise<void> {
  if (!trip) {
    await setJson(ACTIVE_TRIP_STORAGE_KEY, null);
    return;
  }
  await setJson(ACTIVE_TRIP_STORAGE_KEY, trip);
}
