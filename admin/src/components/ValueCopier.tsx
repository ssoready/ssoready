import { offset, useFloating, useTransitionStyles } from "@floating-ui/react";
import { CopyIcon } from "lucide-react";
import React, { useCallback, useEffect, useState } from "react";

export function ValueCopier({ value }: { value: string }) {
  const [open, setOpen] = useState(false);
  const { refs, floatingStyles, context } = useFloating({
    open,
    onOpenChange: setOpen,
    placement: "top",
    middleware: [offset(5)],
  });
  const { isMounted, styles } = useTransitionStyles(context, {
    duration: 150,
    initial: { opacity: 0, transform: "translateY(0)" },
    open: { opacity: 1, transform: "translateY(-5px)" },
  });

  useEffect(() => {
    if (open) {
      const timeoutId = setTimeout(() => {
        setOpen(false);
      }, 1000);
      return () => clearTimeout(timeoutId);
    }
  }, [open]);

  const handleCopy = useCallback(async () => {
    await navigator.clipboard.writeText(value);
    setOpen(true);
  }, [value, setOpen]);

  return (
    <div
      ref={refs.setReference}
      onClick={handleCopy}
      className="flex select-none cursor-pointer bg-muted font-mono text-xs border border-input rounded-md px-3 py-2"
    >
      <span>{value}</span>
      <span className="ml-auto flex gap-x-2">
        <CopyIcon className="cursor-pointer text-muted-foreground hover:text-foreground h-4 w-4" />
      </span>

      {open && (
        <div ref={refs.setFloating} style={floatingStyles}>
          {isMounted && (
            <div
              style={styles}
              className="font-sans bg-black text-white px-2 py-1 text-xs rounded"
            >
              Copied!
            </div>
          )}
        </div>
      )}
    </div>
  );
}
