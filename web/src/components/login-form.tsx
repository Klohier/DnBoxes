import { useAuth } from "@/AuthContext";
import { useNavigate, useRouter, useRouterState } from "@tanstack/react-router";
import { AuthForm } from "./AuthForm";
import { LoginCredentials } from "@/types/auth";

export function LoginForm() {
  const { login } = useAuth();
  const router = useRouter();
  const isLoading = useRouterState({ select: (s) => s.isLoading });
  const navigate = useNavigate();

  async function onSubmit(values: LoginCredentials) {
    await login(values);
    await router.invalidate();
    await navigate({ to: "/" });
  }

  return (
    <AuthForm
      onSubmitHandler={onSubmit}
      isLoading={isLoading}
      buttonText="Login"
    />
  );
}
