import { Lock, Mail } from 'lucide-react-native';
import { useState } from 'react';
import {
	Image,
	KeyboardAvoidingView,
	Platform,
	Pressable,
	ScrollView,
	Text,
	TextInput,
	View,
} from 'react-native';

export default function LoginPage() {
	const [email, setEmail] = useState('');
	const [password, setPassword] = useState('');

	const canSignIn = email.trim().length > 0 && password.length > 0;

	return (
		<KeyboardAvoidingView
			className="flex-1 bg-goojol-sky"
			behavior={Platform.OS === 'ios' ? 'padding' : 'height'}
		>
			<ScrollView
				className="flex-1"
				contentContainerClassName="flex-grow pb-8"
				keyboardShouldPersistTaps="handled"
			>
				<View className="overflow-hidden rounded-b-3xl bg-goojol-sky">
					<Image
						source={require('@/assets/images/login/pixel-ojek-hero.png')}
						className="h-52 w-full"
						resizeMode="contain"
						accessibilityLabel="Pixel art street scene with a scooter and a street sign"
					/>
				</View>

				<View className="gap-6 px-6 pt-8">
					<View className="gap-1">
						<Text className="text-3xl text-goojol-road" style={{ fontFamily: 'SpaceMono' }}>
							Go-Ojol
						</Text>
						<Text className="text-base text-goojol-muted">Your ride, two taps away</Text>
					</View>

					<View className="gap-4">
						<View className="gap-2">
							<Text className="font-medium text-goojol-road text-sm">Email</Text>
							<View className="relative">
								<TextInput
									className="rounded-xl border border-goojol-border bg-goojol-surface py-3.5 pr-4 pl-11 text-base text-goojol-road"
									value={email}
									onChangeText={setEmail}
									keyboardType="email-address"
									autoCapitalize="none"
									autoComplete="email"
									textContentType="emailAddress"
									placeholder="you@example.com"
									placeholderTextColor="#8892A8"
									accessibilityLabel="Email"
								/>
								<View className="absolute top-0 bottom-0 left-3.5 justify-center">
									<Mail size={18} color="#8892A8" />
								</View>
							</View>
						</View>

						<View className="gap-2">
							<Text className="font-medium text-goojol-road text-sm">Password</Text>
							<View className="relative">
								<TextInput
									className="rounded-xl border border-goojol-border bg-goojol-surface py-3.5 pr-4 pl-11 text-base text-goojol-road"
									value={password}
									onChangeText={setPassword}
									secureTextEntry
									autoComplete="password"
									textContentType="password"
									placeholder="Enter your password"
									placeholderTextColor="#8892A8"
									returnKeyType="done"
									accessibilityLabel="Password"
								/>
								<View className="absolute top-0 bottom-0 left-3.5 justify-center">
									<Lock size={18} color="#8892A8" />
								</View>
							</View>
						</View>
					</View>

					<Pressable
						className={`items-center rounded-xl bg-goojol-coral py-4 ${canSignIn ? 'opacity-100' : 'opacity-50'}`}
						disabled={!canSignIn}
						onPress={() => { }}
						accessibilityLabel="Sign in"
						accessibilityState={{ disabled: !canSignIn }}
					>
						<Text className="font-semibold text-base text-white">Sign in</Text>
					</Pressable>

					<View className="items-center gap-1 pt-2">
						<Text className="text-goojol-muted text-sm">Don't have an account?</Text>
						<Pressable accessibilityLabel="Create account">
							<Text className="font-semibold text-goojol-teal text-sm">Create one</Text>
						</Pressable>
					</View>
				</View>
			</ScrollView>
		</KeyboardAvoidingView>
	);
}
