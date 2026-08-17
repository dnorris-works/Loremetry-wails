import { useEffect, useState } from 'react';
import { api } from '@/api/client';
import { useAppState } from '@/lib/app-state';
import { useConfirm } from '@/components/confirm-dialog';
import { Button } from '@/components/ui/button';
import { Pencil, Plus, Settings, Database, Moon, Sun, Trash2, Eye, EyeOff } from 'lucide-react';
import { persistTheme, readTheme } from '@/lib/theme';
import { isAuthenticated, signOut } from '@/auth/session';
import { filesFromList, markHtmlFileDrop } from '@/lib/import-docs';

const nest = 'ml-[2ch] border-l border-border pl-2';

export function Sidebar({ onNewSeries, onEditSeries, onNewStory, onEditStory, onSettings, onAdmin, }) {
    const { setSelection, analysisId, setAnalysis, refreshKey, bumpRefresh, openSeries, openStories, openFolders, toggleSeries, toggleStory, toggleFolder } = useAppState();
    const confirm = useConfirm();
    const [tree, setTree] = useState({ pens: [], problems: [] });
    const [analysisGroups, setAnalysisGroups] = useState([]);
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
    async function onToggleTheme() {
        const next = theme === 'dark' ? 'light' : 'dark';
        setTheme(next);
        await persistTheme(next, isAuthenticated());
    }
    const [tab, setTab] = useState('projects');
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
        <button type="button" className={`flex-1 px-2 py-1.5 text-xs font-medium ${tab === 'analysis' ? 'border-b-2 border-primary text-foreground' : 'text-muted-foreground hover:text-foreground'}`} onClick={() => setTab('analysis')}>
          Analysis
        </button>
        <button type="button" className={`flex-1 px-2 py-1.5 text-xs font-medium ${tab === 'reports' ? 'border-b-2 border-primary text-foreground' : 'text-muted-foreground hover:text-foreground'}`} onClick={() => setTab('reports')}>
          Reports
        </button>
      </div>
      <div className="flex-1 overflow-auto p-2">
        {tab === 'analysis' && analysisGroups.map((g) => (<div key={g.id} className="mb-3">
          <div className="px-1 pb-1 text-[10px] font-semibold uppercase text-muted-foreground">{g.label}</div>
          {(g.items || []).map((item) => (<button key={item.id} type="button" className={`w-full truncate px-1 py-0.5 text-left text-xs leading-tight hover:bg-accent ${analysisId === item.id ? 'bg-accent text-foreground' : 'text-foreground'}`} onClick={() => setAnalysis(item.id)}>
            {item.label}
          </button>))}
        </div>))}
        {tab === 'reports' && (<p className="px-1 text-xs text-muted-foreground">No saved reports yet.</p>)}
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
              {seriesOpen && (pen.series || []).map((s) => (<div key={s.path}>
                  <div className="flex h-6 items-center">
                    <button type="button" className="flex-1 truncate rounded px-1 py-0 text-left text-sm font-normal leading-tight text-teal-700 hover:bg-accent dark:text-teal-400" onClick={() => toggleSeries(s.path)}>
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
                </div>))}
              <div className="flex h-5 items-center justify-between">
                <button type="button" className="min-w-0 flex-1 truncate text-left text-[10px] uppercase leading-none text-muted-foreground hover:bg-accent" onClick={() => toggleFolder(pen.path + '|stories')}>
                  Stories
                </button>
                <Button size="icon" variant="ghost" className="h-5 w-5" title="New story" onClick={() => onNewStory('', pen.name, pen.path)}>
                  <Plus className="h-3 w-3"/>
                </Button>
              </div>
              {storiesOpen && (pen.books || []).map((st) => (<div key={st.path}>
                  <div className="flex h-6 items-center">
                    <button type="button" className="flex-1 truncate rounded px-1 py-0 text-left text-sm font-normal leading-tight text-blue-800 hover:bg-accent dark:text-blue-300" onClick={() => toggleStory(st.path)}>
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
                </div>))}
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
    const { setSelection, bumpRefresh, openFolders, toggleFolder, ensureFolder } = useAppState();
    const confirm = useConfirm();
    const files = node.files || [];
    const folders = node.folders || [];
    const hasKids = files.length > 0 || folders.length > 0;
    const open = openFolders.includes(node.path);
    const key = node.key || node.name;
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
    async function setHidden(hide) {
        await api.setFolderOverride(projectPath, key, hide ? 'hide' : 'show');
        bumpRefresh();
    }
    return (<div className="rounded [--wails-drop-target:drop]" data-folder-path={node.path} onDragOver={(e) => {
            if (e.dataTransfer.types.includes('Files'))
                e.preventDefault();
        }} onDrop={(e) => {
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

function FileRow({ dir, file, onDelete }) {
    const { setSelection } = useAppState();
    return (<div className="flex h-5 items-center">
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
