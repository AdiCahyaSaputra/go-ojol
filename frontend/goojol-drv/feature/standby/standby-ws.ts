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

type StandbyWsHandlers = {
  onStandbyOk?: () => void;
  onTripOffer?: (offer: TripOffer) => void;
  onOfferTaken?: (transactionId: string) => void;
  onOfferExpired?: (transactionId: string) => void;
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

    // NOTES: workaround as we can't send Authorization header in websocket connection
    // https://github.com/kubernetes/kubernetes/commit/714f97d7baf4975ad3aa47735a868a81a984d1f0#diff-43f0e0ab1c89ddde1a59685bcdbe8403d5db98da2c5b7de7ad5191e2e8665e3aR15-R34
    const socket = new WebSocket(url, [accessToken]); // Sec-WebSocket-Protocol
    this.socket = socket;

    socket.onmessage = (event) => {
      try {
        const message = JSON.parse(String(event.data)) as {
          type?: string;
          message?: string;
          transaction_id?: string;
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
