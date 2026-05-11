import * as React from "react";

import { cn } from "@/lib/utils";

function Badge({ className, ...props }: React.HTMLAttributes<HTMLSpanElement>) {
  return (
    <span
      className={cn(
        "inline-flex items-center rounded border border-green-200 bg-green-50 px-2.5 py-1 text-xs font-medium text-green-800",
        className,
      )}
      {...props}
    />
  );
}

export { Badge };
