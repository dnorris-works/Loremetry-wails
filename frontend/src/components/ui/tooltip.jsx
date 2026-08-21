import { useEffect, useRef, useState } from 'react';

const DEFAULT_DELAY_MS = 2000;

const SIDE_CLASS = {
    right: 'left-full top-0 ml-2',
    top: 'bottom-full left-0 mb-1.5',
};

/** Hover tip. Pass `title` (string) or `content` (node). Optional `delayMs` (default 2000). */
export function DelayedTooltip({ title, content, children, className, delayMs = DEFAULT_DELAY_MS, side = 'right' }) {
    const [open, setOpen] = useState(false);
    const timer = useRef(0);
    useEffect(() => () => window.clearTimeout(timer.current), []);
    const body = content ?? title;
    if (!body) {
        return children;
    }
    function showLater() {
        window.clearTimeout(timer.current);
        timer.current = window.setTimeout(() => setOpen(true), delayMs);
    }
    function hide() {
        window.clearTimeout(timer.current);
        setOpen(false);
    }
    const pos = SIDE_CLASS[side] || SIDE_CLASS.right;
    return (<span className={className || 'relative inline-flex'} onMouseEnter={showLater} onMouseLeave={hide} onFocus={showLater} onBlur={hide}>
      {children}
      {open && (<span className={`pointer-events-none absolute z-50 ${pos} rounded border border-border bg-accent px-2 py-1.5 text-[10px] text-muted-foreground shadow-md ${content ? 'min-w-[10rem] whitespace-normal' : 'whitespace-nowrap'}`}>
          {body}
        </span>)}
    </span>);
}
