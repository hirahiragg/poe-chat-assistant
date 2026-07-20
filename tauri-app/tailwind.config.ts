import type { Config } from "tailwindcss";

export default {
  content: ["./index.html", "./src/**/*.{ts,tsx}"],
  theme: {
    extend: {
      colors: {
        bg: "#1a1a20",
        surface: "#22222a",
        card: "#2a2a34",
        border: "#3a3a44",
        selected: "#304050",
        text: "#ccc8c0",
        "text-dim": "#888580",
        accent: "#4a9e6a",
        "btn-bg": "#335a40",
        "btn-text": "#e0e0e0",
        translated: "#a0d0a0",
        "ch-whisper": "#b070d8",
        "ch-trade": "#bf9a4a",
        "ch-party": "#5ec4eb",
        "ch-guild": "#6aaa64",
        "ch-global": "#d04040",
      },
    },
  },
  plugins: [],
} satisfies Config;
