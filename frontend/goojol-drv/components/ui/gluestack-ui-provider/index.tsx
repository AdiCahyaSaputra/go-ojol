import { OverlayProvider } from '@gluestack-ui/core/overlay/creator';
import { ToastProvider } from '@gluestack-ui/core/toast/creator';
import type React from 'react';
import { useEffect } from 'react';
import { Appearance, type ColorSchemeName, View, type ViewProps } from 'react-native';

export type ModeType = 'light' | 'dark' | 'system';

function toColorScheme(mode: ModeType): ColorSchemeName {
  // RN 0.82+: follow system with 'unspecified', not 'system' / null
  if (mode === 'system') return 'unspecified';
  return mode;
}

export function GluestackUIProvider({
  mode = 'system',
  ...props
}: {
  mode?: ModeType;
  children?: React.ReactNode;
  style?: ViewProps['style'];
}) {
  useEffect(() => {
    Appearance.setColorScheme(toColorScheme(mode));
  }, [mode]);

  return (
    <View style={[{ flex: 1, height: '100%', width: '100%' }, props.style]}>
      <OverlayProvider>
        <ToastProvider>{props.children}</ToastProvider>
      </OverlayProvider>
    </View>
  );
}
