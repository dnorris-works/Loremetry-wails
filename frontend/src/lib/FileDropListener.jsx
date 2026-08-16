import { useEffect } from 'react';
import { OnFileDrop, OnFileDropOff } from '../../wailsjs/runtime/runtime';
import { api } from '@/api/client';
import { importDroppedOnHeader } from '@/lib/drop-import';
import { importStoryFolder } from '@/lib/folder-import';
import { basename, recentHtmlFileDrop } from '@/lib/import-docs';
import { currentSidebarDrag } from '@/lib/sidebar-drag';
import { useAppState } from '@/lib/app-state';
import { useConfirm, useNotice } from '@/components/confirm-dialog';
export function FileDropListener() {
    const { bumpRefresh, setSelection, ensureStoryFolder, ensureSeriesFolder } = useAppState();
    const confirm = useConfirm();
    const notice = useNotice();
    useEffect(() => {
        let cancelled = false;
        let attached = false;
        const timer = window.setInterval(() => {
            const runtime = window.runtime;
            if (cancelled || attached || !runtime?.OnFileDrop)
                return;
            attached = true;
            window.clearInterval(timer);
            OnFileDrop((x, y, paths) => {
                if (recentHtmlFileDrop())
                    return;
                if (currentSidebarDrag())
                    return;
                const real = (paths || []).filter((p) => typeof p === 'string' && p.trim());
                if (!real.length)
                    return;
                const el = document.elementFromPoint(x, y);
                const target = el?.closest('[data-drop-kind], [data-drop-story-path]');
                if (!target) {
                    void notice('Drop a folder on a story, or files on Chapters, Characters, or Bible Docs.');
                    return;
                }
                const kind = target.dataset.dropKind;
                const projectPath = target.dataset.projectPath || target.dataset.dropStoryPath;
                const projectKind = target.dataset.projectKind || (target.dataset.dropStoryPath ? 'story' : '');
                const onStory = Boolean(target.dataset.dropStoryPath);
                void (async () => {
                    try {
                        const files = await api.readImportFiles(real);
                        const tree = files.some((f) => (f.rel || '').includes('/'));
                        let created;
                        if (projectPath && projectKind === 'story' && (onStory || tree) && !kind) {
                            created = await importStoryFolder(projectPath, files, basename(real[0]), notice);
                        }
                        else if (kind && projectPath) {
                            created = await importDroppedOnHeader(kind, { projectPath, projectKind }, files, notice, confirm);
                        }
                        else {
                            await notice('Drop a folder on a story, or files on a folder header.');
                            return;
                        }
                        if (!created)
                            return;
                        if (projectKind === 'story' && kind)
                            ensureStoryFolder(projectPath, kind);
                        if (projectKind === 'series' && kind)
                            ensureSeriesFolder(projectPath, kind);
                        bumpRefresh();
                        const rel = created.rel;
                        const fileKind = created.kind || kind;
                        if (rel && fileKind) {
                            setSelection({ type: 'file', projectPath, projectKind: projectKind || 'story', kind: fileKind, rel, title: created.title });
                        }
                    }
                    catch (err) {
                        await notice(err instanceof Error ? err.message : 'Could not import');
                    }
                })();
            }, false);
        }, 100);
        return () => {
            cancelled = true;
            window.clearInterval(timer);
            if (attached)
                OnFileDropOff();
        };
    }, [bumpRefresh, setSelection, ensureStoryFolder, ensureSeriesFolder, confirm, notice]);
    return null;
}
