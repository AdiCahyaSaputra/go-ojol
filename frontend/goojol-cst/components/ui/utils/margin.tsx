import { type DimensionValue, View } from 'react-native';

type Props = {
  value: DimensionValue;
};

const Margin = ({ value }: Props) => {
  return (
    <View
      style={{
        height: value,
      }}
    />
  );
};

export default Margin;
