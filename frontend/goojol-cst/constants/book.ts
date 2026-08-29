export const DEFAULT_PICKUP = {
  name: 'Current location',
  lat: '-6.2088',
  lng: '106.8456',
} as const;

export const LOCATION_HEARTBEAT_MS = 10_000;

export const VEHICLE_OPTIONS = [
  {
    id: 'fd1dae62-002f-4a1b-9c5b-ea74f54b7168',
    label: 'Motorcycle',
    type: 'motorcycle' as const,
  },
  {
    id: '5bd64683-675b-4de7-ade3-e380e129c820',
    label: 'Car',
    type: 'car' as const,
  },
] as const;

export const DESTINATION_PRESETS = [
  { name: 'Sudirman', lat: '-6.2088', lng: '106.8256' },
  { name: 'Monas', lat: '-6.1754', lng: '106.8272' },
] as const;
