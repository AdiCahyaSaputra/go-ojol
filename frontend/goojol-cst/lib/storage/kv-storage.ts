import * as SecureStore from 'expo-secure-store';
import { Platform } from 'react-native';

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

export async function getJson<T>(key: string): Promise<T | null> {
  const raw = await getItem(key);
  if (!raw) {
    return null;
  }
  return JSON.parse(raw) as T;
}

export async function setJson<T>(key: string, value: T): Promise<void> {
  await setItem(key, JSON.stringify(value));
}
