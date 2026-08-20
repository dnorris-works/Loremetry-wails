import { useCallback, useEffect, useRef, useState } from 'react';
import { marked } from 'marked';

const PAGE = 32_000;
const LOAD_THRESHOLD = 400;
export const REPORT_INLINE_MAX = 50_000;

function MarkdownChunk({ text }) {
    return (
        <div
            className="prose prose-sm dark:prose-invert max-w-3xl"
            dangerouslySetInnerHTML={{ __html: marked.parse(text || '', { breaks: true }) }}
        />
    );
}

export function MarkdownReport({ text }) {
    if ((text?.length || 0) > REPORT_INLINE_MAX) {
        return <LazyTextReport text={text} />;
    }
    return <MarkdownChunk text={text} />;
}

function TextChunk({ text }) {
    return (
        <pre className="whitespace-pre-wrap break-words font-sans text-sm leading-relaxed text-foreground">
            {text}
        </pre>
    );
}

function LazyScrollReport({ loadKey, bodySize, loadRange, rich }) {
    const containerRef = useRef(null);
    const [pages, setPages] = useState([]);
    const [loadedEnd, setLoadedEnd] = useState(0);
    const [loading, setLoading] = useState(false);
    const loadingRef = useRef(false);
    const loadedEndRef = useRef(0);
    const userScrolledRef = useRef(false);
    const loadRangeRef = useRef(loadRange);
    loadRangeRef.current = loadRange;

    const loadMore = useCallback(async () => {
        if (loadingRef.current || loadedEndRef.current >= bodySize)
            return;
        loadingRef.current = true;
        setLoading(true);
        try {
            const range = await loadRangeRef.current(loadedEndRef.current, PAGE);
            const chunk = range?.text || '';
            if (!chunk)
                return;
            const end = range.end ?? range.End ?? (loadedEndRef.current + chunk.length);
            setPages((prev) => [...prev, chunk]);
            loadedEndRef.current = end;
            setLoadedEnd(end);
        } finally {
            loadingRef.current = false;
            setLoading(false);
        }
    }, [bodySize]);

    useEffect(() => {
        setPages([]);
        loadedEndRef.current = 0;
        setLoadedEnd(0);
        loadingRef.current = false;
        userScrolledRef.current = false;
        void loadMore();
    }, [loadKey, bodySize, loadMore]);

    useEffect(() => {
        const el = containerRef.current;
        if (!el)
            return;
        function onScroll() {
            userScrolledRef.current = true;
            const remaining = el.scrollHeight - el.scrollTop - el.clientHeight;
            if (remaining < LOAD_THRESHOLD) {
                void loadMore();
            }
        }
        el.addEventListener('scroll', onScroll, { passive: true });
        return () => el.removeEventListener('scroll', onScroll);
    }, [loadMore]);

    const pct = bodySize > 0 ? Math.round((loadedEnd / bodySize) * 100) : 100;
    const Chunk = rich ? MarkdownChunk : TextChunk;

    return (
        <div ref={containerRef} className="h-full overflow-auto">
            <div className="space-y-4 p-6">
                {pages.map((chunk, i) => (
                    <Chunk key={i} text={chunk}/>
                ))}
            </div>
            {loadedEnd < bodySize && (
                <div className="px-6 pb-4 text-xs text-muted-foreground">
                    {loading ? 'Loading…' : userScrolledRef.current ? `${pct}% loaded — scroll for more` : 'Scroll down for more'}
                </div>
            )}
        </div>
    );
}

export function LazyTextReport({ text }) {
    const loadRange = useCallback((offset, limit) => {
        const end = Math.min(offset + limit, text.length);
        return Promise.resolve({ text: text.slice(offset, end), end });
    }, [text]);
    return <LazyScrollReport loadKey={text.length} bodySize={text.length} loadRange={loadRange} rich={false}/>;
}

export function LazySavedReport({ reportId, bodySize, loadRange }) {
    return <LazyScrollReport loadKey={reportId} bodySize={bodySize} loadRange={loadRange} rich={false}/>;
}
