type StandbyCoords = {
  lat: number;
  lng: number;
};

type StandbyWsHandlers = {
  onStandbyOk?: () => void;
  onError?: (message: string) => void;
  onClose?: () => void;
};

function toWsBaseUrl(httpBase: string): string {
  const trimmed = httpBase.replace(/\/$/, '');
  if (trimmed.startsWith('https://')) {
    return `wss://${trimmed.slice('https://'.length)}`;
  }
  if (trimmed.startsWith('http://')) {
    return `ws://${trimmed.slice('http://'.length)}`;
  }
  return trimmed;
}

export class StandbySocket {
  private socket: WebSocket | null = null;
  private closedByClient = false;

  connect(accessToken: string, handlers: StandbyWsHandlers = {}) {
    this.disconnect();
    this.closedByClient = false;

    const apiBase = process.env.EXPO_PUBLIC_API_URL ?? '';
    const wsBase = toWsBaseUrl(apiBase);
    const url = `${wsBase}/api/trip/dispatch/ws?token=${encodeURIComponent(accessToken)}`;

    const socket = new WebSocket(url);
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
