import { useEffect, useState } from 'react';
import { api } from '@/api/client';
import { Dialog, DialogContent, DialogTitle } from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { Input, Label } from '@/components/ui/input';
import { useConfirm } from '@/components/confirm-dialog';
import { PenPick, lastPen, rememberPen } from '@/lib/PenPick';
export function SeriesDialog({ open, onOpenChange, projectPath, projectName, defaultPen, defaultPenPath, onSaved, }) {
    const confirm = useConfirm();
    const [name, setName] = useState('');
    const [pen, setPen] = useState('');
    useEffect(() => {
        if (!open)
            return;
        setName(projectName || '');
        setPen(defaultPen || lastPen());
    }, [open, projectPath, projectName, defaultPen]);
    async function save() {
        if (!name.trim())
            return;
        if (projectPath) {
            await api.renameWritingProject(projectPath, name.trim());
        }
        else {
            if (!pen)
                return;
            rememberPen(pen);
            const penPath = pen === defaultPen ? defaultPenPath : '';
            await api.createWritingSeries({ name: name.trim(), pen_name: pen, pen_path: penPath });
        }
        onSaved();
        onOpenChange(false);
    }
    async function remove() {
        if (!projectPath || !(await confirm(`Delete series folder "${name}"?`)))
            return;
        await api.deleteWritingProject(projectPath);
        onSaved();
        onOpenChange(false);
    }
    return (<Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogTitle>{projectPath ? 'Rename series' : 'New series'}</DialogTitle>
        <div className="mt-4 space-y-3">
          {!projectPath && <PenPick value={pen} onChange={setPen}/>}
          <div>
            <Label>Name</Label>
            <Input value={name} onChange={(e) => setName(e.target.value)}/>
          </div>
          <div className="flex gap-2">
            <Button onClick={() => void save()}>Save</Button>
            {projectPath && (<Button variant="destructive" onClick={() => void remove()}>
                Delete
              </Button>)}
          </div>
        </div>
      </DialogContent>
    </Dialog>);
}
export function StoryDialog({ open, onOpenChange, projectPath, projectName, defaultSeriesPath, defaultPen, defaultPenPath, onSaved, }) {
    const confirm = useConfirm();
    const [name, setName] = useState('');
    const [seriesPath, setSeriesPath] = useState('');
    const [tree, setTree] = useState({ pens: [] });
    const [pen, setPen] = useState('');
    useEffect(() => {
        if (!open)
            return;
        void api.listWritingTree().then(setTree);
        setName(projectName || '');
        setSeriesPath(defaultSeriesPath || '');
        setPen(defaultPen || lastPen());
    }, [open, projectPath, projectName, defaultSeriesPath, defaultPen]);
    const seriesOptions = (tree.pens || []).flatMap((p) => (p.series || []).map((s) => ({ path: s.path, name: `${p.name} / ${s.name}`, pen: p.name })));
    async function save() {
        if (!name.trim())
            return;
        if (projectPath) {
            const next = await api.renameWritingProject(projectPath, name.trim());
            onSaved({ path: next.path, seriesPath });
            onOpenChange(false);
            return;
        }
        const chosen = seriesOptions.find((s) => s.path === seriesPath);
        const penName = chosen?.pen || pen;
        if (!seriesPath && !penName)
            return;
        rememberPen(penName);
        const penPath = !seriesPath && penName === defaultPen ? defaultPenPath : '';
        const created = await api.createWritingBook({ name: name.trim(), pen_name: penName, pen_path: penPath, series_path: seriesPath });
        onSaved({ path: created.path, seriesPath });
        onOpenChange(false);
    }
    async function remove() {
        if (!projectPath || !(await confirm(`Delete book folder "${name}"?`)))
            return;
        await api.deleteWritingProject(projectPath);
        onSaved();
        onOpenChange(false);
    }
    return (<Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogTitle>{projectPath ? 'Rename story' : 'New story'}</DialogTitle>
        <div className="mt-4 space-y-3">
          {!projectPath && <PenPick value={seriesPath ? (seriesOptions.find((s) => s.path === seriesPath)?.pen || pen) : pen} onChange={setPen} disabled={!!seriesPath}/>}
          <div>
            <Label>Name</Label>
            <Input value={name} onChange={(e) => setName(e.target.value)}/>
          </div>
          {!projectPath && (<div>
            <Label>Series</Label>
            <select className="h-9 w-full rounded-md border border-input bg-background px-2 text-sm" value={seriesPath} onChange={(e) => setSeriesPath(e.target.value)}>
              <option value="">None</option>
              {seriesOptions.filter((s) => !pen || s.pen === pen).map((s) => (<option key={s.path} value={s.path}>
                  {s.name}
                </option>))}
            </select>
          </div>)}
          <div className="flex gap-2">
            <Button onClick={() => void save()}>Save</Button>
            {projectPath && (<Button variant="destructive" onClick={() => void remove()}>
                Delete
              </Button>)}
          </div>
        </div>
      </DialogContent>
    </Dialog>);
}
