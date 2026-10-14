import type { Config } from "tailwindcss";

const config: Config = {
  darkMode: "class",
  content: [
    "./src/pages/**/*.{js,ts,jsx,tsx,mdx}",
    "./src/components/**/*.{js,ts,jsx,tsx,mdx}",
    "./src/app/**/*.{js,ts,jsx,tsx,mdx}",
  ],
  theme: {
    extend: {
      colors: {
        background: "var(--background)",
        foreground: "var(--foreground)",
        card: "var(--card)",
        border: "var(--border)",
        muted: "var(--muted)",
        subtle: "var(--hover-subtle)",
        brand: {
          DEFAULT: "#1d9bf0",
          hover: "#1a8cd8",
        },
        like: {
          DEFAULT: "#f91880",
          bg: "rgba(249, 24, 128, 0.1)",
        },
        repost: {
          DEFAULT: "#00ba7c",
          bg: "rgba(0, 186, 124, 0.1)",
        },
      },
      width: {
        "rail-icon": "88px",
        "rail-labeled": "275px",
        "primary-feed": "600px",
        "right-sidebar": "420px",
      },
      height: {
        "sticky-header": "53px",
      },
      fontFamily: {
        chirp: ["TwitterChirp", "-apple-system", "BlinkMacSystemFont", "Segoe UI", "Roboto", "sans-serif"],
      },
    },
  },
  plugins: [],
};
export default config;
