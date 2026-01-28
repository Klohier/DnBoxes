// @/components/BotGameModal.tsx
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import * as z from "zod";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  Form,
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "@/components/ui/form";
import { s } from "node_modules/vite/dist/node/chunks/moduleRunnerTransport";

const botGameSchema = z.object({
  board_size: z.number().min(3).max(7),
  num_bots: z.number().min(1).max(3),
});

type BotGameFormValues = z.infer<typeof botGameSchema>;

interface BotGameModalProps {
  open: boolean;
  onClose: () => void;
  onSubmit: (values: BotGameFormValues) => Promise<void>;
}

export function BotGameModal({ open, onClose, onSubmit }: BotGameModalProps) {
  const form = useForm<BotGameFormValues>({
    resolver: zodResolver(botGameSchema),
    defaultValues: {
      board_size: 5,
      num_bots: 1,
    },
  });

  const handleSubmit = async (values: BotGameFormValues) => {
    try {
      await onSubmit(values);
      form.reset();
    } catch (error) {
      console.error("Error creating bot game:", error);
    }
  };

  const handleClose = () => {
    form.reset();
    onClose();
  };

  const boardSize = form.watch("board_size");
  const numBots = form.watch("num_bots");

  return (
    <Dialog open={open} onOpenChange={handleClose}>
      <DialogContent className="sm:max-w-[425px]">
        <DialogHeader>
          <DialogTitle>Create Bot Game</DialogTitle>
          <DialogDescription>
            Configure your game against AI opponents
          </DialogDescription>
        </DialogHeader>

        <Form {...form}>
          <form
            onSubmit={form.handleSubmit(handleSubmit)}
            className="space-y-4"
          >
            {/* Board Size */}
            <FormField
              control={form.control}
              name="board_size"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Board Size</FormLabel>
                  <Select
                    onValueChange={(value: string) => {
                      field.onChange(parseInt(value));
                    }}
                    value={field.value.toString()}
                  >
                    <FormControl>
                      <SelectTrigger>
                        <SelectValue placeholder="Select board size" />
                      </SelectTrigger>
                    </FormControl>
                    <SelectContent>
                      <SelectItem value="3">3×3 (Small)</SelectItem>
                      <SelectItem value="4">4×4 (Medium)</SelectItem>
                      <SelectItem value="5">5×5 (Standard)</SelectItem>
                      <SelectItem value="6">6×6 (Large)</SelectItem>
                      <SelectItem value="7">7×7 (Extra Large)</SelectItem>
                    </SelectContent>
                  </Select>
                  <FormDescription>
                    {boardSize}×{boardSize} grid
                  </FormDescription>
                  <FormMessage />
                </FormItem>
              )}
            />

            {/* Number of Bots */}
            <FormField
              control={form.control}
              name="num_bots"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Number of Bots</FormLabel>
                  <Select
                    onValueChange={(value: string) => {
                      field.onChange(parseInt(value));
                    }}
                    value={field.value.toString()}
                  >
                    <FormControl>
                      <SelectTrigger>
                        <SelectValue placeholder="Select number of bots" />
                      </SelectTrigger>
                    </FormControl>
                    <SelectContent>
                      <SelectItem value="1">1 Bot (1v1)</SelectItem>
                      <SelectItem value="2">2 Bots (3 Players)</SelectItem>
                      <SelectItem value="3">3 Bots (4 Players)</SelectItem>
                    </SelectContent>
                  </Select>
                  <FormDescription>
                    You + {numBots} bot{numBots > 1 ? "s" : ""}
                  </FormDescription>
                  <FormMessage />
                </FormItem>
              )}
            />

            <DialogFooter>
              <Button
                type="button"
                variant="outline"
                onClick={handleClose}
                disabled={form.formState.isSubmitting}
              >
                Cancel
              </Button>
              <Button type="submit" disabled={form.formState.isSubmitting}>
                {form.formState.isSubmitting ? "Creating..." : "Create Game"}
              </Button>
            </DialogFooter>
          </form>
        </Form>
      </DialogContent>
    </Dialog>
  );
}
