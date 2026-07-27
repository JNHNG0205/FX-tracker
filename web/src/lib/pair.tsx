import {
  createContext,
  useContext,
  useEffect,
  useState,
  type ReactNode,
} from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { fetchSettings, updateSettings } from "@/api/settings";

const DEFAULT_HOME = "MYR";
const DEFAULT_TARGET = "USD";
const TARGET_STORAGE_KEY = "target";

export type ActivePair = {
  home: string;
  target: string;
  setHome: (code: string) => void;
  setTarget: (code: string) => void;
};

function differentFrom(code: string): string {
  return code === "USD" ? "MYR" : "USD";
}

const ActivePairContext = createContext<ActivePair | null>(null);

export function ActivePairProvider({ children }: { children: ReactNode }) {
  const queryClient = useQueryClient();
  const [target, setTargetState] = useState<string>(
    () => localStorage.getItem(TARGET_STORAGE_KEY) ?? DEFAULT_TARGET,
  );

  const settings = useQuery({
    queryKey: ["settings"],
    queryFn: fetchSettings,
  });
  const home = settings.data?.home_currency ?? DEFAULT_HOME;

  useEffect(() => {
    localStorage.setItem(TARGET_STORAGE_KEY, target);
  }, [target]);

  useEffect(() => {
    if (home === target) {
      setTargetState(differentFrom(home));
    }
  }, [home, target]);

  const mutation = useMutation({
    mutationFn: updateSettings,
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ["settings"] });
    },
  });

  const setHome = (code: string) => {
    mutation.mutate(code);
  };

  const setTarget = (code: string) => {
    setTargetState(code === home ? differentFrom(home) : code);
  };

  return (
    <ActivePairContext.Provider value={{ home, target, setHome, setTarget }}>
      {children}
    </ActivePairContext.Provider>
  );
}

export function useActivePair(): ActivePair {
  const context = useContext(ActivePairContext);
  if (!context) {
    throw new Error("useActivePair must be used within an ActivePairProvider");
  }
  return context;
}
