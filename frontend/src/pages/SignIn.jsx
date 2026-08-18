import { useState } from 'react';
import { api } from '@/api/client';
import { setCurrentUser } from '@/auth/session';
import { loadThemeFromServer } from '@/lib/theme';
import { Button } from '@/components/ui/button';
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
        <p className="mb-4 text-sm text-muted-foreground">This desktop app uses a local account on this machine. A cloud plan is only needed for AI analyses.</p>
        <Button disabled={busy} onClick={() => void onContinue()}>
          {busy ? 'Opening…' : 'Continue'}
        </Button>
        <p className="mt-4 text-xs text-muted-foreground">Optional: paste a device token in Settings after you have an account, to run AI analyses.</p>
        {error && <p className="mt-3 text-sm text-destructive">{error}</p>}
      </div>
    </div>);
}
