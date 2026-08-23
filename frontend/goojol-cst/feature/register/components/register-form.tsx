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
import { useRegisterMutation } from '../register.mutation';
import { type RegisterFormSchema, registerFormSchema } from '../register.schema';

export default function RegisterForm() {
  const [showPassword, setShowPassword] = useState(false);
  const [showConfirmPassword, setShowConfirmPassword] = useState(false);
  const registerMutation = useRegisterMutation();

  const {
    control,
    handleSubmit,
    formState: { errors, isValid },
  } = useForm<RegisterFormSchema>({
    resolver: zodResolver(registerFormSchema),
    mode: 'onChange',
    defaultValues: {
      name: '',
      phone_number: '',
      email: '',
      password: '',
      confirmPassword: '',
    },
  });

  const onSubmit = handleSubmit((data) => {
    registerMutation.mutate({
      email: data.email,
      password: data.password,
      role: 'customer',
      name: data.name,
      phone_number: data.phone_number,
    });
  });

  return (
    <VStack space="lg" className="w-full">
      <Controller
        control={control}
        name="name"
        render={({ field: { onChange, onBlur, value } }) => (
          <FormControl isInvalid={!!errors.name} className="w-full">
            <FormControlLabel>
              <FormControlLabelText className="text-goojol-road text-sm">Name</FormControlLabelText>
            </FormControlLabel>
            <Input className="border-goojol-border bg-goojol-surface">
              <InputField
                type="text"
                placeholder="Your name"
                placeholderTextColor="#8892A8"
                className="text-goojol-road"
                value={value}
                onChangeText={onChange}
                onBlur={onBlur}
                autoCapitalize="words"
                autoComplete="name"
                textContentType="name"
                accessibilityLabel="Name"
              />
            </Input>
            {errors.name ? (
              <FormControlError>
                <FormControlErrorIcon as={AlertCircleIcon} />
                <FormControlErrorText>{errors.name.message}</FormControlErrorText>
              </FormControlError>
            ) : null}
          </FormControl>
        )}
      />

      <Controller
        control={control}
        name="phone_number"
        render={({ field: { onChange, onBlur, value } }) => (
          <FormControl isInvalid={!!errors.phone_number} className="w-full">
            <FormControlLabel>
              <FormControlLabelText className="text-goojol-road text-sm">
                Phone number
              </FormControlLabelText>
            </FormControlLabel>
            <Input className="border-goojol-border bg-goojol-surface">
              <InputField
                type="text"
                placeholder="+62 812 3456 7890"
                placeholderTextColor="#8892A8"
                className="text-goojol-road"
                value={value}
                onChangeText={onChange}
                onBlur={onBlur}
                keyboardType="phone-pad"
                autoComplete="tel"
                textContentType="telephoneNumber"
                accessibilityLabel="Phone number"
              />
            </Input>
            {errors.phone_number ? (
              <FormControlError>
                <FormControlErrorIcon as={AlertCircleIcon} />
                <FormControlErrorText>{errors.phone_number.message}</FormControlErrorText>
              </FormControlError>
            ) : null}
          </FormControl>
        )}
      />

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
                placeholder="At least 8 characters"
                placeholderTextColor="#8892A8"
                className="text-goojol-road"
                value={value}
                onChangeText={onChange}
                onBlur={onBlur}
                autoComplete="new-password"
                textContentType="newPassword"
                accessibilityLabel="Password"
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

      <Controller
        control={control}
        name="confirmPassword"
        render={({ field: { onChange, onBlur, value } }) => (
          <FormControl isInvalid={!!errors.confirmPassword} className="w-full">
            <FormControlLabel>
              <FormControlLabelText className="text-goojol-road text-sm">
                Confirm password
              </FormControlLabelText>
            </FormControlLabel>
            <Input className="border-goojol-border bg-goojol-surface">
              <InputField
                type={showConfirmPassword ? 'text' : 'password'}
                placeholder="Re-enter your password"
                placeholderTextColor="#8892A8"
                className="text-goojol-road"
                value={value}
                onChangeText={onChange}
                onBlur={onBlur}
                autoComplete="new-password"
                textContentType="newPassword"
                returnKeyType="done"
                accessibilityLabel="Confirm password"
                onSubmitEditing={onSubmit}
              />
              <InputSlot className="pr-1" onPress={() => setShowConfirmPassword((prev) => !prev)}>
                <InputIcon
                  as={showConfirmPassword ? EyeIcon : EyeOffIcon}
                  className="text-goojol-muted"
                />
              </InputSlot>
            </Input>
            {errors.confirmPassword ? (
              <FormControlError>
                <FormControlErrorIcon as={AlertCircleIcon} />
                <FormControlErrorText>{errors.confirmPassword.message}</FormControlErrorText>
              </FormControlError>
            ) : null}
          </FormControl>
        )}
      />

      <Button
        className="w-full bg-goojol-coral data-[active=true]:bg-goojol-coral/90 data-[hover=true]:bg-goojol-coral/90"
        onPress={onSubmit}
        isDisabled={!isValid || registerMutation.isPending}
        accessibilityLabel="Create account"
      >
        {registerMutation.isPending ? <ButtonSpinner /> : null}
        <ButtonText className="font-semibold text-white">Create account</ButtonText>
      </Button>

      <VStack className="items-center gap-1 pt-2">
        <Text size="sm" className="text-goojol-muted">
          Already have an account?
        </Text>
        <Link href="/(public)/login" accessibilityLabel="Sign in" replace>
          <LinkText size="sm" className="font-semibold text-goojol-teal">
            Sign in
          </LinkText>
        </Link>
      </VStack>
    </VStack>
  );
}
