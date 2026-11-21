import { useEffect, useMemo, useState } from "react";
import { checkUpdate, fail, installUpdate } from "../actions";
import { server } from "../api";
import { useT } from "../i18n";
import { Icon } from "../icons";
import { setState, useStore } from "../store";
import type { ConnectionDetail, MemoryEntry, SkillView } from "../types";

export function UpdatesSection() {
  const t = useT();
  const info = useStore((s) => s.info);
  const upd = useStore((s) => s.update);
  const machine = useStore((s) => s.machine);
  const [checking, setChecking] = useState(false);

  if (machine) return null;

  const runCheck = async () => {
    setChecking(true);
    await checkUpdate(true);
    setChecking(false);
  };

  return (
    <>
      <div className="set-group">{t("Updates")}</div>
      <div className="set-list">
        <div className="set-row static upd-row">
          <span className={"state" + (upd?.available ? " warn" : upd?.installed || (upd && !upd.error) ? " ok" : "")} />
          <span className="grow">
            <span className="name">
              Macron v{info.version || upd?.current}
              {upd?.available && upd.latest ? ` → v${upd.latest}` : ""}
            </span>
            <span className={"meta" + (upd?.error ? " bad-text" : "")}>
              {upd?.installing
                ? upd.step || t("Updating…")
                : upd?.installed
                  ? t("Updated — restart Macron")
                  : upd?.error
                    ? upd.error
                    : upd?.available
                      ? t("Update available")
                      : upd
                        ? t("Up to date")
                        : "GitHub Releases · Metixpro/Macron-Agent"}
            </span>
          </span>
          {upd?.available && !upd.installed ? (
            <button type="button" className="btn primary sm" disabled={upd.installing} onClick={() => installUpdate()}>
              {upd.installing ? t("Updating…") : t("Install update")}
            </button>
          ) : (
            <button type="button" className="btn ghost sm" disabled={checking || upd?.installing} onClick={runCheck}>
              {checking ? t("Checking…") : t("Check for updates")}
            </button>
          )}
        </div>
        {upd?.available && upd.notes && (
          <details className="upd-notes">
            <summary>{t("Release notes")}</summary>
            <pre dir="auto">{upd.notes}</pre>
          </details>
        )}
      </div>
    </>
  );
}

export function SkillsSection() {
  const t = useT();
  const [skills, setSkills] = useState<SkillView[] | null>(null);
  const [q, setQ] = useState("");

  const load = () => {
    server
      .skills()
      .then((r) => setSkills(r.skills || []))
      .catch(() => setSkills([]));
  };

  useEffect(() => {
    load();
  }, []);

  const shown = useMemo(() => {
    const f = q.trim().toLowerCase();
    if (!f) return skills || [];
    return (skills || []).filter((s) => (s.name + " " + s.category + " " + s.description).toLowerCase().includes(f));
  }, [skills, q]);

  const toggle = async (sk: SkillView) => {
    setSkills((list) => (list || []).map((x) => (x.name === sk.name ? { ...x, enabled: !sk.enabled } : x)));
    try {
      const res = await server.enableSkill(sk.name, !sk.enabled);
      setSkills(res.skills || []);
    } catch (e) {
      fail(e);
      load();
    }
  };

  const remove = async (sk: SkillView) => {
    try {
      const res = await server.deleteSkill(sk.name);
      setSkills(res.skills || []);
    } catch (e) {
      fail(e);
    }
  };

  return (
    <>
      <div className="set-group">
        {t("Skills")} {skills ? `(${skills.filter((s) => s.enabled).length}/${skills.length})` : ""}
      </div>
      <div className="set-search">
        <Icon name="search" size={13} />
        <input type="text" value={q} placeholder={t("Filter skills…")} onChange={(e) => setQ(e.target.value)} />
      </div>
      <div className="set-list scroll-box">
        {skills === null && <div className="empty-row">{t("Loading…")}</div>}
        {skills && !shown.length && <div className="empty-row">{t("No skills loaded.")}</div>}
        {shown.map((sk) => (
          <div key={sk.name} className="set-row static">
            <span className={"state" + (sk.enabled ? " ok" : "")} />
            <span className="grow">
              <span className="name" dir="ltr">
                {sk.name} <span className="tag">{sk.category}</span>
              </span>
              <span className="meta" dir="auto">
                {sk.description}
              </span>
            </span>
            <button
              type="button"
              role="switch"
              aria-checked={sk.enabled}
              className={"switch" + (sk.enabled ? " on" : "")}
              onClick={() => toggle(sk)}
            >
              <span />
            </button>
            <button type="button" className="icon-btn sm" title={t("Remove")} onClick={() => remove(sk)}>
              <Icon name="x" size={13} />
            </button>
          </div>
        ))}
      </div>
    </>
  );
}

export function MemorySection() {
  const t = useT();
  const [items, setItems] = useState<MemoryEntry[] | null>(null);
  const [scope, setScope] = useState<"memory" | "user">("memory");
  const [draft, setDraft] = useState("");
  const [saving, setSaving] = useState(false);

  const load = () => {
    server
      .memories()
      .then((r) => setItems(r.memories || []))
      .catch(() => setItems([]));
  };

  useEffect(() => {
    load();
  }, []);

  const filtered = useMemo(() => (items || []).filter((m) => m.scope === scope), [items, scope]);

  const add = async () => {
    const text = draft.trim();
    if (!text) return;
    setSaving(true);
    try {
      await server.addMemory(scope, text);
      setDraft("");
      load();
    } catch (e) {
      fail(e);
    } finally {
      setSaving(false);
    }
  };

  const del = async (entry: MemoryEntry) => {
    try {
      await server.deleteMemory(entry.scope, entry.id);
      load();
    } catch (e) {
      fail(e);
    }
  };

  return (
    <>
      <div className="set-group">{t("Memory")}</div>
      <div className="mem-tabs">
        <button type="button" className={"mem-tab" + (scope === "memory" ? " on" : "")} onClick={() => setScope("memory")}>
          {t("Agent memory")}
        </button>
        <button type="button" className={"mem-tab" + (scope === "user" ? " on" : "")} onClick={() => setScope("user")}>
          {t("About you")}
        </button>
      </div>
      <div className="mem-add">
        <input
          type="text"
          dir="auto"
          value={draft}
          placeholder={t("Add a memory note…")}
          onChange={(e) => setDraft(e.target.value)}
          onKeyDown={(e) => {
            if (e.key === "Enter") {
              e.preventDefault();
              void add();
            }
          }}
        />
        <button type="button" className="btn primary sm" disabled={saving || !draft.trim()} onClick={add}>
          {t("Add")}
        </button>
      </div>
      <div className="set-list scroll-box">
        {items === null && <div className="empty-row">{t("Loading…")}</div>}
        {items && !filtered.length && <div className="empty-row">{t("No saved memories yet.")}</div>}
        {filtered.map((m) => (
          <div key={m.scope + ":" + m.id} className="set-row static">
            <span className="grow">
              <span className="mem-text" dir="auto">
                {m.text}
              </span>
            </span>
            <button type="button" className="icon-btn sm" title={t("Remove")} onClick={() => del(m)}>
              <Icon name="x" size={13} />
            </button>
          </div>
        ))}
      </div>
    </>
  );
}

export function ConnectionsSection() {
  const t = useT();
  const [list, setList] = useState<ConnectionDetail[] | null>(null);
  const [openId, setOpenId] = useState("");
  const [values, setValues] = useState<Record<string, string>>({});
  const [saving, setSaving] = useState(false);

  const load = () => {
    server
      .connectionDetails()
      .then((r) => setList(r.connections || []))
      .catch(() => setList([]));
  };

  useEffect(() => {
    load();
  }, []);

  const toggleOpen = (id: string) => {
    if (openId === id) {
      setOpenId("");
      setValues({});
    } else {
      setOpenId(id);
      setValues({});
    }
  };

  const save = async (c: ConnectionDetail) => {
    const patch: Record<string, string> = {};
    for (const f of c.fields) {
      if (Object.prototype.hasOwnProperty.call(values, f.key)) {
        patch[f.key] = values[f.key];
      }
    }
    if (!Object.keys(patch).length) {
      setOpenId("");
      return;
    }
    setSaving(true);
    try {
      const res = await server.saveConnections(patch);
      setList(res.connections || []);
      if (res.info) setState({ info: res.info });
      setOpenId("");
      setValues({});
    } catch (e) {
      fail(e);
    } finally {
      setSaving(false);
    }
  };

  return (
    <>
      <div className="set-group">{t("Connections")}</div>
      <div className="set-list">
        {list === null && <div className="empty-row">{t("Loading…")}</div>}
        {(list || []).map((c) => {
          const isOpen = openId === c.id;
          return (
            <div key={c.id} className="conn-wrap">
              <div className="set-row static">
                <span className={"state" + (c.ready ? " ok" : "")} />
                <span className="grow">
                  <span className="name">{c.name}</span>
                  {c.capability && <span className="meta">{c.capability}</span>}
                </span>
                <span className="meta">{c.ready ? t("Connected") : t("Not set up")}</span>
                <button type="button" className="btn ghost sm" onClick={() => toggleOpen(c.id)}>
                  {isOpen ? t("Close") : t("Configure")}
                </button>
              </div>
              {isOpen && (
                <div className="conn-form">
                  {c.fields.map((f) => (
                    <label key={f.key} className="conn-field">
                      <span className="conn-key" dir="ltr">
                        {f.key} {f.set && <em className="set-ok">✓</em>}
                      </span>
                      <input
                        type={f.secret ? "password" : "text"}
                        dir="ltr"
                        placeholder={f.set ? "•••••••• (leave blank to keep, or enter new value)" : f.default || f.purpose}
                        value={values[f.key] ?? ""}
                        onChange={(e) => setValues((prev) => ({ ...prev, [f.key]: e.target.value }))}
                      />
                      {f.purpose && <span className="conn-hint">{f.purpose}</span>}
                    </label>
                  ))}
                  <div className="conn-actions">
                    <button type="button" className="btn primary sm" disabled={saving} onClick={() => save(c)}>
                      {saving ? t("Saving…") : t("Save")}
                    </button>
                  </div>
                </div>
              )}
            </div>
          );
        })}
      </div>
    </>
  );
}
