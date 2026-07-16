import { useEffect, useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { fetchSettings, updateSettings } from "@/api/settings";

const DEFAULT_HOME = "MYR";
const DEFAULT_TARGET = "USD";

export type ActivePair = {
  home: string;
  target: string;
  setHome: (code: string) => void;
  setTarget: (code: string) => void;
};

function differentFrom(code: string): string {
  return code === "USD" ? "MYR" : "USD";
}

export function useActivePair(): ActivePair {
  const queryClient = useQueryClient();
  const [target, setTargetState] = useState<string>(DEFAULT_TARGET);

  const settings = useQuery({
    queryKey: ["settings"],
    queryFn: fetchSettings,
  });
  const home = settings.data?.home_currency ?? DEFAULT_HOME;

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

  return { home, target, setHome, setTarget };
}
