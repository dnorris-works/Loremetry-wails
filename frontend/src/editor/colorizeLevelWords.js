/** Level / severity words colored in report HTML. */

const CLASS_HIGH = 'text-green-600 dark:text-green-400 font-medium';
const CLASS_MID = 'text-amber-600 dark:text-amber-400 font-medium';
const CLASS_LOW = 'text-red-600 dark:text-red-400 font-medium';

const LEVEL_RE = /\b(High|Strong|Good|Excellent|Medium|Moderate|Fair|Mixed|Low|Weak|Poor|Critical|Severe)\b/gi;

function classForLevel(word) {
    switch (word.toLowerCase()) {
        case 'high':
        case 'strong':
        case 'good':
        case 'excellent':
            return CLASS_HIGH;
        case 'medium':
        case 'moderate':
        case 'fair':
        case 'mixed':
            return CLASS_MID;
        case 'low':
        case 'weak':
        case 'poor':
        case 'critical':
        case 'severe':
            return CLASS_LOW;
        default:
            return '';
    }
}

/**
 * Wrap level words in colored spans. Only touches text outside tags;
 * skips content inside <code> and <pre>.
 */
export function colorizeLevelWords(html) {
    if (!html) return '';
    let out = '';
    let i = 0;
    let skipDepth = 0;
    while (i < html.length) {
        if (html[i] === '<') {
            const end = html.indexOf('>', i);
            if (end < 0) {
                out += html.slice(i);
                break;
            }
            const tag = html.slice(i, end + 1);
            if (/^<(code|pre)(\s|>|\/)/i.test(tag) && !/^<\//.test(tag)) {
                skipDepth += 1;
            }
            else if (/^<\/(code|pre)>/i.test(tag) && skipDepth > 0) {
                skipDepth -= 1;
            }
            out += tag;
            i = end + 1;
            continue;
        }
        let next = html.indexOf('<', i);
        if (next < 0) next = html.length;
        let text = html.slice(i, next);
        if (skipDepth === 0) {
            text = text.replace(LEVEL_RE, (match) => {
                const cls = classForLevel(match);
                if (!cls) return match;
                return `<span class="${cls}">${match}</span>`;
            });
        }
        out += text;
        i = next;
    }
    return out;
}
