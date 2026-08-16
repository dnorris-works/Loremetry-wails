import { useEffect, useState } from 'react';
import { api } from '@/api/client';
import { Button } from '@/components/ui/button';
import { Input, Label } from '@/components/ui/input';

const LAST_PEN = 'loremetry_last_pen';

export function lastPen() {
    return localStorage.getItem(LAST_PEN) || '';
}

export function rememberPen(name) {
    if (name)
        localStorage.setItem(LAST_PEN, name);
}

export function PenPick({ value, onChange, disabled }) {
    const [pens, setPens] = useState([]);
    const [adding, setAdding] = useState(false);
    const [newName, setNewName] = useState('');
    useEffect(() => {
        void api.listPens().then((list) => setPens(list || [])).catch(() => setPens([]));
    }, []);
    async function add() {
        const name = newName.trim();
        if (!name)
            return;
        try {
            const pen = await api.createPen(name);
            setPens((list) => {
                if (list.some((p) => p.name === pen.name))
                    return list;
                return [...list, pen].sort((a, b) => a.name.localeCompare(b.name));
            });
            onChange(pen.name);
            rememberPen(pen.name);
            setNewName('');
            setAdding(false);
        }
        catch (err) {
            alert(err instanceof Error ? err.message : 'Could not create pen name');
        }
    }
    return (<div>
      <Label>Pen name</Label>
      <div className="flex gap-2">
        <select className="h-9 min-w-0 flex-1 rounded-md border border-input bg-background px-2 text-sm" value={value} disabled={disabled} onChange={(e) => {
            onChange(e.target.value);
            rememberPen(e.target.value);
        }}>
          <option value="">Choose…</option>
          {pens.map((p) => (<option key={p.name} value={p.name}>{p.name}</option>))}
        </select>
        {!disabled && (<Button type="button" size="sm" variant="outline" onClick={() => setAdding((v) => !v)}>
            New
          </Button>)}
      </div>
      {adding && !disabled && (<div className="mt-2 flex gap-2">
          <Input value={newName} onChange={(e) => setNewName(e.target.value)} placeholder="Pen name"/>
          <Button type="button" size="sm" onClick={() => void add()}>Add</Button>
        </div>)}
    </div>);
}
