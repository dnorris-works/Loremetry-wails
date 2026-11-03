import { useEffect, useMemo } from 'react';
import { useIncrementalDomSearch } from '@/editor/useIncrementalDomSearch';
import { renderReportHtml } from '@/editor/renderReportHtml';

export function MarkdownReport({
    markdown: text = '',
    searchQuery = '',
    searchActiveIndex = -1,
    setSearchActiveIndex,
    onSearchMatchCount,
    searchApiRef,
    onStickyChapterClick,
}) {
    const html = useMemo(() => renderReportHtml(text), [text]);
    const containerRef = useIncrementalDomSearch(searchApiRef, {
        query: searchQuery,
        activeIndex: searchActiveIndex,
        setActiveIndex: setSearchActiveIndex || (() => {}),
        onMatchCount: onSearchMatchCount || (() => {}),
        rescanDeps: [text],
    });

    useEffect(() => {
        const root = containerRef.current;
        if (!root || !onStickyChapterClick) return undefined;

        root.querySelectorAll('a[href^="#sticky-chapter-"]').forEach((a) => {
            a.setAttribute('role', 'button');
            a.className =
                'inline-flex items-center rounded border border-border bg-background px-2 py-0.5 text-xs font-medium text-foreground no-underline hover:bg-accent';
        });

        function onClick(e) {
            const a = e.target.closest?.('a[href^="#sticky-chapter-"]');
            if (!a || !root.contains(a)) return;
            e.preventDefault();
            const href = a.getAttribute('href') || '';
            const m = href.match(/#sticky-chapter-(\d+)/);
            if (!m) return;
            onStickyChapterClick(Number(m[1]));
        }
        root.addEventListener('click', onClick);
        return () => root.removeEventListener('click', onClick);
    }, [containerRef, onStickyChapterClick, html]);

    return (
        <div ref={containerRef} className="report-view h-full overflow-auto p-6">
            <div
                className="prose prose-sm dark:prose-invert max-w-3xl"
                dangerouslySetInnerHTML={{ __html: html }}
            />
        </div>
    );
}
