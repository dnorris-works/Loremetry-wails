import { useEffect, useState } from 'react';
import { api } from '@/api/client';
import { useAppState } from '@/lib/app-state';
import { Button } from '@/components/ui/button';
import { X } from 'lucide-react';

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
        bumpRefresh();
    }
    async function runAI() {
        let job = await api.startAnalysisJob(detail.id, projectPath);
        if (job.cached || job.status === 'done') {
            const report = await api.saveAnalysisJobReport(detail.id, projectPath, job.body || '');
            setBody(report.body || '');
            bumpRefresh();
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
        const report = await api.saveAnalysisJobReport(detail.id, projectPath, job.body || '');
        setBody(report.body || '');
        bumpRefresh();
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
        <Button size="sm" onClick={() => void run()} disabled={busy}>
          {busyLabel}
        </Button>
        <Button size="icon" variant="ghost" onClick={() => setAnalysis('')} title="Close">
          <X className="h-4 w-4"/>
        </Button>
      </div>
      <div className="flex-1 overflow-auto p-6">
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
        {body && <pre className="mt-6 max-w-3xl whitespace-pre-wrap text-sm">{body}</pre>}
      </div>
    </div>);
}

export function ReportPane() {
    const { reportId, setReport } = useAppState();
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
    if (!report) {
        return <div className="p-6 text-sm text-muted-foreground">Loading…</div>;
    }
    const ai = !!(report.uses_ai ?? report.usesAI);
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
        <Button size="icon" variant="ghost" onClick={() => setReport(0)} title="Close">
          <X className="h-4 w-4"/>
        </Button>
      </div>
      <div className="flex-1 overflow-auto p-6">
        <pre className="max-w-3xl whitespace-pre-wrap text-sm">{report.body}</pre>
      </div>
    </div>);
}
