import { useMemo } from 'react';
import { useIncrementalDomSearch } from '@/editor/useIncrementalDomSearch';
import { renderReportHtml } from '@/editor/renderReportHtml';

export function MarkdownReport({
    markdown: text = '',
    searchQuery = '',
    searchActiveIndex = -1,
    setSearchActiveIndex,
    onSearchMatchCount,
    searchApiRef,
}) {
    const html = useMemo(() => renderReportHtml(text), [text]);
    const containerRef = useIncrementalDomSearch(searchApiRef, {
        query: searchQuery,
        activeIndex: searchActiveIndex,
        setActiveIndex: setSearchActiveIndex || (() => {}),
        onMatchCount: onSearchMatchCount || (() => {}),
        rescanDeps: [text],
    });

    return (
        <div ref={containerRef} className="report-view h-full overflow-auto p-6">
            <div
                className="prose prose-sm dark:prose-invert max-w-3xl"
                dangerouslySetInnerHTML={{ __html: html }}
            />
        </div>
    );
}
