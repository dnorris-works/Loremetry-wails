import { useCallback, useEffect, useState } from 'react';
import { api } from '@/api/client';
import { useAppState } from '@/lib/app-state';
import { useConfirm } from '@/components/confirm-dialog';
import { Button } from '@/components/ui/button';
import { Save, Trash2, X } from 'lucide-react';
import { LazySavedReport, MarkdownReport, REPORT_INLINE_MAX } from '@/editor/LazyReportViewer';

function usesAI(detail) {
    return !!(detail?.uses_ai ?? detail?.usesAI);
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
    const { analysisId, setAnalysis, selection, bumpRefresh } = useAppState();
    const [detail, setDetail] = useState(null);
    const [busy, setBusy] = useState(false);
    const [progress, setProgress] = useState('');
    const [body, setBody] = useState('');
    const [saved, setSaved] = useState(false);
    const [notice, setNotice] = useState('');
    const [sources, setSources] = useState(null);
    const projectPath = selection?.dir || '';
    useEffect(() => {
        if (!analysisId) {
            setDetail(null);
            return;
        }
        setBody('');
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
    if (!detail) {
        return <div className="p-6 text-sm text-muted-foreground">Loading…</div>;
    }
    const ai = usesAI(detail);
    async function runLocal() {
        const report = await api.runAnalysis(detail.id, projectPath);
        setBody(report.body || '');
        setSaved(false);
    }
    async function runAI() {
        let job = await api.startAnalysisJob(detail.id, projectPath);
        if (job.cached || job.status === 'done') {
            setBody(job.body || '');
            setSaved(false);
            return;
        }
        const jobID = job.job_id || job.jobId;
        while (job.status !== 'done' && job.status !== 'failed') {
            setProgress(progressLabel(job.status, job.step, job.step_total || job.stepTotal));
            await sleep(2000);
            job = await api.getAnalysisJobStatus(jobID);
        }
        if (job.status === 'failed') {
            throw new Error(job.error || 'Analysis failed');
        }
        setBody(job.body || '');
        setSaved(false);
    }
    async function run() {
        setBusy(true);
        setNotice('');
        setProgress('');
        try {
            if (ai) {
                await runAI();
            } else {
                await runLocal();
            }
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
        await api.saveAnalysisJobReport(detail.id, projectPath, body);
        setSaved(true);
        bumpRefresh();
    }
    const busyLabel = progress || (busy ? 'Running…' : 'Run');
    return (<div className="flex h-full flex-col">
      <div className="flex items-center gap-2 border-b border-border px-4 py-2">
        <div className="min-w-0 flex-1">
          <div className="flex items-center gap-2">
            <div className="truncate text-sm font-medium">{detail.label}</div>
            <span className="shrink-0 rounded border border-border px-1.5 py-0.5 text-[10px] uppercase text-muted-foreground">
              {ai ? 'Uses AI' : 'Local'}
            </span>
          </div>
          <div className="truncate text-[11px] text-muted-foreground">{detail.group}</div>
        </div>
        {body && !saved && (
          <Button size="sm" variant="outline" onClick={() => void saveReport()} className="gap-1.5">
            <Save className="h-3.5 w-3.5"/>Save
          </Button>
        )}
        {body && saved && (
          <span className="text-xs text-green-600 dark:text-green-400">Saved</span>
        )}
        <Button size="sm" onClick={() => void run()} disabled={busy}>
          {busyLabel}
        </Button>
        <Button size="icon" variant="ghost" onClick={() => { setBody(''); setAnalysis(''); }} title="Close">
          <X className="h-4 w-4"/>
        </Button>
      </div>
      <div className="flex-1 overflow-auto p-6">
        {body ? (
          <MarkdownReport text={body}/>
        ) : (<>
          <p className="max-w-2xl text-sm leading-relaxed">{detail.description || 'No description.'}</p>
          {ai && <p className="mt-2 max-w-2xl text-xs text-muted-foreground">This analysis uses AI and needs a plan or credits.</p>}
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
        </>)}
      </div>
    </div>);
}

export function ReportPane() {
    const { reportId, setReport, bumpRefresh } = useAppState();
    const confirm = useConfirm();
    const [report, setData] = useState(null);
    useEffect(() => {
        if (!reportId) {
            setData(null);
            return;
        }
        let cancelled = false;
        void api.getAnalysisReport(reportId).then((r) => {
            if (!cancelled)
                setData(r);
        }).catch(() => {
            if (!cancelled)
                setData(null);
        });
        return () => {
            cancelled = true;
        };
    }, [reportId]);
    const loadRange = useCallback(
        (offset, limit) => api.getAnalysisReportBodyRange(reportId, offset, limit),
        [reportId],
    );
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
    const bodySize = report.body_size ?? report.bodySize ?? report.body?.length ?? 0;
    const large = !!(report.large || bodySize > REPORT_INLINE_MAX);
    return (<div className="flex h-full flex-col">
      <div className="flex items-center gap-2 border-b border-border px-4 py-2">
        <div className="min-w-0 flex-1">
          <div className="flex items-center gap-2">
            <div className="truncate text-sm font-medium">{report.analysis_label || report.analysisLabel}</div>
            <span className="shrink-0 rounded border border-border px-1.5 py-0.5 text-[10px] uppercase text-muted-foreground">
              {ai ? 'Uses AI' : 'Local'}
            </span>
          </div>
          <div className="truncate text-[11px] text-muted-foreground">{report.created_at || report.createdAt}</div>
        </div>
        <Button size="icon" variant="ghost" onClick={() => void handleDelete()} title="Delete report" className="text-destructive hover:text-destructive">
          <Trash2 className="h-4 w-4"/>
        </Button>
        <Button size="icon" variant="ghost" onClick={() => setReport(0)} title="Close">
          <X className="h-4 w-4"/>
        </Button>
      </div>
      <div className="flex-1 overflow-hidden">
        {large
            ? <LazySavedReport reportId={reportId} bodySize={bodySize} loadRange={loadRange}/>
            : <div className="h-full overflow-auto p-6"><MarkdownReport text={report.body}/></div>}
      </div>
    </div>);
}
