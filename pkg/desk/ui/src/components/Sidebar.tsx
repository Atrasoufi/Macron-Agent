import { memo, useEffect, useRef, useState } from "react";
import { deleteThread, installUpdate, loadServers, openDialog, openProject, removeProject, switchMachine, switchThread } from "../actions";
import { ago, shortPath } from "../format";
import { useMod } from "../hooks";
import { useT } from "../i18n";
import { Icon } from "../icons";
import { useStore } from "../store";
import type { ServerView, Summary, ThreadSummary } from "../types";

function serverLine(s: ServerView): string {
  if (s.state === "failed" && s.error) return s.error;
  if (s.state === "connecting") return "Connecting…";
  if (s.state === "installing") return "Installing Macron…";
  if (s.kind === "wsl") return "WSL" + (s.host ? " · " + s.host : "");
  return (s.user ? s.user + "@" : "") + (s.host || "") + (s.port && s.port !== 22 ? ":" + s.port : "");
}

function MachineMenu() {
  const t = useT();
  const machine = useStore((s) => s.machine);
  const servers = useStore((s) => s.servers);
  const [open, setOpen] = useState(false);
  const ref = useRef<HTMLDivElement>(null);
  useEffect(() => {
    if (!open) return;
    loadServers();
    const away = (e: MouseEvent) => {
      if (!ref.current?.contains(e.target as Node)) setOpen(false);
    };
    const esc = (e: KeyboardEvent) => {
      if (e.key !== "Escape") return;
      e.preventDefault();
      e.stopPropagation();
      setOpen(false);
    };
    window.addEventListener("mousedown", away);
    window.addEventListener("keydown", esc, true);
    return () => {
      window.removeEventListener("mousedown", away);
      window.removeEventListener("keydown", esc, true);
    };
  }, [open]);
  const cur = servers.find((x) => x.id === machine);
  const pick = (id: string) => {
    setOpen(false);
    switchMachine(id);
  };
  return (
    <div className="machine" ref={ref}>
      <button type="button" className={"machine-btn" + (open ? " open" : "")} onClick={() => setOpen(!open)} aria-expanded={open}>
        <Icon name={cur ? (cur.kind === "wsl" ? "terminal" : "server") : "monitor"} size={15} />
        <span className="mname">{cur ? cur.name : t("This PC")}</span>
        {cur && <span className={"sdot " + cur.state} />}
        <Icon name="down" size={12} />
      </button>
      {open && (
        <div className="menu" role="menu">
          <button type="button" role="menuitem" className={"mrow" + (machine === "" ? " on" : "")} onClick={() => pick("")}>
            <Icon name="monitor" size={15} />
            <span className="grow">
              <span className="name">{t("This PC")}</span>
              <span className="meta">{t("Work on this computer")}</span>
            </span>
            {machine === "" && <Icon name="check" size={14} />}
          </button>
          {servers.map((s) => (
            <button key={s.id} type="button" role="menuitem" className={"mrow" + (machine === s.id ? " on" : "")} onClick={() => pick(s.id)}>
              <Icon name={s.kind === "wsl" ? "terminal" : "server"} size={15} />
              <span className="grow">
                <span className="name">{s.name}</span>
                <span className={"meta" + (s.state === "failed" ? " bad" : "")}>{serverLine(s)}</span>
              </span>
              {machine === s.id ? <Icon name="check" size={14} /> : <span className={"sdot " + s.state} />}
            </button>
          ))}
          <div className="msep" />
          <button
            type="button"
            role="menuitem"
            className="mrow add"
            onClick={() => {
              setOpen(false);
              openDialog({ kind: "server-add" });
            }}
          >
            <Icon name="plus" size={15} />
            <span className="grow">
              <span className="name">{t("Add a server…")}</span>
            </span>
          </button>
        </div>
      )}
    </div>
  );
}

const ThreadList = memo(function ThreadList({ threads, active }: { threads: ThreadSummary[]; active: string }) {
  const t = useT();
  if (threads.length <= 1) return null;
  return (
    <div className="thread-list" aria-label={t("Threads")}>
      {threads.map((th) => (
        <div key={th.id} className={"thread-row" + (th.id === active ? " on" : "")}>
          <button type="button" className="thread-hit" onClick={() => switchThread(th.id)} title={th.title}>
            <span className="thread-title" dir="auto">
              {th.title === "New conversation" ? t("New conversation") : th.title}
            </span>
            {th.updated > 0 && <span className="thread-ago">{ago(th.updated)}</span>}
          </button>
          <button
            type="button"
            className="thread-del"
            title={t("Remove")}
            onClick={(e) => {
              e.stopPropagation();
              deleteThread(th.id);
            }}
          >
            <Icon name="x" size={12} />
          </button>
        </div>
      ))}
    </div>
  );
});

const Row = memo(function Row({ p, on, isHome, home }: { p: Summary; on: boolean; isHome: boolean; home: string }) {
  const t = useT();
  const threads = useStore((s) => (on ? s.threads : []));
  const activeThread = useStore((s) => s.activeThread);
  const preview = p.preview || (isHome ? t("General conversations") : shortPath(p.path, home));
  return (
    <>
      <div className={"proj" + (on ? " on" : "")}>
        <button type="button" className="proj-hit" title={p.path} onClick={() => openProject(p.id)}>
          <span className="pn">{isHome ? t("Home") : p.name}</span>
          {p.asking ? (
            <span className="badge ask" title="Waiting for your approval" />
          ) : p.running ? (
            <span className="badge run" title={t("Working")} />
          ) : (
            <span className="ago">{ago(p.updated)}</span>
          )}
          <span className="pv" dir="auto">
            {preview}
          </span>
        </button>
        {!isHome && (
          <button type="button" className="x" title="Remove from the list (the folder stays)" onClick={() => removeProject(p.id)}>
            <Icon name="x" size={14} />
          </button>
        )}
      </div>
      {on && <ThreadList threads={threads} active={activeThread} />}
    </>
  );
});

function UpdateCard() {
  const t = useT();
  const upd = useStore((s) => s.update);
  const machine = useStore((s) => s.machine);
  if (machine || !upd || (!upd.available && !upd.installing && !upd.installed)) return null;
  return (
    <div className={"upd-card" + (upd.installed ? " done" : "")}>
      <div className="upd-top">
        <Icon name="download" size={14} />
        <span className="grow">
          {upd.installed ? t("Updated — restart Macron") : `${t("Update available")} · v${upd.latest}`}
        </span>
      </div>
      {!upd.installed && (
        <button type="button" className="upd-btn" disabled={upd.installing} onClick={() => installUpdate()}>
          {upd.installing ? upd.step || t("Updating…") : t("Update now")}
        </button>
      )}
    </div>
  );
}

export function Sidebar() {
  const t = useT();
  const projects = useStore((s) => s.projects);
  const current = useStore((s) => s.current);
  const defaultId = useStore((s) => s.defaultId);
  const home = useStore((s) => s.home);
  const online = useStore((s) => s.online);
  const version = useStore((s) => s.info.version);
  const mod = useMod();
  return (
    <aside className="side">
      <div className="brand">Macron</div>
      <MachineMenu />
      <button type="button" className="nav" onClick={() => openDialog({ kind: "folder" })}>
        <Icon name="plus" />
        <span>{t("Open folder")}</span>
        <kbd>{mod} O</kbd>
      </button>
      <div className="side-label">{t("Workspaces")}</div>
      <nav className="projects">
        {projects.map((p) => (
          <Row key={p.id} p={p} on={p.id === current} isHome={p.id === defaultId} home={home} />
        ))}
      </nav>
      <div className="side-foot">
        <UpdateCard />
        <button type="button" className="nav" onClick={() => openDialog({ kind: "settings" })}>
          <Icon name="sliders" />
          <span>{t("Settings")}</span>
          <kbd>{mod} ,</kbd>
        </button>
        <div className={"engine" + (online === true ? " ok" : online === false ? " bad" : "")}>
          <span className="dot" />
          <span>{online === null ? t("Connecting…") : online ? t("Engine ready") + (version ? " · " + version : "") : t("Reconnecting…")}</span>
        </div>
      </div>
    </aside>
  );
}
