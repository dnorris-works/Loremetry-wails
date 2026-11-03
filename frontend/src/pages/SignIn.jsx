import { useEffect, useState } from 'react';
import { api } from '@/api/client';
import { setCurrentUser } from '@/auth/session';
import { loadThemeFromServer } from '@/lib/theme';
import { Button } from '@/components/ui/button';

function LocalAILine() {
    const [status, setStatus] = useState(null);
    useEffect(() => {
        let cancelled = false;
        const tick = () => {
            void api.localAIStatus().then((s) => {
                if (!cancelled) setStatus(s);
            }).catch(() => {
                if (!cancelled) setStatus(null);
            });
        };
        tick();
        const id = setInterval(tick, 2000);
        return () => {
            cancelled = true;
            clearInterval(id);
        };
    }, []);
    if (!status) return null;
    let label = 'Local AI: unavailable';
    if (status.ready) label = `Local AI: ready (${status.model || 'model'})`;
    else if (status.starting) label = 'Local AI: starting…';
    else if (status.error) label = `Local AI: ${status.error}`;
    return <p className="mt-4 text-xs text-muted-foreground">{label}</p>;
}

export function SignIn({ onReady }) {
    const [error, setError] = useState('');
    const [busy, setBusy] = useState(false);

    async function onContinue() {
        setError('');
        setBusy(true);
        try {
            const session = await api.getSession();
            if (!session.authenticated) {
                setError(session.reason || 'Could not start a local session.');
                return;
            }
            setCurrentUser({
                id: session.id,
                email: session.email || '',
                firstName: session.firstName || '',
                lastName: session.lastName || '',
                isAdmin: session.isAdmin || false,
                breakGlass: session.breakGlass || false,
            });
            await loadThemeFromServer();
            onReady();
        }
        catch (err) {
            setError(err instanceof Error ? err.message : 'Sign in failed.');
        }
        finally {
            setBusy(false);
        }
    }

    return (<div className="flex h-full items-center justify-center p-8">
      <div className="w-full max-w-sm rounded-lg border border-border bg-card p-6">
        <h1 className="mb-2 text-xl font-semibold">Loremetry</h1>
        <p className="mb-4 text-sm text-muted-foreground">Writing and AI analyses run on this machine with the bundled local model. Your novel files stay here.</p>
        <Button disabled={busy} onClick={() => void onContinue()}>
          {busy ? 'Opening…' : 'Continue'}
        </Button>
        <LocalAILine />
        {error && <p className="mt-3 text-sm text-destructive">{error}</p>}
      </div>
    </div>);
}
