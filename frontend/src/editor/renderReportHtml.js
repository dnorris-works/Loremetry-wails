import { marked } from 'marked';
import { colorizeLevelWords } from '@/editor/colorizeLevelWords';

marked.setOptions({ breaks: true, gfm: true });

/** Render report markdown for visual display. */
export function renderReportHtml(text) {
    return colorizeLevelWords(marked.parse(text || ''));
}
