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
import { fetchUser } from "@/api/fetchUser";

const fallback = "/" as const;

export const Route = createFileRoute("/register")({
  validateSearch: z.object({
    redirect: z.string().optional().catch(""),
  }),
  beforeLoad: async ({ context, search }) => {
    let user;
    try {
      user = await context.queryClient.ensureQueryData({
        queryKey: ["me"],
        queryFn: fetchUser,
      });
    } catch {
      user = null;
    }

    if (user) {
      redirect({ to: search.redirect ?? fallback, throw: true });
    }
  },
  component: Register,
  head: () => ({
    meta: [
      {
        title: "Create Account - Dots & Boxes Online",
      },
      {
        name: "description",
        content:
          "Create a free Dots & Boxes account. Play multiplayer games, track your stats, and climb the leaderboard.",
      },
      {
        name: "robots",
        content: "noindex, nofollow",
      },
    ],
    links: [
      {
        rel: "canonical",
        href: "https://dotsandboxesonline.com/register",
      },
    ],
  }),
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
    <div className="min-h-screen bg-gray-900 flex items-center justify-center p-4">
      <div className="w-full max-w-md">
        <Card className="bg-gray-800 border-gray-700">
          <CardHeader className="text-center">
            <CardTitle className="text-3xl font-bold text-white mb-2">
              Create Account
            </CardTitle>
            <CardDescription className="text-gray-400">
              Join the Dots & Boxes community
            </CardDescription>
          </CardHeader>
          <CardContent>
            <AuthForm
              onSubmitHandler={onSubmit}
              isLoading={isLoading}
              buttonText="Sign Up"
              formType="register"
            />
          </CardContent>
        </Card>

        <p className="text-center mt-4 text-gray-400 text-sm">
          Already have an account?{" "}
          <Link
            to="/login"
            className="text-blue-400 hover:text-blue-300 font-medium transition-colors"
          >
            Sign in
          </Link>
        </p>
      </div>
    </div>
  );
}
