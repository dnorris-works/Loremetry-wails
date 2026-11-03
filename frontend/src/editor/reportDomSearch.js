function clearHighlights(root) {
    if (!root)
        return;
    root.querySelectorAll('mark.search-hit').forEach((mark) => {
        const parent = mark.parentNode;
        if (!parent)
            return;
        while (mark.firstChild)
            parent.insertBefore(mark.firstChild, mark);
        parent.removeChild(mark);
        parent.normalize();
    });
}

function collectHits(root, query) {
    const q = query.trim();
    if (!q || !root)
        return [];
    const qLower = q.toLowerCase();
    const hits = [];
    const walker = document.createTreeWalker(root, NodeFilter.SHOW_TEXT);
    let textNode = walker.nextNode();
    while (textNode) {
        const text = textNode.textContent || '';
        const lower = text.toLowerCase();
        let pos = 0;
        while (pos < lower.length) {
            const idx = lower.indexOf(qLower, pos);
            if (idx < 0)
                break;
            hits.push({ textNode, start: idx, end: idx + q.length });
            pos = idx + q.length;
        }
        textNode = walker.nextNode();
    }
    return hits;
}

function compareHit(a, b) {
    if (a.textNode === b.textNode)
        return a.start - b.start;
    const pos = a.textNode.compareDocumentPosition(b.textNode);
    if (pos & Node.DOCUMENT_POSITION_FOLLOWING)
        return -1;
    if (pos & Node.DOCUMENT_POSITION_PRECEDING)
        return 1;
    return 0;
}

export function applyDomSearch(root, query, activeIndex) {
    clearHighlights(root);
    if (!query.trim() || !root)
        return 0;
    const hits = collectHits(root, query);
    if (!hits.length)
        return 0;
    const toApply = [...hits].sort((a, b) => {
        const doc = compareHit(a, b);
        return doc !== 0 ? -doc : b.start - a.start;
    });
    hits.forEach((hit, index) => {
        hit.index = index;
    });
    for (const hit of toApply) {
        try {
            const range = document.createRange();
            range.setStart(hit.textNode, hit.start);
            range.setEnd(hit.textNode, hit.end);
            const mark = document.createElement('mark');
            mark.className = hit.index === activeIndex ? 'search-active search-hit' : 'search-hit';
            mark.dataset.searchIndex = String(hit.index);
            range.surroundContents(mark);
        } catch {
            // skip ranges that cross element boundaries
        }
    }
    return hits.length;
}

export function scrollToActiveMatch(root) {
    root?.querySelector('mark.search-active')?.scrollIntoView({ block: 'center', behavior: 'auto' });
}

export function clearDomSearch(root) {
    clearHighlights(root);
}
