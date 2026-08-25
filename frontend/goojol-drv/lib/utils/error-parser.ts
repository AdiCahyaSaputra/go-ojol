import { AxiosError } from 'axios';

type ApiErrorBody = {
  message?: string | string[];
  error?: string | Record<string, string>;
};

export const getErrorMessage = (error: unknown): string => {
  const defaultMessage = 'An unexpected error occurred';

  if (error instanceof AxiosError) {
    const status = error.response?.status || 500;
    if (status >= 500) {
      return defaultMessage;
    }

    const data = error.response?.data as ApiErrorBody | undefined;

    if (typeof data?.error === 'string' && data.error) {
      return data.error;
    }

    if (data?.error && typeof data.error === 'object') {
      const fieldErrors = Object.values(data.error).filter(Boolean);
      if (fieldErrors.length > 0) {
        return fieldErrors.join(', ');
      }
    }

    const message = data?.message;
    if (Array.isArray(message)) {
      return message.join(', ') || defaultMessage;
    }

    if (typeof message === 'string' && message) {
      return message;
    }

    return defaultMessage;
  }

  if (error instanceof Error && error.message) {
    return error.message;
  }

  return defaultMessage;
};
