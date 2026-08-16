import * as React from 'react';
import { cn } from '@/lib/utils';
export function Input({ className, ...props }) {
    return (<input className={cn('flex h-9 w-full rounded-md border border-input bg-background px-3 py-1 text-sm shadow-sm outline-none focus-visible:ring-2 focus-visible:ring-ring', className)} {...props}/>);
}
export function Textarea({ className, ...props }) {
    return (<textarea className={cn('flex min-h-24 w-full rounded-md border border-input bg-background px-3 py-2 text-sm shadow-sm outline-none focus-visible:ring-2 focus-visible:ring-ring', className)} {...props}/>);
}
export function Label({ className, ...props }) {
    return <label className={cn('text-sm font-medium', className)} {...props}/>;
}
