import { Image, View, type ViewStyle } from 'react-native';

const PIXEL = 4;

const COLORS = {
  pixel: '#FFD166',
  pixelDim: '#C9A84F',
  road: '#E8ECF4',
  teal: '#3DDBA8',
  surface: '#1A2238',
  coral: '#FF6B4A',
} as const;

type PixelProps = {
  col: number;
  row: number;
  color: string;
  originLeft: number;
  originBottom: number;
};

function Pixel({ col, row, color, originLeft, originBottom }: PixelProps) {
  return (
    <View
      style={{
        position: 'absolute',
        left: originLeft + col * PIXEL,
        bottom: originBottom + row * PIXEL,
        width: PIXEL,
        height: PIXEL,
        backgroundColor: color,
      }}
    />
  );
}

function LampOverlay({ style }: { style?: ViewStyle }) {
  const left = 0;
  const bottom = 0;

  const lampPixels: Array<[number, number, string]> = [
    [2, 0, COLORS.pixelDim],
    [2, 1, COLORS.pixel],
    [2, 2, COLORS.pixel],
    [2, 3, COLORS.pixel],
    [1, 4, COLORS.pixelDim],
    [2, 4, COLORS.pixel],
    [3, 4, COLORS.pixelDim],
    [2, 5, COLORS.surface],
    [2, 6, COLORS.surface],
    [2, 7, COLORS.surface],
    [1, 8, COLORS.surface],
    [2, 8, COLORS.surface],
    [3, 8, COLORS.surface],
  ];

  return (
    <View style={[{ position: 'absolute', width: 24, height: 40 }, style]} pointerEvents="none">
      {lampPixels.map(([col, row, color]) => (
        <Pixel
          key={`lamp-${col}-${row}`}
          col={col}
          row={row}
          color={color}
          originLeft={left}
          originBottom={bottom}
        />
      ))}
    </View>
  );
}

function PassengerOverlay({ style }: { style?: ViewStyle }) {
  const left = 0;
  const bottom = 0;

  const personPixels: Array<[number, number, string]> = [
    [2, 8, COLORS.road],
    [1, 7, COLORS.teal],
    [2, 7, COLORS.teal],
    [3, 7, COLORS.teal],
    [2, 6, COLORS.coral],
    [1, 5, COLORS.road],
    [2, 5, COLORS.road],
    [3, 5, COLORS.road],
    [1, 4, COLORS.surface],
    [3, 4, COLORS.surface],
    [0, 3, COLORS.road],
    [4, 3, COLORS.road],
  ];

  return (
    <View style={[{ position: 'absolute', width: 24, height: 40 }, style]} pointerEvents="none">
      {personPixels.map(([col, row, color]) => (
        <Pixel
          key={`person-${col}-${row}`}
          col={col}
          row={row}
          color={color}
          originLeft={left}
          originBottom={bottom}
        />
      ))}
    </View>
  );
}

export default function RegisterHero() {
  return (
    <View
      className="relative h-[208px] w-full bg-goojol-sky"
      accessibilityLabel="Pixel art street scene with a scooter, street sign, waiting passenger, and stop lamp"
    >
      <Image
        source={require('@/assets/images/login/pixel-ojek-hero.png')}
        style={{ height: 208, width: '100%' }}
        resizeMode="contain"
      />
      <LampOverlay style={{ left: '12%', bottom: 52 }} />
      <PassengerOverlay style={{ right: '18%', bottom: 48 }} />
    </View>
  );
}
