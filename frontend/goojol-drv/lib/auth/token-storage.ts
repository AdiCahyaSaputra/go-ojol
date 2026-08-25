import * as SecureStore from 'expo-secure-store';
import { Platform } from 'react-native';

const ACCESS_TOKEN_KEY = 'goojol.access_token';
const REFRESH_TOKEN_KEY = 'goojol.refresh_token';
const ROLE_KEY = 'goojol.role';

export type AuthSession = {
  accessToken: string;
  refreshToken: string;
  role: string;
};

async function setItem(key: string, value: string): Promise<void> {
  if (Platform.OS === 'web') {
    if (typeof localStorage !== 'undefined') {
      localStorage.setItem(key, value);
    }
    return;
  }
  await SecureStore.setItemAsync(key, value);
}

async function getItem(key: string): Promise<string | null> {
  if (Platform.OS === 'web') {
    if (typeof localStorage === 'undefined') {
      return null;
    }
    return localStorage.getItem(key);
  }
  return SecureStore.getItemAsync(key);
}

async function deleteItem(key: string): Promise<void> {
  if (Platform.OS === 'web') {
    if (typeof localStorage !== 'undefined') {
      localStorage.removeItem(key);
    }
    return;
  }
  await SecureStore.deleteItemAsync(key);
}

export async function saveSession(session: AuthSession): Promise<void> {
  await Promise.all([
    setItem(ACCESS_TOKEN_KEY, session.accessToken),
    setItem(REFRESH_TOKEN_KEY, session.refreshToken),
    setItem(ROLE_KEY, session.role),
  ]);
}

export async function getSession(): Promise<AuthSession | null> {
  const [accessToken, refreshToken, role] = await Promise.all([
    getItem(ACCESS_TOKEN_KEY),
    getItem(REFRESH_TOKEN_KEY),
    getItem(ROLE_KEY),
  ]);

  if (!accessToken || !refreshToken || !role) {
    return null;
  }

  return { accessToken, refreshToken, role };
}

export async function clearSession(): Promise<void> {
  await Promise.all([
    deleteItem(ACCESS_TOKEN_KEY),
    deleteItem(REFRESH_TOKEN_KEY),
    deleteItem(ROLE_KEY),
  ]);
}
