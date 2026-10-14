import type { Metadata } from "next";
import "./globals.css";

export const metadata: Metadata = {
  title: "X. It’s what’s happening / X",
  description: "From breaking news and entertainment to sports and politics, get the full story with all the live commentary.",
  icons: {
    icon: "/favicon.ico",
  },
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="en" className="dark">
      <body className="bg-background text-foreground min-h-screen antialiased selection:bg-brand selection:text-white">
        {children}
      </body>
    </html>
  );
}
