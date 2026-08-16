import { folderLabel } from '@/lib/folder-label';
import { analysisNotice, analyzeFileKind } from '@/lib/file-kind';
import { importHeaderFiles, isImportableName, } from '@/lib/import-docs';
export async function importDroppedOnHeader(kind, parent, files, notice, confirm) {
    const usable = files.filter((f) => isImportableName(f.name));
    if (!usable.length) {
        await notice('Nothing to import. Use a .md, .txt, .json, or .markdown file.');
        return null;
    }
    const results = usable.map((f) => analyzeFileKind(kind, f.name, f.text));
    const folder = folderLabel(kind, kind);
    await notice(analysisNotice(folder, results));
    const matches = usable.filter((_, i) => results[i].match);
    const misses = usable.filter((_, i) => !results[i].match);
    let take = matches;
    if (misses.length) {
        const ok = await confirm(matches.length
            ? `${misses.length} file(s) do not look like ${folder}. Import those too?`
            : `None of these look like ${folder}. Import anyway?`);
        if (ok)
            take = usable;
        else if (!matches.length)
            return null;
    }
    if (!parent.projectPath) {
        await notice('Drop files on a series or story header.');
        return null;
    }
    return importHeaderFiles(parent.projectPath, parent.projectKind, kind, take);
}
