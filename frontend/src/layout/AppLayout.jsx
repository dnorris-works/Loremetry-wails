import { useEffect, useState } from 'react';
import { EventsOn } from '../../wailsjs/runtime/runtime';
import { Panel, PanelGroup, PanelResizeHandle } from 'react-resizable-panels';
import { Sidebar } from '@/panels/Sidebar';
import { DocumentPane } from '@/panels/DocumentPane';
import { ChatPane } from '@/panels/ChatPane';
import { useAppState } from '@/lib/app-state';
import { FileDropListener } from '@/lib/FileDropListener';
import { SeriesDialog, StoryDialog } from '@/dialogs/EntityDialogs';
import { AdminDialog, SettingsDialog } from '@/dialogs/AppDialogs';
const LAYOUT_KEY = 'loremetry_panel_layout';
export function AppLayout() {
    const { bumpRefresh, ensureOpenSeries, ensureOpenStory } = useAppState();
    useEffect(() => {
        return EventsOn('folders-changed', () => bumpRefresh());
    }, [bumpRefresh]);
    const [seriesOpen, setSeriesOpen] = useState(false);
    const [editSeriesPath, setEditSeriesPath] = useState('');
    const [editSeriesName, setEditSeriesName] = useState('');
    const [storyOpen, setStoryOpen] = useState(false);
    const [editStoryPath, setEditStoryPath] = useState('');
    const [editStoryName, setEditStoryName] = useState('');
    const [storySeriesPath, setStorySeriesPath] = useState('');
    const [defaultPen, setDefaultPen] = useState('');
    const [defaultPenPath, setDefaultPenPath] = useState('');
    const [settingsOpen, setSettingsOpen] = useState(false);
    const [adminOpen, setAdminOpen] = useState(false);
    const defaultLayout = (() => {
        try {
            const raw = localStorage.getItem(LAYOUT_KEY);
            if (raw)
                return JSON.parse(raw);
        }
        catch {
            /* ignore */
        }
        return [22, 53, 25];
    })();
    return (<div className="h-full">
      <FileDropListener />
      <PanelGroup className="h-full" direction="horizontal" onLayout={(sizes) => localStorage.setItem(LAYOUT_KEY, JSON.stringify(sizes))}>
        <Panel defaultSize={defaultLayout[0]} minSize={14}>
          <Sidebar onNewSeries={(pen, penPath) => {
            setDefaultPen(pen || '');
            setDefaultPenPath(penPath || '');
            setEditSeriesPath('');
            setEditSeriesName('');
            setSeriesOpen(true);
        }} onEditSeries={(path, name) => {
            setEditSeriesPath(path);
            setEditSeriesName(name);
            setSeriesOpen(true);
        }} onNewStory={(seriesPath, pen, penPath) => {
            setDefaultPen(pen || '');
            setDefaultPenPath(penPath || '');
            setEditStoryPath('');
            setEditStoryName('');
            setStorySeriesPath(seriesPath || '');
            setStoryOpen(true);
        }} onEditStory={(path, name) => {
            setEditStoryPath(path);
            setEditStoryName(name);
            setStoryOpen(true);
        }} onSettings={() => setSettingsOpen(true)} onAdmin={() => setAdminOpen(true)}/>
        </Panel>
        <PanelResizeHandle className="w-1 bg-border hover:bg-primary"/>
        <Panel defaultSize={defaultLayout[1]} minSize={30}>
          <DocumentPane />
        </Panel>
        <PanelResizeHandle className="w-1 bg-border hover:bg-primary"/>
        <Panel defaultSize={defaultLayout[2]} minSize={16}>
          <ChatPane />
        </Panel>
      </PanelGroup>

      <SeriesDialog open={seriesOpen} onOpenChange={setSeriesOpen} projectPath={editSeriesPath} projectName={editSeriesName} defaultPen={defaultPen} defaultPenPath={defaultPenPath} onSaved={bumpRefresh}/>
      <StoryDialog open={storyOpen} onOpenChange={setStoryOpen} projectPath={editStoryPath} projectName={editStoryName} defaultSeriesPath={storySeriesPath} defaultPen={defaultPen} defaultPenPath={defaultPenPath} onSaved={(created) => {
            bumpRefresh();
            if (!created?.path)
                return;
            if (created.seriesPath)
                ensureOpenSeries(created.seriesPath);
            ensureOpenStory(created.path);
        }}/>
      <SettingsDialog open={settingsOpen} onOpenChange={setSettingsOpen}/>
      <AdminDialog open={adminOpen} onOpenChange={setAdminOpen}/>
    </div>);
}
