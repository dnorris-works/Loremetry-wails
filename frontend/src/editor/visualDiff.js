/** Split markdown into stable blocks for visual diff. */
export function splitBlocks(text) {
    const raw = (text || '').replace(/\r\n/g, '\n').trim();
    if (!raw)
        return [];
    return raw.split(/\n{2,}/).map((b) => b.trim()).filter(Boolean);
}

export function isHeadingBlock(text) {
    return /^#{1,6}\s+\S/.test((text || '').trim());
}

export function headingTitle(text) {
    const m = (text || '').trim().match(/^#{1,6}\s+(.+)$/m);
    return m ? m[1].trim() : '';
}

/**
 * LCS-based block diff.
 * @returns {{ type: 'equal' | 'insert' | 'delete', value: string }[]}
 */
export function diffBlocks(original, proposed) {
    const a = splitBlocks(original);
    const b = splitBlocks(proposed);
    const n = a.length;
    const m = b.length;
    const dp = Array.from({ length: n + 1 }, () => new Array(m + 1).fill(0));
    for (let i = n - 1; i >= 0; i--) {
        for (let j = m - 1; j >= 0; j--) {
            if (a[i] === b[j])
                dp[i][j] = dp[i + 1][j + 1] + 1;
            else
                dp[i][j] = Math.max(dp[i + 1][j], dp[i][j + 1]);
        }
    }
    const ops = [];
    let i = 0;
    let j = 0;
    while (i < n && j < m) {
        if (a[i] === b[j]) {
            ops.push({ type: 'equal', value: a[i] });
            i++;
            j++;
        }
        else if (dp[i + 1][j] >= dp[i][j + 1]) {
            ops.push({ type: 'delete', value: a[i] });
            i++;
        }
        else {
            ops.push({ type: 'insert', value: b[j] });
            j++;
        }
    }
    while (i < n) {
        ops.push({ type: 'delete', value: a[i++] });
    }
    while (j < m) {
        ops.push({ type: 'insert', value: b[j++] });
    }
    return ops;
}

const CONTEXT_CHARS = 240;

/** Join prose blocks and clip from the start or end. Headings are not included. */
export function clipProse(parts, { fromEnd = false, maxChars = CONTEXT_CHARS } = {}) {
    const prose = (parts || []).filter((p) => p && !isHeadingBlock(p));
    if (prose.length === 0)
        return '';
    // Prefer a single nearby block so context stays local to the proposal.
    const raw = (fromEnd ? prose[prose.length - 1] : prose[0]).trim();
    if (raw.length <= maxChars)
        return raw;
    if (fromEnd) {
        const slice = raw.slice(-maxChars);
        const sentence = slice.match(/[.!?]["']?\s+(\S[\s\S]*)$/);
        const body = sentence ? sentence[1] : slice.replace(/^\S*\s+/, '');
        return '…' + body.trim();
    }
    const slice = raw.slice(0, maxChars);
    const sentence = slice.match(/^([\s\S]*[.!?])["']?(\s|$)/);
    const body = sentence ? sentence[1] : slice.replace(/\s+\S*$/, '');
    return body.trim() + '…';
}

function chapterAt(ops, index) {
    for (let k = index; k >= 0; k--) {
        if (ops[k].type === 'equal' && isHeadingBlock(ops[k].value))
            return headingTitle(ops[k].value);
    }
    return '';
}

/**
 * Navigable hunks: chapter label (not context) + short prose before + changes + short prose after.
 * After may come from the next chapter's body; headings themselves are never context.
 * @returns {{ chapter: string, before: string, after: string, deletes: string[], inserts: string[] }[]}
 */
export function proposalHunks(ops) {
    const hunks = [];
    let i = 0;
    while (i < ops.length) {
        if (ops[i].type === 'equal') {
            i++;
            continue;
        }
        const start = i;
        while (i < ops.length && ops[i].type !== 'equal')
            i++;
        const end = i;
        const deletes = [];
        const inserts = [];
        for (let k = start; k < end; k++) {
            if (ops[k].type === 'delete')
                deletes.push(ops[k].value);
            else if (ops[k].type === 'insert')
                inserts.push(ops[k].value);
        }

        const beforeParts = [];
        for (let k = start - 1; k >= 0; k--) {
            if (ops[k].type !== 'equal')
                continue;
            if (isHeadingBlock(ops[k].value))
                break;
            beforeParts.unshift(ops[k].value);
            break; // only the block immediately before the change
        }

        const afterParts = [];
        for (let k = end; k < ops.length; k++) {
            if (ops[k].type !== 'equal')
                continue;
            if (isHeadingBlock(ops[k].value))
                continue; // skip headings; keep looking for prose after
            afterParts.push(ops[k].value);
            break; // only the first prose block after the change
        }

        hunks.push({
            chapter: chapterAt(ops, start - 1),
            before: clipProse(beforeParts, { fromEnd: true }),
            after: clipProse(afterParts, { fromEnd: false }),
            deletes,
            inserts,
        });
    }
    return hunks;
}
