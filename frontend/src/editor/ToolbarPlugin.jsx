import { useCallback, useEffect, useState } from 'react';
import { useLexicalComposerContext } from '@lexical/react/LexicalComposerContext';
import { $createParagraphNode, $getSelection, $isRangeSelection, CAN_REDO_COMMAND, CAN_UNDO_COMMAND, COMMAND_PRIORITY_CRITICAL, FORMAT_TEXT_COMMAND, REDO_COMMAND, SELECTION_CHANGE_COMMAND, UNDO_COMMAND, } from 'lexical';
import { $setBlocksType } from '@lexical/selection';
import { $createHeadingNode, $createQuoteNode, $isHeadingNode, $isQuoteNode } from '@lexical/rich-text';
import { $createCodeNode, $isCodeNode } from '@lexical/code';
import { $isListNode, INSERT_ORDERED_LIST_COMMAND, INSERT_UNORDERED_LIST_COMMAND, ListNode, REMOVE_LIST_COMMAND, } from '@lexical/list';
import { TOGGLE_LINK_COMMAND } from '@lexical/link';
import { $getNearestNodeOfType, mergeRegister } from '@lexical/utils';
import { Button } from '@/components/ui/button';
import { cn } from '@/lib/utils';
import { Bold, Code, Heading1, Heading2, Heading3, Italic, Link, List, ListOrdered, Quote, Redo, Strikethrough, Underline, Undo, } from 'lucide-react';
function Divider() {
    return <span className="mx-1 h-5 w-px bg-border"/>;
}
export function ToolbarPlugin() {
    const [editor] = useLexicalComposerContext();
    const [bold, setBold] = useState(false);
    const [italic, setItalic] = useState(false);
    const [underline, setUnderline] = useState(false);
    const [strike, setStrike] = useState(false);
    const [code, setCode] = useState(false);
    const [block, setBlock] = useState('paragraph');
    const [canUndo, setCanUndo] = useState(false);
    const [canRedo, setCanRedo] = useState(false);
    const sync = useCallback(() => {
        const selection = $getSelection();
        if (!$isRangeSelection(selection))
            return;
        setBold(selection.hasFormat('bold'));
        setItalic(selection.hasFormat('italic'));
        setUnderline(selection.hasFormat('underline'));
        setStrike(selection.hasFormat('strikethrough'));
        setCode(selection.hasFormat('code'));
        const node = selection.anchor.getNode();
        const element = node.getKey() === 'root' ? node : node.getTopLevelElementOrThrow();
        if ($isHeadingNode(element)) {
            setBlock(element.getTag());
        }
        else if ($isQuoteNode(element)) {
            setBlock('quote');
        }
        else if ($isCodeNode(element)) {
            setBlock('code');
        }
        else if ($isListNode(element)) {
            const parent = $getNearestNodeOfType(node, ListNode);
            setBlock(parent ? parent.getListType() : 'paragraph');
        }
        else {
            setBlock('paragraph');
        }
    }, []);
    useEffect(() => {
        return mergeRegister(editor.registerUpdateListener(({ editorState }) => {
            editorState.read(sync);
        }), editor.registerCommand(SELECTION_CHANGE_COMMAND, () => {
            sync();
            return false;
        }, COMMAND_PRIORITY_CRITICAL), editor.registerCommand(CAN_UNDO_COMMAND, (payload) => {
            setCanUndo(payload);
            return false;
        }, COMMAND_PRIORITY_CRITICAL), editor.registerCommand(CAN_REDO_COMMAND, (payload) => {
            setCanRedo(payload);
            return false;
        }, COMMAND_PRIORITY_CRITICAL));
    }, [editor, sync]);
    function formatHeading(tag) {
        editor.update(() => {
            const selection = $getSelection();
            if (!$isRangeSelection(selection))
                return;
            if (block === tag) {
                $setBlocksType(selection, () => $createParagraphNode());
            }
            else {
                $setBlocksType(selection, () => $createHeadingNode(tag));
            }
        });
    }
    function formatQuote() {
        editor.update(() => {
            const selection = $getSelection();
            if (!$isRangeSelection(selection))
                return;
            if (block === 'quote') {
                $setBlocksType(selection, () => $createParagraphNode());
            }
            else {
                $setBlocksType(selection, () => $createQuoteNode());
            }
        });
    }
    function formatCodeBlock() {
        editor.update(() => {
            const selection = $getSelection();
            if (!$isRangeSelection(selection))
                return;
            if (block === 'code') {
                $setBlocksType(selection, () => $createParagraphNode());
            }
            else {
                $setBlocksType(selection, () => $createCodeNode());
            }
        });
    }
    function formatList(type) {
        if (block === type) {
            editor.dispatchCommand(REMOVE_LIST_COMMAND, undefined);
        }
        else if (type === 'bullet') {
            editor.dispatchCommand(INSERT_UNORDERED_LIST_COMMAND, undefined);
        }
        else {
            editor.dispatchCommand(INSERT_ORDERED_LIST_COMMAND, undefined);
        }
    }
    function formatLink() {
        const url = window.prompt('Link URL');
        if (url === null)
            return;
        editor.dispatchCommand(TOGGLE_LINK_COMMAND, url.trim() === '' ? null : url.trim());
    }
    return (<div className="flex flex-wrap items-center gap-0.5 border-b border-border bg-card px-2 py-1">
      <ToolBtn title="Undo" disabled={!canUndo} onClick={() => editor.dispatchCommand(UNDO_COMMAND, undefined)}>
        <Undo className="h-4 w-4"/>
      </ToolBtn>
      <ToolBtn title="Redo" disabled={!canRedo} onClick={() => editor.dispatchCommand(REDO_COMMAND, undefined)}>
        <Redo className="h-4 w-4"/>
      </ToolBtn>
      <Divider />
      <ToolBtn title="Heading 1" active={block === 'h1'} onClick={() => formatHeading('h1')}>
        <Heading1 className="h-4 w-4"/>
      </ToolBtn>
      <ToolBtn title="Heading 2" active={block === 'h2'} onClick={() => formatHeading('h2')}>
        <Heading2 className="h-4 w-4"/>
      </ToolBtn>
      <ToolBtn title="Heading 3" active={block === 'h3'} onClick={() => formatHeading('h3')}>
        <Heading3 className="h-4 w-4"/>
      </ToolBtn>
      <Divider />
      <ToolBtn title="Bold" active={bold} onClick={() => editor.dispatchCommand(FORMAT_TEXT_COMMAND, 'bold')}>
        <Bold className="h-4 w-4"/>
      </ToolBtn>
      <ToolBtn title="Italic" active={italic} onClick={() => editor.dispatchCommand(FORMAT_TEXT_COMMAND, 'italic')}>
        <Italic className="h-4 w-4"/>
      </ToolBtn>
      <ToolBtn title="Underline" active={underline} onClick={() => editor.dispatchCommand(FORMAT_TEXT_COMMAND, 'underline')}>
        <Underline className="h-4 w-4"/>
      </ToolBtn>
      <ToolBtn title="Strikethrough" active={strike} onClick={() => editor.dispatchCommand(FORMAT_TEXT_COMMAND, 'strikethrough')}>
        <Strikethrough className="h-4 w-4"/>
      </ToolBtn>
      <ToolBtn title="Inline code" active={code} onClick={() => editor.dispatchCommand(FORMAT_TEXT_COMMAND, 'code')}>
        <Code className="h-4 w-4"/>
      </ToolBtn>
      <Divider />
      <ToolBtn title="Bullet list" active={block === 'bullet'} onClick={() => formatList('bullet')}>
        <List className="h-4 w-4"/>
      </ToolBtn>
      <ToolBtn title="Numbered list" active={block === 'number'} onClick={() => formatList('number')}>
        <ListOrdered className="h-4 w-4"/>
      </ToolBtn>
      <ToolBtn title="Quote" active={block === 'quote'} onClick={formatQuote}>
        <Quote className="h-4 w-4"/>
      </ToolBtn>
      <ToolBtn title="Code block" active={block === 'code'} onClick={formatCodeBlock}>
        <Code className="h-4 w-4"/>
      </ToolBtn>
      <ToolBtn title="Link" onClick={formatLink}>
        <Link className="h-4 w-4"/>
      </ToolBtn>
    </div>);
}
function ToolBtn({ title, active, disabled, onClick, children, }) {
    return (<Button type="button" size="icon" variant="ghost" title={title} disabled={disabled} className={cn('h-7 w-7', active && 'bg-accent text-accent-foreground')} onClick={onClick}>
      {children}
    </Button>);
}
