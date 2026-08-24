import { getJson, setJson } from '@/lib/storage/kv-storage';
import {
  type SavedAddress,
  savedAddressListSchema,
  savedAddressSchema,
} from './saved-address.schema';

const STORAGE_KEY = 'goojol.saved_addresses';

const DEFAULT_ADDRESSES: SavedAddress[] = [
  { id: 'seed-home', name: 'Home', lat: '-6.2088', lng: '106.8456' },
  { id: 'seed-office', name: 'Office', lat: '-6.1754', lng: '106.8272' },
];

export async function listSavedAddresses(): Promise<SavedAddress[]> {
  const stored = await getJson<unknown>(STORAGE_KEY);
  if (!stored) {
    await setJson(STORAGE_KEY, DEFAULT_ADDRESSES);
    return DEFAULT_ADDRESSES;
  }

  const parsed = savedAddressListSchema.safeParse(stored);
  if (!parsed.success || parsed.data.length === 0) {
    await setJson(STORAGE_KEY, DEFAULT_ADDRESSES);
    return DEFAULT_ADDRESSES;
  }

  return parsed.data;
}

export async function upsertSavedAddress(
  input: Omit<SavedAddress, 'id'> & { id?: string },
): Promise<SavedAddress[]> {
  const current = await listSavedAddresses();
  const nextAddress = savedAddressSchema.parse({
    id: input.id ?? `addr-${Date.now()}`,
    name: input.name,
    lat: input.lat,
    lng: input.lng,
  });

  const withoutDuplicate = current.filter((item) => item.id !== nextAddress.id);
  const next = [...withoutDuplicate, nextAddress];
  await setJson(STORAGE_KEY, next);
  return next;
}
