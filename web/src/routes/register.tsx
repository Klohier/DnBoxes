import { useAuth } from "@/AuthContext";
import { AuthForm } from "@/components/AuthForm";
import {
  createFileRoute,
  useNavigate,
  useRouter,
  useRouterState,
} from "@tanstack/react-router";

export const Route = createFileRoute("/register")({
  component: RouteComponent,
});

function RouteComponent() {
  const { register } = useAuth();
  const router = useRouter();
  const navigate = useNavigate();
  const isLoading = useRouterState({ select: (s) => s.isLoading });

  async function handleRegister(values: {
    username: string;
    password: string;
  }) {
    await register(values);
    await router.invalidate();
    await navigate({ to: "/" });
  }
  return (
    <div className="flex min-h-svh w-full items-center justify-center p-6 md:p-10">
      <div className="w-full max-w-sm">
        <AuthForm
          onSubmitHandler={handleRegister}
          isLoading={isLoading}
          buttonText="Register"
          formType="register"
        />
      </div>
    </div>
  );
}
