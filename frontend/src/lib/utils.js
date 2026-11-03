import { clsx } from 'clsx';
import { twMerge } from 'tailwind-merge';
export function cn(...inputs) {
    return twMerge(clsx(inputs));
}

/** Format a duration in milliseconds as `Xs` or `Mm SSs`. */
export function formatElapsed(ms) {
    const sec = Math.max(0, Math.floor(ms / 1000));
    if (sec < 60)
        return `${sec}s`;
    const m = Math.floor(sec / 60);
    const r = sec % 60;
    return `${m}m ${String(r).padStart(2, '0')}s`;
}
