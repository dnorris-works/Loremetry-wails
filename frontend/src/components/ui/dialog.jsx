import * as React from 'react';
import { useEffect, useRef, useState } from 'react';
import * as DialogPrimitive from '@radix-ui/react-dialog';
import { X } from 'lucide-react';
import { cn } from '@/lib/utils';

const DragCtx = React.createContext(null);

export const Dialog = DialogPrimitive.Root;
export const DialogTrigger = DialogPrimitive.Trigger;
export const DialogClose = DialogPrimitive.Close;

export function DialogContent({ className, children, movable, ...props }) {
    const [pos, setPos] = useState({ x: 0, y: 0 });
    useEffect(() => {
        setPos({ x: 0, y: 0 });
    }, []);
    return (<DialogPrimitive.Portal>
      <DialogPrimitive.Overlay className="fixed inset-0 z-50 bg-black/50"/>
      <DragCtx.Provider value={movable ? { pos, setPos } : null}>
        <DialogPrimitive.Content className={cn('fixed left-1/2 top-1/2 z-50 w-full max-w-lg rounded-lg border border-border bg-background p-6 shadow-lg', className)} style={{ transform: `translate(calc(-50% + ${pos.x}px), calc(-50% + ${pos.y}px))` }} {...props}>
          {children}
          <DialogPrimitive.Close className="absolute right-4 top-4 opacity-70 hover:opacity-100">
            <X className="h-4 w-4"/>
          </DialogPrimitive.Close>
        </DialogPrimitive.Content>
      </DragCtx.Provider>
    </DialogPrimitive.Portal>);
}

export function DialogHeader({ className, children, ...props }) {
    const drag = React.useContext(DragCtx);
    const origin = useRef(null);
    function onPointerDown(e) {
        if (!drag || e.button !== 0)
            return;
        origin.current = { x: e.clientX - drag.pos.x, y: e.clientY - drag.pos.y };
        e.currentTarget.setPointerCapture(e.pointerId);
    }
    function onPointerMove(e) {
        if (!drag || !origin.current)
            return;
        drag.setPos({ x: e.clientX - origin.current.x, y: e.clientY - origin.current.y });
    }
    function onPointerUp() {
        origin.current = null;
    }
    return (<div {...props} className={cn('-mx-6 -mt-6 mb-4 border-b border-border px-6 py-3 pr-12', drag && 'cursor-grab select-none active:cursor-grabbing', className)} onPointerDown={onPointerDown} onPointerMove={onPointerMove} onPointerUp={onPointerUp} onPointerCancel={onPointerUp}>
      {children}
    </div>);
}

export function DialogTitle({ className, ...props }) {
    return <DialogPrimitive.Title className={cn('text-lg font-semibold', className)} {...props}/>;
}
