import { useEffect, useRef, useState } from 'react';
import { api } from '@/api/client';
import { useAppState } from '@/lib/app-state';
import { useConfirm } from '@/components/confirm-dialog';
import { Button } from '@/components/ui/button';
import { Save, Trash2, X, FileDown } from 'lucide-react';
import { MarkdownReport } from '@/editor/MarkdownReport';
import { VisualCompareView } from '@/editor/VisualCompareView';
import { ReportSearchBar, useReportSearch } from '@/editor/ReportSearch';

function usesAI(detail) {
    return !!(detail?.uses_ai ?? detail?.usesAI);
}

function usesMerge(detail) {
    return !!(detail?.uses_merge ?? detail?.usesMerge);
}

function progressLabel(status, step, stepTotal) {
    if (status === 'queued')
        return 'Queued…';
    if (status === 'running' && stepTotal > 1)
        return `Step ${step} of ${stepTotal}…`;
    if (status === 'running')
        return 'Running…';
    return 'Running…';
}

function sleep(ms) {
    return new Promise((resolve) => setTimeout(resolve, ms));
}

export function AnalysisPane() {
    const { analysisId, analysisQueue, setAnalysis, selection, bumpRefresh } = useAppState();
    const [detail, setDetail] = useState(null);
    const [queueLabels, setQueueLabels] = useState([]);
    const [busy, setBusy] = useState(false);
    const [progress, setProgress] = useState('');
    const [body, setBody] = useState('');
    const [mergeOriginal, setMergeOriginal] = useState('');
    const [mergeProposed, setMergeProposed] = useState('');
    const [view, setView] = useState('report');
    const [saved, setSaved] = useState(false);
    const [notice, setNotice] = useState('');
    const [sources, setSources] = useState(null);
    const searchApiRef = useRef(null);
    const search = useReportSearch(searchApiRef);
    const projectPath = selection?.dir || '';
    useEffect(() => {
        if (!analysisId) {
            setDetail(null);
            setQueueLabels([]);
            return;
        }
        setBody('');
        setMergeOriginal('');
        setMergeProposed('');
        setView('report');
        setNotice('');
        setProgress('');
        let cancelled = false;
        void api.getAnalysis(analysisId).then((d) => {
            if (!cancelled)
                setDetail(d);
        }).catch(() => {
            if (!cancelled)
                setDetail(null);
        });
        if (projectPath) {
            void api.matchAnalysisSources(projectPath).then((s) => {
                if (!cancelled)
                    setSources(s);
            }).catch(() => {
                if (!cancelled)
                    setSources(null);
            });
        } else {
            setSources(null);
        }
        return () => {
            cancelled = true;
        };
    }, [analysisId, projectPath]);
    useEffect(() => {
        const ids = analysisQueue?.length ? analysisQueue : (analysisId ? [analysisId] : []);
        if (!ids.length) {
            setQueueLabels([]);
            return;
        }
        let cancelled = false;
        void Promise.all(ids.map((id) => api.getAnalysis(id).then((d) => d.label || id).catch(() => id))).then((labels) => {
            if (!cancelled)
                setQueueLabels(labels);
        });
        return () => {
            cancelled = true;
        };
    }, [analysisQueue, analysisId]);
    if (!detail) {
        return <div className="p-6 text-sm text-muted-foreground">Loading…</div>;
    }
    const ai = usesAI(detail);
    const mergeFlag = usesMerge(detail);
    const queue = (analysisQueue?.length ? analysisQueue : [detail.id]);
    function applyReport(report) {
        const orig = report.original || report.Original || '';
        const prop = report.proposed || report.Proposed || '';
        const text = report.body || '';
        setBody(text);
        setMergeOriginal(orig);
        setMergeProposed(prop);
        setView(usesMerge(detail) || (orig && prop) ? 'merge' : 'report');
        setSaved(false);
    }
    async function runOne(id) {
        const info = await api.getAnalysis(id);
        if (usesAI(info)) {
            let job = await api.startAnalysisJob(id, projectPath);
            if (!(job.cached || job.status === 'done')) {
                const jobID = job.job_id || job.jobId;
                while (job.status !== 'done' && job.status !== 'failed') {
                    setProgress(progressLabel(job.status, job.step, job.step_total || job.stepTotal));
                    await sleep(2000);
                    job = await api.getAnalysisJobStatus(jobID);
                }
                if (job.status === 'failed') {
                    throw new Error(job.error || 'Analysis failed');
                }
            }
            const bodyText = job.body || '';
            const orig = job.original || job.Original || '';
            const prop = job.proposed || job.Proposed || '';
            return api.persistAnalysisResult(id, projectPath, bodyText, orig, prop, '');
        }
        return api.runAnalysis(id, projectPath);
    }
    async function run() {
        setBusy(true);
        setNotice('');
        setProgress('');
        try {
            let primary = null;
            for (let i = 0; i < queue.length; i++) {
                const id = queue[i];
                const label = queueLabels[i] || id;
                setProgress(queue.length > 1 ? `${label} (${i + 1}/${queue.length})…` : 'Running…');
                const report = await runOne(id);
                if (id === detail.id)
                    primary = report;
            }
            if (primary)
                applyReport(primary);
        }
        catch (err) {
            const msg = err instanceof Error ? err.message : 'Could not run analysis';
            if (msg.includes('PLAN_REQUIRED')) {
                setNotice('This analysis uses AI. Choose a plan / add credits.');
            } else if (msg.includes('CREDITS_EMPTY')) {
                setNotice('This month’s credits are used up.');
            } else if (msg.includes('CONCURRENT_LIMIT')) {
                setNotice('Another analysis is already running. Wait or upgrade your plan.');
            } else {
                setNotice(msg);
            }
        }
        finally {
            setBusy(false);
            setProgress('');
        }
    }
    async function saveReport() {
        // Compare-only analyses persist sides only — no report body.
        const saveBody = mergeFlag ? '' : body;
        await api.saveAnalysisJobReport(detail.id, projectPath, saveBody, mergeOriginal, mergeProposed);
        setSaved(true);
        bumpRefresh();
    }
    async function exportDocx() {
        setNotice('');
        try {
            const path = await api.exportMarkdownDocx(body, 'vellum-prep.docx');
            if (path)
                setNotice(`Exported to ${path}`);
        }
        catch (err) {
            setNotice(err instanceof Error ? err.message : 'Export failed');
        }
    }
    const hasMerge = !!(mergeOriginal && mergeProposed);
    const hasOutput = mergeFlag ? hasMerge : !!body;
    const isVellum = detail.id === 'vellum_prep';
    const busyLabel = progress || (busy ? 'Running…' : (queue.length > 1 ? `Run ${queue.length}` : 'Run'));
    return (<div className="flex h-full flex-col">
      <div className="border-b border-border">
        <div className="flex items-center gap-2 px-4 py-2">
          <div className="min-w-0 shrink-0">
            <div className="flex items-center gap-2">
              <div className="truncate text-sm font-medium">{detail.label}</div>
              <span className="shrink-0 rounded border border-border px-1.5 py-0.5 text-[10px] uppercase text-muted-foreground">
                {ai ? 'Uses AI' : 'Local'}
              </span>
              {mergeFlag && (
                <span className="shrink-0 rounded border border-border px-1.5 py-0.5 text-[10px] uppercase text-muted-foreground">
                  Compare
                </span>
              )}
            </div>
            <div className="truncate text-[11px] text-muted-foreground">{detail.group}</div>
          </div>
          {hasOutput && !mergeFlag && view === 'report' && (
            <ReportSearchBar
              query={search.query}
              setQuery={search.setQuery}
              matchCount={search.matchCount}
              activeIndex={search.activeIndex}
              onPrev={() => search.step(-1)}
              onNext={() => search.step(1)}
            />
          )}
          {hasOutput && isVellum && (
            <Button size="sm" variant="outline" onClick={() => void exportDocx()} className="shrink-0 gap-1.5">
              <FileDown className="h-3.5 w-3.5"/>Export DOCX
            </Button>
          )}
          {hasOutput && !saved && (
            <Button size="sm" variant="outline" onClick={() => void saveReport()} className="shrink-0 gap-1.5">
              <Save className="h-3.5 w-3.5"/>Save
            </Button>
          )}
          {hasOutput && saved && (
            <span className="shrink-0 text-xs text-green-600 dark:text-green-400">Saved</span>
          )}
          <Button size="sm" className="shrink-0" onClick={() => void run()} disabled={busy}>
            {busyLabel}
          </Button>
          <Button size="icon" variant="ghost" className="shrink-0" onClick={() => { setBody(''); setMergeOriginal(''); setMergeProposed(''); setAnalysis(''); }} title="Close">
            <X className="h-4 w-4"/>
          </Button>
        </div>
      </div>
      <div className="flex-1 overflow-hidden">
        {hasOutput ? (
          <>
            {notice && !busy && (
              <div className="border-b border-border px-4 py-1.5 text-xs text-muted-foreground">{notice}</div>
            )}
            {mergeFlag || (view === 'merge' && hasMerge) ? (
              <VisualCompareView original={mergeOriginal} proposed={mergeProposed}/>
            ) : (
              <MarkdownReport
                markdown={body}
                searchQuery={search.query}
                searchActiveIndex={search.activeIndex}
                setSearchActiveIndex={search.setActiveIndex}
                onSearchMatchCount={search.onMatchCount}
                searchApiRef={searchApiRef}
              />
            )}
          </>
        ) : (<div className="overflow-auto p-6">
          <p className="max-w-2xl text-sm leading-relaxed">{detail.description || 'No description.'}</p>
          {mergeFlag && <p className="mt-2 max-w-2xl text-xs text-muted-foreground">Run to open a Compare view of the manuscript with findings marked in place. No separate report is created.</p>}
          {ai && <p className="mt-2 max-w-2xl text-xs text-muted-foreground">This analysis uses AI and needs a plan or credits.</p>}
          {queue.length > 1 && (
            <div className="mt-4 max-w-2xl">
              <div className="mb-1 text-[10px] font-semibold uppercase text-muted-foreground">Run queue</div>
              <ol className="list-decimal space-y-0.5 pl-4 text-sm text-muted-foreground">
                {queueLabels.map((label, i) => (
                  <li key={queue[i] || label} className={queue[i] === detail.id ? 'text-foreground' : ''}>{label}</li>
                ))}
              </ol>
            </div>
          )}
          {notice && (<div className="mt-3 max-w-2xl space-y-2">
            <p className="text-sm text-destructive">{notice}</p>
            {ai && notice.includes('AI') && (<div className="flex gap-2">
              <Button size="sm" variant="outline" onClick={() => void api.connectCloudAccount('').then(() => setNotice('Connected. Try Run again.')).catch((e) => setNotice(e instanceof Error ? e.message : 'Connect failed'))}>Connect for AI</Button>
              <Button size="sm" variant="outline" onClick={() => void api.openBillingCheckout().catch((e) => setNotice(e instanceof Error ? e.message : 'Checkout failed'))}>Choose a plan</Button>
            </div>)}
          </div>)}
          {(detail.needs || []).length > 0 && (<div className="mt-6">
            <div className="mb-2 text-[10px] font-semibold uppercase text-muted-foreground">Uses</div>
            <ul className="space-y-1 text-sm">
              {(detail.needs || []).map((n) => {
                  const role = (sources?.roles || []).find((r) => r.role === n);
                  const ok = role?.present && (role.files || []).length > 0;
                  return (<li key={n} className="text-muted-foreground">{n}{projectPath ? (ok ? ' — ready' : ' — missing') : ''}</li>);
              })}
            </ul>
            {!projectPath && <p className="mt-2 text-xs text-muted-foreground">Select a book or series folder in Projects first.</p>}
          </div>)}
        </div>)}
      </div>
    </div>);
}

export function ReportPane() {
    const { reportId, setReport, bumpRefresh } = useAppState();
    const confirm = useConfirm();
    const [report, setData] = useState(null);
    const [view, setView] = useState('report');
    useEffect(() => {
        if (!reportId) {
            setData(null);
            return;
        }
        setView('report');
        let cancelled = false;
        void api.getAnalysisReport(reportId).then((r) => {
            if (cancelled)
                return;
            setData(r);
            const orig = r.original || r.Original || '';
            const prop = r.proposed || r.Proposed || '';
            if (orig && prop)
                setView('merge');
        }).catch(() => {
            if (!cancelled)
                setData(null);
        });
        return () => {
            cancelled = true;
        };
    }, [reportId]);
    const searchApiRef = useRef(null);
    const search = useReportSearch(searchApiRef);
    async function handleDelete() {
        if (!(await confirm('Delete this report?')))
            return;
        await api.deleteAnalysisReport(reportId);
        setReport(0);
        bumpRefresh();
    }
    if (!report) {
        return <div className="p-6 text-sm text-muted-foreground">Loading…</div>;
    }
    const ai = !!(report.uses_ai ?? report.usesAI);
    const mergeOriginal = report.original || report.Original || '';
    const mergeProposed = report.proposed || report.Proposed || '';
    const hasMerge = !!(mergeOriginal && mergeProposed);
    const compareOnly = hasMerge && !String(report.body || '').trim();
    const isVellum = (report.analysis_id || report.analysisId) === 'vellum_prep';
    async function exportDocx() {
        try {
            await api.exportMarkdownDocx(report.body || '', 'vellum-prep.docx');
        }
        catch {
            /* ignore cancel */
        }
    }
    return (<div className="flex h-full flex-col">
      <div className="border-b border-border">
        <div className="flex items-center gap-2 px-4 py-2">
          <div className="min-w-0 shrink-0">
            <div className="flex items-center gap-2">
              <div className="truncate text-sm font-medium">{report.analysis_label || report.analysisLabel}</div>
              <span className="shrink-0 rounded border border-border px-1.5 py-0.5 text-[10px] uppercase text-muted-foreground">
                {ai ? 'Uses AI' : 'Local'}
              </span>
              {compareOnly && (
                <span className="shrink-0 rounded border border-border px-1.5 py-0.5 text-[10px] uppercase text-muted-foreground">
                  Compare
                </span>
              )}
            </div>
            <div className="truncate text-[11px] text-muted-foreground">{report.created_at || report.createdAt}</div>
          </div>
          {!compareOnly && view === 'report' && (
            <ReportSearchBar
              query={search.query}
              setQuery={search.setQuery}
              matchCount={search.matchCount}
              activeIndex={search.activeIndex}
              onPrev={() => search.step(-1)}
              onNext={() => search.step(1)}
            />
          )}
          {isVellum && !compareOnly && (
            <Button size="sm" variant="outline" onClick={() => void exportDocx()} className="shrink-0 gap-1.5">
              <FileDown className="h-3.5 w-3.5"/>Export DOCX
            </Button>
          )}
          {hasMerge && !compareOnly && (
            <div className="flex shrink-0 rounded border border-border text-[10px]">
              <button
                type="button"
                className={`px-2 py-1 ${view === 'merge' ? 'bg-muted text-foreground' : 'text-muted-foreground'}`}
                onClick={() => setView('merge')}
              >
                Compare
              </button>
              <button
                type="button"
                className={`px-2 py-1 ${view === 'report' ? 'bg-muted text-foreground' : 'text-muted-foreground'}`}
                onClick={() => setView('report')}
              >
                Report
              </button>
            </div>
          )}
          <Button size="icon" variant="ghost" className="shrink-0 text-destructive hover:text-destructive" onClick={() => void handleDelete()} title="Delete report">
            <Trash2 className="h-4 w-4"/>
          </Button>
          <Button size="icon" variant="ghost" className="shrink-0" onClick={() => setReport(0)} title="Close">
            <X className="h-4 w-4"/>
          </Button>
        </div>
      </div>
      <div className="flex-1 overflow-hidden">
        {compareOnly || (view === 'merge' && hasMerge) ? (
          <VisualCompareView original={mergeOriginal} proposed={mergeProposed}/>
        ) : (
          <MarkdownReport
            markdown={report.body || ''}
            searchQuery={search.query}
            searchActiveIndex={search.activeIndex}
            setSearchActiveIndex={search.setActiveIndex}
            onSearchMatchCount={search.onMatchCount}
            searchApiRef={searchApiRef}
          />
        )}
      </div>
    </div>);
}
