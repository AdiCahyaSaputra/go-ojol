type StandbyCoords = {
  lat: number;
  lng: number;
};

export type TripOffer = {
  transactionId: string;
  customerName: string;
  pickup: [number, number];
  destination: [number, number];
  distanceM: number;
  totalFare: number;
  expiresInSec: number;
};

type TripOfferPayload = {
  transaction_id?: string;
  customer_name?: string;
  pickup?: [number, number];
  destination?: [number, number];
  distance_m?: number;
  total_fare?: number;
  expires_in_sec?: number;
};

export type TripLocationEvent = {
  transactionId: string;
  lat: number;
  lng: number;
};

export type TripStatusEvent = {
  transactionId: string;
  status: string;
};

export type TripCompletedEvent = {
  transactionId: string;
  status: string;
  totalFare: number;
  paidAt: string;
};

type StandbyWsHandlers = {
  onStandbyOk?: () => void;
  onTripOffer?: (offer: TripOffer) => void;
  onOfferTaken?: (transactionId: string) => void;
  onOfferExpired?: (transactionId: string) => void;
  onCustomerLocation?: (event: TripLocationEvent) => void;
  onTripStatus?: (event: TripStatusEvent) => void;
  onTripCompleted?: (event: TripCompletedEvent) => void;
  onError?: (message: string) => void;
  onClose?: () => void;
};

export class StandbySocket {
  private socket: WebSocket | null = null;
  private closedByClient = false;

  connect(accessToken: string, handlers: StandbyWsHandlers = {}) {
    this.disconnect();
    this.closedByClient = false;

    const wsUrl = process.env.EXPO_PUBLIC_WS_URL ?? '';
    const url = `${wsUrl}/api/trip/dispatch/ws`;

    const socket = new WebSocket(url, [accessToken]);
    this.socket = socket;

    socket.onmessage = (event) => {
      try {
        const message = JSON.parse(String(event.data)) as {
          type?: string;
          message?: string;
          transaction_id?: string;
          lat?: number;
          lng?: number;
          status?: string;
          total_fare?: number;
          paid_at?: string;
          offer?: TripOfferPayload;
        };

        if (message.type === 'standby_ok') {
          handlers.onStandbyOk?.();
          return;
        }

        if (message.type === 'trip_offer') {
          const offer = message.offer;
          const transactionId = offer?.transaction_id ?? message.transaction_id ?? '';
          if (!offer || !transactionId || !offer.pickup || !offer.destination) {
            handlers.onError?.('Invalid trip offer payload');
            return;
          }
          handlers.onTripOffer?.({
            transactionId,
            customerName: offer.customer_name ?? 'Customer',
            pickup: offer.pickup,
            destination: offer.destination,
            distanceM: offer.distance_m ?? 0,
            totalFare: offer.total_fare ?? 0,
            expiresInSec: offer.expires_in_sec ?? 30,
          });
          return;
        }

        if (message.type === 'offer_taken') {
          handlers.onOfferTaken?.(message.transaction_id ?? '');
          return;
        }

        if (message.type === 'offer_expired') {
          handlers.onOfferExpired?.(message.transaction_id ?? '');
          return;
        }

        if (message.type === 'customer_location') {
          handlers.onCustomerLocation?.({
            transactionId: message.transaction_id ?? '',
            lat: message.lat ?? 0,
            lng: message.lng ?? 0,
          });
          return;
        }

        if (message.type === 'trip_status') {
          handlers.onTripStatus?.({
            transactionId: message.transaction_id ?? '',
            status: message.status ?? '',
          });
          return;
        }

        if (message.type === 'trip_completed') {
          handlers.onTripCompleted?.({
            transactionId: message.transaction_id ?? '',
            status: message.status ?? 'completed',
            totalFare: message.total_fare ?? 0,
            paidAt: message.paid_at ?? '',
          });
          return;
        }

        if (message.type === 'error') {
          handlers.onError?.(message.message ?? 'WebSocket error');
        }
      } catch {
        handlers.onError?.('Invalid WebSocket message');
      }
    };

    socket.onerror = () => {
      handlers.onError?.('WebSocket connection failed');
    };

    socket.onclose = () => {
      if (this.socket === socket) {
        this.socket = null;
      }
      if (!this.closedByClient) {
        handlers.onClose?.();
      }
    };
  }

  private send(payload: Record<string, unknown>) {
    if (!this.socket || this.socket.readyState !== WebSocket.OPEN) {
      return false;
    }
    this.socket.send(JSON.stringify(payload));
    return true;
  }

  sendStandby(coords: StandbyCoords) {
    return this.send({ type: 'standby', lat: coords.lat, lng: coords.lng });
  }

  sendLocation(coords: StandbyCoords) {
    return this.send({ type: 'location', lat: coords.lat, lng: coords.lng });
  }

  sendTripLocation(transactionId: string, coords: StandbyCoords) {
    return this.send({
      type: 'trip_location',
      transaction_id: transactionId,
      lat: coords.lat,
      lng: coords.lng,
    });
  }

  whenOpen(callback: () => void) {
    if (!this.socket) {
      return;
    }
    if (this.socket.readyState === WebSocket.OPEN) {
      callback();
      return;
    }
    this.socket.addEventListener('open', callback, { once: true });
  }

  disconnect() {
    this.closedByClient = true;
    if (this.socket) {
      this.socket.close();
      this.socket = null;
    }
  }

  get isConnected() {
    return this.socket?.readyState === WebSocket.OPEN;
  }
}
