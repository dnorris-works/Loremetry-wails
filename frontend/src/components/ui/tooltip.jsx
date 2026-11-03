import { useEffect, useLayoutEffect, useRef, useState } from 'react';
import { createPortal } from 'react-dom';

const DEFAULT_DELAY_MS = 2000;

/** Hover tip. Pass `title` (string) or `content` (node). Optional `delayMs` (default 2000). */
export function DelayedTooltip({ title, content, children, className, delayMs = DEFAULT_DELAY_MS, side = 'right' }) {
    const [open, setOpen] = useState(false);
    const [coords, setCoords] = useState(null);
    const timer = useRef(0);
    const triggerRef = useRef(null);
    const tipRef = useRef(null);
    const body = content ?? title;

    useEffect(() => () => window.clearTimeout(timer.current), []);

    useLayoutEffect(() => {
        if (!open || !triggerRef.current) {
            setCoords(null);
            return;
        }
        function place() {
            const el = triggerRef.current;
            const tip = tipRef.current;
            if (!el) return;
            const r = el.getBoundingClientRect();
            const tipH = tip?.offsetHeight || 0;
            const tipW = tip?.offsetWidth || 0;
            const gap = 6;
            let top;
            let left;
            if (side === 'top') {
                top = r.top - tipH - gap;
                left = r.left;
                if (top < 8) {
                    top = r.bottom + gap;
                }
            } else {
                top = r.top;
                left = r.right + gap;
                if (left + tipW > window.innerWidth - 8) {
                    left = Math.max(8, r.left - tipW - gap);
                }
            }
            if (top + tipH > window.innerHeight - 8) {
                top = Math.max(8, window.innerHeight - tipH - 8);
            }
            if (left < 8) left = 8;
            setCoords({ top, left });
        }
        place();
        // Re-measure after tip mounts so height/width are known
        const id = requestAnimationFrame(place);
        window.addEventListener('scroll', place, true);
        window.addEventListener('resize', place);
        return () => {
            cancelAnimationFrame(id);
            window.removeEventListener('scroll', place, true);
            window.removeEventListener('resize', place);
        };
    }, [open, side, body]);

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
    const tip = open && createPortal(
        <span
            ref={tipRef}
            className={`pointer-events-none fixed z-[9999] rounded border border-border bg-accent px-2 py-1.5 text-[10px] text-muted-foreground shadow-md ${content ? 'min-w-[10rem] max-w-[16rem] whitespace-normal' : 'whitespace-nowrap'} ${coords ? 'opacity-100' : 'opacity-0'}`}
            style={coords ? { top: coords.top, left: coords.left } : { top: 0, left: 0 }}
        >
            {body}
        </span>,
        document.body,
    );
    return (
        <span
            ref={triggerRef}
            className={className || 'relative inline-flex'}
            onMouseEnter={showLater}
            onMouseLeave={hide}
            onFocus={showLater}
            onBlur={hide}
        >
            {children}
            {tip}
        </span>
    );
}
