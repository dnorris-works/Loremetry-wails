import { folderLabel } from '@/lib/folder-label';
const KINDS = ['character', 'chapter', 'act', 'location', 'research', 'bible'];
const SIGNALS = {
    character: [
        'character', 'protagonist', 'antagonist', 'npc', 'pov', 'want', 'need', 'core wound',
        'defining trait', 'role in story', 'character_name', 'arc', 'backstory', 'personality',
    ],
    chapter: [
        'chapter', 'scene', 'manuscript', 'he said', 'she said', 'they said',
        'chapter 1', 'chapter 2', 'chapter 3',
    ],
    act: ['act 1', 'act 2', 'act 3', 'act i', 'act ii', 'act iii', 'act one', 'act two'],
    location: [
        'location', 'setting', 'geography', 'town', 'city', 'village', 'castle', 'room',
        'landscape', 'place', 'region', 'map',
    ],
    research: [
        'research', 'source', 'citation', 'bibliography', 'reference', 'according to',
        'notes', 'interview', 'article',
    ],
    bible: [
        'bible', 'canon', 'lore', 'worldbuilding', 'magic system', 'timeline', 'mythology',
        'rules of', 'world rules',
    ],
};
export function analyzeFileKind(expected, name, text) {
    const hay = `${name}\n${text}`.toLowerCase();
    const scores = new Map();
    for (const kind of KINDS) {
        let score = 0;
        for (const word of SIGNALS[kind]) {
            if (hay.includes(word))
                score += word.length > 8 ? 2 : 1;
        }
        scores.set(kind, score);
    }
    const base = name.toLowerCase();
    if (/\bch(apter)?[\s._-]?\d/.test(base) || /^0?\d+[-_]/.test(base)) {
        scores.set('chapter', (scores.get('chapter') || 0) + 4);
    }
    if (/\bact[\s._-]?\d/.test(base))
        scores.set('act', (scores.get('act') || 0) + 4);
    if (/\b(char|character|npc)\b/.test(base))
        scores.set('character', (scores.get('character') || 0) + 4);
    if (/\b(loc|location|place|setting)\b/.test(base))
        scores.set('location', (scores.get('location') || 0) + 3);
    if (/\b(research|notes|sources?)\b/.test(base))
        scores.set('research', (scores.get('research') || 0) + 3);
    if (/\b(bible|lore|canon|world)\b/.test(base))
        scores.set('bible', (scores.get('bible') || 0) + 3);
    if (text.trim().startsWith('{') && /character_name|"name"/.test(text)) {
        scores.set('character', (scores.get('character') || 0) + 5);
    }
    const words = text.split(/\s+/).filter(Boolean).length;
    const quoteHeavy = (text.match(/["“]/g) || []).length > 8;
    if (words > 800 || quoteHeavy)
        scores.set('chapter', (scores.get('chapter') || 0) + 3);
    if (words < 120 && /\bact\b/i.test(hay))
        scores.set('act', (scores.get('act') || 0) + 2);
    let guessed = expected || 'bible';
    let best = -1;
    for (const kind of KINDS) {
        const n = scores.get(kind) || 0;
        if (n > best) {
            best = n;
            guessed = kind;
        }
    }
    if (best <= 0) {
        guessed = (['bible', 'location', 'research', 'chapter'].includes(expected) ? expected : 'chapter');
    }
    const expectedScore = scores.get(expected) || 0;
    const match = guessed === expected || (expectedScore > 0 && expectedScore >= best - 1);
    return {
        file: name,
        expected,
        guessed,
        match,
        score: best,
        detail: best <= 0
            ? 'No strong signals; treating as a generic text file.'
            : `Strongest signals point to ${folderLabel(guessed, guessed)}.`,
    };
}
export function guessFileKind(name, text) {
    const r = analyzeFileKind('', name, text);
    return r.score > 0 ? r.guessed : null;
}
export function analysisNotice(folder, results) {
    const lines = results.map((r) => {
        const verdict = r.match ? 'matches' : 'does not match';
        return `${r.file}: looks like ${folderLabel(r.guessed, r.guessed)} (${verdict}). ${r.detail}`;
    });
    return `Dropped on ${folder}\n\n${lines.join('\n\n')}`;
}
