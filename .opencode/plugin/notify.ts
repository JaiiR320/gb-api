import type { Plugin } from "@opencode-ai/plugin";

// Global state shared across all plugin instances
const globalState = globalThis as typeof globalThis & {
  __opencode_notify_last?: number;
};

export const Notify: Plugin = async ({ $ }) => {
  const COOLDOWN_MS = 5000; // Only allow one notification per 5 seconds

  return {
    async event(input) {
      if (input.event.type === "session.idle") {
        const now = Date.now();
        const last = globalState.__opencode_notify_last || 0;
        
        // Only notify if 5+ seconds since last notification
        if (now - last > COOLDOWN_MS) {
          globalState.__opencode_notify_last = now;
          // Small delay to let duplicate events pass
          setTimeout(async () => {
            await $`notify-send "OpenCode" "Finished!"`;
          }, 500);
        }
      }
    },
  };
};
