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
function folderList(map, parentId) {
    return map[String(parentId)] || [];
}
function withFolder(map, parentId, code, on) {
    const key = String(parentId);
    const cur = map[key] || [];
    const has = cur.includes(code);
    if (on && has)
        return map;
    if (!on && !has)
        return map;
    const next = on ? [...cur, code] : cur.filter((c) => c !== code);
    return { ...map, [key]: next };
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
    const [seriesFolders, setSeriesFolders] = useState({});
    const [storyFolders, setStoryFolders] = useState({});
    const [seriesFoldersHidden, setSeriesFoldersHidden] = useState([]);
    const [restoreOpen, setRestoreOpenState] = useState(true);
    const [refreshKey, setRefreshKey] = useState(0);
    const ready = useRef(false);
    const snapshot = useRef({
        selection,
        openSeries,
        openStories,
        seriesFolders,
        storyFolders,
        seriesFoldersHidden,
        restoreOpen,
    });
    snapshot.current = { selection, openSeries, openStories, seriesFolders, storyFolders, seriesFoldersHidden, restoreOpen };
    async function saveSession() {
        const s = snapshot.current;
        if (!s.restoreOpen)
            return;
        const body = {
            selection: s.selection,
            openSeries: s.openSeries,
            openStories: s.openStories,
            seriesFolders: s.seriesFolders,
            storyFolders: s.storyFolders,
            seriesFoldersHidden: s.seriesFoldersHidden,
        };
        await api.putSetting(SESSION_KEY, JSON.stringify(body));
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
                        setSeriesFolders(saved.seriesFolders || {});
                        setStoryFolders(saved.storyFolders || {});
                        if (Array.isArray(saved.seriesFoldersHidden))
                            setSeriesFoldersHidden(saved.seriesFoldersHidden);
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
    }, [selection, openSeries, openStories, seriesFolders, storyFolders, seriesFoldersHidden, restoreOpen]);
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
        seriesFolderOpen: (seriesId, code) => folderList(seriesFolders, seriesId).includes(code),
        storyFolderOpen: (storyId, code) => folderList(storyFolders, storyId).includes(code),
        toggleSeriesFolder: (seriesId, code) => setSeriesFolders((m) => withFolder(m, seriesId, code, !folderList(m, seriesId).includes(code))),
        toggleStoryFolder: (storyId, code) => setStoryFolders((m) => withFolder(m, storyId, code, !folderList(m, storyId).includes(code))),
        ensureSeriesFolder: (seriesId, code) => {
            setSeriesFolders((m) => withFolder(m, seriesId, code, true));
            setSeriesFoldersHidden((ids) => ids.filter((id) => id !== seriesId));
        },
        ensureStoryFolder: (storyId, code) => setStoryFolders((m) => withFolder(m, storyId, code, true)),
        seriesFoldersOpen: (seriesId) => !seriesFoldersHidden.includes(seriesId),
        toggleSeriesFolders: (seriesId) => setSeriesFoldersHidden((ids) => toggleId(ids, seriesId)),
        restoreOpen,
        setRestoreOpen,
        refreshKey,
        bumpRefresh: () => setRefreshKey((n) => n + 1),
    }), [selection, openSeries, openStories, seriesFolders, storyFolders, seriesFoldersHidden, restoreOpen, refreshKey]);
    return <Ctx.Provider value={value}>{children}</Ctx.Provider>;
}
export function useAppState() {
    const ctx = useContext(Ctx);
    if (!ctx)
        throw new Error('useAppState outside provider');
    return ctx;
}
