import { api } from '@/api/client';
const TEXT_EXT = /\.(md|txt|markdown|text|fountain|json)$/i;
export function isImportableName(name) {
    if (TEXT_EXT.test(name))
        return true;
    return !name.includes('.');
}
export function docTitleFromName(name) {
    return name.replace(/\.(md|txt|markdown|text|fountain|json)$/i, '');
}
export function basename(path) {
    const parts = path.split(/[/\\]/);
    return parts[parts.length - 1] || path;
}
export async function importHeaderFiles(projectPath, projectKind, kind, files) {
    let last;
    const existing = (await api.listHeaderFiles(projectPath, projectKind, kind)).files.map((f) => f.name.replace(/\.[^.]+$/, ''));
    for (const file of files) {
        if (!isImportableName(file.name))
            continue;
        const title = uniqueTitle(existing, docTitleFromName(file.name));
        existing.push(title);
        last = await api.createHeaderFile({
            project_path: projectPath,
            project_kind: projectKind,
            kind,
            name: title,
            text: file.text,
        });
    }
    return last ? { rel: last.rel, title: last.name } : null;
}
export async function importTextDocs(storyId, kind, files) {
    return null;
}
export async function importSeriesTextDocs(seriesId, category, files) {
    return null;
}
export async function importDroppedPaths(storyId, kind, paths) {
    return null;
}
export async function importDroppedSeriesPaths(seriesId, category, paths) {
    return null;
}
export async function filesFromList(list) {
    const files = [];
    for (const file of Array.from(list)) {
        if (!TEXT_EXT.test(file.name))
            continue;
        files.push({ name: file.name, text: await file.text() });
    }
    return files;
}
export async function importCharacters(parent, files) {
    if (!parent.projectPath)
        return null;
    const mapped = files.map((file) => {
        const parsed = parseCharacterFile(file.name, file.text);
        return { name: parsed.character_name + '.md', text: parsed.character_voice_notes || file.text };
    });
    return importHeaderFiles(parent.projectPath, parent.projectKind, 'character', mapped);
}
export async function importDroppedCharacters(parent, paths) {
    return importCharacters(parent, await api.readImportFiles(paths));
}
function parseCharacterFile(filename, text) {
    const trimmed = text.trim();
    if (trimmed.startsWith('{')) {
        try {
            const raw = JSON.parse(trimmed);
            const name = String(raw.character_name || raw.name || '').trim();
            if (name) {
                return {
                    character_name: name,
                    role_in_story: strField(raw, 'role_in_story', 'role'),
                    character_want: strField(raw, 'character_want', 'want'),
                    character_need: strField(raw, 'character_need', 'need'),
                    character_voice_notes: strField(raw, 'character_voice_notes', 'notes') || trimmed,
                };
            }
        }
        catch {
            /* use filename */
        }
    }
    const heading = text.match(/^#\s+(.+)$/m);
    return {
        character_name: (heading?.[1] || docTitleFromName(filename)).trim(),
        character_voice_notes: text,
    };
}
function strField(raw, ...keys) {
    for (const key of keys) {
        const v = raw[key];
        if (typeof v === 'string' && v.trim())
            return v;
    }
    return undefined;
}
let htmlDropAt = 0;
export function markHtmlFileDrop() {
    htmlDropAt = Date.now();
}
export function recentHtmlFileDrop() {
    return Date.now() - htmlDropAt < 800;
}
function uniqueTitle(existing, title) {
    if (!existing.includes(title))
        return title;
    let n = 2;
    while (existing.includes(`${title} ${n}`))
        n += 1;
    return `${title} ${n}`;
}
