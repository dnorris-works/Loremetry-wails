import { useCallback, useEffect, useRef, useState } from 'react';
import { api } from '@/api/client';

const PAGE = 32_000;
const LOAD_THRESHOLD = 400;

export function LargeFileViewer({ dir, name, fileSize }) {
    const containerRef = useRef(null);
    const [pages, setPages] = useState([]);
    const [loadedEnd, setLoadedEnd] = useState(0);
    const [loading, setLoading] = useState(false);
    const loadingRef = useRef(false);
    const loadedEndRef = useRef(0);

    const loadMore = useCallback(async () => {
        if (loadingRef.current || loadedEndRef.current >= fileSize)
            return;
        loadingRef.current = true;
        setLoading(true);
        try {
            const range = await api.readDiskFileRange(dir, name, loadedEndRef.current, PAGE);
            if (range.text) {
                setPages((prev) => [...prev, range.text]);
                loadedEndRef.current = range.end;
                setLoadedEnd(range.end);
            }
        } finally {
            loadingRef.current = false;
            setLoading(false);
        }
    }, [dir, name, fileSize]);

    useEffect(() => {
        setPages([]);
        loadedEndRef.current = 0;
        setLoadedEnd(0);
        loadingRef.current = false;
        void loadMore();
    }, [dir, name, fileSize, loadMore]);

    useEffect(() => {
        const el = containerRef.current;
        if (!el)
            return;
        function onScroll() {
            const remaining = el.scrollHeight - el.scrollTop - el.clientHeight;
            if (remaining < LOAD_THRESHOLD) {
                void loadMore();
            }
        }
        el.addEventListener('scroll', onScroll, { passive: true });
        return () => el.removeEventListener('scroll', onScroll);
    }, [loadMore]);

    const pct = fileSize > 0 ? Math.round((loadedEnd / fileSize) * 100) : 100;

    return (
        <div ref={containerRef} className="h-full overflow-auto">
            <pre className="whitespace-pre-wrap break-words p-6 text-sm leading-relaxed">
                {pages.map((text, i) => (
                    <span key={i}>{text}</span>
                ))}
            </pre>
            {loadedEnd < fileSize && (
                <div className="px-6 pb-4 text-xs text-muted-foreground">
                    {loading ? 'Loading…' : `${pct}% loaded — scroll for more`}
                </div>
            )}
        </div>
    );
}
