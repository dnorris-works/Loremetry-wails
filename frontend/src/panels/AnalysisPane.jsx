import { useEffect, useState } from 'react';
import { api } from '@/api/client';
import { useAppState } from '@/lib/app-state';
import { Button } from '@/components/ui/button';
import { X } from 'lucide-react';

export function AnalysisPane() {
    const { analysisId, setAnalysis } = useAppState();
    const [detail, setDetail] = useState(null);
    useEffect(() => {
        if (!analysisId) {
            setDetail(null);
            return;
        }
        let cancelled = false;
        void api.getAnalysis(analysisId).then((d) => {
            if (!cancelled)
                setDetail(d);
        }).catch(() => {
            if (!cancelled)
                setDetail(null);
        });
        return () => {
            cancelled = true;
        };
    }, [analysisId]);
    if (!detail) {
        return <div className="p-6 text-sm text-muted-foreground">Loading…</div>;
    }
    return (<div className="flex h-full flex-col">
      <div className="flex items-center gap-2 border-b border-border px-4 py-2">
        <div className="min-w-0 flex-1">
          <div className="truncate text-sm font-medium">{detail.label}</div>
          <div className="truncate text-[11px] text-muted-foreground">{detail.group}</div>
        </div>
        <Button size="icon" variant="ghost" onClick={() => setAnalysis('')} title="Close">
          <X className="h-4 w-4"/>
        </Button>
      </div>
      <div className="flex-1 overflow-auto p-6">
        <p className="max-w-2xl text-sm leading-relaxed">{detail.description || 'No description.'}</p>
        {(detail.needs || []).length > 0 && (<div className="mt-6">
          <div className="mb-2 text-[10px] font-semibold uppercase text-muted-foreground">Uses</div>
          <ul className="space-y-1 text-sm">
            {(detail.needs || []).map((n) => (<li key={n} className="text-muted-foreground">{n}</li>))}
          </ul>
        </div>)}
      </div>
    </div>);
}
