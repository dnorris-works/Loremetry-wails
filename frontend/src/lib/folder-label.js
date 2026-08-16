export function folderLabel(code, fallback) {
    switch (code) {
        case 'bible':
            return 'Bible Docs';
        case 'location':
            return 'Locations';
        case 'research':
            return 'Research Docs';
        case 'character':
            return 'Characters';
        case 'act':
            return 'Acts';
        case 'chapter':
            return 'Chapters';
        default:
            return fallback;
    }
}
export function folderHeader(code, fallback, count) {
    return `${folderLabel(code, fallback)} (${count})`;
}
