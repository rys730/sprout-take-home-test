"use client";

import { Separator } from "@/components/ui/separator";
import { SidebarTrigger } from "@/components/ui/sidebar";

export function HeaderBar() {
  return (
    <header className="z-20 flex h-14 shrink-0 items-center justify-between border-b bg-white px-4">
      {/* Logo / App title */}
      <div className="flex items-center gap-2">
        <SidebarTrigger />
        <Separator orientation="vertical" className="h-6" />
        <span className="text-lg font-semibold text-foreground">
          Sprout Accounting
        </span>
      </div>

      {/* Right side actions */}
      <div className="flex items-center gap-4">
        {/* Placeholder avatar */}
        <div className="flex h-8 w-8 items-center justify-center rounded-full bg-primary text-xs font-semibold text-primary-foreground">
          U
        </div>
      </div>
    </header>
  );
}
