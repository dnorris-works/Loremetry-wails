import { useEffect, useState } from 'react';
import { api } from '@/api/client';
import { useAppState } from '@/lib/app-state';
import { Button } from '@/components/ui/button';
import { X } from 'lucide-react';

function usesAI(detail) {
    return !!(detail?.uses_ai ?? detail?.usesAI);
}

export function AnalysisPane() {
    const { analysisId, setAnalysis, selection, bumpRefresh } = useAppState();
    const [detail, setDetail] = useState(null);
    const [busy, setBusy] = useState(false);
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
    async function run() {
        setBusy(true);
        setNotice('');
        try {
            const report = await api.runAnalysis(detail.id, projectPath);
            setBody(report.body || '');
            bumpRefresh();
        }
        catch (err) {
            const msg = err instanceof Error ? err.message : 'Could not run analysis';
            if (msg.includes('PLAN_REQUIRED')) {
                setNotice('This analysis uses AI. Choose a plan / add credits.');
            } else if (msg.includes('CREDITS_EMPTY')) {
                setNotice('This month’s credits are used up.');
            } else {
                setNotice(msg);
            }
        }
        finally {
            setBusy(false);
        }
    }
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
          {busy ? 'Running…' : 'Run'}
        </Button>
        <Button size="icon" variant="ghost" onClick={() => setAnalysis('')} title="Close">
          <X className="h-4 w-4"/>
        </Button>
      </div>
      <div className="flex-1 overflow-auto p-6">
        <p className="max-w-2xl text-sm leading-relaxed">{detail.description || 'No description.'}</p>
        {ai && <p className="mt-2 max-w-2xl text-xs text-muted-foreground">This analysis uses AI and needs a plan or credits.</p>}
        {notice && <p className="mt-3 max-w-2xl text-sm text-destructive">{notice}</p>}
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
