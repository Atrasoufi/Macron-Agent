import { useEffect, useState } from "react";
import { loadChanges, revertFile } from "../actions";
import { server } from "../api";
import { useT } from "../i18n";
import { Icon } from "../icons";
import { useStore } from "../store";
import type { DiffResponse, GitChange } from "../types";

function DiffLines({ diff }: { diff: string }) {
  const lines = diff.split("\n");
  return (
    <pre className="diff-code" dir="ltr">
      {lines.map((line, i) => {
        let cls = "dl";
        if (line.startsWith("+++") || line.startsWith("---")) cls += " meta";
        else if (line.startsWith("@@")) cls += " hunk";
        else if (line.startsWith("+")) cls += " add";
        else if (line.startsWith("-")) cls += " del";
        return (
          <span key={i} className={cls}>
            {line || " "}
          </span>
        );
      })}
    </pre>
  );
}

export function ChangesBody() {
  const t = useT();
  const current = useStore((s) => s.current);
  const running = useStore((s) => s.live.running);
  const changes = useStore((s) => s.changes);
  const [selected, setSelected] = useState("");
  const [diff, setDiff] = useState<DiffResponse | null>(null);
  const [loadingDiff, setLoadingDiff] = useState(false);
  const [confirmRevert, setConfirmRevert] = useState("");
  const [busy, setBusy] = useState(false);

  const files = changes?.files || [];

  useEffect(() => {
    setSelected("");
    setDiff(null);
    setConfirmRevert("");
  }, [current]);

  useEffect(() => {
    if (!selected && files.length > 0) {
      setSelected(files[0].path);
    } else if (selected && !files.some((f) => f.path === selected)) {
      setSelected(files[0]?.path || "");
    }
  }, [files, selected]);

  useEffect(() => {
    if (!current || !selected) {
      setDiff(null);
      return;
    }
    let alive = true;
    setLoadingDiff(true);
    server
      .diff(current, selected)
      .then((res) => {
        if (alive) setDiff(res);
      })
      .catch(() => {
        if (alive) setDiff(null);
      })
      .finally(() => {
        if (alive) setLoadingDiff(false);
      });
    return () => {
      alive = false;
    };
  }, [current, selected, changes]);

  const doRevert = async (path: string) => {
    setBusy(true);
    const ok = await revertFile(path);
    setBusy(false);
    if (ok) {
      setConfirmRevert("");
      if (selected === path) setDiff(null);
    }
  };

  if (!changes) {
    return <div className="ch-empty">{t("Loading…")}</div>;
  }

  if (!changes.git) {
    return (
      <div className="ch-empty">
        <Icon name="gitBranch" size={20} />
        <span>{t("No git repository in this folder")}</span>
      </div>
    );
  }

  return (
    <div className="ch-wrap">
      <div className="ch-bar">
        <span className="ch-branch" dir="ltr">
          <Icon name="gitBranch" size={12} />
          {changes.branch || "HEAD"}
        </span>
        <span className="grow" />
        <button type="button" className="icon-btn sm" title={t("Refresh")} onClick={() => loadChanges(current)}>
          <Icon name="refresh" size={13} />
        </button>
      </div>
      {files.length === 0 ? (
        <div className="ch-empty">
          <Icon name="check" size={18} />
          <span>{t("Working tree is clean")}</span>
        </div>
      ) : (
        <>
          <div className="ch-files">
            {files.map((f: GitChange) => (
              <div key={f.path} className={"ch-row" + (selected === f.path ? " on" : "")}>
                <button type="button" className="ch-hit" onClick={() => setSelected(f.path)} title={f.path}>
                  <span className={"ch-st st-" + (f.status === "?" ? "u" : f.status.toLowerCase())}>{f.status}</span>
                  <span className="ch-path" dir="ltr">
                    {f.path}
                  </span>
                  <span className="ch-stat" dir="ltr">
                    {f.added > 0 && <em className="add">+{f.added}</em>}
                    {f.deleted > 0 && <em className="del">-{f.deleted}</em>}
                  </span>
                </button>
                {confirmRevert === f.path ? (
                  <span className="ch-confirm">
                    <button type="button" className="ch-btn bad" disabled={busy || running} onClick={() => doRevert(f.path)}>
                      {t("Revert")}
                    </button>
                    <button type="button" className="ch-btn" onClick={() => setConfirmRevert("")}>
                      {t("Cancel")}
                    </button>
                  </span>
                ) : (
                  <button
                    type="button"
                    className="icon-btn sm ch-rev"
                    disabled={running}
                    title={t("Revert")}
                    onClick={() => setConfirmRevert(f.path)}
                  >
                    <Icon name="rotateCcw" size={12} />
                  </button>
                )}
              </div>
            ))}
          </div>
          <div className="ch-diff">
            {!selected ? (
              <div className="ch-diff-empty">{t("Select a file to inspect diff")}</div>
            ) : loadingDiff ? (
              <div className="ch-diff-empty">{t("Loading diff…")}</div>
            ) : diff && diff.diff ? (
              <DiffLines diff={diff.diff} />
            ) : (
              <div className="ch-diff-empty">{t("No diff available")}</div>
            )}
          </div>
        </>
      )}
    </div>
  );
}
