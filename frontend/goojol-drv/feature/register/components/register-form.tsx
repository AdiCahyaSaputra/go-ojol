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
import { HStack } from '@/components/ui/hstack';
import { AlertCircleIcon, EyeIcon, EyeOffIcon } from '@/components/ui/icon';
import { Input, InputField, InputIcon, InputSlot } from '@/components/ui/input';
import { LinkText } from '@/components/ui/link';
import { Text } from '@/components/ui/text';
import { VStack } from '@/components/ui/vstack';
import { useRegisterMutation } from '../register.mutation';
import {
  type RegisterFormSchema,
  registerFormSchema,
  registerStepOneSchema,
} from '../register.schema';
import RegisterStepIndicator from './register-step-indicator';

export default function RegisterForm() {
  const [step, setStep] = useState<1 | 2>(1);
  const [showPassword, setShowPassword] = useState(false);
  const [showConfirmPassword, setShowConfirmPassword] = useState(false);
  const registerMutation = useRegisterMutation();

  const {
    control,
    handleSubmit,
    trigger,
    getValues,
    formState: { errors },
  } = useForm<RegisterFormSchema>({
    resolver: zodResolver(registerFormSchema),
    mode: 'onChange',
    defaultValues: {
      name: '',
      phone_number: '',
      email: '',
      vehicle_name: '',
      vehicle_license_number: '',
      vehicle_max_size: 1,
      vehicle_type: 'motorcycle',
      password: '',
      confirmPassword: '',
    },
  });

  const goNext = async () => {
    const values = getValues();
    const parsed = registerStepOneSchema.safeParse({
      name: values.name,
      phone_number: values.phone_number,
      email: values.email,
    });
    if (!parsed.success) {
      await trigger(['name', 'phone_number', 'email']);
      return;
    }
    setStep(2);
  };

  const onSubmit = handleSubmit((data) => {
    registerMutation.mutate({
      email: data.email,
      password: data.password,
      role: 'driver',
      name: data.name,
      phone_number: data.phone_number,
      vehicle_name: data.vehicle_name,
      vehicle_license_number: data.vehicle_license_number,
      vehicle_max_size: data.vehicle_max_size,
      vehicle_type: data.vehicle_type,
    });
  });

  return (
    <VStack space="lg" className="w-full">
      <RegisterStepIndicator step={step} />

      {step === 1 ? (
        <>
          <Controller
            control={control}
            name="name"
            render={({ field: { onChange, onBlur, value } }) => (
              <FormControl isInvalid={!!errors.name} className="w-full">
                <FormControlLabel>
                  <FormControlLabelText className="text-goojol-road text-sm">
                    Name
                  </FormControlLabelText>
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
                    Phone
                  </FormControlLabelText>
                </FormControlLabel>
                <Input className="border-goojol-border bg-goojol-surface">
                  <InputField
                    type="text"
                    placeholder="6281234567890"
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
                    <InputIcon
                      as={showPassword ? EyeIcon : EyeOffIcon}
                      className="text-goojol-muted"
                    />
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
                    Confirm Password
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
                  <InputSlot
                    className="pr-1"
                    onPress={() => setShowConfirmPassword((prev) => !prev)}
                  >
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
            onPress={goNext}
            accessibilityLabel="Continue to vehicle details"
          >
            <ButtonText className="font-semibold text-white">Continue</ButtonText>
          </Button>
        </>
      ) : (
        <>
          <Controller
            control={control}
            name="vehicle_type"
            render={({ field: { onChange, value } }) => (
              <FormControl isInvalid={!!errors.vehicle_type} className="w-full">
                <FormControlLabel>
                  <FormControlLabelText className="text-goojol-road text-sm">
                    Vehicle Type
                  </FormControlLabelText>
                </FormControlLabel>
                <HStack space="md">
                  {(
                    [
                      { value: 'motorcycle', label: 'Motorcycle' },
                      { value: 'car', label: 'Car' },
                    ] as const
                  ).map((option) => {
                    const selected = value === option.value;
                    return (
                      <Button
                        key={option.value}
                        variant="outline"
                        className={`flex-1 ${
                          selected
                            ? 'border-goojol-coral bg-goojol-coral/15 data-[active=true]:bg-goojol-coral/20'
                            : 'border-goojol-border bg-goojol-surface data-[active=true]:bg-goojol-surface'
                        }`}
                        onPress={() => onChange(option.value)}
                        accessibilityLabel={option.label}
                        accessibilityState={{ selected }}
                      >
                        <ButtonText
                          className={`font-semibold text-sm ${
                            selected ? 'text-goojol-coral' : 'text-goojol-muted'
                          }`}
                        >
                          {option.label}
                        </ButtonText>
                      </Button>
                    );
                  })}
                </HStack>
                {errors.vehicle_type ? (
                  <FormControlError>
                    <FormControlErrorIcon as={AlertCircleIcon} />
                    <FormControlErrorText>{errors.vehicle_type.message}</FormControlErrorText>
                  </FormControlError>
                ) : null}
              </FormControl>
            )}
          />

          <Controller
            control={control}
            name="vehicle_name"
            render={({ field: { onChange, onBlur, value } }) => (
              <FormControl isInvalid={!!errors.vehicle_name} className="w-full">
                <FormControlLabel>
                  <FormControlLabelText className="text-goojol-road text-sm">
                    Vehicle Name
                  </FormControlLabelText>
                </FormControlLabel>
                <Input className="border-goojol-border bg-goojol-surface">
                  <InputField
                    type="text"
                    placeholder="Honda Beat"
                    placeholderTextColor="#8892A8"
                    className="text-goojol-road"
                    value={value}
                    onChangeText={onChange}
                    onBlur={onBlur}
                    autoCapitalize="words"
                    accessibilityLabel="Vehicle name"
                  />
                </Input>
                {errors.vehicle_name ? (
                  <FormControlError>
                    <FormControlErrorIcon as={AlertCircleIcon} />
                    <FormControlErrorText>{errors.vehicle_name.message}</FormControlErrorText>
                  </FormControlError>
                ) : null}
              </FormControl>
            )}
          />

          <Controller
            control={control}
            name="vehicle_license_number"
            render={({ field: { onChange, onBlur, value } }) => (
              <FormControl isInvalid={!!errors.vehicle_license_number} className="w-full">
                <FormControlLabel>
                  <FormControlLabelText className="text-goojol-road text-sm">
                    License Plate
                  </FormControlLabelText>
                </FormControlLabel>
                <Input className="border-goojol-border bg-goojol-surface">
                  <InputField
                    type="text"
                    placeholder="B 1234 XYZ"
                    placeholderTextColor="#8892A8"
                    className="text-goojol-road"
                    value={value}
                    onChangeText={onChange}
                    onBlur={onBlur}
                    autoCapitalize="characters"
                    accessibilityLabel="License plate"
                  />
                </Input>
                {errors.vehicle_license_number ? (
                  <FormControlError>
                    <FormControlErrorIcon as={AlertCircleIcon} />
                    <FormControlErrorText>
                      {errors.vehicle_license_number.message}
                    </FormControlErrorText>
                  </FormControlError>
                ) : null}
              </FormControl>
            )}
          />

          <Controller
            control={control}
            name="vehicle_max_size"
            render={({ field: { onChange, onBlur, value } }) => (
              <FormControl isInvalid={!!errors.vehicle_max_size} className="w-full">
                <FormControlLabel>
                  <FormControlLabelText className="text-goojol-road text-sm">
                    Seat Capacity
                  </FormControlLabelText>
                </FormControlLabel>
                <Input className="border-goojol-border bg-goojol-surface">
                  <InputField
                    type="text"
                    placeholder="1"
                    placeholderTextColor="#8892A8"
                    className="text-goojol-road"
                    value={value ? String(value) : ''}
                    onChangeText={(text) => {
                      const digits = text.replace(/[^0-9]/g, '');
                      onChange(digits ? Number(digits) : 0);
                    }}
                    onBlur={onBlur}
                    keyboardType="number-pad"
                    accessibilityLabel="Seat capacity"
                  />
                </Input>
                {errors.vehicle_max_size ? (
                  <FormControlError>
                    <FormControlErrorIcon as={AlertCircleIcon} />
                    <FormControlErrorText>{errors.vehicle_max_size.message}</FormControlErrorText>
                  </FormControlError>
                ) : null}
              </FormControl>
            )}
          />

          {registerMutation.isError ? (
            <Text size="sm" className="text-center text-goojol-coral">
              {registerMutation.error instanceof Error
                ? registerMutation.error.message
                : 'Could not create account. Try again.'}
            </Text>
          ) : null}

          <HStack space="md" className="mt-16">
            <Button
              variant="outline"
              className="border-goojol-border bg-goojol-surface data-[active=true]:bg-goojol-surface"
              onPress={() => setStep(1)}
              accessibilityLabel="Back to about you"
            >
              <ButtonText className="font-semibold text-goojol-road">Back</ButtonText>
            </Button>
            <Button
              className="flex-1 bg-goojol-coral data-[active=true]:bg-goojol-coral/90 data-[hover=true]:bg-goojol-coral/90"
              onPress={onSubmit}
              isDisabled={registerMutation.isPending}
              accessibilityLabel="Create account"
            >
              {registerMutation.isPending ? <ButtonSpinner /> : null}
              <ButtonText className="font-semibold text-white">Create account</ButtonText>
            </Button>
          </HStack>
        </>
      )}

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
