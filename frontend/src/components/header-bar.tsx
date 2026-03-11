"use client";

import { Separator } from "@/components/ui/separator";

export function HeaderBar() {
  return (
    <header className="z-20 flex h-14 shrink-0 items-center justify-between border-b bg-background px-6">
      {/* Logo / App title */}
      <div className="flex items-center gap-3">
        <span className="text-lg font-semibold text-foreground">
          Sprout Accounting
        </span>
      </div>

      <Separator orientation="vertical" className="h-6" />

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
