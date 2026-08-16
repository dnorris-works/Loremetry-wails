import { Folder, X } from 'lucide-react';
import { api } from '@/api/client';
import { Button } from '@/components/ui/button';
import { DelayedTooltip } from '@/components/ui/tooltip';

export function HeaderFolderLink({ projectPath, kind, path, missing, onChanged }) {
    async function pick() {
        try {
            const next = await api.pickImportFolder();
            if (!next)
                return;
            await api.setHeaderOverride(projectPath, kind, next);
            onChanged?.();
        }
        catch (err) {
            alert(err instanceof Error ? err.message : 'Could not link folder');
        }
    }
    async function clear() {
        try {
            await api.clearHeaderOverride(projectPath, kind);
            onChanged?.();
        }
        catch (err) {
            alert(err instanceof Error ? err.message : 'Could not unlink folder');
        }
    }
    const label = path ? path.split(/[/\\]/).filter(Boolean).at(-1) : 'Link folder';
    const hint = missing ? 'Header folder is missing' : (path || 'Point this header at a folder of markdown files');
    return (<div className="flex items-center gap-0.5 pl-1">
      <DelayedTooltip title={hint} className="relative min-w-0 flex-1">
        <button type="button" className="w-full truncate text-left text-[10px] text-muted-foreground hover:text-foreground" onClick={() => void pick()}>
          {missing ? `${label} (missing)` : label}
        </button>
      </DelayedTooltip>
      <Button type="button" size="icon" variant="ghost" className="h-5 w-5 shrink-0" title={path ? 'Change folder' : 'Link folder'} onClick={() => void pick()}>
        <Folder className="h-3 w-3"/>
      </Button>
      {path && (<Button type="button" size="icon" variant="ghost" className="h-5 w-5 shrink-0" title="Use the template folder again" onClick={() => void clear()}>
          <X className="h-3 w-3"/>
        </Button>)}
    </div>);
}

export function folderPathMap(links) {
    const m = {};
    for (const l of links || [])
        m[l.kind] = l.path;
    return m;
}
