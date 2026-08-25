export const DEFAULT_LOCATION = {
  name: 'Current location',
  lat: -6.2088,
  lng: 106.8456,
} as const;

export const LOCATION_HEARTBEAT_MS = 10_000;
export const OFFER_COUNTDOWN_SEC = 30;

/** Mock offer near the Jakarta demo area (frontend-only until trip offers exist). */
export const MOCK_OFFER = {
  id: 'mock-offer-001',
  customerName: 'Raka Pratama',
  pickup: {
    name: 'Plaza Indonesia',
    lat: -6.1935,
    lng: 106.8219,
  },
  destination: {
    name: 'Monas',
    lat: -6.1754,
    lng: 106.8272,
  },
  distanceM: 3200,
  durationSec: 720,
  totalFare: 28000,
  path: [
    [-6.1935, 106.8219],
    [-6.188, 106.823],
    [-6.182, 106.825],
    [-6.1754, 106.8272],
  ] as [number, number][],
};
