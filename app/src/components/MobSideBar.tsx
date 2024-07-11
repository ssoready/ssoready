import React from "react";

import { Button } from "@/components/ui/button";

import {
  Sheet,
  SheetClose,
  SheetContent,
  SheetDescription,
  SheetFooter,
  SheetHeader,
  SheetTitle,
  SheetTrigger,
} from "@/components/ui/sheet";
import { Menu } from "lucide-react";
import { PageComponent } from "./PageComponent";

export function MobSideBar() {
  return (
    <div className="block md:hidden">
      <Sheet>
        <SheetTrigger asChild>
          <Button variant="outline" className="mt-7 ml-4">
            <Menu />
          </Button>
        </SheetTrigger>
        <SheetContent className="flex flex-col justify-between">
          <PageComponent />
        </SheetContent>
      </Sheet>
    </div>
  );
}
