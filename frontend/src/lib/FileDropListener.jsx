import { useEffect } from 'react';
import { OnFileDrop, OnFileDropOff } from '../../wailsjs/runtime/runtime';
import { api } from '@/api/client';
import { isImportableName, recentHtmlFileDrop } from '@/lib/import-docs';
import { currentSidebarDrag } from '@/lib/sidebar-drag';
import { useAppState } from '@/lib/app-state';
import { useNotice } from '@/components/confirm-dialog';
export function FileDropListener() {
    const { bumpRefresh, setSelection, ensureFolder } = useAppState();
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
                const target = el?.closest('[data-folder-path]');
                const dir = target?.dataset.folderPath;
                if (!dir) {
                    void notice('Drop files on a folder header.');
                    return;
                }
                void (async () => {
                    try {
                        const files = await api.readImportFiles(real);
                        let last;
                        for (const f of files) {
                            if (!isImportableName(f.name))
                                continue;
                            last = await api.createDiskFile(dir, f.name, f.text);
                        }
                        if (!last) {
                            await notice('Nothing to import.');
                            return;
                        }
                        ensureFolder(dir);
                        bumpRefresh();
                        setSelection({ type: 'file', dir, name: last.name });
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
    }, [bumpRefresh, setSelection, ensureFolder, notice]);
    return null;
}
