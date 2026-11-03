import { useMemo } from 'react';
import CodeMirrorMerge from 'react-codemirror-merge';
import { markdown } from '@codemirror/lang-markdown';
import { EditorState } from '@codemirror/state';
import { EditorView } from '@codemirror/view';
import { vscodeDark, vscodeLight } from '@uiw/codemirror-theme-vscode';
import { readTheme } from '@/lib/theme';

const Original = CodeMirrorMerge.Original;
const Modified = CodeMirrorMerge.Modified;

const sharedExtensions = [markdown(), EditorView.lineWrapping];

export function ProposalMergeView({ original = '', proposed = '' }) {
    const dark = readTheme() === 'dark';
    const theme = useMemo(() => (dark ? vscodeDark : vscodeLight), [dark]);
    const extensions = useMemo(() => [theme, ...sharedExtensions], [theme]);
    const readOnly = useMemo(() => [EditorState.readOnly.of(true)], []);

    return (
        <div className="proposal-merge h-full min-h-0">
            <CodeMirrorMerge
                orientation="a-b"
                highlightChanges
                gutter
                collapseUnchanged={{ margin: 3, minSize: 4 }}
                revertControls="a-to-b"
            >
                <Original value={original} extensions={[...extensions, ...readOnly]}/>
                <Modified value={proposed} extensions={[...extensions, ...readOnly]}/>
            </CodeMirrorMerge>
        </div>
    );
}
