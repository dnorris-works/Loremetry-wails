import { createContext, useContext, useEffect, useMemo, useState } from 'react';
import { api } from '@/api/client';

const emptySession = {
    selection: { type: 'empty', dir: '', name: '' },
    open_series: [],
    open_stories: [],
    open_folders: [],
    restore_open: true,
    last_pen: '',
};

function fromGo(sess) {
    if (!sess)
        return emptySession;
    const sel = sess.selection || emptySession.selection;
    return {
        selection: {
            type: sel.type || 'empty',
            dir: sel.dir || '',
            name: sel.name || '',
        },
        open_series: sess.open_series || sess.openSeries || [],
        open_stories: sess.open_stories || sess.openStories || [],
        open_folders: sess.open_folders || sess.openFolders || [],
        restore_open: sess.restore_open ?? sess.restoreOpen ?? true,
        last_pen: sess.last_pen || sess.lastPen || '',
    };
}

const Ctx = createContext(null);

export function AppStateProvider({ children }) {
    const [session, setSession] = useState(emptySession);
    const [refreshKey, setRefreshKey] = useState(0);
    useEffect(() => {
        let cancelled = false;
        void api.getUISession().then((sess) => {
            if (cancelled)
                return;
            const next = fromGo(sess);
            if (next.restore_open === false) {
                setSession({
                    ...next,
                    selection: emptySession.selection,
                    open_series: [],
                    open_stories: [],
                    open_folders: [],
                });
                return;
            }
            setSession(next);
        }).catch(() => { });
        return () => {
            cancelled = true;
        };
    }, []);
    async function apply(next) {
        setSession(fromGo(next));
        return next;
    }
    const value = useMemo(() => ({
        selection: session.selection || emptySession.selection,
        openSeries: session.open_series || [],
        openStories: session.open_stories || [],
        openFolders: session.open_folders || [],
        restoreOpen: session.restore_open !== false,
        lastPen: session.last_pen || '',
        setSelection: (sel) => {
            const file = sel?.type === 'file';
            return api.setUISelection(file ? sel.dir : '', file ? sel.name : '').then(apply);
        },
        toggleSeries: (path) => api.toggleOpenSeries(path).then(apply),
        toggleStory: (path) => api.toggleOpenStory(path).then(apply),
        ensureOpenSeries: (path) => api.ensureOpenSeries(path).then(apply),
        ensureOpenStory: (path) => api.ensureOpenStory(path).then(apply),
        toggleFolder: (path) => api.toggleOpenFolder(path).then(apply),
        ensureFolder: (path) => api.ensureOpenFolder(path).then(apply),
        setRestoreOpen: (on) => api.setRestoreOpen(on).then(apply),
        setLastPen: (name) => api.setLastPen(name).then(apply),
        refreshKey,
        bumpRefresh: () => setRefreshKey((n) => n + 1),
    }), [session, refreshKey]);
    return <Ctx.Provider value={value}>{children}</Ctx.Provider>;
}
export function useAppState() {
    const ctx = useContext(Ctx);
    if (!ctx)
        throw new Error('useAppState outside provider');
    return ctx;
}
