import { SafeAreaView } from 'react-native-safe-area-context';
import RegisterPage from '@/feature/register/components/register-page';

const Register = () => {
  return (
    <SafeAreaView edges={['top']} style={{ flex: 1, backgroundColor: '#0F1729' }}>
      <RegisterPage />
    </SafeAreaView>
  );
};

export default Register;
