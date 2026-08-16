import { useEffect, useState } from 'react';
import { api } from '@/api/client';
import { useAppState } from '@/lib/app-state';
import { useConfirm, useNotice } from '@/components/confirm-dialog';
import { StoryFolders } from '@/panels/StoryFolders';
import { Button } from '@/components/ui/button';
import { Pencil, Plus, Settings, Database, Moon, Sun, Trash2 } from 'lucide-react';
import { persistTheme, readTheme } from '@/lib/theme';
import { folderHeader } from '@/lib/folder-label';
import { filesFromList, markHtmlFileDrop } from '@/lib/import-docs';
import { importDroppedOnHeader } from '@/lib/drop-import';
import { isAuthenticated, signOut } from '@/auth/session';
import { clearSidebarDrag, currentSidebarDrag, moveBefore, setSidebarDrag, sidebarDrag, } from '@/lib/sidebar-drag';
import { HeaderFolderLink } from '@/lib/HeaderFolderLink';
export function Sidebar({ onNewSeries, onEditSeries, onNewStory, onEditStory, onSettings, onAdmin, }) {
    const { setSelection, refreshKey, bumpRefresh, openSeries, openStories, toggleSeries, toggleStory, ensureSeriesFolder, seriesFoldersOpen, toggleSeriesFolders, } = useAppState();
    const [tree, setTree] = useState({ pens: [], problems: [] });
    const [types, setTypes] = useState([]);
    const [theme, setTheme] = useState(readTheme());
    useEffect(() => {
        void (async () => {
            try {
                setTree(await api.listWritingTree());
                setTypes(await api.listDocumentTypes());
            }
            catch {
                /* ignore */
            }
        })();
    }, [refreshKey]);
    async function onToggleTheme() {
        const next = theme === 'dark' ? 'light' : 'dark';
        setTheme(next);
        await persistTheme(next, isAuthenticated());
    }
    const seriesTypes = types.filter((t) => t.applies_to_series);
    const storyTypes = types.filter((t) => t.applies_to_story);
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
      <div className="flex-1 overflow-auto p-2">
        {(tree.problems || []).map((p) => (<p key={p.path || p.message} className="mb-2 px-1 text-[11px] text-destructive">{p.message}</p>))}
        {pens.map((pen) => (<div key={pen.path} className="mb-2">
            <div className="px-1 leading-tight">
              <span className="text-xs font-semibold uppercase text-muted-foreground">{pen.name}</span>
            </div>
            {(pen.problems || []).map((p) => (<p key={p.path || p.message} className="px-1 text-[11px] leading-tight text-destructive">{p.message}</p>))}
            <div className="ml-[2ch] border-l border-border pl-2">
            <div className="flex h-5 items-center justify-between">
              <span className="text-[10px] uppercase leading-none text-muted-foreground">Series</span>
                  <Button size="icon" variant="ghost" className="h-5 w-5" title="New series" onClick={() => onNewSeries(pen.name, pen.path)}>
                <Plus className="h-3 w-3"/>
              </Button>
            </div>
            {(pen.series || []).map((s) => (<div key={s.path}>
                <div className="flex h-6 items-center">
                  <button type="button" className="flex-1 truncate rounded px-1 py-0 text-left text-sm font-normal leading-tight text-teal-700 hover:bg-accent dark:text-teal-400" onClick={() => toggleSeries(s.path)}>
                    {s.name}
                  </button>
                  <AddMenu types={seriesTypes} extra={[{ code: 'story', display_name: 'Story' }]} onPick={(t) => {
                    if (t.code === 'story') {
                        onNewStory(s.path, pen.name, pen.path);
                        return;
                    }
                    void createHeader('series', s.path, t, setSelection, bumpRefresh, () => ensureSeriesFolder(s.path, t.code));
                }}/>
                  <Button size="icon" variant="ghost" className="h-5 w-5" onClick={() => onEditSeries(s.path, s.name)}>
                    <Pencil className="h-3 w-3"/>
                  </Button>
                </div>
                {openSeries.includes(s.path) && (<div className="ml-[2ch] border-l border-border pl-2">
                    <button type="button" className="w-full truncate py-0 text-left text-xs font-normal leading-tight text-muted-foreground hover:text-foreground" onClick={() => toggleSeriesFolders(s.path)}>
                      {seriesFoldersOpen(s.path) ? 'Hide folders' : 'Show folders'}
                    </button>
                    {seriesFoldersOpen(s.path) && <SeriesFolders projectPath={s.path} types={seriesTypes}/>}
                    {(s.problems || []).map((p) => (<p key={p.path || p.message} className="text-[11px] leading-tight text-destructive">{p.message}</p>))}
                    {(s.books || []).map((st) => (<StoryRow key={st.path} story={st} types={storyTypes} expanded={openStories.includes(st.path)} onToggle={() => toggleStory(st.path)} onEdit={() => onEditStory(st.path, st.name)}/>))}
                  </div>)}
              </div>))}
            <div className="flex h-5 items-center justify-between">
              <span className="text-[10px] uppercase leading-none text-muted-foreground">Stories</span>
              <Button size="icon" variant="ghost" className="h-5 w-5" title="New story" onClick={() => onNewStory('', pen.name, pen.path)}>
                <Plus className="h-3 w-3"/>
              </Button>
            </div>
            {(pen.books || []).map((st) => (<StoryRow key={st.path} story={st} types={storyTypes} expanded={openStories.includes(st.path)} onToggle={() => toggleStory(st.path)} onEdit={() => onEditStory(st.path, st.name)}/>))}
            </div>
          </div>))}
      </div>
      <div className="border-t border-border p-2">
        <Button variant="ghost" size="sm" className="w-full" onClick={() => signOut()}>
          Sign out
        </Button>
      </div>
    </div>);
}
function nextName(existing, base) {
    const names = new Set(existing);
    if (!names.has(base))
        return base;
    let n = 2;
    while (names.has(`${base} ${n}`))
        n += 1;
    return `${base} ${n}`;
}
export async function createHeader(projectKind, projectPath, t, setSelection, bumpRefresh, ensureFolder) {
    const list = await api.listHeaderFiles(projectPath, projectKind, t.code);
    if (list.missing) {
        alert('This header folder is missing. Fix the files or link a folder.');
        return;
    }
    const name = nextName(list.files.map((f) => f.name.replace(/\.[^.]+$/, '')), t.display_name);
    const created = await api.createHeaderFile({
        project_path: projectPath,
        project_kind: projectKind,
        kind: t.code,
        name,
        text: '',
    });
    bumpRefresh();
    ensureFolder?.();
    setSelection({ type: 'file', projectPath, projectKind, kind: t.code, rel: created.rel, title: created.name });
}
function AddMenu({ types, extra, onPick }) {
    const [open, setOpen] = useState(false);
    const items = [...(extra || []), ...types];
    return (<div className="relative">
      <Button size="icon" variant="ghost" className="h-5 w-5" onClick={() => setOpen((v) => !v)} title="Add">
        <Plus className="h-3 w-3"/>
      </Button>
      {open && (<div className="absolute right-0 z-20 mt-1 min-w-32 rounded-md border border-border bg-popover py-1 shadow">
          {items.map((t) => (<button key={t.code} type="button" className="block w-full px-3 py-1 text-left text-xs hover:bg-accent" onClick={() => {
                    setOpen(false);
                    onPick(t);
                }}>
              {t.display_name}
            </button>))}
        </div>)}
    </div>);
}
function SeriesFolders({ projectPath, types }) {
    const { selection, setSelection, refreshKey, bumpRefresh, seriesFolderOpen, toggleSeriesFolder, ensureSeriesFolder } = useAppState();
    const confirm = useConfirm();
    const notice = useNotice();
    const [byKind, setByKind] = useState({});
    useEffect(() => {
        void (async () => {
            const next = {};
            for (const t of types) {
                try {
                    next[t.code] = await api.listHeaderFiles(projectPath, 'series', t.code);
                }
                catch {
                    next[t.code] = { files: [], missing: true, folder: '' };
                }
            }
            setByKind(next);
        })();
    }, [projectPath, types, refreshKey]);
    async function onSeriesFiles(kind, files) {
        const list = await filesFromList(files);
        if (!list.length)
            return;
        markHtmlFileDrop();
        const created = await importDroppedOnHeader(kind, { projectPath, projectKind: 'series' }, list, notice, confirm);
        if (!created)
            return;
        ensureSeriesFolder(projectPath, kind);
        bumpRefresh();
        setSelection({ type: 'file', projectPath, projectKind: 'series', kind, rel: created.rel, title: created.title });
    }
    return (<div className="ml-[2ch] border-l border-border pl-2">
      {types.map((t) => {
            const block = byKind[t.code] || { files: [], missing: false, folder: '' };
            const items = block.files || [];
            return (<FolderBlock key={t.code} label={folderHeader(t.code, t.display_name, items.length)} open={seriesFolderOpen(projectPath, t.code)} onToggle={() => toggleSeriesFolder(projectPath, t.code)} folderLink={{ projectPath, kind: t.code, path: block.folder, missing: block.missing }} onFolderChange={bumpRefresh} fileDrop={{ kind: t.code, projectPath, projectKind: 'series' }} onFiles={(files) => void onSeriesFiles(t.code, files)} items={items.map((f) => ({ id: f.rel, title: f.name, rel: f.rel }))} onOpen={(item) => setSelection({ type: 'file', projectPath, projectKind: 'series', kind: t.code, rel: item.rel, title: item.title })} onDelete={async (rel) => {
                    if (!(await confirm('Delete this file?')))
                        return;
                    try {
                        await api.deleteHeaderFile({ project_path: projectPath, project_kind: 'series', kind: t.code, rel });
                        if (selection.type === 'file' && selection.rel === rel && selection.projectPath === projectPath)
                            setSelection({ type: 'empty' });
                        bumpRefresh();
                    }
                    catch (err) {
                        alert(err instanceof Error ? err.message : 'Could not delete');
                    }
                }}/>);
        })}
    </div>);
}
function FolderBlock({ label, open, onToggle, items, onOpen, onDelete, fileDrop, onFiles, folderLink, onFolderChange, }) {
    return (<div className="mb-0.5 rounded [--wails-drop-target:drop]" data-drop-kind={fileDrop?.kind} data-project-path={fileDrop?.projectPath} data-project-kind={fileDrop?.projectKind} onDragOver={(e) => {
            if (e.dataTransfer.types.includes('Files'))
                e.preventDefault();
        }} onDrop={(e) => {
            if (e.dataTransfer.files.length && onFiles) {
                e.preventDefault();
                void onFiles(e.dataTransfer.files);
            }
        }}>
      <button type="button" className="w-full truncate px-1 py-0 text-left text-xs font-normal leading-tight text-foreground" onClick={onToggle}>
        {label}
      </button>
      {open && folderLink && <HeaderFolderLink {...folderLink} onChanged={onFolderChange}/>}
      {open &&
            items.map((item) => (<div key={item.id} className="flex items-center pl-2">
            <button type="button" className="flex-1 truncate px-1 py-0 text-left text-xs font-normal leading-tight text-muted-foreground hover:bg-accent hover:text-foreground" onClick={() => onOpen(item)}>
              {item.title}
            </button>
            <Button type="button" size="icon" variant="ghost" className="h-5 w-5 shrink-0" onClick={(e) => {
                    e.preventDefault();
                    e.stopPropagation();
                    void onDelete(item.id);
                }}>
              <Trash2 className="pointer-events-none h-3 w-3"/>
            </Button>
          </div>))}
    </div>);
}
function StoryRow({ story, types, expanded, onToggle, onEdit }) {
    const { setSelection, bumpRefresh, ensureStoryFolder } = useAppState();
    return (<div>
      <div className="flex h-6 items-center">
        <button type="button" className="flex-1 truncate rounded px-2 py-0 text-left text-sm font-normal leading-tight text-blue-800 hover:bg-accent dark:text-blue-300 [--wails-drop-target:drop]" data-drop-story-path={story.path} onClick={onToggle}>
          {story.name}
        </button>
        <AddMenu types={types} onPick={(t) => void createHeader('story', story.path, t, setSelection, bumpRefresh, () => ensureStoryFolder(story.path, t.code))}/>
        <Button size="icon" variant="ghost" className="h-5 w-5" onClick={onEdit}>
          <Pencil className="h-3 w-3"/>
        </Button>
      </div>
      {expanded && <div className="ml-[2ch] border-l border-border pl-2"><StoryFolders projectPath={story.path} types={types}/></div>}
    </div>);
}
