import { useEffect, useState } from 'react';
import { api } from '@/api/client';
import { setCurrentUser } from '@/auth/session';
import { loadThemeFromServer } from '@/lib/theme';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';

export function SignIn({ onReady }) {
    const [error, setError] = useState('');
    const [busy, setBusy] = useState(false);
    const [aiBusy, setAiBusy] = useState(false);
    const [email, setEmail] = useState('');
    const [cloudAccount, setCloudAccount] = useState(null);
    const [hasCloud, setHasCloud] = useState(false);

    useEffect(() => {
        void api.hasCloudAccount().then(setHasCloud).catch(() => setHasCloud(false));
        if (hasCloud) {
            void api.getCloudAccount().then(setCloudAccount).catch(() => setCloudAccount(null));
        } else {
            setCloudAccount(null);
        }
    }, [hasCloud]);

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
            if (!email && session.email) {
                setEmail(session.email);
            }
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

    async function onConnectAI() {
        setError('');
        setAiBusy(true);
        try {
            const session = await api.getSession();
            const useEmail = email.trim() || session.email || '';
            const acc = await api.connectCloudAccount(useEmail);
            setCloudAccount(acc);
            setHasCloud(true);
        }
        catch (err) {
            setError(err instanceof Error ? err.message : 'Could not connect for AI.');
        }
        finally {
            setAiBusy(false);
        }
    }

    return (<div className="flex h-full items-center justify-center p-8">
      <div className="w-full max-w-sm rounded-lg border border-border bg-card p-6">
        <h1 className="mb-2 text-xl font-semibold">Loremetry</h1>
        <p className="mb-4 text-sm text-muted-foreground">Local writing and non-AI analyses work on this machine. AI analyses use your plan on our server — your novel files stay here.</p>
        <Button disabled={busy} onClick={() => void onContinue()}>
          {busy ? 'Opening…' : 'Continue'}
        </Button>
        <div className="mt-6 space-y-2 border-t border-border pt-4">
          <div className="text-sm font-medium">AI analyses</div>
          <p className="text-xs text-muted-foreground">Connect once. Your device token is saved in the app database on this Mac.</p>
          <Input type="email" value={email} onChange={(e) => setEmail(e.target.value)} placeholder="Email for your plan"/>
          <div className="flex gap-2">
            <Button size="sm" variant="outline" disabled={aiBusy} onClick={() => void onConnectAI()}>
              {aiBusy ? 'Connecting…' : hasCloud ? 'Reconnect' : 'Connect for AI'}
            </Button>
            <Button size="sm" variant="outline" disabled={!hasCloud} onClick={() => void api.openBillingCheckout().catch((err) => alert(err instanceof Error ? err.message : 'Checkout failed'))}>
              Choose a plan
            </Button>
          </div>
          {cloudAccount && <p className="text-xs text-muted-foreground">{cloudAccount.plan || 'No plan'} — {cloudAccount.remaining || `${cloudAccount.credits ?? 0} credits`}</p>}
        </div>
        {error && <p className="mt-3 text-sm text-destructive">{error}</p>}
      </div>
    </div>);
}
