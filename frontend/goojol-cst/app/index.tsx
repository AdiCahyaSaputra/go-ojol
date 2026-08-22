import { usePathname, useRouter } from 'expo-router';
import { useEffect } from 'react';
import { Text, View } from 'react-native';

const EntryPoint = () => {
	const currentPath = usePathname();
	const router = useRouter();

	useEffect(() => {
		if (currentPath !== '/login') {
			// TODO: need to check for auth
			router.replace('/(public)/login');
		}
	}, [currentPath]);

	// TODO: create spalsh screen
	return (
		<View className="flex-1">
			<Text>Can be splashscreen</Text>
		</View>
	);
};

export default EntryPoint;
