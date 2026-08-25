import { zodResolver } from '@hookform/resolvers/zod';
import { Link } from 'expo-router';
import { useState } from 'react';
import { Controller, useForm } from 'react-hook-form';
import { Button, ButtonSpinner, ButtonText } from '@/components/ui/button';
import {
  FormControl,
  FormControlError,
  FormControlErrorIcon,
  FormControlErrorText,
  FormControlLabel,
  FormControlLabelText,
} from '@/components/ui/form-control';
import { AlertCircleIcon, EyeIcon, EyeOffIcon } from '@/components/ui/icon';
import { Input, InputField, InputIcon, InputSlot } from '@/components/ui/input';
import { LinkText } from '@/components/ui/link';
import { Text } from '@/components/ui/text';
import { VStack } from '@/components/ui/vstack';
import { useLoginMutation } from '../login.mutation';
import { type LoginSchema, loginSchema } from '../login.schema';

export default function LoginForm() {
  const [showPassword, setShowPassword] = useState(false);
  const loginMutation = useLoginMutation();

  const {
    control,
    handleSubmit,
    formState: { errors, isValid },
  } = useForm<LoginSchema>({
    resolver: zodResolver(loginSchema),
    mode: 'onChange',
    defaultValues: {
      email: '',
      password: '',
    },
  });

  const onSubmit = handleSubmit((data) => {
    loginMutation.mutate(data);
  });

  return (
    <VStack space="lg" className="w-full">
      <Controller
        control={control}
        name="email"
        render={({ field: { onChange, onBlur, value } }) => (
          <FormControl isInvalid={!!errors.email} className="w-full">
            <FormControlLabel>
              <FormControlLabelText className="text-goojol-road text-sm">
                Email
              </FormControlLabelText>
            </FormControlLabel>
            <Input className="border-goojol-border bg-goojol-surface">
              <InputField
                type="text"
                placeholder="you@example.com"
                placeholderTextColor="#8892A8"
                className="text-goojol-road"
                value={value}
                onChangeText={onChange}
                onBlur={onBlur}
                keyboardType="email-address"
                autoCapitalize="none"
                autoComplete="email"
                textContentType="emailAddress"
                accessibilityLabel="Email"
              />
            </Input>
            {errors.email ? (
              <FormControlError>
                <FormControlErrorIcon as={AlertCircleIcon} />
                <FormControlErrorText>{errors.email.message}</FormControlErrorText>
              </FormControlError>
            ) : null}
          </FormControl>
        )}
      />

      <Controller
        control={control}
        name="password"
        render={({ field: { onChange, onBlur, value } }) => (
          <FormControl isInvalid={!!errors.password} className="w-full">
            <FormControlLabel>
              <FormControlLabelText className="text-goojol-road text-sm">
                Password
              </FormControlLabelText>
            </FormControlLabel>
            <Input className="border-goojol-border bg-goojol-surface">
              <InputField
                type={showPassword ? 'text' : 'password'}
                placeholder="Enter your password"
                placeholderTextColor="#8892A8"
                className="text-goojol-road"
                value={value}
                onChangeText={onChange}
                onBlur={onBlur}
                autoComplete="password"
                textContentType="password"
                returnKeyType="done"
                accessibilityLabel="Password"
                onSubmitEditing={onSubmit}
              />
              <InputSlot className="pr-1" onPress={() => setShowPassword((prev) => !prev)}>
                <InputIcon as={showPassword ? EyeIcon : EyeOffIcon} className="text-goojol-muted" />
              </InputSlot>
            </Input>
            {errors.password ? (
              <FormControlError>
                <FormControlErrorIcon as={AlertCircleIcon} />
                <FormControlErrorText>{errors.password.message}</FormControlErrorText>
              </FormControlError>
            ) : null}
          </FormControl>
        )}
      />

      <Button
        className="w-full bg-goojol-coral data-[active=true]:bg-goojol-coral/90 data-[hover=true]:bg-goojol-coral/90"
        onPress={onSubmit}
        isDisabled={!isValid || loginMutation.isPending}
        accessibilityLabel="Sign in"
      >
        {loginMutation.isPending ? <ButtonSpinner /> : null}
        <ButtonText className="font-semibold text-white">Sign in</ButtonText>
      </Button>

      <VStack className="items-center gap-1 pt-2">
        <Text size="sm" className="text-goojol-muted">
          Don't have an account?
        </Text>
        <Link href="/(public)/register" accessibilityLabel="Create account">
          <LinkText size="sm" className="font-semibold text-goojol-teal">
            Create one
          </LinkText>
        </Link>
      </VStack>
    </VStack>
  );
}
