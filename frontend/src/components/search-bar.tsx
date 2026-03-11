"use client";

import { Search } from "lucide-react";
import { Input } from "@/components/ui/input";
import { cn } from "@/lib/utils";
import { Button } from "./ui/button";

interface SearchBarProps extends React.ComponentProps<"input"> {
  containerClassName?: string;
}

export function SearchBar({
  containerClassName,
  className,
  ...props
}: SearchBarProps) {
  return (
    <div className={cn("relative flex items-center gap-4", containerClassName)}>
      <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
      <Input
        type="search"
        placeholder="Search..."
        className={cn("pl-9", className)}
        {...props}
      />
    </div>
  );
}
