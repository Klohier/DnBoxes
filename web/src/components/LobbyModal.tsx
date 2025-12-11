import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";

import { z } from "zod";
import { zodResolver } from "@hookform/resolvers/zod";
import { useForm } from "react-hook-form";

const LobbySchema = z.object({
  name: z.string().trim().min(1, "Lobby name is required"),
  player_limit: z
    .number()
    .min(2, "Minimum 2 players")
    .max(10, "Maximum 10 players"),
  is_private: z.boolean(),
});

type LobbyFormValues = z.infer<typeof LobbySchema>;

interface LobbyModalProps {
  open: boolean;
  onClose: () => void;
  onSubmit: (values: LobbyFormValues) => Promise<void>;
}

export function LobbyModal({ open, onClose, onSubmit }: LobbyModalProps) {
  const form = useForm<LobbyFormValues>({
    resolver: zodResolver(LobbySchema),
    defaultValues: {
      name: "",
      player_limit: 4,
      is_private: false,
    },
  });

  return (
    <Dialog open={open} onOpenChange={onClose}>
      <DialogContent className="max-w-md">
        <DialogHeader>
          <DialogTitle>Create a Lobby</DialogTitle>
        </DialogHeader>

        <form className="space-y-6" onSubmit={form.handleSubmit(onSubmit)}>
          {/* Lobby Name */}
          <div className="flex flex-col gap-2">
            <Label htmlFor="name">Lobby Name</Label>
            <Input
              id="name"
              {...form.register("name")}
              placeholder="My Awesome Lobby"
            />
            {form.formState.errors.name && (
              <p className="text-red-500 text-sm">
                {form.formState.errors.name.message}
              </p>
            )}
          </div>

          {/* Player Limit */}
          <div className="flex flex-col gap-2">
            <Label htmlFor="player_limit">Player Limit</Label>
            <Input
              id="player_limit"
              type="number"
              {...form.register("player_limit", { valueAsNumber: true })}
            />
            {form.formState.errors.player_limit && (
              <p className="text-red-500 text-sm">
                {form.formState.errors.player_limit.message}
              </p>
            )}
          </div>

          {/* Private Toggle */}
          <div className="flex items-center justify-between">
            <Label htmlFor="is_private">Private Lobby</Label>
          </div>

          <DialogFooter>
            <Button type="button" variant="outline" onClick={onClose}>
              Cancel
            </Button>
            <Button type="submit">Create Lobby</Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
