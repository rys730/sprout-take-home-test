import type { Metadata } from "next";
import { Inter } from "next/font/google";
import "./globals.css";
import { SidebarProvider, SidebarInset } from "@/components/ui/sidebar";
import { AppSidebar } from "@/components/app-sidebar";
import { HeaderBar } from "@/components/header-bar";

const inter = Inter({
  subsets: ["latin"],
  variable: "--font-inter",
});

export const metadata: Metadata = {
  title: "Sprout Accounting",
  description: "Accounting management system",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en" suppressHydrationWarning>
      <body className={`${inter.variable} antialiased`}>
        <SidebarProvider>
          <div className="flex h-screen w-full flex-col" style={{ "--header-height": "3.5rem" } as React.CSSProperties}>
            <HeaderBar />
            <div className="flex flex-1 overflow-hidden">
              <AppSidebar />
              <SidebarInset>
                <main className="flex-1 overflow-auto p-4">{children}</main>
              </SidebarInset>
            </div>
          </div>
        </SidebarProvider>
      </body>
    </html>
  );
}
