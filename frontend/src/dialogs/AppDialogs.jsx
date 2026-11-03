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
    const [tab, setTab] = useState('catalog');
    return (<Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-3xl max-h-[90vh] overflow-y-auto">
        <DialogTitle>Admin</DialogTitle>
        <div className="mt-2 flex gap-2">
          <Button size="sm" variant={tab === 'catalog' ? 'default' : 'outline'} onClick={() => setTab('catalog')}>Catalog</Button>
          <Button size="sm" variant={tab === 'sql' ? 'default' : 'outline'} onClick={() => setTab('sql')}>SQL</Button>
        </div>
        {tab === 'catalog' ? <CatalogEditor/> : <AdminSQLPanel/>}
      </DialogContent>
    </Dialog>);
}
function CatalogEditor() {
    const [items, setItems] = useState([]);
    const [selectedId, setSelectedId] = useState('');
    const [form, setForm] = useState(null);
    const [error, setError] = useState('');
    const [saved, setSaved] = useState('');
    useEffect(() => {
        void api.listAnalysisCatalog().then((groups) => {
            const flat = [];
            for (const g of groups || []) {
                for (const it of g.items || []) {
                    flat.push({ ...it, group: g.label });
                }
            }
            setItems(flat);
        });
    }, []);
    useEffect(() => {
        if (!selectedId) {
            setForm(null);
            return;
        }
        setError('');
        setSaved('');
        void api.getAnalysis(selectedId).then((d) => {
            setForm({
                id: d.id,
                label: d.label || '',
                description: d.description || '',
                usage: d.usage || '',
                needs: (d.needs || []).join(', '),
                depends_on: (d.depends_on || []).join(', '),
                uses_ai: !!d.uses_ai,
                uses_merge: !!d.uses_merge,
                ai_profile: d.ai_profile || 'single',
                chapter_instruction: d.chapter_instruction || '',
                chapter_headings: (d.chapter_headings || []).join(', '),
                final_instruction: d.final_instruction || '',
                source_injection: d.source_injection || 'selective',
                local_runner: d.local_runner || '',
                local_config: d.local_config || '{}',
            });
        }).catch((e) => setError(e instanceof Error ? e.message : String(e)));
    }, [selectedId]);
    function patch(key, value) {
        setForm((prev) => (prev ? { ...prev, [key]: value } : prev));
    }
    async function save() {
        if (!form)
            return;
        setError('');
        setSaved('');
        const splitList = (s) => s.split(',').map((x) => x.trim()).filter(Boolean);
        try {
            await api.updateAnalysisCatalog({
                id: form.id,
                label: form.label,
                description: form.description,
                usage: form.usage,
                needs: splitList(form.needs),
                depends_on: splitList(form.depends_on),
                uses_ai: form.uses_ai,
                uses_merge: form.uses_merge,
                ai_profile: form.ai_profile,
                chapter_instruction: form.chapter_instruction,
                chapter_headings: splitList(form.chapter_headings),
                final_instruction: form.final_instruction,
                source_injection: form.source_injection,
                local_runner: form.local_runner,
                local_config: form.local_config,
            });
            setSaved('Saved.');
        }
        catch (e) {
            setError(e instanceof Error ? e.message : String(e));
        }
    }
    return (<div className="mt-3 space-y-3">
      <select className="h-9 w-full rounded-md border border-input bg-background px-2 text-sm" value={selectedId} onChange={(e) => setSelectedId(e.target.value)}>
        <option value="">Select analysis…</option>
        {items.map((it) => (<option key={it.id} value={it.id}>{it.group} — {it.label} ({it.id})</option>))}
      </select>
      {form && (<>
          <div className="grid gap-2 sm:grid-cols-2">
            <label className="text-xs">Label<input className="mt-1 h-8 w-full rounded-md border border-input bg-background px-2 text-sm" value={form.label} onChange={(e) => patch('label', e.target.value)}/></label>
            <label className="text-xs">AI profile<input className="mt-1 h-8 w-full rounded-md border border-input bg-background px-2 text-sm" value={form.ai_profile} onChange={(e) => patch('ai_profile', e.target.value)}/></label>
            <label className="text-xs sm:col-span-2">Needs (comma-separated)<input className="mt-1 h-8 w-full rounded-md border border-input bg-background px-2 text-sm" value={form.needs} onChange={(e) => patch('needs', e.target.value)}/></label>
            <label className="text-xs sm:col-span-2">Depends on<input className="mt-1 h-8 w-full rounded-md border border-input bg-background px-2 text-sm" value={form.depends_on} onChange={(e) => patch('depends_on', e.target.value)}/></label>
            <label className="text-xs">Source injection<select className="mt-1 h-8 w-full rounded-md border border-input bg-background px-2 text-sm" value={form.source_injection} onChange={(e) => patch('source_injection', e.target.value)}>
                <option value="selective">selective</option>
                <option value="all">all</option>
              </select></label>
            <label className="text-xs">Local runner<input className="mt-1 h-8 w-full rounded-md border border-input bg-background px-2 text-sm" value={form.local_runner} onChange={(e) => patch('local_runner', e.target.value)}/></label>
            <label className="flex items-center gap-2 text-xs"><input type="checkbox" checked={form.uses_ai} onChange={(e) => patch('uses_ai', e.target.checked)}/>Uses AI</label>
            <label className="flex items-center gap-2 text-xs"><input type="checkbox" checked={form.uses_merge} onChange={(e) => patch('uses_merge', e.target.checked)}/>Uses merge</label>
          </div>
          <label className="block text-xs">Description<textarea className="mt-1 h-16 w-full rounded-md border border-input bg-background p-2 text-sm" value={form.description} onChange={(e) => patch('description', e.target.value)}/></label>
          <label className="block text-xs">Usage<textarea className="mt-1 h-16 w-full rounded-md border border-input bg-background p-2 text-sm" value={form.usage} onChange={(e) => patch('usage', e.target.value)}/></label>
          <label className="block text-xs">Chapter instruction<textarea className="mt-1 h-20 w-full rounded-md border border-input bg-background p-2 text-xs font-mono" value={form.chapter_instruction} onChange={(e) => patch('chapter_instruction', e.target.value)}/></label>
          <label className="block text-xs">Chapter headings (comma-separated)<input className="mt-1 h-8 w-full rounded-md border border-input bg-background px-2 text-sm" value={form.chapter_headings} onChange={(e) => patch('chapter_headings', e.target.value)}/></label>
          <label className="block text-xs">Final instruction<textarea className="mt-1 h-20 w-full rounded-md border border-input bg-background p-2 text-xs font-mono" value={form.final_instruction} onChange={(e) => patch('final_instruction', e.target.value)}/></label>
          <label className="block text-xs">Local config (JSON)<textarea className="mt-1 h-20 w-full rounded-md border border-input bg-background p-2 text-xs font-mono" value={form.local_config} onChange={(e) => patch('local_config', e.target.value)}/></label>
          <Button size="sm" onClick={() => void save()}>Save catalog row</Button>
        </>)}
      {error && <p className="text-sm text-destructive">{error}</p>}
      {saved && <p className="text-sm text-muted-foreground">{saved}</p>}
    </div>);
}
function AdminSQLPanel() {
    const [tables, setTables] = useState([]);
    const [table, setTable] = useState('');
    const [columns, setColumns] = useState([]);
    const [sql, setSql] = useState('');
    const [result, setResult] = useState(null);
    const [error, setError] = useState('');
    useEffect(() => {
        void api.adminListTables().then((d) => setTables(d.tables || []));
    }, []);
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
    return (<div className="mt-3 space-y-3">
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
    </div>);
}
