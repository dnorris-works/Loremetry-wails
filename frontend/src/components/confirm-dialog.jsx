import { createContext, useCallback, useContext, useRef, useState } from 'react';
import { Dialog, DialogContent, DialogTitle } from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
const Ctx = createContext(null);
export function ConfirmProvider({ children }) {
    const [pending, setPending] = useState(null);
    const pendingRef = useRef(null);
    const confirm = useCallback((message) => {
        return new Promise((resolve) => {
            const next = { mode: 'confirm', message, resolve };
            pendingRef.current = next;
            setPending(next);
        });
    }, []);
    const notice = useCallback((message) => {
        return new Promise((resolve) => {
            const next = { mode: 'notice', message, resolve };
            pendingRef.current = next;
            setPending(next);
        });
    }, []);
    function close(ok) {
        const current = pendingRef.current;
        if (!current)
            return;
        pendingRef.current = null;
        if (current.mode === 'confirm')
            current.resolve(ok);
        else
            current.resolve();
        setPending(null);
    }
    return (<Ctx.Provider value={{ confirm, notice }}>
      {children}
      <Dialog open={!!pending} onOpenChange={(open) => { if (!open)
        close(false); }}>
        <DialogContent className="max-w-md">
          <DialogTitle>{pending?.mode === 'notice' ? 'Drop analysis' : 'Confirm'}</DialogTitle>
          <p className="mt-3 whitespace-pre-wrap text-sm">{pending?.message}</p>
          <div className="mt-6 flex justify-end gap-2">
            {pending?.mode === 'confirm' && (<Button variant="outline" onClick={() => close(false)}>
                Cancel
              </Button>)}
            <Button variant={pending?.mode === 'confirm' ? 'destructive' : 'default'} onClick={() => close(true)}>
              OK
            </Button>
          </div>
        </DialogContent>
      </Dialog>
    </Ctx.Provider>);
}
export function useConfirm() {
    const ctx = useContext(Ctx);
    if (!ctx)
        throw new Error('useConfirm outside provider');
    return ctx.confirm;
}
export function useNotice() {
    const ctx = useContext(Ctx);
    if (!ctx)
        throw new Error('useNotice outside provider');
    return ctx.notice;
}
