import { z } from "zod";
import { zodResolver } from "@hookform/resolvers/zod";
import { useForm } from "react-hook-form";
import React, { useState } from "react";
import axios from "axios";
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
  formType: "login" | "register";
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
    if (axios.isAxiosError(err) && typeof err.response?.data?.message === "string") {
      return err.response.data.message;
    }
    if (err instanceof Error) {
      return err.message;
    }
    return "Something went wrong";
  }

  const isProcessing = isLoading || isSubmitting;

  async function onSubmit(values: z.infer<typeof formSchema>) {
    setIsSubmitting(true);
    try {
      await onSubmitHandler(values);
    } catch (err: unknown) {
      const message = getErrorMessage(err);
      form.setError("root", {
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

        {form.formState.errors.root && (
          <p className="text-sm font-medium text-destructive">
            {form.formState.errors.root.message}
          </p>
        )}

        <Button type="submit">
          {isProcessing ? "Loading..." : buttonText}
        </Button>
      </form>
    </Form>
  );
}
