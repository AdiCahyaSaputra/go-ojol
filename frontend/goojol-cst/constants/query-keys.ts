export const QUERY_KEYS = {
  home: ['home'] as const,
  userMe: ['user', 'me'] as const,
  savedAddresses: ['saved-addresses'] as const,
  calculateArgo: (parts: {
    pickupLat: string;
    pickupLng: string;
    destinationLat?: string;
    destinationLng?: string;
    vehicleType: 'car' | 'motorcycle';
  }) =>
    [
      'calculate-argo',
      parts.pickupLat,
      parts.pickupLng,
      parts.destinationLat,
      parts.destinationLng,
      parts.vehicleType,
    ] as const,
};
