import { useEffect, useState } from 'react';
import { api } from '@/api/client';
import { useAppState } from '@/lib/app-state';
import { useConfirm } from '@/components/confirm-dialog';
import { Button } from '@/components/ui/button';
import { DelayedTooltip } from '@/components/ui/tooltip';
import { Check, Pencil, Plus, Settings, Database, Moon, Sun, Trash2, Eye, EyeOff } from 'lucide-react';
import { persistTheme, readTheme } from '@/lib/theme';
import { isAuthenticated, signOut } from '@/auth/session';
import { filesFromList, markHtmlFileDrop } from '@/lib/import-docs';
import { formatElapsed } from '@/lib/utils';
import { clearSidebarDrag, hasFiles, setSidebarDrag, sidebarDrag } from '@/lib/sidebar-drag';

const nest = 'ml-[0.5ch] border-l border-border pl-2';

function isCharacterTypeFolder(node) {
    const name = (node?.name || '').toLowerCase();
    return name === 'main' || name === 'supporting' || name === 'minor' || name === 'characters';
}

function needReady(sources, need) {
    const role = (sources?.roles || []).find((r) => r.role === need);
    return !!(role?.present && (role.files || []).length > 0);
}

function AnalysisUsesHover({ item, sources, hasProject, estimateSec, children }) {
    const needs = item.needs || [];
    const description = (item.description || '').trim();
    const content = (
        <div className="text-left">
            {description && (
                <div className="mb-1.5 max-w-[16rem] leading-snug opacity-90">{description}</div>
            )}
            {estimateSec > 0 && (
                <div className="mb-1.5 opacity-80">Typical: ~{formatElapsed(estimateSec * 1000)}</div>
            )}
            <div className="mb-1 font-semibold uppercase tracking-wide opacity-80">Uses</div>
            {needs.length === 0 ? (
                <div className="opacity-80">No sources required</div>
            ) : (
                <ul className="space-y-0.5">
                    {needs.map((n) => {
                        let status = '';
                        if (hasProject) {
                            status = needReady(sources, n) ? ' — ready' : ' — missing';
                        }
                        return <li key={n}>{n}{status}</li>;
                    })}
                </ul>
            )}
            {!hasProject && needs.length > 0 && (
                <div className="mt-1.5 opacity-70">Select a book or series first</div>
            )}
        </div>
    );
    return (
        <DelayedTooltip delayMs={300} side="right" content={content} className="relative block w-full">
            {children}
        </DelayedTooltip>
    );
}

export function Sidebar({ onNewSeries, onEditSeries, onNewStory, onEditStory, onSettings, onAdmin, }) {
    const { selection, setSelection, analysisId, analysisQueue, setAnalysis, reportId, setReport, refreshKey, bumpRefresh, openSeries, openStories, openFolders, toggleSeries, toggleStory, toggleFolder } = useAppState();
    const hasProject = !!(selection?.dir);
    const confirm = useConfirm();
    const [tree, setTree] = useState({ pens: [], problems: [] });
    const [analysisGroups, setAnalysisGroups] = useState([]);
    const [reports, setReports] = useState([]);
    const [theme, setTheme] = useState(readTheme());
    useEffect(() => {
        void (async () => {
            try {
                setTree(await api.listWritingTree());
            }
            catch {
                /* ignore */
            }
        })();
    }, [refreshKey]);
    useEffect(() => {
        void api.listAnalysisCatalog().then(setAnalysisGroups).catch(() => setAnalysisGroups([]));
    }, []);
    useEffect(() => {
        void api.listAnalysisReports().then(setReports).catch(() => setReports([]));
    }, [refreshKey]);
    async function onToggleTheme() {
        const next = theme === 'dark' ? 'light' : 'dark';
        setTheme(next);
        await persistTheme(next, isAuthenticated());
    }
    const [tab, setTab] = useState('projects');
    const [analysisFilter, setAnalysisFilter] = useState('all');
    const [analysisSources, setAnalysisSources] = useState(null);
    const [runEstimates, setRunEstimates] = useState({});
    useEffect(() => {
        if (!hasProject && tab === 'analysis') setTab('projects');
    }, [hasProject, tab]);
    useEffect(() => {
        if (tab !== 'analysis') {
            return;
        }
        let cancelled = false;
        void api.listAnalysisRunEstimates().then((m) => {
            if (!cancelled) setRunEstimates(m && typeof m === 'object' ? m : {});
        }).catch(() => {
            if (!cancelled) setRunEstimates({});
        });
        return () => { cancelled = true; };
    }, [tab, refreshKey]);
    useEffect(() => {
        if (tab !== 'analysis' || !selection?.dir) {
            setAnalysisSources(null);
            return;
        }
        let cancelled = false;
        void api.matchAnalysisSources(selection.dir).then((s) => {
            if (!cancelled) setAnalysisSources(s);
        }).catch(() => {
            if (!cancelled) setAnalysisSources(null);
        });
        return () => { cancelled = true; };
    }, [tab, selection?.dir, refreshKey]);
    const pens = tree.pens || [];
    return (<div className="flex h-full flex-col bg-card">
      <div className="flex items-center justify-between border-b border-border px-3 py-2">
        <span className="text-sm font-semibold">Loremetry</span>
        <div className="flex gap-1">
          <Button size="icon" variant="ghost" onClick={() => void onToggleTheme()} title="Theme">
            {theme === 'dark' ? <Sun className="h-4 w-4"/> : <Moon className="h-4 w-4"/>}
          </Button>
          <Button size="icon" variant="ghost" onClick={onSettings} title="Settings">
            <Settings className="h-4 w-4"/>
          </Button>
          <Button size="icon" variant="ghost" onClick={onAdmin} title="Admin">
            <Database className="h-4 w-4"/>
          </Button>
        </div>
      </div>
      <div className="flex border-b border-border">
        <button type="button" className={`flex-1 px-2 py-1.5 text-xs font-medium ${tab === 'projects' ? 'border-b-2 border-primary text-foreground' : 'text-muted-foreground hover:text-foreground'}`} onClick={() => setTab('projects')}>
          Projects
        </button>
        <button type="button" disabled={!hasProject} title={hasProject ? undefined : 'Select a book or series in Projects first'} className={`flex-1 px-2 py-1.5 text-xs font-medium ${!hasProject ? 'cursor-not-allowed opacity-40' : ''} ${tab === 'analysis' ? 'border-b-2 border-primary text-foreground' : 'text-muted-foreground hover:text-foreground'}`} onClick={() => hasProject && setTab('analysis')}>
          Analysis
        </button>
        <button type="button" className={`flex-1 px-2 py-1.5 text-xs font-medium ${tab === 'reports' ? 'border-b-2 border-primary text-foreground' : 'text-muted-foreground hover:text-foreground'}`} onClick={() => setTab('reports')}>
          Reports
        </button>
      </div>
      <div className="flex-1 overflow-auto p-2">
        {tab === 'analysis' && (<>
          <div className="mb-2 flex rounded border border-border text-[10px]">
            {[['all', 'All'], ['local', 'Local'], ['ai', 'AI']].map(([value, label]) => (
              <button
                key={value}
                type="button"
                className={`flex-1 px-2 py-1 ${analysisFilter === value ? 'bg-primary text-primary-foreground' : 'text-muted-foreground hover:text-foreground'}`}
                onClick={() => setAnalysisFilter(value)}
              >
                {label}
              </button>
            ))}
          </div>
          {analysisGroups.map((g) => {
            const items = (g.items || []).filter((item) => {
              const ai = !!(item.uses_ai ?? item.usesAI);
              if (analysisFilter === 'local') return !ai;
              if (analysisFilter === 'ai') return ai;
              return true;
            });
            if (items.length === 0) return null;
            return (<div key={g.id} className="mb-3">
              <div className="px-1 pb-1 text-[10px] font-semibold uppercase text-muted-foreground">{g.label}</div>
              {items.map((item) => {
                const ai = !!(item.uses_ai ?? item.usesAI);
                const selected = (analysisQueue || []).includes(item.id);
                const primary = analysisId === item.id;
                return (
                  <AnalysisUsesHover key={item.id} item={item} sources={analysisSources} hasProject={hasProject} estimateSec={runEstimates[item.id] || 0}>
                    <button type="button" className={`flex w-full items-center gap-1 truncate px-1 py-0.5 text-left text-xs leading-tight hover:bg-accent ${primary ? 'bg-accent text-foreground' : selected ? 'bg-muted/60 text-foreground' : 'text-foreground'}`} onClick={() => setAnalysis(item.id)}>
                      {selected ? <Check className="h-3 w-3 shrink-0 text-primary"/> : <span className="inline-block h-3 w-3 shrink-0"/>}
                      <span className="min-w-0 flex-1 truncate">{item.label}</span>
                      <span className="shrink-0 text-[9px] uppercase text-muted-foreground">{ai ? 'AI' : 'Local'}</span>
                    </button>
                  </AnalysisUsesHover>
                );
              })}
            </div>);
          })}
        </>)}
        {tab === 'reports' && (reports.length === 0 ? (<p className="px-1 text-xs text-muted-foreground">No saved reports yet.</p>) : reports.map((r) => (
          <div key={r.id} className="group mb-1 flex items-start">
            <button type="button" className={`min-w-0 flex-1 truncate px-1 py-0.5 text-left text-xs hover:bg-accent ${reportId === r.id ? 'bg-accent text-foreground' : 'text-foreground'}`} onClick={() => setReport(r.id)}>
              {r.analysis_label || r.analysisLabel}
              <span className="block truncate text-[10px] text-muted-foreground">{r.created_at || r.createdAt}</span>
            </button>
            <button type="button" className="shrink-0 p-1 text-muted-foreground opacity-0 hover:text-destructive group-hover:opacity-100" title="Delete report" onClick={async () => { if (await confirm('Delete this report?')) { await api.deleteAnalysisReport(r.id); if (reportId === r.id) setReport(0); bumpRefresh(); } }}>
              <Trash2 className="h-3 w-3"/>
            </button>
          </div>
        )))}
        {tab === 'projects' && (<>
        {(tree.problems || []).map((p) => (<p key={p.path || p.message} className="mb-2 px-1 text-[11px] text-destructive">{p.message}</p>))}
        {pens.map((pen) => {
          const penOpen = !openFolders.includes(pen.path);
          const seriesOpen = !openFolders.includes(pen.path + '|series');
          const storiesOpen = !openFolders.includes(pen.path + '|stories');
          return (<div key={pen.path} className="mb-2">
            <div className="px-1 leading-tight">
              <button type="button" className="w-full truncate text-left text-xs font-semibold uppercase text-muted-foreground hover:bg-accent" onClick={() => toggleFolder(pen.path)}>
                {pen.name}
              </button>
            </div>
            {(pen.problems || []).map((p) => (<p key={p.path || p.message} className="px-1 text-[11px] leading-tight text-destructive">{p.message}</p>))}
            {penOpen && <div className={nest}>
              <div className="flex h-5 items-center justify-between">
                <button type="button" className="min-w-0 flex-1 truncate text-left text-[10px] uppercase leading-none text-muted-foreground hover:bg-accent" onClick={() => toggleFolder(pen.path + '|series')}>
                  Series
                </button>
                <Button size="icon" variant="ghost" className="h-5 w-5" title="New series" onClick={() => onNewSeries(pen.name, pen.path)}>
                  <Plus className="h-3 w-3"/>
                </Button>
              </div>
              {seriesOpen && (pen.series || []).map((s) => {
                const selected = selection?.dir === s.path;
                return (<div key={s.path}>
                  <div className="flex h-6 items-center">
                    {selected && <Check className="h-3 w-3 shrink-0 text-green-600 dark:text-green-400"/>}
                    <button type="button" className={`flex-1 truncate rounded px-1 py-0 text-left text-sm font-normal leading-tight hover:bg-accent ${selected ? 'font-medium text-teal-800 dark:text-teal-300' : 'text-teal-700 dark:text-teal-400'}`} onClick={async () => { await setSelection({ type: 'file', dir: s.path, name: s.name }); toggleSeries(s.path); }}>
                      {s.name}
                    </button>
                    <Button size="icon" variant="ghost" className="h-5 w-5" title="Add story" onClick={() => onNewStory(s.path, pen.name, pen.path)}>
                      <Plus className="h-3 w-3"/>
                    </Button>
                    <Button size="icon" variant="ghost" className="h-5 w-5" onClick={() => onEditSeries(s.path, s.name)}>
                      <Pencil className="h-3 w-3"/>
                    </Button>
                  </div>
                  {openSeries.includes(s.path) && (<div className={nest}>
                      {(s.problems || []).map((p) => (<p key={p.path || p.message} className="text-[11px] leading-tight text-destructive">{p.message}</p>))}
                      {(s.tree?.folders || []).map((node) => (<FolderNode key={node.path} node={node} projectPath={s.path}/>))}
                      {(s.tree?.files || []).map((f) => (<FileRow key={f.path} dir={s.path} file={f} onDelete={async () => {
                          if (!(await confirm('Delete this file?')))
                              return;
                          await api.deleteDiskFile(s.path, f.name);
                          bumpRefresh();
                      }}/>))}
                    </div>)}
                </div>);
              })}
              <div className="flex h-5 items-center justify-between">
                <button type="button" className="min-w-0 flex-1 truncate text-left text-[10px] uppercase leading-none text-muted-foreground hover:bg-accent" onClick={() => toggleFolder(pen.path + '|stories')}>
                  Stories
                </button>
                <Button size="icon" variant="ghost" className="h-5 w-5" title="New story" onClick={() => onNewStory('', pen.name, pen.path)}>
                  <Plus className="h-3 w-3"/>
                </Button>
              </div>
              {storiesOpen && (pen.books || []).map((st) => {
                const selected = selection?.dir === st.path;
                return (<div key={st.path}>
                  <div className="flex h-6 items-center">
                    {selected && <Check className="h-3 w-3 shrink-0 text-green-600 dark:text-green-400"/>}
                    <button type="button" className={`flex-1 truncate rounded px-1 py-0 text-left text-sm font-normal leading-tight hover:bg-accent ${selected ? 'font-medium text-blue-900 dark:text-blue-200' : 'text-blue-800 dark:text-blue-300'}`} onClick={async () => { await setSelection({ type: 'file', dir: st.path, name: st.name }); toggleStory(st.path); }}>
                      {st.name}
                    </button>
                    <Button size="icon" variant="ghost" className="h-5 w-5" onClick={() => onEditStory(st.path, st.name)}>
                      <Pencil className="h-3 w-3"/>
                    </Button>
                  </div>
                  {openStories.includes(st.path) && (<div className={nest}>
                      {(st.tree?.folders || []).map((node) => (<FolderNode key={node.path} node={node} projectPath={st.path}/>))}
                      {(st.tree?.files || []).map((f) => (<FileRow key={f.path} dir={st.path} file={f} onDelete={async () => {
                          if (!(await confirm('Delete this file?')))
                              return;
                          await api.deleteDiskFile(st.path, f.name);
                          bumpRefresh();
                      }}/>))}
                    </div>)}
                </div>);
              })}
            </div>}
          </div>);
        })}
        </>)}
      </div>
      <div className="border-t border-border p-2">
        <Button variant="ghost" size="sm" className="w-full" onClick={() => signOut()}>
          Sign out
        </Button>
      </div>
    </div>);
}

function FolderNode({ node, projectPath }) {
    const { selection, setSelection, bumpRefresh, openFolders, toggleFolder, ensureFolder } = useAppState();
    const confirm = useConfirm();
    const files = node.files || [];
    const folders = node.folders || [];
    const hasKids = files.length > 0 || folders.length > 0;
    const open = openFolders.includes(node.path);
    const key = node.key || node.name;
    const acceptCharacterDrop = isCharacterTypeFolder(node);
    async function addDoc() {
        const created = await api.createDiskFile(node.path, '', '');
        bumpRefresh();
        await ensureFolder(node.path);
        setSelection({ type: 'file', dir: node.path, name: created.name });
    }
    async function dropFiles(list) {
        const incoming = await filesFromList(list);
        if (!incoming.length)
            return;
        markHtmlFileDrop();
        const last = await api.createDiskFiles(node.path, incoming);
        bumpRefresh();
        await ensureFolder(node.path);
        setSelection({ type: 'file', dir: node.path, name: last.name });
    }
    async function dropSidebarFile(item) {
        if (!item || item.t !== 'file' || !item.dir || !item.name)
            return;
        if (item.dir === node.path)
            return;
        try {
            const moved = await api.moveDiskFile(item.dir, item.name, node.path);
            if (selection?.type === 'file' && selection.dir === item.dir && selection.name === item.name) {
                setSelection({ type: 'file', dir: moved.dir || node.path, name: moved.name || item.name });
            }
            bumpRefresh();
            await ensureFolder(node.path);
        }
        catch (err) {
            alert(err instanceof Error ? err.message : 'Could not move');
        }
    }
    async function setHidden(hide) {
        await api.setFolderOverride(projectPath, key, hide ? 'hide' : 'show');
        bumpRefresh();
    }
    return (<div className="rounded [--wails-drop-target:drop]" data-folder-path={node.path} onDragOver={(e) => {
            if (hasFiles(e) || (acceptCharacterDrop && sidebarDrag(e)))
                e.preventDefault();
        }} onDrop={(e) => {
            const item = sidebarDrag(e);
            if (item?.t === 'file' && acceptCharacterDrop) {
                e.preventDefault();
                clearSidebarDrag();
                void dropSidebarFile(item);
                return;
            }
            if (e.dataTransfer.files.length) {
                e.preventDefault();
                void dropFiles(e.dataTransfer.files);
            }
        }}>
      <div className="flex h-5 items-center">
        <button type="button" className={`min-w-0 flex-1 truncate py-0 text-left text-xs font-normal leading-tight hover:bg-accent ${node.hidden ? 'text-muted-foreground/70' : 'text-foreground'}`} onClick={() => toggleFolder(node.path)}>
          {node.required && <span className="mr-1 text-red-600 dark:text-red-400" aria-hidden="true">*</span>}
          {node.label || `${node.name} (${node.count ?? files.length})`}
        </button>
        <Button size="icon" variant="ghost" className="h-5 w-5" title={node.hidden ? 'Show in this story' : 'Hide in this story'} onClick={() => void setHidden(!node.hidden)}>
          {node.hidden ? <Eye className="h-3 w-3"/> : <EyeOff className="h-3 w-3"/>}
        </Button>
        <Button size="icon" variant="ghost" className="h-5 w-5" title="Add document" onClick={() => void addDoc()}>
          <Plus className="h-3 w-3"/>
        </Button>
      </div>
      {open && hasKids && (<div className={nest}>
        {files.map((f) => (<FileRow key={f.path} dir={node.path} file={f} onDelete={async () => {
                if (!(await confirm('Delete this file?')))
                    return;
                await api.deleteDiskFile(node.path, f.name);
                bumpRefresh();
            }}/>))}
        {folders.map((child) => (<FolderNode key={child.path} node={child} projectPath={projectPath}/>))}
      </div>)}
    </div>);
}

function underCharacters(dir) {
    return /[/\\]Characters([/\\]|$)/i.test(dir || '');
}

function FileRow({ dir, file, onDelete }) {
    const { setSelection } = useAppState();
    const allowDrag = underCharacters(dir);
    return (<div className="flex h-5 items-center" draggable={allowDrag} onDragStart={allowDrag ? (e) => setSidebarDrag(e, { t: 'file', dir, name: file.name }) : undefined} onDragEnd={allowDrag ? () => clearSidebarDrag() : undefined}>
      <button type="button" className="min-w-0 flex-1 truncate py-0 text-left text-xs font-normal leading-tight text-muted-foreground hover:bg-accent hover:text-foreground" onClick={() => setSelection({ type: 'file', dir, name: file.name })}>
        {file.name}
      </button>
      {onDelete && (<Button type="button" size="icon" variant="ghost" className="h-5 w-5 shrink-0" onClick={(e) => {
                e.preventDefault();
                e.stopPropagation();
                void onDelete();
            }}>
          <Trash2 className="pointer-events-none h-3 w-3"/>
        </Button>)}
    </div>);
}
