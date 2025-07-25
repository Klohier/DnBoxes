import { createFileRoute, redirect } from "@tanstack/react-router";

export const Route = createFileRoute("/about")({
  beforeLoad: async ({ context }) => {
    const { isAuthenticated } = context.authentication;

    await context.queryClient.invalidateQueries({ queryKey: ["me"] });

    if (!isAuthenticated) {
      // eslint-disable-next-line @typescript-eslint/only-throw-error
      throw redirect({
        to: "/login",
      });
    }
  },
  component: About,
});

function About() {
  return <div className="p-2">Hello from About!</div>;
}
