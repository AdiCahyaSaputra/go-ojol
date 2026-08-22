import { SafeAreaView } from 'react-native-safe-area-context';
import LoginPage from '@/feature/login/components/login-page';

const Login = () => {
  return (
    <SafeAreaView edges={['top']} className="flex-1 bg-goojol-sky">
      <LoginPage />
    </SafeAreaView>
  );
};

export default Login;
