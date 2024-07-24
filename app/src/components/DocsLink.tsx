import React from "react";
import { Link } from "react-router-dom";
import { ArrowUpRightIcon } from "lucide-react";

export function DocsLink({ to }: { to: string }) {
  return (
    <Link
      className="ml-4 hover:underline active:text-blue-800 text-xs font-semibold text-blue-600 inline-flex items-center"
      to={to}
    >
      Docs
      <ArrowUpRightIcon className="ml-0.5 h-3 w-3" />
    </Link>
  );
}
