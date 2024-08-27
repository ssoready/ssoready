import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Input, InputProps } from "@/components/ui/input";
import { XIcon } from "lucide-react";
import React, { Dispatch, SetStateAction, forwardRef, useState } from "react";

type InputTagsProps = InputProps & {
  value: string[];
  onChange: Dispatch<SetStateAction<string[]>>;
};

export const InputTags = forwardRef<HTMLInputElement, InputTagsProps>(
  ({ value, onChange, onBlur, ...props }, ref) => {
    const [pendingDataPoint, setPendingDataPoint] = useState("");

    const addPendingDataPoint = () => {
      if (pendingDataPoint) {
        // trim() because a copy-pasted input may still contain leading/trailing whitespace
        const newDataPoints = new Set([...value, pendingDataPoint.trim()]);
        onChange(Array.from(newDataPoints));
        setPendingDataPoint("");
      }
    };

    return (
      <>
        <div className="flex">
          <Input
            value={pendingDataPoint}
            onChange={(e) => setPendingDataPoint(e.target.value)}
            onKeyDown={(e) => {
              if (e.key === "Enter") {
                e.preventDefault();
                addPendingDataPoint();
              } else if (e.key === "," || e.key === " ") {
                e.preventDefault();
                addPendingDataPoint();
              }
            }}
            onBlur={(e) => {
              if (pendingDataPoint !== "") {
                addPendingDataPoint();
              }
              if (onBlur) {
                onBlur(e);
              }
            }}
            className="pr-[4.1rem]"
            {...props}
            ref={ref}
          />
          <Button
            type="button"
            variant="secondary"
            className="absolute h-[3.5vh] mt-[3px] rounded-lg right-7 p-[0.9rem]"
            onClick={addPendingDataPoint}
          >
            Add
          </Button>
        </div>
        {value.length > 0 && 
          <div className="rounded-md min-h-[2.5rem] overflow-y-auto p-2 flex gap-2 flex-wrap items-center">
            {value.map((item, idx) => (
              <Badge key={idx} variant="secondary">
                {item}
                <button
                  type="button"
                  className="w-3 ml-2"
                  onClick={() => {
                    onChange(value.filter((i) => i !== item));
                  }}
                >
                  <XIcon className="w-3" />
                </button>
              </Badge>
            ))}
          </div>
        }
      </>
    );
  },
);
InputTags.displayName = "InputTags";
