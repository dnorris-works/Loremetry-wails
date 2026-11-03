import { api } from '@/api/client';
const STORAGE_KEY = 'loremetry_theme';
const SETTING_KEY = 'theme';
function getStoredTheme() {
    try {
        const stored = localStorage.getItem(STORAGE_KEY);
        if (stored === 'dark' || stored === 'light')
            return stored;
    }
    catch {
        /* ignore */
    }
    if (window.matchMedia('(prefers-color-scheme: dark)').matches)
        return 'dark';
    return 'light';
}
export function applyTheme(t) {
    document.documentElement.classList.toggle('dark', t === 'dark');
    document.documentElement.setAttribute('data-theme', t);
}
export function readTheme() {
    return getStoredTheme();
}
export async function persistTheme(t, authenticated) {
    applyTheme(t);
    try {
        localStorage.setItem(STORAGE_KEY, t);
    }
    catch {
        /* ignore */
    }
    if (authenticated) {
        try {
            await api.putSetting(SETTING_KEY, t);
        }
        catch {
            /* ignore */
        }
    }
}
export async function loadThemeFromServer() {
    try {
        const data = await api.getSetting(SETTING_KEY);
        if (data.value === 'dark' || data.value === 'light') {
            applyTheme(data.value);
            localStorage.setItem(STORAGE_KEY, data.value);
            return data.value;
        }
    }
    catch {
        /* ignore */
    }
    return null;
}
