import type { MatchedDriver } from './dispatch.schema';

export type DispatchWaitingEvent = {
  transactionId: string;
  expiresInSec: number;
};

export type DispatchMatchedEvent = {
  transactionId: string;
  matchedDriver: MatchedDriver;
};

export type DispatchTransactionEvent = {
  transactionId: string;
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

type DispatchWsHandlers = {
  onWaiting?: (event: DispatchWaitingEvent) => void;
  onDriverMatched?: (event: DispatchMatchedEvent) => void;
  onOfferExpired?: (event: DispatchTransactionEvent) => void;
  onOfferRejected?: (event: DispatchTransactionEvent) => void;
  onNoDrivers?: () => void;
  onDriverLocation?: (event: TripLocationEvent) => void;
  onTripStatus?: (event: TripStatusEvent) => void;
  onTripCompleted?: (event: TripCompletedEvent) => void;
  onError?: (message: string) => void;
  onClose?: () => void;
};

type ServerMessage = {
  type?: string;
  message?: string;
  transaction_id?: string;
  expires_in_sec?: number;
  lat?: number;
  lng?: number;
  status?: string;
  total_fare?: number;
  paid_at?: string;
  matched_driver?: MatchedDriver;
};

export class CustomerDispatchSocket {
  private socket: WebSocket | null = null;
  private closedByClient = false;

  connect(accessToken: string, handlers: DispatchWsHandlers = {}) {
    this.disconnect();
    this.closedByClient = false;

    const wsUrl = process.env.EXPO_PUBLIC_WS_URL ?? '';
    const url = `${wsUrl}/api/trip/dispatch/customer/ws`;

    const socket = new WebSocket(url, [accessToken]);
    this.socket = socket;

    socket.onmessage = (event) => {
      try {
        const message = JSON.parse(String(event.data)) as ServerMessage;
        switch (message.type) {
          case 'waiting':
            handlers.onWaiting?.({
              transactionId: message.transaction_id ?? '',
              expiresInSec: message.expires_in_sec ?? 30,
            });
            return;
          case 'driver_matched':
            if (!message.matched_driver) {
              handlers.onError?.('Matched driver payload missing');
              return;
            }
            handlers.onDriverMatched?.({
              transactionId: message.transaction_id ?? '',
              matchedDriver: message.matched_driver,
            });
            return;
          case 'offer_expired':
            handlers.onOfferExpired?.({
              transactionId: message.transaction_id ?? '',
            });
            return;
          case 'offer_rejected':
            handlers.onOfferRejected?.({
              transactionId: message.transaction_id ?? '',
            });
            return;
          case 'no_drivers':
            handlers.onNoDrivers?.();
            return;
          case 'driver_location':
            handlers.onDriverLocation?.({
              transactionId: message.transaction_id ?? '',
              lat: message.lat ?? 0,
              lng: message.lng ?? 0,
            });
            return;
          case 'trip_status':
            handlers.onTripStatus?.({
              transactionId: message.transaction_id ?? '',
              status: message.status ?? '',
            });
            return;
          case 'trip_completed':
            handlers.onTripCompleted?.({
              transactionId: message.transaction_id ?? '',
              status: message.status ?? 'completed',
              totalFare: message.total_fare ?? 0,
              paidAt: message.paid_at ?? '',
            });
            return;
          case 'error':
            handlers.onError?.(message.message ?? 'WebSocket error');
            return;
          default:
            break;
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

  sendRetry() {
    if (!this.socket || this.socket.readyState !== WebSocket.OPEN) {
      return false;
    }
    this.socket.send(JSON.stringify({ type: 'retry' }));
    return true;
  }

  sendTripLocation(transactionId: string, coords: { lat: number; lng: number }) {
    if (!this.socket || this.socket.readyState !== WebSocket.OPEN) {
      return false;
    }
    this.socket.send(
      JSON.stringify({
        type: 'trip_location',
        transaction_id: transactionId,
        lat: coords.lat,
        lng: coords.lng,
      }),
    );
    return true;
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
