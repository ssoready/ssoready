import React, { ReactNode } from "react";
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover";
import { InfoIcon } from "lucide-react";

export function InfoTooltip({ children }: { children: ReactNode }) {
  return (
    <Popover>
      <PopoverTrigger>
        <InfoIcon className="h-4 w-4" />
      </PopoverTrigger>
      <PopoverContent side="top" className="text-xs">
        {children}
      </PopoverContent>
    </Popover>
  );
}
