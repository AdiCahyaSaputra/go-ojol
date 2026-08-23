import { useColorScheme as useColorSchemeCore } from 'react-native';

export const useColorScheme = () => {
  const coreScheme = useColorSchemeCore();
  // RN 0.82+ can report 'unspecified' while following system; treat as light fallback
  // only when truly unknown — prefer dark/light when available.
  if (coreScheme === 'unspecified' || coreScheme == null) {
    return 'light';
  }
  return coreScheme;
};
