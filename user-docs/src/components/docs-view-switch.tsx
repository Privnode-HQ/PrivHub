"use client";

import {
  Popover,
  PopoverClose,
  PopoverContent,
  PopoverTrigger,
} from "fumadocs-ui/components/ui/popover";
import { Check, ChevronDown } from "lucide-react";
import {
  createContext,
  type ReactNode,
  useContext,
  useEffect,
  useMemo,
  useState,
} from "react";
import { cn } from "@/lib/cn";

const storageKey = "51api-doc-view";
const defaultView = "cli";

const docViews = [
  { value: "cch", label: "CCH", mark: "CH" },
  { value: "cc-switch", label: "CC Switch", mark: "CC" },
  { value: "cli", label: "CLI", mark: "CLI" },
] as const;

type DocViewValue = (typeof docViews)[number]["value"];

type DocViewContextValue = {
  value: DocViewValue;
  setValue: (value: DocViewValue) => void;
};

const DocViewContext = createContext<DocViewContextValue | null>(null);

function isDocViewValue(value: string | null): value is DocViewValue {
  return docViews.some((item) => item.value === value);
}

function normalizeDocView(value: string | undefined) {
  const normalized = value?.trim().toLowerCase().replace(/\s+/g, "-");

  if (!normalized) return undefined;
  if (isDocViewValue(normalized)) return normalized;

  return docViews.find(
    (item) => item.label.toLowerCase() === value?.trim().toLowerCase(),
  )?.value;
}

export function DocViewProvider({ children }: { children: ReactNode }) {
  const [value, setValueState] = useState<DocViewValue>(defaultView);

  useEffect(() => {
    const saved = window.localStorage.getItem(storageKey);

    if (isDocViewValue(saved)) {
      setValueState(saved);
    }
  }, []);

  const contextValue = useMemo<DocViewContextValue>(
    () => ({
      value,
      setValue(nextValue) {
        setValueState(nextValue);
        window.localStorage.setItem(storageKey, nextValue);
      },
    }),
    [value],
  );

  return (
    <DocViewContext.Provider value={contextValue}>
      {children}
    </DocViewContext.Provider>
  );
}

export function DocViewSwitcher({ enabled }: { enabled: boolean }) {
  const context = useContext(DocViewContext);

  if (!enabled || !context) return null;

  const selectedView = docViews.find((item) => item.value === context.value);

  return (
    <div className="mb-4">
      <Popover>
        <PopoverTrigger
          type="button"
          aria-label="选择文档版本"
          className="group flex h-9 w-full items-center gap-2 rounded-lg border border-fd-border/70 bg-transparent px-2.5 text-left text-fd-muted-foreground outline-none hover:bg-fd-accent/30 hover:text-fd-accent-foreground focus-visible:ring-2 focus-visible:ring-fd-ring"
        >
          <span className="min-w-0 flex-1 truncate text-sm font-medium">
            {selectedView?.label ?? "CLI"}
          </span>
          <ChevronDown className="size-4 shrink-0 text-fd-muted-foreground group-data-[state=open]:rotate-180" />
        </PopoverTrigger>
        <PopoverContent
          align="start"
          className="min-w-(--radix-popover-trigger-width) rounded-2xl p-2"
        >
          <div className="grid gap-1">
            {docViews.map((item) => {
              const selected = item.value === context.value;

              return (
                <PopoverClose asChild key={item.value}>
                  <button
                    type="button"
                    className={cn(
                      "flex h-9 w-full items-center gap-2 rounded-xl px-2 text-left text-sm text-fd-popover-foreground hover:bg-fd-accent hover:text-fd-accent-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-fd-ring",
                      selected && "bg-fd-accent text-fd-accent-foreground",
                    )}
                    onClick={() => {
                      context.setValue(item.value);
                    }}
                  >
                    <span className="min-w-0 flex-1 truncate">
                      {item.label}
                    </span>
                    {selected ? <Check className="size-4 shrink-0" /> : null}
                  </button>
                </PopoverClose>
              );
            })}
          </div>
        </PopoverContent>
      </Popover>
    </div>
  );
}

export function View({
  children,
  title,
  value,
}: {
  children: ReactNode;
  title?: string;
  value?: string;
}) {
  const context = useContext(DocViewContext);
  const target = normalizeDocView(value ?? title);
  const selected = context?.value ?? defaultView;

  if (!target || target !== selected) return null;

  return <div data-doc-view={target}>{children}</div>;
}
