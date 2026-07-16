import { useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { fetchSettings, updateSettings } from "@/api/settings";

const DEFAULT_HOME = "MYR";
const DEFAULT_TARGET = "USD";
const FALLBACK_TARGET = "EUR";

export type ActivePair = {
  home: string;
  target: string;
  setHome: (code: string) => void;
  setTarget: (code: string) => void;
};

export function useActivePair(): ActivePair {
  const queryClient = useQueryClient();
  const [target, setTargetState] = useState<string>(DEFAULT_TARGET);

  const settings = useQuery({
    queryKey: ["settings"],
    queryFn: fetchSettings,
  });
  const home = settings.data?.home_currency ?? DEFAULT_HOME;

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
    setTargetState(code === home ? (home === FALLBACK_TARGET ? DEFAULT_TARGET : FALLBACK_TARGET) : code);
  };

  return { home, target, setHome, setTarget };
}
