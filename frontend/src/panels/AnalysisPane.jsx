import { useEffect, useRef, useState } from 'react';
import { api } from '@/api/client';
import { useAppState } from '@/lib/app-state';
import { useConfirm } from '@/components/confirm-dialog';
import { Button } from '@/components/ui/button';
import { Save, Trash2, X, FileDown } from 'lucide-react';
import { MarkdownReport } from '@/editor/MarkdownReport';
import { StickyChapterDialog } from '@/editor/StickyChapterDialog';
import { VisualCompareView } from '@/editor/VisualCompareView';
import { ReportSearchBar, useReportSearch } from '@/editor/ReportSearch';

function usesAI(detail) {
    return !!(detail?.uses_ai ?? detail?.usesAI);
}

function usesMerge(detail) {
    return !!(detail?.uses_merge ?? detail?.usesMerge);
}

function parseStickyData(raw) {
    if (!raw || typeof raw !== 'string') return null;
    try {
        const data = JSON.parse(raw);
        if (data?.kind !== 'sticky_sentences') return null;
        return data;
    }
    catch {
        return null;
    }
}

async function openStickyChapter(projectPath, chapterIndex) {
    const result = await api.getAnalysisResult(projectPath, 'sticky_sentences');
    const data = parseStickyData(result.data_json || result.DataJSON || '');
    const chapters = data?.chapters || [];
    const ch = chapters[chapterIndex];
    if (!ch?.rel) {
        throw new Error('Chapter data missing — run Sticky Sentences again');
    }
    return api.manuscriptChapterStickyContext(projectPath, ch.rel);
}

function progressLabel(status, step, stepTotal, message) {
    const msg = (message || '').trim();
    if (status === 'queued')
        return msg || 'Queued…';
    if (msg && stepTotal > 1 && !msg.includes('(') && !msg.includes('·'))
        return `${msg} (${step}/${stepTotal})`;
    if (msg && stepTotal > 1 && msg.includes('·') && !/\(\d+\/\d+\)/.test(msg))
        return msg.replace('·', `(${step}/${stepTotal}) ·`);
    if (msg)
        return msg;
    if (status === 'running' && stepTotal > 1)
        return `Step ${step} of ${stepTotal}…`;
    if (status === 'running')
        return 'Running…';
    return 'Running…';
}

function formatElapsed(ms) {
    const sec = Math.max(0, Math.floor(ms / 1000));
    if (sec < 60)
        return `${sec}s`;
    const m = Math.floor(sec / 60);
    const r = sec % 60;
    return `${m}m ${String(r).padStart(2, '0')}s`;
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
    const [progressTick, setProgressTick] = useState(0);
    const progressStepStarted = useRef(0);
    const lastProgressText = useRef('');
    const [body, setBody] = useState('');
    const [mergeOriginal, setMergeOriginal] = useState('');
    const [mergeProposed, setMergeProposed] = useState('');
    const [view, setView] = useState('report');
    const [saved, setSaved] = useState(false);
    const [notice, setNotice] = useState('');
    const [stickyOpen, setStickyOpen] = useState(false);
    const [stickyContext, setStickyContext] = useState(null);
    const [stickyError, setStickyError] = useState('');
    const searchApiRef = useRef(null);
    const search = useReportSearch(searchApiRef);
    const projectPath = selection?.dir || '';
    const autoRanFor = useRef('');
    const busyRef = useRef(false);
    const detailRef = useRef(null);
    const queueRef = useRef([]);
    const queueLabelsRef = useRef([]);

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
        setStickyOpen(false);
        setStickyContext(null);
        setStickyError('');
        autoRanFor.current = '';
        let cancelled = false;
        void api.getAnalysis(analysisId).then((d) => {
            if (!cancelled)
                setDetail(d);
        }).catch(() => {
            if (!cancelled)
                setDetail(null);
        });
        return () => {
            cancelled = true;
        };
    }, [analysisId]);

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

    const queue = (analysisQueue?.length ? analysisQueue : (detail?.id ? [detail.id] : []));
    detailRef.current = detail;
    queueRef.current = queue;
    queueLabelsRef.current = queueLabels;
    busyRef.current = busy;

    useEffect(() => {
        if (!busy) {
            progressStepStarted.current = 0;
            lastProgressText.current = '';
            return undefined;
        }
        const id = setInterval(() => setProgressTick((n) => n + 1), 1000);
        return () => clearInterval(id);
    }, [busy]);

    function setLiveProgress(text) {
        const next = text || '';
        // Strip trailing elapsed from server messages so step changes reset the timer.
        const key = next.replace(/\s*·\s*\d+m?\s*\d*s?\s*$/, '').replace(/\s*·\s*\d+s\s*$/, '');
        if (key !== lastProgressText.current) {
            lastProgressText.current = key;
            progressStepStarted.current = Date.now();
        }
        else if (!progressStepStarted.current) {
            progressStepStarted.current = Date.now();
        }
        setProgress(next);
    }

    async function handleStickyChapterClick(index) {
        if (!projectPath) {
            setStickyError('Select a book or series folder first');
            setStickyContext(null);
            setStickyOpen(true);
            return;
        }
        setStickyError('');
        setStickyOpen(true);
        setStickyContext(null);
        try {
            const ctx = await openStickyChapter(projectPath, index);
            setStickyContext(ctx);
        }
        catch (err) {
            setStickyError(err instanceof Error ? err.message : 'Could not load chapter');
        }
    }

    function applyReport(report, forDetail) {
        const orig = report.original || report.Original || '';
        const prop = report.proposed || report.Proposed || '';
        const text = report.body || '';
        setBody(text);
        setMergeOriginal(orig);
        setMergeProposed(prop);
        setView(usesMerge(forDetail) || (orig && prop) ? 'merge' : 'report');
        setSaved(false);
    }

    async function runOne(id) {
        const info = await api.getAnalysis(id);
        if (usesAI(info)) {
            let job = await api.startAnalysisJob(id, projectPath);
            if (!(job.cached || job.status === 'done')) {
                const jobID = job.job_id || job.jobId;
                while (job.status !== 'done' && job.status !== 'failed') {
                    setLiveProgress(progressLabel(job.status, job.step, job.step_total || job.stepTotal, job.message));
                    await sleep(1000);
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
        const current = detailRef.current;
        if (!current || busyRef.current) return;
        const runQueue = queueRef.current.length ? queueRef.current : [current.id];
        const labels = queueLabelsRef.current;
        setBusy(true);
        setNotice('');
        setLiveProgress('');
        try {
            let primary = null;
            for (let i = 0; i < runQueue.length; i++) {
                const id = runQueue[i];
                const label = labels[i] || id;
                setLiveProgress(runQueue.length > 1 ? `${label} (${i + 1}/${runQueue.length})…` : 'Running…');
                const report = await runOne(id);
                if (id === current.id)
                    primary = report;
            }
            if (primary)
                applyReport(primary, current);
        }
        catch (err) {
            const msg = err instanceof Error ? err.message : 'Could not run analysis';
            setNotice(msg);
        }
        finally {
            setBusy(false);
            setProgress('');
            lastProgressText.current = '';
            progressStepStarted.current = 0;
        }
    }

    useEffect(() => {
        if (!detail || detail.id !== analysisId) return;
        if (!projectPath) {
            setNotice('Select a book or series folder in Projects first.');
            return;
        }
        if (autoRanFor.current === analysisId) return;
        autoRanFor.current = analysisId;
        void run();
        // eslint-disable-next-line react-hooks/exhaustive-deps -- run once per analysis open
    }, [detail, analysisId, projectPath]);

    if (!detail) {
        return <div className="p-6 text-sm text-muted-foreground">Loading…</div>;
    }
    const ai = usesAI(detail);
    const mergeFlag = usesMerge(detail);
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
                onStickyChapterClick={detail.id === 'sticky_sentences' ? handleStickyChapterClick : undefined}
              />
            )}
          </>
        ) : (<div className="overflow-auto p-6">
          {busy || progress ? (
            <div className="space-y-1">
              <p className="text-sm text-muted-foreground">{progress || 'Running…'}</p>
              {busy && progressStepStarted.current > 0 && (
                <p className="text-xs text-muted-foreground/80">
                  Elapsed on this step: {formatElapsed(Date.now() - progressStepStarted.current)}
                  <span className="sr-only">{progressTick}</span>
                </p>
              )}
            </div>
          ) : notice ? (
            <div className="max-w-2xl space-y-2">
              <p className="text-sm text-destructive">{notice}</p>
            </div>
          ) : (
            <p className="text-sm text-muted-foreground">Starting…</p>
          )}
        </div>)}
      </div>
      <StickyChapterDialog
        open={stickyOpen}
        onOpenChange={setStickyOpen}
        context={stickyContext}
        error={stickyError}
      />
    </div>);
}

export function ReportPane() {
    const { reportId, setReport, bumpRefresh } = useAppState();
    const confirm = useConfirm();
    const [report, setData] = useState(null);
    const [view, setView] = useState('report');
    const [stickyOpen, setStickyOpen] = useState(false);
    const [stickyContext, setStickyContext] = useState(null);
    const [stickyError, setStickyError] = useState('');
    useEffect(() => {
        if (!reportId) {
            setData(null);
            return;
        }
        setView('report');
        setStickyOpen(false);
        setStickyContext(null);
        setStickyError('');
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
    async function handleStickyChapterClick(index) {
        const projectPath = report?.project_path || report?.projectPath || '';
        if (!projectPath) {
            setStickyError('Project path missing on this report');
            setStickyContext(null);
            setStickyOpen(true);
            return;
        }
        setStickyError('');
        setStickyOpen(true);
        setStickyContext(null);
        try {
            const ctx = await openStickyChapter(projectPath, index);
            setStickyContext(ctx);
        }
        catch (err) {
            setStickyError(err instanceof Error ? err.message : 'Could not load chapter');
        }
    }
    if (!report) {
        return <div className="p-6 text-sm text-muted-foreground">Loading…</div>;
    }
    const ai = !!(report.uses_ai ?? report.usesAI);
    const mergeOriginal = report.original || report.Original || '';
    const mergeProposed = report.proposed || report.Proposed || '';
    const hasMerge = !!(mergeOriginal && mergeProposed);
    const compareOnly = hasMerge && !String(report.body || '').trim();
    const analysisId = report.analysis_id || report.analysisId;
    const isVellum = analysisId === 'vellum_prep';
    const isSticky = analysisId === 'sticky_sentences';
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
            onStickyChapterClick={isSticky ? handleStickyChapterClick : undefined}
          />
        )}
      </div>
      <StickyChapterDialog
        open={stickyOpen}
        onOpenChange={setStickyOpen}
        context={stickyContext}
        error={stickyError}
      />
    </div>);
}
