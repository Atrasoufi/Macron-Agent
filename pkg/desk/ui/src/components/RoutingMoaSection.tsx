import { useEffect, useState } from "react";
import { errorText, fail, refreshState } from "../actions";
import { server } from "../api";
import { useT } from "../i18n";
import { Icon } from "../icons";
import type { MoaState, ProviderDetail, RoutingConfig, RoutingTestResult } from "../types";

export function RoutingMoaSection() {
  const t = useT();
  const [providers, setProviders] = useState<ProviderDetail[]>([]);
  const [routing, setRouting] = useState<RoutingConfig | null>(null);
  const [testText, setTestText] = useState("");
  const [testRes, setTestRes] = useState<RoutingTestResult | null>(null);
  const [testing, setTesting] = useState(false);

  const [moa, setMoa] = useState<MoaState | null>(null);
  const [addingMoa, setAddingMoa] = useState(false);
  const [mName, setMName] = useState("");
  const [mAgg, setMAgg] = useState("");
  const [mRefs, setMRefs] = useState<string[]>([]);
  const [mFanout, setMFanout] = useState("user_turn");
  const [mPrivacy, setMPrivacy] = useState("");
  const [mErr, setMErr] = useState("");
  const [savingMoa, setSavingMoa] = useState(false);

  const load = () => {
    server.providers().then((r) => setProviders((r.providers || []).filter((p) => p.enabled))).catch(() => setProviders([]));
    server.routing().then(setRouting).catch(() => setRouting(null));
    server.moa().then(setMoa).catch(() => setMoa(null));
  };

  useEffect(() => {
    load();
  }, []);

  const updateRouting = async (patch: Parameters<typeof server.saveRouting>[0]) => {
    try {
      const next = await server.saveRouting(patch);
      setRouting(next);
      await refreshState();
    } catch (e) {
      fail(e);
    }
  };

  const toggleKind = (kind: string) => {
    if (!routing) return;
    const set = new Set(routing.kinds || []);
    if (set.has(kind)) set.delete(kind);
    else set.add(kind);
    void updateRouting({ kinds: Array.from(set), kindsSet: true });
  };

  const runRouteTest = async () => {
    if (!testText.trim()) return;
    setTesting(true);
    try {
      const r = await server.testRouting(testText.trim());
      setTestRes(r);
    } catch (e) {
      fail(e);
    } finally {
      setTesting(false);
    }
  };

  const toggleRef = (name: string) => {
    setMRefs((prev) => (prev.includes(name) ? prev.filter((x) => x !== name) : [...prev, name]));
  };

  const savePreset = async () => {
    setMErr("");
    if (!mName.trim() || !mAgg || mRefs.length === 0) {
      setMErr(t("Enter a preset name, pick an aggregator, and choose at least one reference model."));
      return;
    }
    setSavingMoa(true);
    try {
      const st = await server.saveMoaPreset({
        name: mName.trim(),
        aggregator: { provider: mAgg },
        references: mRefs.map((p) => ({ provider: p })),
        fanout: mFanout,
        privacyFilter: mPrivacy,
        enabled: true,
      });
      setMoa(st);
      setAddingMoa(false);
      setMName("");
      setMRefs([]);
      await refreshState();
    } catch (e) {
      setMErr(errorText(e));
    } finally {
      setSavingMoa(false);
    }
  };

  const setMoaActive = async (name: string) => {
    try {
      const st = await server.updateMoaConfig({ active: name });
      setMoa(st);
      await refreshState();
    } catch (e) {
      fail(e);
    }
  };

  const setMoaDefault = async (name: string) => {
    try {
      const st = await server.updateMoaConfig({ defaultPreset: name });
      setMoa(st);
    } catch (e) {
      fail(e);
    }
  };

  const delPreset = async (name: string) => {
    try {
      const st = await server.deleteMoaPreset(name);
      setMoa(st);
      await refreshState();
    } catch (e) {
      fail(e);
    }
  };

  return (
    <>
      <div className="set-group">{t("Smart model routing")}</div>
      {routing && (
        <div className="rm-box">
          <div className="rm-row">
            <div className="grow">
              <div className="set-title">
                {t("Automatic light / strong routing")}
                {routing.active ? (
                  <span className="pv-pill active">{t("Active")}</span>
                ) : (
                  <span className="pv-pill muted">{t("Inactive")}</span>
                )}
              </div>
              <div className="set-help">
                {t("Route simple prompts to your active model and heavy tasks (code, reasoning, research) to a strong model.")}
              </div>
            </div>
            <button
              type="button"
              role="switch"
              aria-checked={routing.enabled}
              className={"switch" + (routing.enabled ? " on" : "")}
              onClick={() => void updateRouting({ enabled: !routing.enabled })}
            >
              <span />
            </button>
          </div>

          <div className="rm-grid">
            <label className="field">
              <span>{t("Strong model provider")}</span>
              <select
                className="rm-select"
                value={String(routing.strongProviderId || 0)}
                onChange={(e) => void updateRouting({ strongProviderId: Number(e.target.value) })}
              >
                <option value="0">{t("None (use active model only)")}</option>
                {providers.map((p) => (
                  <option key={p.id} value={String(p.id)}>
                    {p.name} — {p.model}
                  </option>
                ))}
              </select>
            </label>
            <label className="field">
              <span>{t("Routing sensitivity")}</span>
              <div className="choice" role="radiogroup">
                {(
                  [
                    ["cheap", t("Cheap")],
                    ["balanced", t("Balanced")],
                    ["quality", t("Quality")],
                  ] as const
                ).map(([val, label]) => (
                  <button
                    key={val}
                    type="button"
                    role="radio"
                    aria-checked={routing.sensitivity === val}
                    className={routing.sensitivity === val ? "on" : ""}
                    onClick={() => void updateRouting({ sensitivity: val })}
                  >
                    {label}
                  </button>
                ))}
              </div>
            </label>
          </div>

          <div className="rm-kinds-wrap">
            <div className="rm-label">{t("Send these task kinds to the strong model")}</div>
            <div className="rm-kinds" dir="ltr">
              {(routing.allKinds || []).map((k) => {
                const on = (routing.kinds || []).includes(k);
                return (
                  <button key={k} type="button" className={"pv-qpill" + (on ? " on" : "")} onClick={() => toggleKind(k)}>
                    {k}
                  </button>
                );
              })}
            </div>
          </div>

          <div className="rm-test">
            <input
              type="text"
              placeholder={t("Test routing with a sample prompt…")}
              value={testText}
              onChange={(e) => setTestText(e.target.value)}
              onKeyDown={(e) => e.key === "Enter" && (e.preventDefault(), void runRouteTest())}
            />
            <button type="button" className="btn ghost sm" disabled={testing || !testText.trim()} onClick={() => void runRouteTest()}>
              {testing ? t("Testing…") : t("Test")}
            </button>
          </div>
          {testRes && (
            <div className={"pv-test-res " + (testRes.strong ? "ok" : "")} dir="ltr">
              {testRes.strong ? "→ Strong" : "→ Regular"}
              {testRes.targetProvider ? ` (${testRes.targetProvider})` : ""} · kind: {testRes.kind}
              {testRes.hard ? " · hard" : ""}
            </div>
          )}
        </div>
      )}

      <div className="set-group">{t("Multi-model voting (MoA)")}</div>
      <div className="set-list">
        {moa && moa.presets.length === 0 && !addingMoa && (
          <div className="empty-row">{t("No MoA presets yet. Combine multiple models to vote and synthesize better answers.")}</div>
        )}
        {(moa?.presets || []).map((p) => {
          const isActive = moa?.active === p.name;
          const isDef = moa?.defaultPreset === p.name;
          const hasProbs = (p.problems || []).length > 0;
          return (
            <div key={p.name} className={"pv-card" + (isActive ? " on" : "")}>
              <div className="pv-top">
                <div className="pv-ident">
                  <div className="pv-name-row">
                    <span className="name">{p.name}</span>
                    {isActive && <span className="pv-pill active">{t("Always on")}</span>}
                    {isDef && <span className="pv-pill api">{t("Default")}</span>}
                    {hasProbs && <span className="pv-pill warn">{t("Needs setup")}</span>}
                  </div>
                  <div className="pv-chips moa-chips" dir="ltr">
                    <span className="pv-chip">agg: {p.aggregator.provider}</span>
                    <span className="pv-chip">refs: {p.references.map((r) => r.provider).join(", ")}</span>
                    <span className="pv-chip">{p.fanout}</span>
                  </div>
                </div>
                <div className="pv-actions">
                  {!isDef && (
                    <button type="button" className="btn ghost sm" onClick={() => void setMoaDefault(p.name)}>
                      {t("Set default")}
                    </button>
                  )}
                  <button
                    type="button"
                    className={"btn sm " + (isActive ? "primary" : "ghost")}
                    disabled={hasProbs}
                    onClick={() => void setMoaActive(isActive ? "" : p.name)}
                  >
                    {isActive ? t("Turn off") : t("Always on")}
                  </button>
                  <button type="button" className="icon-btn" title={t("Remove")} onClick={() => void delPreset(p.name)}>
                    <Icon name="x" size={14} />
                  </button>
                </div>
              </div>
              {hasProbs && (
                <div className="pv-test-res bad" dir="ltr">
                  {(p.problems || []).join(" · ")}
                </div>
              )}
            </div>
          );
        })}

        {addingMoa ? (
          <div className="conn-form">
            <div className="pv-grid2">
              <label className="field">
                <span>{t("Preset name")}</span>
                <input placeholder="e.g. Council" value={mName} onChange={(e) => setMName(e.target.value)} />
              </label>
              <label className="field">
                <span>{t("Aggregator model")}</span>
                <select className="rm-select" value={mAgg} onChange={(e) => setMAgg(e.target.value)}>
                  <option value="">{t("Choose provider…")}</option>
                  {providers.map((p) => (
                    <option key={p.id} value={p.name}>
                      {p.name} ({p.model})
                    </option>
                  ))}
                </select>
              </label>
            </div>
            <div className="rm-kinds-wrap">
              <div className="rm-label">{t("Reference models (advisors)")}</div>
              <div className="rm-kinds" dir="ltr">
                {providers.map((p) => {
                  const on = mRefs.includes(p.name);
                  return (
                    <button key={p.id} type="button" className={"pv-qpill" + (on ? " on" : "")} onClick={() => toggleRef(p.name)}>
                      {p.name}
                    </button>
                  );
                })}
              </div>
            </div>
            <div className="pv-grid2">
              <label className="field">
                <span>{t("Fanout mode")}</span>
                <select className="rm-select" value={mFanout} onChange={(e) => setMFanout(e.target.value)}>
                  <option value="user_turn">Once per user turn</option>
                  <option value="per_iteration">Every reasoning step</option>
                </select>
              </label>
              <label className="field">
                <span>{t("Privacy filter")}</span>
                <select className="rm-select" value={mPrivacy} onChange={(e) => setMPrivacy(e.target.value)}>
                  <option value="">None</option>
                  <option value="display">Scrub displayed advice</option>
                  <option value="full">Scrub prompts & advice</option>
                </select>
              </label>
            </div>
            {mErr && <div className="form-error">{mErr}</div>}
            <div className="soul-bar">
              <button type="button" className="btn ghost sm" onClick={() => setAddingMoa(false)}>
                {t("Cancel")}
              </button>
              <button type="button" className="btn primary sm" disabled={savingMoa} onClick={() => void savePreset()}>
                {savingMoa ? t("Saving…") : t("Create preset")}
              </button>
            </div>
          </div>
        ) : (
          <button
            type="button"
            className="set-row"
            onClick={() => {
              setAddingMoa(true);
              if (!mAgg && providers[0]) setMAgg(providers[0].name);
            }}
          >
            <Icon name="plus" size={14} />
            <span className="grow">
              <span className="name">{t("New MoA voting preset…")}</span>
            </span>
          </button>
        )}
      </div>
    </>
  );
}
