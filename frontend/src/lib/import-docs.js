export async function filesFromList(list) {
    const files = [];
    for (const file of Array.from(list)) {
        files.push({ name: file.name, text: await file.text() });
    }
    return files;
}

let htmlDropAt = 0;
export function markHtmlFileDrop() {
    htmlDropAt = Date.now();
}
export function recentHtmlFileDrop() {
    return Date.now() - htmlDropAt < 800;
}
