import { createContext, useContext, useEffect, useMemo, useRef, useState } from 'react';
import { api } from '@/api/client';
const RESTORE_KEY = 'restore_open_items';
const SESSION_KEY = 'last_open_session';
const Ctx = createContext(null);
function toggleId(ids, id) {
    return ids.includes(id) ? ids.filter((x) => x !== id) : [...ids, id];
}
function addId(ids, id) {
    return ids.includes(id) ? ids : [...ids, id];
}
async function readSetting(key) {
    try {
        const row = await api.getSetting(key);
        return row.value;
    }
    catch {
        return null;
    }
}
export function AppStateProvider({ children }) {
    const [selection, setSelection] = useState({ type: 'empty' });
    const [openSeries, setOpenSeries] = useState([]);
    const [openStories, setOpenStories] = useState([]);
    const [openFolders, setOpenFolders] = useState([]);
    const [restoreOpen, setRestoreOpenState] = useState(true);
    const [refreshKey, setRefreshKey] = useState(0);
    const ready = useRef(false);
    const snapshot = useRef({ selection, openSeries, openStories, openFolders, restoreOpen });
    snapshot.current = { selection, openSeries, openStories, openFolders, restoreOpen };
    async function saveSession() {
        const s = snapshot.current;
        if (!s.restoreOpen)
            return;
        await api.putSetting(SESSION_KEY, JSON.stringify({
            selection: s.selection,
            openSeries: s.openSeries,
            openStories: s.openStories,
            openFolders: s.openFolders,
        }));
    }
    useEffect(() => {
        let cancelled = false;
        void (async () => {
            const restoreVal = await readSetting(RESTORE_KEY);
            const restore = restoreVal !== 'false';
            if (cancelled)
                return;
            setRestoreOpenState(restore);
            if (restore) {
                const raw = await readSetting(SESSION_KEY);
                if (raw && !cancelled) {
                    try {
                        const saved = JSON.parse(raw);
                        if (saved.selection && saved.selection.type === 'file')
                            setSelection(saved.selection);
                        setOpenSeries(Array.isArray(saved.openSeries) ? saved.openSeries : []);
                        setOpenStories(Array.isArray(saved.openStories) ? saved.openStories : []);
                        setOpenFolders(Array.isArray(saved.openFolders) ? saved.openFolders : []);
                    }
                    catch {
                        /* ignore */
                    }
                }
            }
            ready.current = true;
        })();
        return () => {
            cancelled = true;
        };
    }, []);
    useEffect(() => {
        if (!ready.current || !restoreOpen)
            return;
        const t = window.setTimeout(() => {
            void saveSession();
        }, 300);
        return () => window.clearTimeout(t);
    }, [selection, openSeries, openStories, openFolders, restoreOpen]);
    useEffect(() => {
        function flush() {
            if (document.visibilityState === 'hidden')
                void saveSession();
        }
        document.addEventListener('visibilitychange', flush);
        window.addEventListener('beforeunload', flush);
        return () => {
            document.removeEventListener('visibilitychange', flush);
            window.removeEventListener('beforeunload', flush);
        };
    }, []);
    async function setRestoreOpen(v) {
        setRestoreOpenState(v);
        snapshot.current.restoreOpen = v;
        await api.putSetting(RESTORE_KEY, v ? 'true' : 'false');
        if (v)
            await saveSession();
    }
    const value = useMemo(() => ({
        selection,
        setSelection,
        openSeries,
        openStories,
        toggleSeries: (id) => setOpenSeries((ids) => toggleId(ids, id)),
        toggleStory: (id) => setOpenStories((ids) => toggleId(ids, id)),
        ensureOpenSeries: (id) => setOpenSeries((ids) => addId(ids, id)),
        ensureOpenStory: (id) => setOpenStories((ids) => addId(ids, id)),
        folderOpen: (path) => openFolders.includes(path),
        toggleFolder: (path) => setOpenFolders((ids) => toggleId(ids, path)),
        ensureFolder: (path) => setOpenFolders((ids) => addId(ids, path)),
        restoreOpen,
        setRestoreOpen,
        refreshKey,
        bumpRefresh: () => setRefreshKey((n) => n + 1),
    }), [selection, openSeries, openStories, openFolders, restoreOpen, refreshKey]);
    return <Ctx.Provider value={value}>{children}</Ctx.Provider>;
}
export function useAppState() {
    const ctx = useContext(Ctx);
    if (!ctx)
        throw new Error('useAppState outside provider');
    return ctx;
}
