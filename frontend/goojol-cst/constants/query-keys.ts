export const QUERY_KEYS = {
  home: ['home'] as const,
  userMe: ['user', 'me'] as const,
  savedAddresses: ['saved-addresses'] as const,
  calculateArgo: (parts: {
    pickupLat: string;
    pickupLng: string;
    destinationLat?: string;
    destinationLng?: string;
  }) =>
    [
      'calculate-argo',
      parts.pickupLat,
      parts.pickupLng,
      parts.destinationLat,
      parts.destinationLng,
    ] as const,
};
