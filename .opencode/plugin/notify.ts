import type { Plugin } from "@opencode-ai/plugin";

// Global state shared across all plugin instances
const globalState = globalThis as typeof globalThis & {
  __opencode_notify_last?: number;
};

const MESSAGES = [
  "All done!",
  "Finished!",
  "Task complete!",
  "Ready when you are!",
  "Your wish is my command!",
  "At your service!",
];

const SOUNDS = [
  "/usr/share/sounds/custom/a.mp3",
  "/usr/share/sounds/custom/b.mp3",
];

function getRandomItem<T>(arr: T[]): T {
  return arr[Math.floor(Math.random() * arr.length)] as T;
}

export const Notify: Plugin = async ({ $ }) => {
  const COOLDOWN_MS = 5000; // Only allow one notification per 5 seconds
  const ICON = "~/Pictures/icons/opencode.png";

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
            const message = getRandomItem(MESSAGES);
            const sound = getRandomItem(SOUNDS);

            await Promise.all([
              $`notify-send -i ${ICON} "OpenCode" "${message}"`,
              $`paplay --volume=45000 ${sound}`,
            ]);
          }, 500);
        }
      }
    },
  };
};
