import { useSyncExternalStore } from "react";
import { local } from "./format";
import type { Lang } from "./i18n";
import type { Ask, Block, ChangesResponse, Info, Live, Mode, ProviderDetail, ServerView, Summary, ThreadSummary, Tuning, UpdateStatus } from "./types";

export type Theme = "system" | "dark" | "light";
export type RightTab = "plan" | "changes";

export type Dialog =
  | { kind: "ask"; ask: Ask }
  | { kind: "palette"; mode: "root" | "models" }
  | { kind: "folder" }
  | { kind: "settings" }
  | { kind: "status" }
  | { kind: "server-add" }
  | { kind: "model-connect" }
  | { kind: "provider-edit"; provider?: ProviderDetail }
  | { kind: "mcp-add" }
  | { kind: "server-auth"; id: string; need: "trust" | "password" | "passphrase"; fingerprint?: string; message?: string; install?: boolean }
  | { kind: "server-install"; id: string; update: boolean; secrets?: { password?: string; passphrase?: string; trust?: string } }
  | null;

export interface State {
  online: boolean | null;
  machine: string;
  tuning: Tuning | null;
  servers: ServerView[];
  wsl: boolean;
  info: Info;
  mode: Mode;
  projects: Summary[];
  defaultId: string;
  home: string;
  current: string;
  blocks: Block[];
  fresh: number;
  live: Live;
  lastTook: number;
  dialog: Dialog;
  errorAlert: string | null;
  theme: Theme;
  lang: Lang;
  sideOpen: boolean;
  planOpen: boolean;
  changesOpen: boolean;
  rightTab: RightTab;
  changes: ChangesResponse | null;
  activeThread: string;
  threads: ThreadSummary[];
  update: UpdateStatus | null;
  unseen: boolean;
}

const emptyInfo: Info = {
  version: "",
  platform: "",
  host: "",
  model: "",
  provider: "",
  effort: "",
  approvals: "smart",
  executionMode: "best-effort",
  sandboxed: false,
  yolo: false,
  contextWindow: 0,
  connections: 0,
  services: 0,
};

let state: State = {
  online: null,
  machine: "",
  tuning: null,
  servers: [],
  wsl: false,
  info: emptyInfo,
  mode: "build",
  projects: [],
  defaultId: "",
  home: "",
  current: "",
  blocks: [],
  fresh: 0,
  live: { running: false },
  lastTook: 0,
  dialog: null,
  errorAlert: null,
  theme: local.get<Theme>("theme", "system"),
  lang: local.get<Lang>("lang", "en"),
  sideOpen: false,
  planOpen: local.get("planOpen", true),
  changesOpen: local.get("changesOpen", false),
  rightTab: "plan",
  changes: null,
  activeThread: "default",
  threads: [],
  update: null,
  unseen: false,
};

const listeners = new Set<() => void>();

function subscribe(fn: () => void) {
  listeners.add(fn);
  return () => {
    listeners.delete(fn);
  };
}

export function getState(): State {
  return state;
}

export function setState(patch: Partial<State> | ((s: State) => Partial<State>)) {
  const next = typeof patch === "function" ? patch(state) : patch;
  state = { ...state, ...next };
  listeners.forEach((fn) => fn());
}

export function patchLive(fn: (l: Live) => Partial<Live>) {
  setState((s) => ({ live: { ...s.live, ...fn(s.live) } }));
}

export function useStore<T>(select: (s: State) => T): T {
  return useSyncExternalStore(subscribe, () => select(state));
}
