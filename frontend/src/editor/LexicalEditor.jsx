import { LexicalComposer } from '@lexical/react/LexicalComposer';
import { RichTextPlugin } from '@lexical/react/LexicalRichTextPlugin';
import { ContentEditable } from '@lexical/react/LexicalContentEditable';
import { HistoryPlugin } from '@lexical/react/LexicalHistoryPlugin';
import { OnChangePlugin } from '@lexical/react/LexicalOnChangePlugin';
import { ListPlugin } from '@lexical/react/LexicalListPlugin';
import { LinkPlugin } from '@lexical/react/LexicalLinkPlugin';
import { MarkdownShortcutPlugin } from '@lexical/react/LexicalMarkdownShortcutPlugin';
import { TabIndentationPlugin } from '@lexical/react/LexicalTabIndentationPlugin';
import { LexicalErrorBoundary } from '@lexical/react/LexicalErrorBoundary';
import { HeadingNode, QuoteNode } from '@lexical/rich-text';
import { ListItemNode, ListNode } from '@lexical/list';
import { CodeNode } from '@lexical/code';
import { LinkNode } from '@lexical/link';
import { TRANSFORMERS, $convertFromMarkdownString, $convertToMarkdownString } from '@lexical/markdown';
import { useRef } from 'react';
import { ToolbarPlugin } from '@/editor/ToolbarPlugin';
const theme = {
    paragraph: 'editor-paragraph',
    quote: 'editor-quote',
    heading: {
        h1: 'editor-h1',
        h2: 'editor-h2',
        h3: 'editor-h3',
    },
    list: {
        ul: 'editor-ul',
        ol: 'editor-ol',
        listitem: 'editor-li',
        nested: { listitem: 'editor-li' },
    },
    link: 'editor-link',
    text: {
        bold: 'editor-bold',
        italic: 'editor-italic',
        underline: 'editor-underline',
        strikethrough: 'editor-strike',
        code: 'editor-inline-code',
    },
    code: 'editor-code',
};
function onError(error) {
    console.error(error);
}
export function LexicalEditor({ docKey, markdown, onMarkdown }) {
    const initialConfig = {
        namespace: 'Loremetry',
        theme,
        onError,
        nodes: [HeadingNode, QuoteNode, ListNode, ListItemNode, CodeNode, LinkNode],
        editorState: () => {
            $convertFromMarkdownString(markdown || '', TRANSFORMERS);
        },
    };
    return (<LexicalComposer key={docKey} initialConfig={initialConfig}>
      <div className="flex h-full min-h-0 flex-col">
        <ToolbarPlugin />
        <div className="relative min-h-0 flex-1">
          <RichTextPlugin contentEditable={<ContentEditable className="lexical-editor h-full overflow-auto"/>} placeholder={<div className="pointer-events-none absolute left-6 top-5 text-muted-foreground">Start writing…</div>} ErrorBoundary={LexicalErrorBoundary}/>
        </div>
        <HistoryPlugin />
        <ListPlugin />
        <LinkPlugin />
        <TabIndentationPlugin />
        <MarkdownShortcutPlugin transformers={TRANSFORMERS}/>
        <SaveMarkdownPlugin onMarkdown={onMarkdown}/>
      </div>
    </LexicalComposer>);
}
function SaveMarkdownPlugin({ onMarkdown }) {
    const skipFirst = useRef(true);
    return (<OnChangePlugin ignoreSelectionChange onChange={(state) => {
            if (skipFirst.current) {
                skipFirst.current = false;
                return;
            }
            state.read(() => {
                onMarkdown($convertToMarkdownString(TRANSFORMERS));
            });
        }}/>);
}
