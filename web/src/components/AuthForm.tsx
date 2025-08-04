/* eslint-disable @typescript-eslint/no-misused-promises */
//TODO: Figure out how to correct on submit for login
// components/AuthForm.tsx
import { z } from "zod";
import { zodResolver } from "@hookform/resolvers/zod";
import { useForm } from "react-hook-form";
import React, { useState } from "react";
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
import { Button } from "@/components/ui/button";

const formSchema = z.object({
  username: z
    .string()
    .min(2, { message: "Username must have at least 2 characters" })
    .max(50),
  password: z.string().min(5, {
    message: "Password must be at least 5 characters.",
  }),
});

interface AuthFormProps {
  onSubmitHandler: (values: z.infer<typeof formSchema>) => Promise<void>;
  isLoading?: boolean;
  buttonText?: string;
}

export function AuthForm({
  onSubmitHandler,
  isLoading = false,
  buttonText = "Submit",
}: AuthFormProps) {
  const [isSubmitting, setIsSubmitting] = useState(false);
  const form = useForm<z.infer<typeof formSchema>>({
    resolver: zodResolver(formSchema),
    defaultValues: {
      username: "",
      password: "",
    },
  });

  function getErrorMessage(err: unknown): string {
    return typeof err === "object" &&
      err !== null &&
      "message" in err &&
      typeof err.message === "string"
      ? err.message
      : "Something went wrong";
  }

  const isProcessing = isLoading || isSubmitting;

  async function onSubmit(values: z.infer<typeof formSchema>) {
    setIsSubmitting(true);
    try {
      await onSubmitHandler(values);
    } catch (err: unknown) {
      const message = getErrorMessage(err);
      form.setError("username", {
        type: "manual",
        message,
      });

      form.setError("password", {
        type: "manual",
        message,
      });
    } finally {
      setIsSubmitting(false);
    }
  }

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
        <Button type="submit">
          {isProcessing ? "Loading..." : buttonText}
        </Button>
      </form>
    </Form>
  );
}
