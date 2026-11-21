import { newChat, openDialog, toggleChanges } from "../actions";
import { shortPath } from "../format";
import { useMod, useProjectName } from "../hooks";
import { useT } from "../i18n";
import { Icon } from "../icons";
import { setState, useStore } from "../store";
import { togglePlan, usePanelPlan } from "./PlanPanel";

export function Header() {
  const t = useT();
  const name = useProjectName();
  const path = useStore((s) => s.projects.find((p) => p.id === s.current)?.path || "");
  const home = useStore((s) => s.home);
  const remote = useStore((s) => s.servers.find((x) => x.id === s.machine)?.name || "");
  const mod = useMod();
  return (
    <header className="top">
      <button type="button" className="icon-btn side-toggle" title={t("Workspaces")} onClick={() => setState((s) => ({ sideOpen: !s.sideOpen }))}>
        <Icon name="menu" />
      </button>
      <div className="title">
        <h1>{name === "Home" ? t("Home") : name}</h1>
        {remote && (
          <span className="on-server" title="Macron works on this server">
            <Icon name="server" size={12} />
            {remote}
          </span>
        )}
        <span className="path" title={path} dir="ltr">
          {"‎" + shortPath(path, home) + "‎"}
        </span>
      </div>
      <ChangesToggle />
      <PlanToggle />
      <button type="button" className="search" onClick={() => openDialog({ kind: "palette", mode: "root" })}>
        <Icon name="search" size={14} />
        <span>{t("Search or run a command")}</span>
        <kbd>{mod} K</kbd>
      </button>
      <button type="button" className="icon-btn" title={`${t("New conversation")} (${mod}+N)`} onClick={newChat}>
        <Icon name="edit" />
      </button>
    </header>
  );
}

function ChangesToggle() {
  const t = useT();
  const changes = useStore((s) => s.changes);
  const open = useStore((s) => s.changesOpen && s.rightTab === "changes");
  const count = changes?.files?.length || 0;
  if (!changes?.git) return null;
  return (
    <button
      type="button"
      className={"plan-toggle" + (open ? " on" : "")}
      title={t("Show changes")}
      onClick={() => toggleChanges()}
    >
      <Icon name="gitBranch" size={14} />
      <span>{count > 0 ? count : t("Changes")}</span>
    </button>
  );
}

function PlanToggle() {
  const t = useT();
  const plan = usePanelPlan();
  const open = useStore((s) => s.planOpen && s.rightTab === "plan");
  if (!plan.length) return null;
  const done = plan.filter((i) => i.status === "done").length;
  return (
    <button type="button" className={"plan-toggle" + (open ? " on" : "")} title={t("Hide the plan")} onClick={() => togglePlan()}>
      <Icon name="steps" size={14} />
      <span>
        {done}/{plan.length}
      </span>
    </button>
  );
}
