import { StrictMode } from 'react';
import { createRoot } from 'react-dom/client';
import { App } from '@/App';
import { api } from '@/api/client';
import { setCurrentUser } from '@/auth/session';
import { applyTheme, loadThemeFromServer, readTheme } from '@/lib/theme';
import './index.css';
applyTheme(readTheme());
async function bootstrap() {
    try {
        const session = await api.getSession();
        if (session.authenticated) {
            setCurrentUser({
                id: session.id,
                email: session.email || '',
                firstName: session.firstName || '',
                lastName: session.lastName || '',
                isAdmin: session.isAdmin || false,
                breakGlass: session.breakGlass || false,
            });
            await loadThemeFromServer();
        }
    }
    catch (err) {
        console.warn('[auth] session restore failed:', err);
    }
    const el = document.getElementById('root') ?? document.getElementById('app');
    if (!el)
        throw new Error('root element missing');
    createRoot(el).render(<StrictMode>
      <App />
    </StrictMode>);
}
void bootstrap();
