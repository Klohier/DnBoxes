import { useAuth } from "@/AuthContext";
import { AuthForm } from "@/components/AuthForm";
import { LoginCredentials } from "@/types/auth";
import {
  createFileRoute,
  redirect,
  useRouter,
  useRouterState,
  Link,
} from "@tanstack/react-router";
import z from "zod";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";

const fallback = "/" as const;

export const Route = createFileRoute("/login")({
  validateSearch: z.object({
    redirect: z.string().optional().catch(""),
  }),
  beforeLoad: ({ context, search }) => {
    if (context.authentication.isAuthenticated) {
      redirect({ to: search.redirect ?? fallback, throw: true });
    }
  },
  component: Login,
});

function Login() {
  const { login } = useAuth();
  const search = Route.useSearch();
  const router = useRouter();
  const isLoading = useRouterState({ select: (s) => s.isLoading });
  const navigate = Route.useNavigate();

  async function onSubmit(values: LoginCredentials) {
    await login(values);
    await router.invalidate();
    await navigate({ to: search.redirect ?? fallback });
  }

  return (
    <div className="min-h-screen bg-gray-900 flex items-center justify-center p-4">
      <div className="w-full max-w-md">
        <Card className="bg-gray-800 border-gray-700">
          <CardHeader className="text-center">
            <CardTitle className="text-3xl font-bold text-white mb-2">
              Welcome Back
            </CardTitle>
            <CardDescription className="text-gray-400">
              Sign in to continue playing Dots & Boxes
            </CardDescription>
          </CardHeader>
          <CardContent>
            <AuthForm
              onSubmitHandler={onSubmit}
              isLoading={isLoading}
              buttonText="Login"
              formType="login"
            />
          </CardContent>
        </Card>

        <p className="text-center mt-4 text-gray-400 text-sm">
          Don't have an account?{" "}
          <Link
            to="/register"
            className="text-blue-400 hover:text-blue-300 font-medium transition-colors"
          >
            Sign up
          </Link>
        </p>
      </div>
    </div>
  );
}
