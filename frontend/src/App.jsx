import { useEffect, useState } from 'react';
import { HashRouter } from 'react-router';
import { isAuthenticated, subscribeAuth } from '@/auth/session';
import { AppStateProvider } from '@/lib/app-state';
import { ConfirmProvider } from '@/components/confirm-dialog';
import { AppLayout } from '@/layout/AppLayout';
import { SignIn } from '@/pages/SignIn';
export function App() {
    const [, setTick] = useState(0);
    useEffect(() => subscribeAuth(() => setTick((n) => n + 1)), []);
    if (!isAuthenticated()) {
        return (<HashRouter>
        <SignIn onReady={() => setTick((n) => n + 1)}/>
      </HashRouter>);
    }
    return (<HashRouter>
      <AppStateProvider>
        <ConfirmProvider>
          <AppLayout />
        </ConfirmProvider>
      </AppStateProvider>
    </HashRouter>);
}
