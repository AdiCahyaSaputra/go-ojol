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

type DispatchWsHandlers = {
  onWaiting?: (event: DispatchWaitingEvent) => void;
  onDriverMatched?: (event: DispatchMatchedEvent) => void;
  onOfferExpired?: (event: DispatchTransactionEvent) => void;
  onOfferRejected?: (event: DispatchTransactionEvent) => void;
  onNoDrivers?: () => void;
  onError?: (message: string) => void;
  onClose?: () => void;
};

type ServerMessage = {
  type?: string;
  message?: string;
  transaction_id?: string;
  expires_in_sec?: number;
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

    // NOTES: workaround as we can't send Authorization header in websocket connection
    // https://github.com/kubernetes/kubernetes/commit/714f97d7baf4975ad3aa47735a868a81a984d1f0#diff-43f0e0ab1c89ddde1a59685bcdbe8403d5db98da2c5b7de7ad5191e2e8665e3aR15-R34
    const socket = new WebSocket(url, [accessToken]); // Sec-WebSocket-Protocol
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
