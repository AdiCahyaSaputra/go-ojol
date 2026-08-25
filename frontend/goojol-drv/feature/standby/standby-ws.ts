type StandbyCoords = {
  lat: number;
  lng: number;
};

type StandbyWsHandlers = {
  onStandbyOk?: () => void;
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
        };
        if (message.type === 'standby_ok') {
          handlers.onStandbyOk?.();
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
