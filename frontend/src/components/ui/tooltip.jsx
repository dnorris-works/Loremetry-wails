import { useEffect, useRef, useState } from 'react';

const DEFAULT_DELAY_MS = 2000;

/** Hover tip. Pass `title` (string) or `content` (node). Optional `delayMs` (default 2000). */
export function DelayedTooltip({ title, content, children, className, delayMs = DEFAULT_DELAY_MS }) {
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
    return (<span className={className || 'relative inline-flex'} onMouseEnter={showLater} onMouseLeave={hide} onFocus={showLater} onBlur={hide}>
      {children}
      {open && (<span className={`pointer-events-none absolute left-full top-0 z-50 ml-2 rounded bg-foreground px-2 py-1.5 text-[10px] text-background shadow ${content ? 'min-w-[10rem] whitespace-normal' : 'whitespace-nowrap'}`}>
          {body}
        </span>)}
    </span>);
}

