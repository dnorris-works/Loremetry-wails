import { useEffect, useRef, useState } from 'react';

const DELAY_MS = 2000;

export function DelayedTooltip({ title, children, className }) {
    const [open, setOpen] = useState(false);
    const timer = useRef(0);
    useEffect(() => () => window.clearTimeout(timer.current), []);
    if (!title) {
        return children;
    }
    function showLater() {
        window.clearTimeout(timer.current);
        timer.current = window.setTimeout(() => setOpen(true), DELAY_MS);
    }
    function hide() {
        window.clearTimeout(timer.current);
        setOpen(false);
    }
    return (<span className={className || 'relative inline-flex'} onMouseEnter={showLater} onMouseLeave={hide} onFocus={showLater} onBlur={hide}>
      {children}
      {open && (<span className="pointer-events-none absolute left-1/2 top-full z-50 mt-1 -translate-x-1/2 whitespace-nowrap rounded bg-foreground px-2 py-1 text-[10px] text-background shadow">
          {title}
        </span>)}
    </span>);
}
