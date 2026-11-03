import { useCallback, useEffect, useImperativeHandle, useRef } from 'react';
import { applyDomSearch, clearDomSearch, scrollToActiveMatch } from '@/editor/reportDomSearch';

export function useIncrementalDomSearch(ref, {
    query,
    activeIndex,
    setActiveIndex,
    onMatchCount,
    rescanDeps = [],
}) {
    const containerRef = useRef(null);
    const queryRef = useRef(query);
    const activeIndexRef = useRef(activeIndex);
    queryRef.current = query;
    activeIndexRef.current = activeIndex;

    const rescan = useCallback(() => {
        const root = containerRef.current;
        const q = queryRef.current.trim();
        if (!q || !root) {
            clearDomSearch(root);
            onMatchCount(0);
            return 0;
        }
        const idx = Math.max(0, activeIndexRef.current);
        const count = applyDomSearch(root, q, idx);
        onMatchCount(count);
        if (count > 0 && activeIndexRef.current < 0)
            setActiveIndex(0);
        requestAnimationFrame(() => scrollToActiveMatch(root));
        return count;
    }, [onMatchCount, setActiveIndex]);

    useEffect(() => {
        const root = containerRef.current;
        if (!query.trim()) {
            clearDomSearch(root);
            onMatchCount(0);
            setActiveIndex(-1);
            return;
        }
        const timer = window.setTimeout(() => {
            rescan();
        }, 200);
        return () => window.clearTimeout(timer);
    }, [query, ...rescanDeps, rescan, onMatchCount, setActiveIndex]);

    useEffect(() => {
        if (!query.trim() || activeIndex < 0)
            return;
        const root = containerRef.current;
        applyDomSearch(root, query, activeIndex);
        requestAnimationFrame(() => scrollToActiveMatch(root));
    }, [activeIndex, query]);

    const step = useCallback((delta) => {
        const q = queryRef.current.trim();
        if (!q)
            return;
        const root = containerRef.current;
        let count = root?.querySelectorAll('mark.search-hit').length || 0;
        let idx = activeIndexRef.current;

        if (count === 0) {
            count = rescan();
            if (count > 0)
                setActiveIndex(delta > 0 ? 0 : count - 1);
            return;
        }

        const nextIdx = delta > 0
            ? (idx < 0 ? 0 : Math.min(idx + 1, count - 1))
            : (idx <= 0 ? count - 1 : idx - 1);

        setActiveIndex(nextIdx);
    }, [rescan, setActiveIndex]);

    useImperativeHandle(ref, () => ({ step, rescan }), [step, rescan]);

    return containerRef;
}
