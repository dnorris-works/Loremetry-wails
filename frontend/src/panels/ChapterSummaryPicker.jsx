import { Button } from '@/components/ui/button';

/** Chapter picker for Chapter Plot Summary — checkbox cards + Run. */
export function ChapterSummaryPicker({
  chapters,
  selected,
  onToggle,
  onSelectAll,
  onClear,
  onRun,
  loading,
  error,
}) {
  const selectedCount = selected.size;
  const allSelected = chapters.length > 0 && selectedCount === chapters.length;
  return (
    <div className="flex h-full flex-col">
      <div className="flex flex-wrap items-center gap-2 border-b border-border px-4 py-2">
        <p className="mr-auto text-sm text-muted-foreground">
          Choose chapters to summarize
          {chapters.length > 0 ? ` (${selectedCount} of ${chapters.length})` : ''}
        </p>
        <Button size="sm" variant="outline" onClick={onSelectAll} disabled={!chapters.length || loading}>
          {allSelected ? 'All selected' : 'Select all'}
        </Button>
        <Button size="sm" variant="outline" onClick={onClear} disabled={!selectedCount || loading}>
          Clear
        </Button>
        <Button size="sm" onClick={onRun} disabled={!selectedCount || loading}>
          Run
        </Button>
      </div>
      <div className="flex-1 overflow-auto p-4">
        {loading && !chapters.length ? (
          <p className="text-sm text-muted-foreground">Loading chapters…</p>
        ) : error ? (
          <p className="text-sm text-destructive">{error}</p>
        ) : !chapters.length ? (
          <p className="text-sm text-muted-foreground">No manuscript chapters found in this book.</p>
        ) : (
          <ul className="mx-auto grid max-w-3xl gap-2">
            {chapters.map((ch) => {
              const rel = ch.rel || '';
              const checked = selected.has(rel);
              return (
                <li key={rel}>
                  <label
                    className={`flex cursor-pointer gap-3 rounded-md border border-border px-3 py-2.5 transition-colors ${
                      checked ? 'bg-muted/40' : 'hover:bg-muted/20'
                    }`}
                  >
                    <input
                      type="checkbox"
                      className="mt-1 shrink-0"
                      checked={checked}
                      onChange={() => onToggle(rel)}
                    />
                    <span className="min-w-0 flex-1">
                      <span className="block text-sm font-medium text-foreground">{ch.name || rel}</span>
                      <span className="mt-0.5 block text-xs leading-relaxed text-muted-foreground">
                        {ch.description || 'No description yet.'}
                      </span>
                    </span>
                  </label>
                </li>
              );
            })}
          </ul>
        )}
      </div>
    </div>
  );
}
