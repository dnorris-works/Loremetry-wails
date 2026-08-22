import { useCallback, useEffect, useRef, useState } from 'react';
import { api } from '@/api/client';
import { useAppState } from '@/lib/app-state';
import { useConfirm } from '@/components/confirm-dialog';
import { LexicalEditor } from '@/editor/LexicalEditor';
import { LargeFileViewer } from '@/editor/LargeFileViewer';
import { AnalysisPane, ReportPane } from '@/panels/AnalysisPane';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { X } from 'lucide-react';
import { EventsOn } from '../../wailsjs/runtime/runtime';

function useDiskChange(selection, dirty, onReload) {
    const confirm = useConfirm();
    const dirtyRef = useRef(dirty);
    dirtyRef.current = dirty;
    const selRef = useRef(selection);
    selRef.current = selection;
    useEffect(() => {
        if (selection.type !== 'file')
            return undefined;
        return EventsOn('folders-changed', (payload) => {
            const sel = selRef.current;
            if (sel.type !== 'file')
                return;
            // Only prompt for explicit changed-file lists from the disk watcher.
            // Generic refreshes (empty string / no list) are for the sidebar only.
            const changed = Array.isArray(payload) ? payload : null;
            if (!changed || !changed.length)
                return;
            const openFull = `${sel.dir}/${sel.name}`.replace(/\\/g, '/');
            const hit = changed.some((rel) => {
                const r = String(rel || '').replace(/\\/g, '/');
                return r && (openFull === r || openFull.endsWith(`/${r}`));
            });
            if (!hit)
                return;
            void (async () => {
                const msg = dirtyRef.current
                    ? 'This file changed on disk. Reload and discard your unsaved edits?'
                    : 'This file changed on disk. Reload it?';
                if (await confirm(msg))
                    onReload();
            })();
        });
    }, [selection.type, selection.name, selection.dir, confirm, onReload]);
}
function selectionKey(selection) {
    if (selection.type === 'file')
        return `${selection.dir}:${selection.name}`;
    return '';
}

function displayFileName(name) {
    const base = String(name || '').trim();
    if (!base)
        return '';
    return base.replace(/\.(md|markdown|txt|text)$/i, '');
}
export function DocumentPane() {
    const { selection, setSelection, bumpRefresh, analysisId, reportId } = useAppState();
    const confirm = useConfirm();
    const [loaded, setLoaded] = useState(null);
    const [title, setTitle] = useState('');
    const [draft, setDraft] = useState('');
    const [editorNonce, setEditorNonce] = useState(0);
    const [reloadTick, setReloadTick] = useState(0);
    const [largeFile, setLargeFile] = useState(null);
    const key = selectionKey(selection);
    const dirty = !!loaded && (title !== loaded.title || draft !== loaded.markdown);
    const reloadFromDisk = useCallback(() => setReloadTick((n) => n + 1), []);
    useDiskChange(selection, dirty, reloadFromDisk);
    useEffect(() => {
        let cancelled = false;
        async function load() {
            if (selection.type !== 'file') {
                if (!cancelled) {
                    setLoaded(null);
                    setLargeFile(null);
                }
                return;
            }
            if (!cancelled) {
                setLoaded(null);
                setLargeFile(null);
            }
            try {
                const opened = await api.openDiskFile(selection.dir, selection.name);
                if (cancelled)
                    return;
                if (opened.large) {
                    setLargeFile({ dir: selection.dir, name: opened.name, fileSize: opened.file_size || opened.fileSize });
                    const label = displayFileName(opened.name);
                    setLoaded({ key, markdown: '', title: label });
                    setTitle(label);
                } else {
                    setLargeFile(null);
                    const label = displayFileName(opened.name);
                    setLoaded({ key, markdown: opened.text || '', title: label });
                    setTitle(label);
                    setDraft(opened.text || '');
                    setEditorNonce((n) => n + 1);
                }
            }
            catch {
                if (!cancelled) {
                    setLoaded(null);
                    setLargeFile(null);
                    if (reloadTick)
                        setSelection({ type: 'empty' });
                }
            }
        }
        void load();
        return () => {
            cancelled = true;
        };
    }, [key, reloadTick]);
    async function close() {
        if (!largeFile && dirty && !(await confirm('Discard unsaved changes?')))
            return;
        setSelection({ type: 'empty' });
    }
    function cancel() {
        if (!loaded)
            return;
        setTitle(loaded.title);
        setDraft(loaded.markdown);
        setEditorNonce((n) => n + 1);
    }
    async function save() {
        const name = title.trim();
        if (!name || !loaded || selection.type !== 'file')
            return;
        try {
            const saved = await api.writeDiskFile({
                dir: selection.dir,
                name: selection.name,
                new_name: name,
                text: draft,
            });
            setSelection({ type: 'file', dir: selection.dir, name: saved.name });
            const label = displayFileName(saved.name);
            setLoaded({ key: `${selection.dir}:${saved.name}`, markdown: draft, title: label });
            setTitle(label);
            bumpRefresh();
        }
        catch (err) {
            alert(err instanceof Error ? err.message : 'Could not save');
        }
    }
    async function remove() {
        const name = title.trim() || loaded?.title || 'this file';
        if (!(await confirm(`Delete "${name}"?`)))
            return;
        try {
            if (selection.type === 'file')
                await api.deleteDiskFile(selection.dir, selection.name);
            setSelection({ type: 'empty' });
            bumpRefresh();
        }
        catch (err) {
            alert(err instanceof Error ? err.message : 'Could not delete');
        }
    }
    if (reportId) {
        return <ReportPane />;
    }
    if (analysisId) {
        return <AnalysisPane />;
    }
    if (selection.type === 'empty') {
        return (<div className="flex h-full items-center justify-center text-sm text-muted-foreground">
        Select a document or analysis from the sidebar.
      </div>);
    }
    if (!loaded || loaded.key !== key) {
        return <div className="h-full"/>;
    }
    if (largeFile) {
        return (<div className="flex h-full flex-col">
          <div className="flex items-center gap-2 border-b border-border px-4 py-2">
            <div className="min-w-0 flex-1">
              <div className="truncate text-sm font-medium">{title}</div>
              <div className="text-[10px] text-muted-foreground">
                {Math.round(largeFile.fileSize / 1024)} KB — read-only (large file)
              </div>
            </div>
            <Button size="sm" variant="destructive" onClick={() => void remove()}>Delete</Button>
            <Button size="icon" variant="ghost" onClick={() => void close()} title="Close">
              <X className="h-4 w-4"/>
            </Button>
          </div>
          <div className="flex-1 overflow-hidden">
            <LargeFileViewer dir={largeFile.dir} name={largeFile.name} fileSize={largeFile.fileSize}/>
          </div>
        </div>);
    }
    return (<div className="flex h-full flex-col">
      <div className="flex items-center gap-2 border-b border-border px-4 py-2">
        <div className="min-w-0 flex-1">
          <Input className="h-8 font-medium" value={title} onChange={(e) => setTitle(e.target.value)} aria-label="Document name"/>
        </div>
        <Button size="sm" onClick={() => void save()} disabled={!dirty}>Save</Button>
        <Button size="sm" variant="outline" onClick={cancel} disabled={!dirty}>Cancel</Button>
        <Button size="sm" variant="destructive" onClick={() => void remove()}>Delete</Button>
        <Button size="icon" variant="ghost" onClick={() => void close()} title="Close">
          <X className="h-4 w-4"/>
        </Button>
      </div>
      <div className="flex-1 overflow-hidden">
        <LexicalEditor docKey={`${loaded.key}:${editorNonce}`} markdown={loaded.markdown} onMarkdown={setDraft}/>
      </div>
    </div>);
}
