import { SafeAreaView } from 'react-native-safe-area-context';
import LoginPage from '@/feature/login/components/login-page';

const Login = () => {
  return (
    <SafeAreaView edges={['top']} style={{ flex: 1, backgroundColor: '#0F1729' }}>
      <LoginPage />
    </SafeAreaView>
  );
};

export default Login;
