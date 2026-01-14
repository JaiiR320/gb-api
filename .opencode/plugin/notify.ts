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

function getRandomMessage(): string {
  return MESSAGES[Math.floor(Math.random() * MESSAGES.length)];
}

export const Notify: Plugin = async ({ $ }) => {
  const COOLDOWN_MS = 5000; // Only allow one notification per 5 seconds
  const ICON = "~/Pictures/icons/opencode.png";
  const SOUND = "/usr/share/sounds/freedesktop/stereo/complete.oga";

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
            const message = getRandomMessage();

            await Promise.all([
              $`notify-send -i ${ICON} "OpenCode" "${message}"`,
              $`paplay --volume=32768 ${SOUND}`,
            ]);
          }, 500);
        }
      }
    },
  };
};
