import {
  createFileRoute,
  Link,
  redirect,
  useRouteContext,
  useRouter,
} from "@tanstack/react-router";
import { Outlet } from "@tanstack/react-router";

export const Route = createFileRoute("/_authenticated")({
  component: RouteComponent,
  beforeLoad: ({ context }) => {
    const { isAuthenticated, logout } = context.authentication;
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
  const { authentication } = useRouteContext({ from: "/_authenticated" });

  return (
    <>
      <div className="p-2 flex gap-2">
        <Link to="/" className="[&.active]:font-bold">
          Home
        </Link>{" "}
        <Link to="/about" className="[&.active]:font-bold">
          About
        </Link>
        <button onClick={authentication.logout}>Log Out</button>
      </div>
      <hr />
      <Outlet />;
    </>
  );
}
