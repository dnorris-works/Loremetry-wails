import { useCallback, useEffect, useState } from 'react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { ChevronDown, ChevronUp, Search } from 'lucide-react';

export function useReportSearch(searchApiRef) {
    const [query, setQuery] = useState('');
    const [activeIndex, setActiveIndex] = useState(-1);
    const [matchCount, setMatchCount] = useState(0);

    useEffect(() => {
        setActiveIndex(-1);
        setMatchCount(0);
    }, [query]);

    const step = useCallback((delta) => {
        searchApiRef.current?.step?.(delta);
    }, [searchApiRef]);

    const onMatchCount = useCallback((count) => {
        setMatchCount(count);
        setActiveIndex((i) => {
            if (count === 0)
                return -1;
            if (i < 0)
                return 0;
            if (i >= count)
                return count - 1;
            return i;
        });
    }, []);

    return {
        query,
        setQuery,
        activeIndex,
        setActiveIndex,
        matchCount,
        onMatchCount,
        step,
    };
}

export function ReportSearchBar({ query, setQuery, matchCount, activeIndex, onPrev, onNext }) {
    const label = !query.trim()
        ? ''
        : matchCount === 0
            ? 'No matches'
            : `${activeIndex + 1} / ${matchCount}`;
    return (
        <div className="flex min-w-0 flex-1 items-center gap-1">
            <Search className="h-3.5 w-3.5 shrink-0 text-muted-foreground"/>
            <Input
                className="h-7 min-w-0 flex-1 text-xs"
                placeholder="Search report"
                value={query}
                onChange={(e) => setQuery(e.target.value)}
                aria-label="Search report"
            />
            <span className="w-14 shrink-0 text-right text-[10px] text-muted-foreground">{label}</span>
            <Button size="icon" variant="ghost" className="h-7 w-7" disabled={!matchCount} onClick={onPrev} title="Previous match">
                <ChevronUp className="h-3.5 w-3.5"/>
            </Button>
            <Button size="icon" variant="ghost" className="h-7 w-7" disabled={!matchCount} onClick={onNext} title="Next match">
                <ChevronDown className="h-3.5 w-3.5"/>
            </Button>
        </div>
    );
}
