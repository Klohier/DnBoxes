import { createFileRoute, redirect } from "@tanstack/react-router";
import { z } from "zod";
import { Outlet } from "@tanstack/react-router";

export const Route = createFileRoute("/_authenticated")({
  component: RouteComponent,
  validateSearch: z.object({
    redirect: z.string().optional().catch(""),
  }),
  beforeLoad: ({ context, location }) => {
    if (!context.authentication.isAuthenticated) {
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
