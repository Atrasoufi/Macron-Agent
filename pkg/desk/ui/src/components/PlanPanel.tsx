import { useMemo } from "react";
import { setRightTab, toggleChanges } from "../actions";
import { local, took } from "../format";
import { useNow } from "../hooks";
import { useT } from "../i18n";
import { Icon } from "../icons";
import { setState, useStore } from "../store";
import type { PlanItem } from "../types";
import { ChangesBody } from "./ChangesPanel";

export function usePanelPlan(): PlanItem[] {
  const running = useStore((s) => s.live.running);
  const livePlan = useStore((s) => s.live.plan);
  const blocks = useStore((s) => s.blocks);
  return useMemo(() => {
    if (running && livePlan && livePlan.length) return livePlan;
    for (let i = blocks.length - 1; i >= 0; i--) {
      const b = blocks[i];
      if (b.kind === "divider" || b.kind === "user") break;
      if (b.kind === "plan") {
        try {
          return (JSON.parse(b.text || "[]") as PlanItem[]) || [];
        } catch {
          return [];
        }
      }
    }
    return [];
  }, [running, livePlan, blocks]);
}

export function togglePlan(open?: boolean) {
  setState((s) => {
    const next = open ?? !(s.planOpen && s.rightTab === "plan");
    local.set("planOpen", next);
    if (next) return { planOpen: true, rightTab: "plan" };
    return { planOpen: false };
  });
}

export function useRightPanelVisible(): boolean {
  const plan = usePanelPlan();
  const planOpen = useStore((s) => s.planOpen);
  const changesOpen = useStore((s) => s.changesOpen);
  const rightTab = useStore((s) => s.rightTab);
  if (rightTab === "changes" && changesOpen) return true;
  if (plan.length > 0 && planOpen) return true;
  if (changesOpen) return true;
  return false;
}

export function PlanPanel() {
  const t = useT();
  const plan = usePanelPlan();
  const planOpen = useStore((s) => s.planOpen);
  const changesOpen = useStore((s) => s.changesOpen);
  const rightTab = useStore((s) => s.rightTab);
  const changes = useStore((s) => s.changes);
  const running = useStore((s) => s.live.running);
  const started = useStore((s) => s.live.started || 0);
  const now = useNow(running);

  const hasPlan = plan.length > 0;
  const showChanges = (rightTab === "changes" && changesOpen) || (!hasPlan && changesOpen) || (!planOpen && changesOpen);
  const showPlan = !showChanges && hasPlan && planOpen;

  if (!showPlan && !showChanges) return null;

  const done = plan.filter((i) => i.status === "done").length;
  const current = plan.find((i) => i.status === "in_progress");
  const dirtyCount = changes?.files?.length || 0;

  const close = () => {
    if (showChanges) toggleChanges(false);
    else togglePlan(false);
  };

  return (
    <aside className={"planside" + (showChanges ? " changes-mode" : "")} aria-label={showChanges ? t("Changes") : t("Plan")}>
      <div className="ps-head">
        {hasPlan ? (
          <div className="ps-tabs" role="tablist">
            <button
              type="button"
              role="tab"
              aria-selected={showPlan}
              className={"ps-tab" + (showPlan ? " on" : "")}
              onClick={() => setRightTab("plan")}
            >
              {t("Plan")}
              <span className="ps-badge">
                {done}/{plan.length}
              </span>
            </button>
            <button
              type="button"
              role="tab"
              aria-selected={showChanges}
              className={"ps-tab" + (showChanges ? " on" : "")}
              onClick={() => setRightTab("changes")}
            >
              {t("Changes")}
              {dirtyCount > 0 && <span className="ps-badge">{dirtyCount}</span>}
            </button>
          </div>
        ) : (
          <>
            <span className="ps-title">{t("Changes")}</span>
            {dirtyCount > 0 && <span className="ps-count">{dirtyCount}</span>}
          </>
        )}
        <button type="button" className="icon-btn sm" title={t("Hide the plan")} aria-label={t("Hide the plan")} onClick={close}>
          <Icon name="x" size={14} />
        </button>
      </div>
      {showPlan ? (
        <>
          <div className="ps-bar">
            <span style={{ width: `${(100 * done) / plan.length}%` }} />
          </div>
          <ol className="ps-list">
            {plan.map((it, i) => {
              const st = it.status === "done" ? "done" : it.status === "in_progress" ? "now" : "todo";
              return (
                <li key={i} className={"ps-item " + st}>
                  <span className="ps-mark">
                    {st === "done" ? <Icon name="check" size={11} /> : st === "now" && running ? <span className="spin" /> : null}
                  </span>
                  <span className="ps-text" dir="auto">
                    {it.text}
                  </span>
                </li>
              );
            })}
          </ol>
          <div className="ps-foot">
            {running
              ? current
                ? "Working on step " + (plan.indexOf(current) + 1) + " · " + took(now - started)
                : t("Working") + " · " + took(now - started)
              : done === plan.length
                ? t("All steps done")
                : done + " of " + plan.length + " steps done"}
          </div>
        </>
      ) : (
        <ChangesBody />
      )}
    </aside>
  );
}
