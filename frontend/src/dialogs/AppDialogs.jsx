import { useEffect, useState } from 'react';
import { api } from '@/api/client';
import { useAppState } from '@/lib/app-state';
import { Dialog, DialogContent, DialogHeader, DialogTitle } from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { persistTheme, readTheme } from '@/lib/theme';
import { isAuthenticated } from '@/auth/session';
import { useConfirm } from '@/components/confirm-dialog';
export function SettingsDialog({ open, onOpenChange }) {
    const [theme, setTheme] = useState(readTheme());
    const { restoreOpen, setRestoreOpen, bumpRefresh } = useAppState();
    const confirm = useConfirm();
    const [writingRoot, setWritingRoot] = useState('');
    const [templateFolders, setTemplateFolders] = useState({ series: [], books: [] });
    const [hiddenNames, setHiddenNames] = useState([]);
    const [showHidden, setShowHidden] = useState(false);
    const [draftSection, setDraftSection] = useState('Act');
    useEffect(() => {
        if (!open)
            return;
        void api.getSetting('writing_root').then((s) => setWritingRoot(s.value || '')).catch(() => setWritingRoot(''));
        void api.listTemplateFolders().then((d) => setTemplateFolders({ series: d.series || [], books: d.books || [] })).catch(() => { });
        void api.getFolderVisibility().then((v) => {
            setHiddenNames(v.hidden_names || v.hiddenNames || []);
            setShowHidden(!!(v.show_hidden ?? v.showHidden));
        }).catch(() => { });
        void api.getDraftSection().then((v) => setDraftSection(v === 'Part' ? 'Part' : 'Act')).catch(() => setDraftSection('Act'));
    }, [open]);
    async function pickRoot() {
        try {
            const path = await api.pickImportFolder();
            if (!path)
                return;
            await api.putSetting('writing_root', path);
            setWritingRoot(path);
            bumpRefresh();
        }
        catch (err) {
            alert(err instanceof Error ? err.message : 'Could not set writing folder');
        }
    }
    async function toggleName(name, visible) {
        const next = visible ? hiddenNames.filter((n) => n !== name) : [...hiddenNames.filter((n) => n !== name), name];
        setHiddenNames(next);
        await api.setHiddenFolderNames(next);
        bumpRefresh();
    }
    async function toggleShowHidden(on) {
        setShowHidden(on);
        await api.setShowHiddenFolders(on);
        bumpRefresh();
    }
    async function chooseDraftSection(next) {
        if (next === draftSection)
            return;
        const other = next === 'Act' ? 'Part' : 'Act';
        const rename = await confirm(`Rename existing ${other} folders on disk to ${next}?`);
        const out = await api.setDraftSection(next, rename);
        setDraftSection(out.section || next);
        bumpRefresh();
    }
    return (<Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-h-[80vh] overflow-y-auto" movable>
        <DialogHeader>
          <DialogTitle>Settings</DialogTitle>
        </DialogHeader>
        <div className="flex gap-2">
          <Button variant={theme === 'light' ? 'default' : 'outline'} onClick={() => {
            setTheme('light');
            void persistTheme('light', isAuthenticated());
        }}>
            Light
          </Button>
          <Button variant={theme === 'dark' ? 'default' : 'outline'} onClick={() => {
            setTheme('dark');
            void persistTheme('dark', isAuthenticated());
        }}>
            Dark
          </Button>
        </div>
        <label className="mt-4 flex items-center gap-2 text-sm">
          <input type="checkbox" checked={restoreOpen} onChange={(e) => void setRestoreOpen(e.target.checked)}/>
          Remember what was open
        </label>
        <div className="mt-4 space-y-1">
          <div className="text-sm">Writing folder</div>
          <p className="truncate text-xs text-muted-foreground" title={writingRoot}>{writingRoot || 'Not set — you will be asked when you create a series or story'}</p>
          <Button size="sm" variant="outline" onClick={() => void pickRoot()}>Choose folder</Button>
        </div>
        <div className="mt-4 space-y-1">
          <div className="text-sm">Use</div>
          <div className="flex gap-2">
            <Button size="sm" variant={draftSection === 'Act' ? 'default' : 'outline'} onClick={() => void chooseDraftSection('Act')}>Act</Button>
            <Button size="sm" variant={draftSection === 'Part' ? 'default' : 'outline'} onClick={() => void chooseDraftSection('Part')}>Part</Button>
          </div>
          <p className="text-xs text-muted-foreground">New stories use {draftSection}-01 under Current Draft. Switching can rename existing Act/Part folders.</p>
        </div>
        <label className="mt-4 flex items-center gap-2 text-sm">
          <input type="checkbox" checked={showHidden} onChange={(e) => void toggleShowHidden(e.target.checked)}/>
          Show hidden folders in the sidebar
        </label>
        <p className="mt-1 text-xs text-muted-foreground">Folders still exist on disk. Unchecked names stay hidden everywhere unless you show them on a story.</p>
        <FolderChecks title="Series folders" names={templateFolders.series} hiddenNames={hiddenNames} onToggle={toggleName}/>
        <FolderChecks title="Story folders" names={templateFolders.books} hiddenNames={hiddenNames} onToggle={toggleName}/>
      </DialogContent>
    </Dialog>);
}
function FolderChecks({ title, names, hiddenNames, onToggle }) {
    return (<div className="mt-4">
      <div className="text-sm">{title}</div>
      <div className="mt-1 max-h-40 space-y-1 overflow-auto text-sm">
        {names.map((name) => (<label key={name} className="flex items-center gap-2">
            <input type="checkbox" checked={!hiddenNames.includes(name)} onChange={(e) => void onToggle(name, e.target.checked)}/>
            <span className="truncate">{name}</span>
          </label>))}
      </div>
    </div>);
}
export function AdminDialog({ open, onOpenChange }) {
    const [tables, setTables] = useState([]);
    const [table, setTable] = useState('');
    const [columns, setColumns] = useState([]);
    const [sql, setSql] = useState('');
    const [result, setResult] = useState(null);
    const [error, setError] = useState('');
    useEffect(() => {
        if (!open)
            return;
        void api.adminListTables().then((d) => setTables(d.tables || []));
    }, [open]);
    useEffect(() => {
        if (!table) {
            setColumns([]);
            return;
        }
        setSql(`SELECT * FROM "${table.replaceAll('"', '""')}" LIMIT 1000`);
        void api.adminTableSchema(table).then((d) => setColumns(d.columns || []));
    }, [table]);
    async function run() {
        setError('');
        try {
            setResult(await api.adminExecSQL(sql));
        }
        catch (e) {
            setResult(null);
            setError(e instanceof Error ? e.message : String(e));
        }
    }
    return (<Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-3xl">
        <DialogTitle>Admin</DialogTitle>
        <div className="mt-3 space-y-3">
          <select className="h-9 w-full rounded-md border border-input bg-background px-2 text-sm" value={table} onChange={(e) => setTable(e.target.value)}>
            <option value="">Select table…</option>
            {tables.map((t) => (<option key={t} value={t}>
                {t}
              </option>))}
          </select>
          {columns.length > 0 && (<table className="w-full text-xs">
              <thead>
                <tr>
                  <th className="text-left">Column</th>
                  <th className="text-left">Type</th>
                </tr>
              </thead>
              <tbody>
                {columns.map((c) => (<tr key={c.name}>
                    <td>{c.name}</td>
                    <td>{c.type}</td>
                  </tr>))}
              </tbody>
            </table>)}
          <textarea className="h-28 w-full rounded-md border border-input bg-background p-2 font-mono text-xs" value={sql} onChange={(e) => setSql(e.target.value)}/>
          <Button size="sm" onClick={() => void run()}>
            Run SQL
          </Button>
          {error && <p className="text-sm text-destructive">{error}</p>}
          {result?.message && <p className="text-sm text-muted-foreground">{result.message}</p>}
          {result && (result.columns?.length ?? 0) > 0 && (<div className="max-h-64 overflow-auto rounded border border-border">
              <table className="w-full text-xs">
                <thead>
                  <tr>
                    {(result.columns || []).map((c) => (<th key={c} className="sticky top-0 bg-muted px-2 py-1 text-left">
                        {c}
                      </th>))}
                  </tr>
                </thead>
                <tbody>
                  {(result.rows || []).map((row, i) => (<tr key={i} className="border-t border-border">
                      {row.map((cell, j) => (<td key={j} className="max-w-48 truncate px-2 py-1 align-top">
                          {cell}
                        </td>))}
                    </tr>))}
                </tbody>
              </table>
              <p className="border-t border-border px-2 py-1 text-muted-foreground">
                {(result.rows || []).length} row(s)
              </p>
            </div>)}
        </div>
      </DialogContent>
    </Dialog>);
}
