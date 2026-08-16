import { useCallback, useEffect, useRef, useState } from 'react';
import { api } from '@/api/client';
import { useAppState } from '@/lib/app-state';
import { useConfirm } from '@/components/confirm-dialog';
import { LexicalEditor } from '@/editor/LexicalEditor';
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
        return EventsOn('folders-changed', (path) => {
            const sel = selRef.current;
            if (sel.type !== 'file')
                return;
            const changed = typeof path === 'string' ? path : '';
            if (changed && sel.rel && !changed.replace(/\\/g, '/').includes(sel.rel.replace(/\\/g, '/')))
                return;
            void (async () => {
                const msg = dirtyRef.current
                    ? 'This file changed on disk. Reload and discard your unsaved edits?'
                    : 'This file changed on disk. Reload it?';
                if (await confirm(msg))
                    onReload();
            })();
        });
    }, [selection.type, selection.rel, confirm, onReload]);
}
function selectionKey(selection) {
    if (selection.type === 'file')
        return `${selection.projectPath}:${selection.kind}:${selection.rel}`;
    return '';
}
export function DocumentPane() {
    const { selection, setSelection, bumpRefresh } = useAppState();
    const confirm = useConfirm();
    const [loaded, setLoaded] = useState(null);
    const [title, setTitle] = useState('');
    const [draft, setDraft] = useState('');
    const [editorNonce, setEditorNonce] = useState(0);
    const [reloadTick, setReloadTick] = useState(0);
    const key = selectionKey(selection);
    const dirty = !!loaded && (title !== loaded.title || draft !== loaded.markdown);
    const reloadFromDisk = useCallback(() => setReloadTick((n) => n + 1), []);
    useDiskChange(selection, dirty, reloadFromDisk);
    useEffect(() => {
        let cancelled = false;
        async function load() {
            if (selection.type !== 'file') {
                if (!cancelled)
                    setLoaded(null);
                return;
            }
            if (!cancelled)
                setLoaded(null);
            try {
                const doc = await api.readHeaderFile({
                    project_path: selection.projectPath,
                    project_kind: selection.projectKind,
                    kind: selection.kind,
                    rel: selection.rel,
                });
                if (cancelled)
                    return;
                setLoaded({ key, markdown: doc.text || '', title: doc.name });
                setTitle(doc.name);
                setDraft(doc.text || '');
                setEditorNonce((n) => n + 1);
            }
            catch {
                if (!cancelled) {
                    setLoaded(null);
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
        if (dirty && !(await confirm('Discard unsaved changes?')))
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
            const saved = await api.writeHeaderFile({
                project_path: selection.projectPath,
                project_kind: selection.projectKind,
                kind: selection.kind,
                rel: selection.rel,
                name,
                text: draft,
            });
            setSelection({ ...selection, rel: saved.rel, title: saved.name });
            setLoaded({ ...loaded, key: `${selection.projectPath}:${selection.kind}:${saved.rel}`, title: saved.name, markdown: draft });
            setTitle(saved.name);
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
            if (selection.type === 'file') {
                await api.deleteHeaderFile({
                    project_path: selection.projectPath,
                    project_kind: selection.projectKind,
                    kind: selection.kind,
                    rel: selection.rel,
                });
            }
            setSelection({ type: 'empty' });
            bumpRefresh();
        }
        catch (err) {
            alert(err instanceof Error ? err.message : 'Could not delete');
        }
    }
    if (selection.type === 'empty') {
        return (<div className="flex h-full items-center justify-center text-sm text-muted-foreground">
        Select a document from the sidebar.
      </div>);
    }
    if (!loaded || loaded.key !== key) {
        return <div className="p-6 text-sm text-muted-foreground">Loading…</div>;
    }
    return (<div className="flex h-full flex-col">
      <EditorActions dirty={dirty} onSave={() => void save()} onCancel={cancel} onClose={close} onDelete={() => void remove()}>
        <Input className="h-8 font-medium" value={title} onChange={(e) => setTitle(e.target.value)} aria-label="Document name"/>
      </EditorActions>
      <div className="flex-1 overflow-hidden">
        <LexicalEditor docKey={`${loaded.key}:${editorNonce}`} markdown={loaded.markdown} onMarkdown={setDraft}/>
      </div>
    </div>);
}
function EditorActions({ dirty, onSave, onCancel, onClose, onDelete, children, }) {
    return (<div className="flex items-center gap-2 border-b border-border px-4 py-2">
      <div className="min-w-0 flex-1">{children}</div>
      <Button size="sm" onClick={onSave} disabled={!dirty}>
        Save
      </Button>
      <Button size="sm" variant="outline" onClick={onCancel} disabled={!dirty}>
        Cancel
      </Button>
      {onDelete && (<Button size="sm" variant="destructive" onClick={onDelete}>
          Delete
        </Button>)}
      <Button size="icon" variant="ghost" onClick={onClose} title="Close">
        <X className="h-4 w-4"/>
      </Button>
    </div>);
}
