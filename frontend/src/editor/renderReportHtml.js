import { marked } from 'marked';

marked.setOptions({ breaks: true, gfm: true });

/** Render report markdown for visual display. */
export function renderReportHtml(text) {
    return marked.parse(text || '');
}
