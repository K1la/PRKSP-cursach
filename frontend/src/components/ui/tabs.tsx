"use client";

import { createContext, useContext } from "react";

import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";

type TabsContextValue = {
  value: string;
  onValueChange: (value: string) => void;
};

const TabsContext = createContext<TabsContextValue | null>(null);

export function Tabs({
  value,
  onValueChange,
  children,
}: {
  value: string;
  onValueChange: (value: string) => void;
  children: React.ReactNode;
}) {
  return <TabsContext.Provider value={{ value, onValueChange }}>{children}</TabsContext.Provider>;
}

export function TabsList({ className, ...props }: React.HTMLAttributes<HTMLDivElement>) {
  return <div className={cn("inline-flex rounded border border-border bg-white p-1", className)} {...props} />;
}

export function TabsTrigger({
  value,
  children,
}: {
  value: string;
  children: React.ReactNode;
}) {
  const context = useContext(TabsContext);
  if (!context) {
    throw new Error("TabsTrigger must be used inside Tabs");
  }
  return (
    <Button
      className={cn(
        "h-8 bg-transparent px-3 text-slate-700 hover:bg-slate-100",
        context.value === value && "bg-primary text-primary-foreground hover:bg-primary",
      )}
      onClick={() => context.onValueChange(value)}
      type="button"
    >
      {children}
    </Button>
  );
}

export function TabsContent({
  value,
  className,
  ...props
}: React.HTMLAttributes<HTMLDivElement> & { value: string }) {
  const context = useContext(TabsContext);
  if (!context || context.value !== value) {
    return null;
  }
  return <div className={cn("mt-4", className)} {...props} />;
}
