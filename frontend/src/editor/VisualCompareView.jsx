import { useEffect, useMemo, useState } from 'react';
import { ChevronDown, ChevronUp } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { renderReportHtml } from '@/editor/renderReportHtml';
import { diffBlocks, proposalHunks } from '@/editor/visualDiff';

function BlockHtml({ markdown, className }) {
    const html = useMemo(() => renderReportHtml(markdown || ''), [markdown]);
    if (!markdown)
        return null;
    return (
        <div
            className={className}
            dangerouslySetInnerHTML={{ __html: html }}
        />
    );
}

function CellStack({ blocks, tone }) {
    if (!blocks?.length) {
        return <div className="min-h-8" />;
    }
    const toneClass =
        tone === 'del' ? 'visual-diff-del rounded-md px-3 py-2' :
            tone === 'ins' ? 'visual-diff-ins rounded-md px-3 py-2' :
                '';
    return (
        <div className="space-y-3">
            {blocks.map((md, idx) => (
                <BlockHtml key={idx} markdown={md} className={toneClass || undefined} />
            ))}
        </div>
    );
}

/** One proposal at a time in a two-column table — one scroll for both sides. */
export function VisualCompareView({ original = '', proposed = '' }) {
    const hunks = useMemo(
        () => proposalHunks(diffBlocks(original, proposed)),
        [original, proposed],
    );
    const [index, setIndex] = useState(0);

    useEffect(() => {
        setIndex(0);
    }, [original, proposed]);

    const count = hunks.length;
    const safeIndex = count === 0 ? 0 : Math.min(index, count - 1);
    const hunk = count > 0 ? hunks[safeIndex] : null;

    function step(delta) {
        if (count === 0)
            return;
        setIndex((i) => (i + delta + count) % count);
    }

    if (!hunk) {
        return (
            <div className="flex h-full items-center justify-center p-6 text-sm text-muted-foreground">
                No proposals to compare.
            </div>
        );
    }

    return (
        <div className="flex h-full min-h-0 flex-col">
            <div className="flex shrink-0 items-center justify-center gap-2 border-b border-border px-3 py-1.5">
                <Button type="button" size="icon" variant="ghost" className="h-7 w-7" onClick={() => step(-1)} title="Previous proposal">
                    <ChevronUp className="h-4 w-4" />
                </Button>
                <span className="min-w-16 text-center text-xs text-muted-foreground">
                    {safeIndex + 1} / {count}
                </span>
                <Button type="button" size="icon" variant="ghost" className="h-7 w-7" onClick={() => step(1)} title="Next proposal">
                    <ChevronDown className="h-4 w-4" />
                </Button>
            </div>
            <div className="report-view min-h-0 flex-1 overflow-auto">
                <table className="visual-compare-table w-full table-fixed border-collapse">
                    <thead className="sticky top-0 z-10 bg-background">
                        <tr>
                            <th className="w-1/2 border-b border-r border-border px-4 py-1.5 text-left text-[10px] font-semibold uppercase tracking-wide text-muted-foreground">
                                Manuscript
                            </th>
                            <th className="w-1/2 border-b border-border px-4 py-1.5 text-left text-[10px] font-semibold uppercase tracking-wide text-muted-foreground">
                                Proposed
                            </th>
                        </tr>
                    </thead>
                    <tbody className="align-top">
                        {hunk.chapter ? (
                            <tr>
                                <td colSpan={2} className="border-b border-border px-4 py-3">
                                    <div className="text-[10px] font-semibold uppercase tracking-wide text-muted-foreground">Chapter</div>
                                    <div className="text-base font-semibold text-foreground">{hunk.chapter}</div>
                                </td>
                            </tr>
                        ) : null}
                        {hunk.before ? (
                            <tr>
                                <td className="border-b border-r border-border px-4 py-3">
                                    <div className="prose prose-sm dark:prose-invert max-w-none">
                                        <BlockHtml markdown={hunk.before} />
                                    </div>
                                </td>
                                <td className="border-b border-border px-4 py-3">
                                    <div className="prose prose-sm dark:prose-invert max-w-none">
                                        <BlockHtml markdown={hunk.before} />
                                    </div>
                                </td>
                            </tr>
                        ) : null}
                        <tr>
                            <td className="border-b border-r border-border px-4 py-3">
                                <div className="prose prose-sm dark:prose-invert max-w-none">
                                    <CellStack blocks={hunk.deletes} tone="del" />
                                </div>
                            </td>
                            <td className="border-b border-border px-4 py-3">
                                <div className="prose prose-sm dark:prose-invert max-w-none">
                                    <CellStack blocks={hunk.inserts} tone="ins" />
                                </div>
                            </td>
                        </tr>
                        {hunk.after ? (
                            <tr>
                                <td className="border-r border-border px-4 py-3">
                                    <div className="prose prose-sm dark:prose-invert max-w-none">
                                        <BlockHtml markdown={hunk.after} />
                                    </div>
                                </td>
                                <td className="px-4 py-3">
                                    <div className="prose prose-sm dark:prose-invert max-w-none">
                                        <BlockHtml markdown={hunk.after} />
                                    </div>
                                </td>
                            </tr>
                        ) : null}
                    </tbody>
                </table>
            </div>
        </div>
    );
}
