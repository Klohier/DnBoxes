import { useAuth } from "@/AuthContext";
import { AuthForm } from "@/components/AuthForm";
import { LoginCredentials } from "@/types/auth";
import {
  createFileRoute,
  redirect,
  useRouter,
  useRouterState,
} from "@tanstack/react-router";
import z from "zod";

const fallback = "/" as const;

export const Route = createFileRoute("/register")({
  validateSearch: z.object({
    redirect: z.string().optional().catch(""),
  }),
  beforeLoad: ({ context, search }) => {
    if (context.authentication.isAuthenticated) {
      redirect({ to: search.redirect ?? fallback, throw: true });
    }
  },
  component: Register,
});

function Register() {
  const { register } = useAuth();
  const search = Route.useSearch();
  const router = useRouter();
  const isLoading = useRouterState({ select: (s) => s.isLoading });
  const navigate = Route.useNavigate();

  async function onSubmit(values: LoginCredentials) {
    await register(values);
    await router.invalidate();
    await navigate({ to: search.redirect ?? fallback });
  }

  return (
    <div className="flex min-h-svh w-full items-center justify-center p-6 md:p-10">
      <div className="w-full max-w-sm">
        <AuthForm
          onSubmitHandler={onSubmit}
          isLoading={isLoading}
          buttonText="Register"
          formType="register"
        />
      </div>
    </div>
  );
}
