import { api } from '@/api/client';
import { guessFileKind } from '@/lib/file-kind';
import { folderLabel } from '@/lib/folder-label';
import { basename, importHeaderFiles } from '@/lib/import-docs';

const FOLDER_KIND = [
    [/^(characters?|chars|people)$/i, 'character'],
    [/^(locations?|places|settings?)$/i, 'location'],
    [/^(research|notes)$/i, 'research'],
    [/^(bible|lore|canon|world)$/i, 'bible'],
    [/^(chapters?|manuscripts?|ms|prose)$/i, 'chapter'],
    [/^(acts?)$/i, 'act'],
];

export function kindFromFolderName(name) {
    const n = String(name || '').replace(/[_-]+/g, ' ').trim();
    for (const [re, kind] of FOLDER_KIND) {
        if (re.test(n))
            return kind;
    }
    return null;
}

export function classifyImportRel(rel, rootFolderName) {
    const parts = String(rel || '').split(/[/\\]/).filter(Boolean);
    const file = parts.pop() || rel;
    const dirs = parts;
    const rootKind = kindFromFolderName(rootFolderName);
    let kind = null;
    for (let i = 0; i < dirs.length; i++) {
        const k = kindFromFolderName(dirs[i]);
        if (k) {
            kind = k;
            break;
        }
    }
    if (!kind && rootKind)
        kind = rootKind;
    return { kind, file };
}

export async function importStoryFolder(projectPath, files, rootFolderName, notice) {
    const rows = [];
    for (const file of files) {
        const rel = file.rel || file.name;
        let { kind } = classifyImportRel(rel, rootFolderName);
        if (!kind)
            kind = guessFileKind(file.name, file.text);
        if (!kind)
            continue;
        rows.push({ ...file, kind });
    }
    if (!rows.length) {
        await notice('No importable text files found in that folder.');
        return null;
    }
    const counts = {};
    for (const r of rows)
        counts[r.kind] = (counts[r.kind] || 0) + 1;
    const summary = Object.entries(counts)
        .map(([k, n]) => `${n} ${folderLabel(k, k)}`)
        .join(', ');
    await notice(`Importing ${rows.length} file(s): ${summary}.`);
    let last = null;
    for (const kind of ['act', 'character', 'location', 'research', 'bible', 'chapter']) {
        const group = rows.filter((r) => r.kind === kind);
        if (!group.length)
            continue;
        const created = await importHeaderFiles(projectPath, 'story', kind, group);
        if (created)
            last = { type: 'file', kind, ...created };
    }
    return last;
}

export async function pickAndImportStoryFolder(projectPath, notice) {
    const dir = await api.pickImportFolder();
    if (!dir)
        return null;
    const files = await api.readImportFiles([dir]);
    return importStoryFolder(projectPath, files, basename(dir), notice);
}
