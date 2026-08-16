import { useEffect, useState } from 'react';
import { api } from '@/api/client';
import { useAppState } from '@/lib/app-state';
import { useConfirm, useNotice } from '@/components/confirm-dialog';
import { filesFromList, markHtmlFileDrop } from '@/lib/import-docs';
import { importDroppedOnHeader } from '@/lib/drop-import';
import { Button } from '@/components/ui/button';
import { Trash2 } from 'lucide-react';
import { cn } from '@/lib/utils';
import { folderHeader } from '@/lib/folder-label';
import { HeaderFolderLink } from '@/lib/HeaderFolderLink';
import { clearSidebarDrag, currentSidebarDrag, hasFiles, moveBefore, setSidebarDrag, sidebarDrag, } from '@/lib/sidebar-drag';
export function StoryFolders({ projectPath, types }) {
    const { selection, setSelection, refreshKey, bumpRefresh, storyFolderOpen, toggleStoryFolder, ensureStoryFolder } = useAppState();
    const confirm = useConfirm();
    const notice = useNotice();
    const [byKind, setByKind] = useState({});
    const [over, setOver] = useState(null);
    useEffect(() => {
        void (async () => {
            const next = {};
            for (const t of types) {
                try {
                    next[t.code] = await api.listHeaderFiles(projectPath, 'story', t.code);
                }
                catch {
                    next[t.code] = { files: [], missing: true, folder: '' };
                }
            }
            setByKind(next);
        })();
    }, [projectPath, types, refreshKey]);
    async function dropFiles(kind, files) {
        if (!files?.length)
            return;
        const list = await filesFromList(files);
        if (!list.length)
            return;
        markHtmlFileDrop();
        const created = await importDroppedOnHeader(kind, { projectPath, projectKind: 'story' }, list, notice, confirm);
        if (!created)
            return;
        ensureStoryFolder(projectPath, kind);
        bumpRefresh();
        setSelection({ type: 'file', projectPath, projectKind: 'story', kind, rel: created.rel, title: created.title });
    }
    async function placeDoc(kind, rel, beforeRel) {
        const items = (byKind[kind]?.files || []).map((f) => f.rel);
        const next = moveBefore(items, rel, beforeRel);
        try {
            await api.placeHeaderFiles({ project_path: projectPath, project_kind: 'story', kind, rels: next });
            ensureStoryFolder(projectPath, kind);
            bumpRefresh();
        }
        catch (err) {
            alert(err instanceof Error ? err.message : 'Could not move');
        }
    }
    return (<div>
      {types.map((t) => {
            const block = byKind[t.code] || { files: [], missing: false, folder: '' };
            const items = block.files || [];
            const dropKey = `doc:${t.code}`;
            return (<div key={t.code} className={cn('rounded [--wails-drop-target:drop]', over === dropKey && 'bg-accent ring-1 ring-primary')} data-project-path={projectPath} data-project-kind="story" data-drop-kind={t.code} onDragOver={(e) => {
                    const drag = currentSidebarDrag();
                    if (hasFiles(e) || drag?.t === 'file') {
                        e.preventDefault();
                        setOver(dropKey);
                    }
                }} onDragLeave={() => setOver((k) => (k === dropKey ? null : k))} onDrop={(e) => {
                    e.preventDefault();
                    setOver(null);
                    if (hasFiles(e) && e.dataTransfer.files.length) {
                        markHtmlFileDrop();
                        void dropFiles(t.code, e.dataTransfer.files);
                        return;
                    }
                    const drag = sidebarDrag(e);
                    clearSidebarDrag();
                    if (drag?.t === 'file' && drag.kind === t.code)
                        void placeDoc(t.code, drag.rel, null);
                }}>
            <button type="button" className="w-full truncate px-1 py-0 text-left text-xs font-normal leading-tight text-foreground" onClick={() => toggleStoryFolder(projectPath, t.code)}>
              {folderHeader(t.code, t.display_name, items.length)}
            </button>
            {storyFolderOpen(projectPath, t.code) && <HeaderFolderLink projectPath={projectPath} kind={t.code} path={block.folder} missing={block.missing} onChanged={bumpRefresh}/>}
            {storyFolderOpen(projectPath, t.code) &&
                    items.map((doc) => (<div key={doc.rel} className="flex items-center pl-2" draggable onDragStart={(e) => setSidebarDrag(e, { t: 'file', rel: doc.rel, kind: t.code, projectPath })} onDragEnd={clearSidebarDrag} onDragOver={(e) => {
                            const drag = currentSidebarDrag();
                            if (drag?.t === 'file')
                                e.preventDefault();
                        }} onDrop={(e) => {
                            e.preventDefault();
                            e.stopPropagation();
                            const drag = sidebarDrag(e);
                            clearSidebarDrag();
                            if (drag?.t === 'file' && drag.rel !== doc.rel)
                                void placeDoc(t.code, drag.rel, doc.rel);
                        }}>
                  <button type="button" className="flex-1 truncate px-1 py-0 text-left text-xs font-normal leading-tight text-muted-foreground hover:bg-accent hover:text-foreground" onClick={() => setSelection({ type: 'file', projectPath, projectKind: 'story', kind: t.code, rel: doc.rel, title: doc.name })}>
                    {doc.name}
                  </button>
                  <Button size="icon" variant="ghost" className="h-5 w-5" onClick={async (e) => {
                            e.preventDefault();
                            e.stopPropagation();
                            if (!(await confirm('Delete this file?')))
                                return;
                            try {
                                await api.deleteHeaderFile({ project_path: projectPath, project_kind: 'story', kind: t.code, rel: doc.rel });
                                if (selection.type === 'file' && selection.rel === doc.rel && selection.projectPath === projectPath)
                                    setSelection({ type: 'empty' });
                                bumpRefresh();
                            }
                            catch (err) {
                                alert(err instanceof Error ? err.message : 'Could not delete');
                            }
                        }}>
                    <Trash2 className="pointer-events-none h-3 w-3"/>
                  </Button>
                </div>))}
          </div>);
        })}
    </div>);
}
