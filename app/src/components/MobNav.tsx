import React from "react";
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
  SheetTrigger,
} from "@/components/ui/sheet";
import { Tally3Icon } from "lucide-react";
import { PageComp } from "./PageComp";
export function MobNav() {
  return (
    <div className="md:hidden fixed top-2 left-4 p-0 ">
      <Sheet>
        <SheetTrigger>
          <Tally3Icon className="rotate-90" />
        </SheetTrigger>
        <SheetContent>
          <SheetHeader className="mt-12">
            <SheetTitle></SheetTitle>
            <SheetDescription>
              <PageComp />
            </SheetDescription>
          </SheetHeader>
        </SheetContent>
      </Sheet>
    </div>
  );
}
