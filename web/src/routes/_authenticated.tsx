// import { Button } from "@/components/ui/button";
import {
  createFileRoute,
  //   Link,
  redirect,
  //   useRouteContext,
} from "@tanstack/react-router";
import { Outlet } from "@tanstack/react-router";

export const Route = createFileRoute("/_authenticated")({
  component: RouteComponent,
  beforeLoad: ({ context }) => {
    const { isAuthenticated } = context.authentication;
    // await context.queryClient.invalidateQueries({ queryKey: ["me"] });

    if (!isAuthenticated) {
      // eslint-disable-next-line @typescript-eslint/only-throw-error
      throw redirect({
        to: "/login",
      });
    }
  },
});

function RouteComponent() {
  //   const { authentication } = useRouteContext({ from: "/_authenticated" });

  return <Outlet />;
}
