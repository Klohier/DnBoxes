import { createFileRoute, redirect } from "@tanstack/react-router";
import { z } from "zod";
import { Outlet } from "@tanstack/react-router";
import { fetchUser } from "@/api/fetchUser";

export const Route = createFileRoute("/_authenticated")({
  component: RouteComponent,
  validateSearch: z.object({
    redirect: z.string().optional().catch(""),
  }),
  beforeLoad: async ({ context, location }) => {
    let user;
    try {
      user = await context.queryClient.ensureQueryData({
        queryKey: ["me"],
        queryFn: fetchUser,
      });
    } catch {
      user = null;
    }

    if (!user) {
      redirect({
        to: "/login",
        search: {
          redirect: location.href,
        },
        throw: true,
      });
    }
  },
});

function RouteComponent() {
  return <Outlet />;
}
