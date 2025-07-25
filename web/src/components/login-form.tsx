// import { cn } from "@/lib/utils";
import { z } from "zod";
import { zodResolver } from "@hookform/resolvers/zod";
import { useForm } from "react-hook-form";
import React, { useState } from "react";
// import { useAuth } from "../hooks/useAuth";
import { useAuth } from "@/AuthContext";
import { Button } from "@/components/ui/button";
import {
  Route,
  useNavigate,
  useRouter,
  useRouterState,
  useSearch,
} from "@tanstack/react-router";
// import { useQueryClient } from "@tanstack/react-query";

import {
  Form,
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "@/components/ui/form";

import { Input } from "@/components/ui/input";
// import { Label } from "@/components/ui/label";

const fallback = "/" as const;

const formSchema = z.object({
  username: z
    .string()
    .min(2, { message: "Username must have at least 2 characters" })
    .max(50),
  password: z.string().min(5, {
    message: "Password must be at least 5 characters.",
  }),
});

export function LoginForm() {
  const { login } = useAuth();
  const router = useRouter();
  const isLoading = useRouterState({ select: (s) => s.isLoading });
  const navigate = useNavigate();
  const [isSubmitting, setIsSubmitting] = useState(false);

  const search = useSearch({ from: "/login" });

  const form = useForm<z.infer<typeof formSchema>>({
    resolver: zodResolver(formSchema),
    defaultValues: {
      username: "",
      password: "",
    },
  });

  // const queryClient = useQueryClient();

  async function onSubmit(values: z.infer<typeof formSchema>) {
    setIsSubmitting(true);
    try {
      await login(values);
      await router.invalidate();
      // await queryClient.ensureQueryData({ queryKey: ["me"] });
      await navigate({ to: search.redirect ?? fallback });
    } catch (err: any) {
      form.setError("username", {
        type: "manual",
        message: err.message ?? "Invalid username or password",
      });

      form.setError("password", {
        type: "manual",
        message: err.message ?? "Invalid username or password",
      });
    } finally {
      setIsSubmitting(false);
    }
  }

  const isLoggingIn = isLoading || isSubmitting;

  return (
    <Form {...form}>
      <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-8">
        <FormField
          control={form.control}
          name="username"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Username</FormLabel>
              <FormControl>
                <Input placeholder="Username" {...field} />
              </FormControl>
              <FormDescription>
                This is your public display name.
              </FormDescription>
              <FormMessage />
            </FormItem>
          )}
        />

        <FormField
          control={form.control}
          name="password"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Password</FormLabel>
              <FormControl>
                <Input type="password" placeholder="••••••••" {...field} />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />
        <Button type="submit">{isLoggingIn ? "Loading..." : "Login"}</Button>
      </form>
    </Form>
  );
}
