import { Dialog, DialogContent, DialogHeader, DialogTitle } from '@/components/ui/dialog';

/** Dialog showing a chapter wrapped at fiction line length with sticky lines highlighted. */
export function StickyChapterDialog({ open, onOpenChange, context, error }) {
    const chapter = context?.chapter || context?.Chapter || '';
    const sticky = context?.sticky_count ?? context?.stickyCount ?? 0;
    const lines = context?.lines || context?.Lines || [];
    const width = context?.line_width ?? context?.lineWidth ?? 60;

    return (
        <Dialog open={open} onOpenChange={onOpenChange}>
            <DialogContent className="flex max-h-[85vh] max-w-3xl flex-col overflow-hidden">
                <DialogHeader>
                    <DialogTitle>
                        {chapter ? `${chapter} — ${sticky} sticky` : 'Sticky sentences'}
                    </DialogTitle>
                    <p className="text-xs text-muted-foreground">
                        Line length {width} characters (~5.5×8 fiction). Highlighted lines contain sticky sentences.
                    </p>
                </DialogHeader>
                <div className="min-h-0 flex-1 overflow-auto">
                    {error && <p className="text-sm text-destructive">{error}</p>}
                    {!error && lines.length === 0 && (
                        <p className="text-sm text-muted-foreground">No lines to show.</p>
                    )}
                    {!error && lines.length > 0 && (
                        <pre className="font-mono text-[12px] leading-5">
                            {lines.map((ln) => {
                                const n = ln.line ?? ln.Line ?? 0;
                                const text = ln.text ?? ln.Text ?? '';
                                const hi = !!(ln.highlight ?? ln.Highlight);
                                return (
                                    <div
                                        key={n}
                                        className={hi ? 'bg-amber-200/70 dark:bg-amber-500/25' : undefined}
                                    >
                                        <span className="inline-block w-10 select-none text-right text-muted-foreground">
                                            {n}
                                        </span>
                                        <span className="px-2">|</span>
                                        <span>{text}</span>
                                    </div>
                                );
                            })}
                        </pre>
                    )}
                </div>
            </DialogContent>
        </Dialog>
    );
}
