const KEY = 'loremetry-sidebar';
let current = null;
export function setSidebarDrag(e, item) {
    current = item;
    e.dataTransfer.setData(KEY, JSON.stringify(item));
    e.dataTransfer.setData('text/plain', JSON.stringify(item));
    e.dataTransfer.effectAllowed = 'move';
}
export function currentSidebarDrag() {
    return current;
}
export function clearSidebarDrag() {
    current = null;
}
export function sidebarDrag(e) {
    const raw = e.dataTransfer.getData(KEY) || e.dataTransfer.getData('text/plain');
    if (raw) {
        try {
            const v = JSON.parse(raw);
            if (v && typeof v === 'object' && 't' in v)
                return v;
        }
        catch {
            /* ignore */
        }
    }
    return current;
}
export function hasFiles(e) {
    return Array.from(e.dataTransfer.types).includes('Files');
}
export function moveBefore(ids, id, beforeId) {
    const next = ids.filter((x) => x !== id);
    if (beforeId == null)
        return [...next, id];
    const i = next.indexOf(beforeId);
    if (i < 0)
        return [...next, id];
    next.splice(i, 0, id);
    return next;
}
export function inStoryGroup(story, seriesId) {
    if (seriesId)
        return story.series_id === seriesId;
    return !story.series_id;
}
export function applyStoryOrder(list, seriesId, ids) {
    const byId = new Map(list.map((s) => [s.id, s]));
    const ordered = ids.map((id) => byId.get(id)).filter(Boolean);
    const next = [];
    let placed = false;
    for (const st of list) {
        if (!inStoryGroup(st, seriesId)) {
            next.push(st);
            continue;
        }
        if (!placed) {
            next.push(...ordered);
            placed = true;
        }
    }
    return next;
}
