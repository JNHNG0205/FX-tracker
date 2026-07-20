import { useState } from "react";
import { NavLink } from "react-router-dom";
import { ArrowLeftRight, Menu, X } from "lucide-react";
import { CurrencySelect } from "@/components/CurrencySelect";
import { useActivePair } from "@/lib/pair";
import { cn } from "@/lib/utils";

const navLinkClassName = ({ isActive }: { isActive: boolean }) =>
  cn(
    "flex items-center rounded-md px-3 py-2 text-sm font-medium transition-colors",
    isActive
      ? "bg-primary/10 text-primary"
      : "text-muted-foreground hover:bg-accent hover:text-accent-foreground",
  );

function SidebarContent({ onNavigate }: { onNavigate?: () => void }) {
  const { home, setHome } = useActivePair();

  return (
    <div className="flex h-full flex-col gap-6 p-4">
      <div className="flex items-center gap-2 px-2 pt-2">
        <ArrowLeftRight className="size-5 text-primary" aria-hidden="true" />
        <span className="text-lg font-bold tracking-tight">FX Tracker</span>
      </div>

      <nav className="flex flex-col gap-1" aria-label="Main navigation">
        <NavLink to="/" end className={navLinkClassName} onClick={onNavigate}>
          Convert
        </NavLink>
        <NavLink to="/conversions" className={navLinkClassName} onClick={onNavigate}>
          Conversions
        </NavLink>
        <NavLink to="/holdings" className={navLinkClassName} onClick={onNavigate}>
          Holdings
        </NavLink>
        <span
          aria-disabled="true"
          className="flex items-center justify-between rounded-md px-3 py-2 text-sm font-medium text-muted-foreground/50"
        >
          Dividends
          <span className="rounded-full bg-muted px-2 py-0.5 text-xs font-semibold uppercase tracking-wide text-muted-foreground">
            Soon
          </span>
        </span>
      </nav>

      <div className="mt-auto">
        <p className="mb-2 px-2 text-xs font-medium uppercase tracking-wide text-muted-foreground">
          Home currency
        </p>
        <CurrencySelect value={home} onChange={setHome} label="Home currency" />
      </div>
    </div>
  );
}

export function Sidebar() {
  const [open, setOpen] = useState(false);

  return (
    <>
      <div className="flex items-center justify-between border-b bg-background px-4 py-3 md:hidden">
        <div className="flex items-center gap-2">
          <ArrowLeftRight className="size-5 text-primary" aria-hidden="true" />
          <span className="text-base font-bold tracking-tight">FX Tracker</span>
        </div>
        <button
          type="button"
          aria-expanded={open}
          aria-label="Toggle navigation menu"
          className="rounded-md p-2 text-muted-foreground hover:bg-accent hover:text-accent-foreground"
          onClick={() => setOpen(true)}
        >
          <Menu className="size-5" aria-hidden="true" />
        </button>
      </div>

      <aside className="hidden w-56 shrink-0 border-r bg-background md:block">
        <SidebarContent />
      </aside>

      {open ? (
        <div className="fixed inset-0 z-50 md:hidden">
          <button
            type="button"
            aria-label="Close navigation menu"
            className="absolute inset-0 bg-black/40"
            onClick={() => setOpen(false)}
          />
          <div
            className="motion-safe:animate-in motion-safe:slide-in-from-left absolute inset-y-0 left-0 w-64 border-r bg-background shadow-lg motion-safe:duration-200"
            onKeyDown={(event) => {
              if (event.key === "Escape") setOpen(false);
            }}
          >
            <div className="flex justify-end p-2">
              <button
                type="button"
                aria-label="Close navigation menu"
                className="rounded-md p-2 text-muted-foreground hover:bg-accent hover:text-accent-foreground"
                onClick={() => setOpen(false)}
              >
                <X className="size-5" aria-hidden="true" />
              </button>
            </div>
            <SidebarContent onNavigate={() => setOpen(false)} />
          </div>
        </div>
      ) : null}
    </>
  );
}
