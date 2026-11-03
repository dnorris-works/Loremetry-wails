import assert from 'node:assert/strict';
import { test } from 'node:test';
import { colorizeLevelWords } from './colorizeLevelWords.js';

test('colors High green', () => {
    const html = colorizeLevelWords('<p>Overall Hook Strength: High</p>');
    assert.match(html, /<span class="text-green-600[^"]*">High<\/span>/);
    assert.ok(html.includes('Overall Hook Strength:'));
});

test('colors Medium amber and Low red', () => {
    const html = colorizeLevelWords('<p>Risk: Medium — Severity: Low</p>');
    assert.match(html, /text-amber-600[^"]*">Medium<\/span>/);
    assert.match(html, /text-red-600[^"]*">Low<\/span>/);
});

test('colors synonyms Critical and Strong', () => {
    const html = colorizeLevelWords('<p>Critical issue; Strong pull</p>');
    assert.match(html, /text-red-600[^"]*">Critical<\/span>/);
    assert.match(html, /text-green-600[^"]*">Strong<\/span>/);
});

test('skips code and pre', () => {
    const html = colorizeLevelWords('<p>High</p><code>High</code><pre>Low</pre>');
    assert.match(html, /<p><span class="text-green-600[^"]*">High<\/span><\/p>/);
    assert.match(html, /<code>High<\/code>/);
    assert.match(html, /<pre>Low<\/pre>/);
});

test('does not match inside longer words', () => {
    const html = colorizeLevelWords('<p>Highway and lowest</p>');
    assert.equal(html, '<p>Highway and lowest</p>');
});
