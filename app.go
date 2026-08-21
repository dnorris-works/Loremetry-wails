package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"loremetry/internal/cloud"
	appdb "loremetry/internal/db"
	"loremetry/internal/store"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type ImportedFile struct {
	Name string `json:"name"`
	Text string `json:"text"`
	Rel  string `json:"rel"`
}

type App struct {
	ctx       context.Context
	db        *sql.DB
	store     *store.Store
	dbPath    string
	dbErr     error
	watchMu   sync.Mutex
	watchStop chan struct{}
	syncing   bool
	busy      bool
}

func NewApp() *App {
	return &App{}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	path, err := appdb.DefaultPath()
	if err != nil {
		a.dbErr = err
		return
	}
	a.dbPath = path
	conn, err := appdb.Open(path)
	if err != nil {
		a.dbErr = err
		return
	}
	store.SetCatalogDB(conn)
	uid, err := appdb.LocalUserID(conn)
	if err != nil {
		a.dbErr = err
		_ = conn.Close()
		return
	}
	a.db = conn
	a.store = &store.Store{DB: conn, UserID: uid}
	a.startFolderWatch()
}

func (a *App) shutdown(ctx context.Context) {
	a.stopFolderWatch()
	if a.db != nil {
		_ = a.db.Close()
	}
}

func (a *App) ready() (*store.Store, error) {
	if a.dbErr != nil {
		return nil, a.dbErr
	}
	if a.store == nil {
		return nil, fmt.Errorf("database is not open")
	}
	return a.store, nil
}

func (a *App) GetSession() (store.AuthSession, error) {
	s, err := a.ready()
	if err != nil {
		return store.AuthSession{Authenticated: false, Reason: err.Error()}, err
	}
	return s.Session()
}

func (a *App) GetUISession() (store.UISession, error) {
	s, err := a.ready()
	if err != nil {
		return store.UISession{}, err
	}
	return s.GetUISession()
}

func (a *App) SetUISelection(dir string, name string) (store.UISession, error) {
	s, err := a.ready()
	if err != nil {
		return store.UISession{}, err
	}
	return s.SetUISelection(dir, name)
}

func (a *App) ToggleOpenSeries(path string) (store.UISession, error) {
	s, err := a.ready()
	if err != nil {
		return store.UISession{}, err
	}
	return s.ToggleOpenSeries(path)
}

func (a *App) ToggleOpenStory(path string) (store.UISession, error) {
	s, err := a.ready()
	if err != nil {
		return store.UISession{}, err
	}
	return s.ToggleOpenStory(path)
}

func (a *App) EnsureOpenSeries(path string) (store.UISession, error) {
	s, err := a.ready()
	if err != nil {
		return store.UISession{}, err
	}
	return s.EnsureOpenSeries(path)
}

func (a *App) EnsureOpenStory(path string) (store.UISession, error) {
	s, err := a.ready()
	if err != nil {
		return store.UISession{}, err
	}
	return s.EnsureOpenStory(path)
}

func (a *App) ToggleOpenFolder(path string) (store.UISession, error) {
	s, err := a.ready()
	if err != nil {
		return store.UISession{}, err
	}
	return s.ToggleOpenFolder(path)
}

func (a *App) EnsureOpenFolder(path string) (store.UISession, error) {
	s, err := a.ready()
	if err != nil {
		return store.UISession{}, err
	}
	return s.EnsureOpenFolder(path)
}

func (a *App) SetRestoreOpen(on bool) (store.UISession, error) {
	s, err := a.ready()
	if err != nil {
		return store.UISession{}, err
	}
	return s.SetRestoreOpen(on)
}

func (a *App) SetLastPen(name string) (store.UISession, error) {
	s, err := a.ready()
	if err != nil {
		return store.UISession{}, err
	}
	return s.SetLastPen(name)
}

func (a *App) ListSeries() ([]store.Series, error) {
	s, err := a.ready()
	if err != nil {
		return nil, err
	}
	return s.ListSeries()
}

func (a *App) CreateSeries(in store.SeriesInput) (store.IDResult, error) {
	s, err := a.ready()
	if err != nil {
		return store.IDResult{}, err
	}
	if strings.TrimSpace(in.PenName) == "" {
		return store.IDResult{}, fmt.Errorf("choose a pen name")
	}
	parent, err := a.penDir(in.PenName)
	if err != nil {
		return store.IDResult{}, err
	}
	out, err := s.CreateSeries(in)
	if err != nil {
		return store.IDResult{}, err
	}
	root, err := store.ApplySeriesTemplate(parent, in.Name)
	if err != nil {
		return out, err
	}
	if err := s.SetRootLink("series", out.ID, root); err != nil {
		return out, err
	}
	if err := s.LinkProjectHeaders("series", out.ID, root); err != nil {
		return out, err
	}
	a.startFolderWatch()
	return out, nil
}

func (a *App) UpdateSeries(id int64, in store.SeriesInput) (store.UpdatedResult, error) {
	s, err := a.ready()
	if err != nil {
		return store.UpdatedResult{}, err
	}
	return s.UpdateSeries(id, in)
}

func (a *App) DeleteSeries(id int64) (store.DeletedResult, error) {
	s, err := a.ready()
	if err != nil {
		return store.DeletedResult{}, err
	}
	return s.DeleteSeries(id)
}

func (a *App) GetSeriesBible(seriesID int64) (store.BibleDoc, error) {
	s, err := a.ready()
	if err != nil {
		return store.BibleDoc{}, err
	}
	return s.GetSeriesBible(seriesID)
}

func (a *App) UpdateSeriesBible(seriesID int64, textContent string) (store.UpdatedResult, error) {
	s, err := a.ready()
	if err != nil {
		return store.UpdatedResult{}, err
	}
	return s.UpdateSeriesBible(seriesID, textContent)
}

func (a *App) ListStories(seriesID int64) ([]store.Story, error) {
	s, err := a.ready()
	if err != nil {
		return nil, err
	}
	return s.ListStories(seriesID)
}

func (a *App) CreateStory(in store.StoryInput) (store.IDResult, error) {
	s, err := a.ready()
	if err != nil {
		return store.IDResult{}, err
	}
	if in.SeriesID == 0 && strings.TrimSpace(in.PenName) == "" {
		return store.IDResult{}, fmt.Errorf("choose a pen name")
	}
	if in.SeriesID > 0 {
		list, _ := s.ListSeries()
		for _, se := range list {
			if se.ID == in.SeriesID && se.PenName != "" {
				in.PenName = se.PenName
				break
			}
		}
	}
	parent, seriesName, err := a.storyTemplateParent(s, in.SeriesID, in.PenName)
	if err != nil {
		return store.IDResult{}, err
	}
	out, err := s.CreateStory(in)
	if err != nil {
		return store.IDResult{}, err
	}
	root, err := store.ApplyBookTemplate(parent, in.Name, seriesName, s.GetDraftSection())
	if err != nil {
		return out, err
	}
	if err := s.SetRootLink("story", out.ID, root); err != nil {
		return out, err
	}
	if err := s.LinkProjectHeaders("story", out.ID, root); err != nil {
		return out, err
	}
	a.startFolderWatch()
	return out, nil
}

func (a *App) ReadDiskFile(dir string, name string) (store.DiskFile, error) {
	return store.ReadDiskFile(dir, name)
}

func (a *App) OpenDiskFile(dir string, name string) (store.OpenedFile, error) {
	return store.OpenDiskFile(dir, name)
}

func (a *App) ReadDiskFileRange(dir string, name string, offset int64, limit int) (store.FileRange, error) {
	return store.ReadDiskFileRange(dir, name, offset, limit)
}

func (a *App) NeedsChunking(dir string, name string) bool {
	return store.NeedsChunking(dir, name)
}

func (a *App) ReadDiskFileSections(dir string, name string) (store.FileSections, error) {
	return store.ReadDiskFileSections(dir, name)
}

func (a *App) ReadDiskFileSection(dir string, name string, index int) (store.FileSectionContent, error) {
	return store.ReadDiskFileSection(dir, name, index)
}

func (a *App) WriteDiskFileSection(dir string, name string, index int, text string) (store.DiskFile, error) {
	return store.WriteDiskFileSection(dir, name, index, text)
}

func (a *App) WriteDiskFile(in store.DiskFileWrite) (store.DiskFile, error) {
	return withBusy(a, func() (store.DiskFile, error) { return store.WriteDiskFile(in) })
}

func (a *App) CreateDiskFile(dir string, name string, text string) (store.DiskFile, error) {
	return withBusy(a, func() (store.DiskFile, error) {
		out, err := store.CreateDiskFile(dir, name, text)
		if err != nil {
			return store.DiskFile{}, err
		}
		a.startFolderWatch()
		return out, nil
	})
}

func (a *App) CreateDiskFiles(dir string, files []store.IncomingFile) (store.DiskFile, error) {
	return withBusy(a, func() (store.DiskFile, error) {
		out, err := store.CreateDiskFiles(dir, files)
		if err != nil {
			return store.DiskFile{}, err
		}
		a.startFolderWatch()
		return out, nil
	})
}

func (a *App) ListAnalysisCatalog() []store.AnalysisGroup {
	return store.AnalysisCatalog()
}

func (a *App) GetAnalysis(id string) (store.AnalysisDetail, error) {
	got, ok := store.GetAnalysisDetail(id)
	if !ok {
		return store.AnalysisDetail{}, fmt.Errorf("unknown analysis")
	}
	return got, nil
}

func (a *App) AnalysisRunQueue(id string) []string {
	return store.AnalysisRunQueue(id)
}

func (a *App) RunAnalysis(id string, projectPath string) (store.AnalysisReport, error) {
	return withBusy(a, func() (store.AnalysisReport, error) { return a.runAnalysis(id, projectPath) })
}

func (a *App) runAnalysis(id string, projectPath string) (store.AnalysisReport, error) {
	detail, ok := store.GetAnalysisDetail(id)
	if !ok {
		return store.AnalysisReport{}, fmt.Errorf("unknown analysis")
	}
	root := store.ResolveProjectRoot(projectPath)
	if root == "" {
		return store.AnalysisReport{}, fmt.Errorf("select a book or series first")
	}
	src := store.MatchAnalysisSources(root)
	for _, need := range detail.Needs {
		role := src.Role(need)
		if !role.Present || len(role.Files) == 0 {
			return store.AnalysisReport{}, fmt.Errorf("missing source: %s", need)
		}
	}
	var body string
	var original, proposed, dataJSON string
	usesAI := detail.UsesAI
	if !usesAI {
		result, runErr := store.RunLocalAnalysis(id, root)
		if runErr != nil {
			return store.AnalysisReport{}, runErr
		}
		body = result.Markdown
		original = result.Original
		proposed = result.Proposed
		dataJSON = result.DataJSON
	} else {
		cli, err := a.cloudClient()
		if err != nil {
			return store.AnalysisReport{}, err
		}
		blobs := store.CollectNeededText(root, detail.Needs)
		src := make([]cloud.RoleText, 0, len(blobs))
		for _, b := range blobs {
			src = append(src, cloud.RoleText{Role: b.Role, Rel: b.Rel, Name: b.Name, Text: b.Text})
		}
		out, err := cli.RunAnalysis(id, src)
		if err != nil {
			return store.AnalysisReport{}, err
		}
		body = store.FormatReportBody(detail.Label, detail.ID, out.Body)
		dataJSON = `{"kind":"ai","analysis_id":"` + detail.ID + `"}`
	}
	if err := a.persistResult(detail, root, body, original, proposed, dataJSON); err != nil {
		return store.AnalysisReport{}, err
	}
	size := len(body)
	return store.AnalysisReport{
		AnalysisID:    detail.ID,
		AnalysisLabel: detail.Label,
		ProjectPath:   root,
		UsesAI:        usesAI,
		Body:          body,
		Original:      original,
		Proposed:      proposed,
		BodySize:      size,
	}, nil
}

func (a *App) persistResult(detail store.AnalysisDetail, root, body, original, proposed, dataJSON string) error {
	s, err := a.ready()
	if err != nil {
		return err
	}
	_, err = s.UpsertAnalysisResult(store.AnalysisResult{
		AnalysisID:  detail.ID,
		Label:       detail.Label,
		ProjectPath: root,
		DataJSON:    dataJSON,
		Markdown:    body,
		Original:    original,
		Proposed:    proposed,
	})
	return err
}

func (a *App) PersistAnalysisResult(id string, projectPath string, body string, original string, proposed string, dataJSON string) (store.AnalysisReport, error) {
	detail, ok := store.GetAnalysisDetail(id)
	if !ok {
		return store.AnalysisReport{}, fmt.Errorf("unknown analysis")
	}
	root := store.ResolveProjectRoot(projectPath)
	if root == "" {
		return store.AnalysisReport{}, fmt.Errorf("select a book or series first")
	}
	if strings.TrimSpace(dataJSON) == "" {
		dataJSON = `{"kind":"ai","analysis_id":"` + detail.ID + `"}`
	}
	md := body
	if !strings.Contains(body, "## About this report") {
		md = store.FormatReportBody(detail.Label, detail.ID, body)
	}
	if err := a.persistResult(detail, root, md, original, proposed, dataJSON); err != nil {
		return store.AnalysisReport{}, err
	}
	return store.AnalysisReport{
		AnalysisID:    detail.ID,
		AnalysisLabel: detail.Label,
		ProjectPath:   root,
		UsesAI:        detail.UsesAI,
		Body:          md,
		Original:      original,
		Proposed:      proposed,
		BodySize:      len(md),
	}, nil
}

func (a *App) ListAnalysisReports() ([]store.AnalysisReportSummary, error) {
	s, err := a.ready()
	if err != nil {
		return nil, err
	}
	return s.ListAnalysisReports()
}

func (a *App) GetAnalysisReport(id int64) (store.AnalysisReport, error) {
	s, err := a.ready()
	if err != nil {
		return store.AnalysisReport{}, err
	}
	return s.GetAnalysisReport(id)
}

func (a *App) GetAnalysisResult(projectPath string, analysisID string) (store.AnalysisResult, error) {
	s, err := a.ready()
	if err != nil {
		return store.AnalysisResult{}, err
	}
	root := store.ResolveProjectRoot(projectPath)
	if root == "" {
		return store.AnalysisResult{}, fmt.Errorf("select a book or series first")
	}
	return s.GetAnalysisResult(root, analysisID)
}

func (a *App) ManuscriptChapterStickyContext(projectPath string, chapterRel string) (store.StickyChapterContext, error) {
	return store.BuildStickyChapterContext(projectPath, chapterRel)
}

func (a *App) DeleteAnalysisReport(id int64) error {
	s, err := a.ready()
	if err != nil {
		return err
	}
	return s.DeleteAnalysisReport(id)
}

func (a *App) StartAnalysisJob(id string, projectPath string) (cloud.JobStatus, error) {
	detail, ok := store.GetAnalysisDetail(id)
	if !ok {
		return cloud.JobStatus{}, fmt.Errorf("unknown analysis")
	}
	if !detail.UsesAI {
		return cloud.JobStatus{}, fmt.Errorf("not an AI analysis")
	}
	if _, err := a.ready(); err != nil {
		return cloud.JobStatus{}, err
	}
	root := store.ResolveProjectRoot(projectPath)
	if root == "" {
		return cloud.JobStatus{}, fmt.Errorf("select a book or series first")
	}
	src := store.MatchAnalysisSources(root)
	for _, need := range detail.Needs {
		role := src.Role(need)
		if !role.Present || len(role.Files) == 0 {
			return cloud.JobStatus{}, fmt.Errorf("missing source: %s", need)
		}
	}
	cli, err := a.cloudClient()
	if err != nil {
		return cloud.JobStatus{}, err
	}
	blobs := store.CollectNeededText(root, detail.Needs)
	roles := make([]cloud.RoleText, 0, len(blobs))
	for _, b := range blobs {
		roles = append(roles, cloud.RoleText{Role: b.Role, Rel: b.Rel, Name: b.Name, Text: b.Text})
	}
	st, err := cli.SubmitAnalysisJob(id, roles)
	if err != nil {
		return cloud.JobStatus{}, err
	}
	return st, nil
}

func (a *App) GetAnalysisJobStatus(jobID string) (cloud.JobStatus, error) {
	cli, err := a.cloudClient()
	if err != nil {
		return cloud.JobStatus{}, err
	}
	st, err := cli.GetAnalysisJob(jobID)
	if err != nil {
		return cloud.JobStatus{}, err
	}
	return st, nil
}

func (a *App) SaveAnalysisJobReport(id string, projectPath string, body string, original string, proposed string) (store.AnalysisReport, error) {
	detail, ok := store.GetAnalysisDetail(id)
	if !ok {
		return store.AnalysisReport{}, fmt.Errorf("unknown analysis")
	}
	s, err := a.ready()
	if err != nil {
		return store.AnalysisReport{}, err
	}
	root := store.ResolveProjectRoot(projectPath)
	if err := a.persistResult(detail, root, body, original, proposed, `{"kind":"saved","analysis_id":"`+detail.ID+`"}`); err != nil {
		return store.AnalysisReport{}, err
	}
	return s.SaveAnalysisReport(store.AnalysisReport{
		AnalysisID:    detail.ID,
		AnalysisLabel: detail.Label,
		ProjectPath:   root,
		UsesAI:        detail.UsesAI,
		Body:          body,
		Original:      original,
		Proposed:      proposed,
	})
}

// ExportMarkdownDocx opens a save dialog and writes a .docx from markdown (report header stripped).
func (a *App) ExportMarkdownDocx(markdown string, defaultName string) (string, error) {
	if a.ctx == nil {
		return "", fmt.Errorf("app not ready")
	}
	name := strings.TrimSpace(defaultName)
	if name == "" {
		name = "manuscript.docx"
	}
	if !strings.HasSuffix(strings.ToLower(name), ".docx") {
		name += ".docx"
	}
	path, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title:           "Export DOCX",
		DefaultFilename: name,
		Filters: []runtime.FileFilter{
			{DisplayName: "Word Document", Pattern: "*.docx"},
		},
	})
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(path) == "" {
		return "", nil
	}
	content := store.ReportContentOnly(markdown)
	raw, err := store.MarkdownToDocx(content)
	if err != nil {
		return "", err
	}
	if err := os.WriteFile(path, raw, 0o644); err != nil {
		return "", err
	}
	return path, nil
}

func (a *App) cloudClient() (*cloud.Client, error) {
	s, err := a.ready()
	if err != nil {
		return nil, err
	}
	base := cloud.DefaultBaseURL
	if got, err := s.GetAppSetting(cloud.AppSettingBaseURL); err == nil && strings.TrimSpace(got.Value) != "" {
		base = got.Value
	}
	return cloud.New(base, s.CloudToken()), nil
}

func (a *App) cloudBaseURL() (string, error) {
	s, err := a.ready()
	if err != nil {
		return "", err
	}
	base := cloud.DefaultBaseURL
	if got, err := s.GetAppSetting(cloud.AppSettingBaseURL); err == nil && strings.TrimSpace(got.Value) != "" {
		base = got.Value
	}
	return base, nil
}

func (a *App) GetCloudAccount() (cloud.Account, error) {
	cli, err := a.cloudClient()
	if err != nil {
		return cloud.Account{}, err
	}
	return cli.GetAccount()
}

func (a *App) ConnectCloudAccount(email string) (cloud.Account, error) {
	s, err := a.ready()
	if err != nil {
		return cloud.Account{}, err
	}
	base, err := a.cloudBaseURL()
	if err != nil {
		return cloud.Account{}, err
	}
	token, err := cloud.RegisterDevice(base, email)
	if err != nil {
		return cloud.Account{}, err
	}
	if _, err := s.PutSetting(cloud.SettingToken, token); err != nil {
		return cloud.Account{}, err
	}
	return cloud.New(base, token).GetAccount()
}

func (a *App) HasCloudAccount() bool {
	s, err := a.ready()
	if err != nil {
		return false
	}
	return s.CloudToken() != ""
}

func (a *App) OpenBillingCheckout() error {
	cli, err := a.cloudClient()
	if err != nil {
		return err
	}
	out, err := cli.Checkout()
	if err != nil {
		return err
	}
	if strings.TrimSpace(out.URL) == "" {
		return fmt.Errorf("no checkout URL")
	}
	runtime.BrowserOpenURL(a.ctx, out.URL)
	return nil
}

func (a *App) MatchAnalysisSources(projectPath string) store.AnalysisSources {
	return store.MatchAnalysisSources(projectPath)
}

func (a *App) DeleteDiskFile(dir string, name string) (store.DeletedResult, error) {
	if err := store.DeleteDiskFile(dir, name); err != nil {
		return store.DeletedResult{}, err
	}
	return store.DeletedResult{Deleted: true}, nil
}

func (a *App) ListWritingTree() (store.WritingTree, error) {
	root, err := a.writingRootIfSet()
	if err != nil {
		return store.ListWritingTree(""), nil
	}
	tree := store.ListWritingTree(root)
	s, storeErr := a.ready()
	if storeErr != nil {
		return tree, nil
	}
	vis, visErr := s.GetFolderVisibility()
	if visErr != nil {
		return tree, nil
	}
	return store.FilterWritingTree(tree, vis), nil
}

func (a *App) GetFolderVisibility() (store.FolderVisibility, error) {
	s, err := a.ready()
	if err != nil {
		return store.FolderVisibility{}, err
	}
	return s.GetFolderVisibility()
}

func (a *App) ListTemplateFolders() (store.TemplateFolderLists, error) {
	return store.TemplateFolderNames(), nil
}

func (a *App) SetHiddenFolderNames(in store.NameList) (store.FolderVisibility, error) {
	s, err := a.ready()
	if err != nil {
		return store.FolderVisibility{}, err
	}
	return s.SetHiddenFolderNames(in.Names)
}

func (a *App) SetShowHiddenFolders(on bool) (store.FolderVisibility, error) {
	s, err := a.ready()
	if err != nil {
		return store.FolderVisibility{}, err
	}
	return s.SetShowHiddenFolders(on)
}

func (a *App) SetFolderOverride(in store.FolderOverrideInput) (store.FolderVisibility, error) {
	s, err := a.ready()
	if err != nil {
		return store.FolderVisibility{}, err
	}
	return s.SetFolderOverride(in)
}

func (a *App) CreateWritingSeries(in store.WritingProjectInput) (store.PathResult, error) {
	root, err := a.ensureWritingRoot()
	if err != nil {
		return store.PathResult{}, err
	}
	path, err := store.CreateSeriesOnDisk(root, in.PenName, in.PenPath, in.Name)
	if err != nil {
		return store.PathResult{}, err
	}
	a.startFolderWatch()
	return store.PathResult{Path: path}, nil
}

func (a *App) CreateWritingBook(in store.WritingProjectInput) (store.PathResult, error) {
	root, err := a.ensureWritingRoot()
	if err != nil {
		return store.PathResult{}, err
	}
	section := "Act"
	if st, err := a.ready(); err == nil {
		section = st.GetDraftSection()
	}
	path, err := store.CreateBookOnDisk(root, in.PenName, in.PenPath, in.SeriesPath, in.Name, section)
	if err != nil {
		return store.PathResult{}, err
	}
	a.startFolderWatch()
	return store.PathResult{Path: path}, nil
}

func (a *App) GetDraftSection() (string, error) {
	s, err := a.ready()
	if err != nil {
		return "Act", err
	}
	return s.GetDraftSection(), nil
}

func (a *App) SetDraftSection(value string, rename bool) (store.RenameCount, error) {
	s, err := a.ready()
	if err != nil {
		return store.RenameCount{}, err
	}
	section, err := s.SetDraftSection(value)
	if err != nil {
		return store.RenameCount{}, err
	}
	out := store.RenameCount{Section: section}
	if !rename {
		return out, nil
	}
	root, err := a.writingRootIfSet()
	if err != nil {
		return out, nil
	}
	n, err := store.RenameDraftSections(root, section)
	if err != nil {
		return out, err
	}
	out.Renamed = n
	a.startFolderWatch()
	a.emitFoldersChanged("")
	return out, nil
}

func (a *App) RenameWritingProject(path string, name string) (store.PathResult, error) {
	next, err := store.RenameProjectDir(path, name)
	if err != nil {
		return store.PathResult{}, err
	}
	a.startFolderWatch()
	return store.PathResult{Path: next}, nil
}

func (a *App) DeleteWritingProject(path string) (store.DeletedResult, error) {
	if err := store.DeleteProjectDir(path); err != nil {
		return store.DeletedResult{}, err
	}
	a.startFolderWatch()
	return store.DeletedResult{Deleted: true}, nil
}

func (a *App) ListHeaderFiles(projectPath, projectKind, kind string) (store.HeaderList, error) {
	s, err := a.ready()
	if err != nil {
		return store.HeaderList{}, err
	}
	return s.ListHeaderFiles(projectPath, projectKind, kind)
}

func (a *App) ReadHeaderFile(in store.HeaderFileRef) (store.HeaderFileContent, error) {
	s, err := a.ready()
	if err != nil {
		return store.HeaderFileContent{}, err
	}
	return s.ReadHeaderFile(in)
}

func (a *App) WriteHeaderFile(in store.HeaderFileWrite) (store.HeaderFile, error) {
	s, err := a.ready()
	if err != nil {
		return store.HeaderFile{}, err
	}
	return s.WriteHeaderFile(in)
}

func (a *App) CreateHeaderFile(in store.HeaderFileWrite) (store.HeaderFile, error) {
	s, err := a.ready()
	if err != nil {
		return store.HeaderFile{}, err
	}
	out, err := s.CreateHeaderFile(in)
	if err != nil {
		return store.HeaderFile{}, err
	}
	a.startFolderWatch()
	return out, nil
}

func (a *App) DeleteHeaderFile(in store.HeaderFileRef) (store.DeletedResult, error) {
	s, err := a.ready()
	if err != nil {
		return store.DeletedResult{}, err
	}
	if err := s.DeleteHeaderFile(in); err != nil {
		return store.DeletedResult{}, err
	}
	return store.DeletedResult{Deleted: true}, nil
}

func (a *App) PlaceHeaderFiles(in store.HeaderPlaceInput) (store.UpdatedResult, error) {
	s, err := a.ready()
	if err != nil {
		return store.UpdatedResult{}, err
	}
	if err := s.PlaceHeaderFiles(in); err != nil {
		return store.UpdatedResult{}, err
	}
	return store.UpdatedResult{Updated: true}, nil
}

func (a *App) ListHeaderOverrides(projectPath string) ([]store.FolderLink, error) {
	s, err := a.ready()
	if err != nil {
		return nil, err
	}
	return s.ListHeaderOverrides(projectPath)
}

func (a *App) SetHeaderOverride(projectPath, kind, path string) (store.FolderLink, error) {
	s, err := a.ready()
	if err != nil {
		return store.FolderLink{}, err
	}
	link, err := s.SetHeaderOverride(projectPath, kind, path)
	if err != nil {
		return store.FolderLink{}, err
	}
	a.startFolderWatch()
	return link, nil
}

func (a *App) ClearHeaderOverride(projectPath, kind string) (store.DeletedResult, error) {
	s, err := a.ready()
	if err != nil {
		return store.DeletedResult{}, err
	}
	if err := s.ClearHeaderOverride(projectPath, kind); err != nil {
		return store.DeletedResult{}, err
	}
	a.startFolderWatch()
	return store.DeletedResult{Deleted: true}, nil
}

func (a *App) ListPens() ([]store.Pen, error) {
	s, err := a.ready()
	if err != nil {
		return nil, err
	}
	root, _ := a.writingRootIfSet()
	series, err := s.ListSeries()
	if err != nil {
		return nil, err
	}
	stories, err := s.ListStories(0)
	if err != nil {
		return nil, err
	}
	return store.ListPens(root, series, stories), nil
}

func (a *App) CreatePen(name string) (store.Pen, error) {
	if strings.TrimSpace(name) == "" {
		return store.Pen{}, fmt.Errorf("pen name is required")
	}
	path, err := a.penDir(name)
	if err != nil {
		return store.Pen{}, err
	}
	penName := filepath.Base(path)
	if s, err := a.ready(); err == nil {
		_, _ = s.SetLastPen(penName)
	}
	return store.Pen{Name: penName}, nil
}

func (a *App) penDir(penName string) (string, error) {
	root, err := a.ensureWritingRoot()
	if err != nil {
		return "", err
	}
	return store.EnsurePenDir(root, penName)
}

func (a *App) writingRootIfSet() (string, error) {
	s, err := a.ready()
	if err != nil {
		return "", err
	}
	got, err := s.GetSetting("writing_root")
	if err != nil || strings.TrimSpace(got.Value) == "" {
		return "", fmt.Errorf("writing folder is not set")
	}
	return got.Value, nil
}

func (a *App) storyTemplateParent(s *store.Store, seriesID int64, penName string) (parent, seriesName string, err error) {
	if seriesID == 0 {
		parent, err = a.penDir(penName)
		return parent, "", err
	}
	seriesName = "Series"
	list, listErr := s.ListSeries()
	if listErr == nil {
		for _, se := range list {
			if se.ID == seriesID {
				seriesName = se.Name
				break
			}
		}
	}
	root := s.RootPath("series", seriesID)
	if root == "" {
		parent, err = a.ensureWritingRoot()
		return parent, seriesName, err
	}
	parent = filepath.Join(root, "Books")
	if err := os.MkdirAll(parent, 0o755); err != nil {
		return "", "", err
	}
	return parent, seriesName, nil
}

func (a *App) ensureWritingRoot() (string, error) {
	s, err := a.ready()
	if err != nil {
		return "", err
	}
	got, err := s.GetSetting("writing_root")
	if err == nil && strings.TrimSpace(got.Value) != "" {
		info, statErr := os.Stat(got.Value)
		if statErr == nil && info.IsDir() {
			return got.Value, nil
		}
	}
	if a.ctx == nil {
		return "", fmt.Errorf("choose a writing folder in Settings")
	}
	path, err := runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Choose folder for series and book files",
	})
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(path) == "" {
		return "", fmt.Errorf("a writing folder is required")
	}
	if _, err := s.PutSetting("writing_root", path); err != nil {
		return "", err
	}
	return path, nil
}

func (a *App) UpdateStory(id int64, in store.StoryInput) (store.UpdatedResult, error) {
	s, err := a.ready()
	if err != nil {
		return store.UpdatedResult{}, err
	}
	return s.UpdateStory(id, in)
}

func (a *App) DeleteStory(id int64) (store.DeletedResult, error) {
	s, err := a.ready()
	if err != nil {
		return store.DeletedResult{}, err
	}
	return s.DeleteStory(id)
}

func (a *App) ListChapters(storyID int64) ([]store.Chapter, error) {
	s, err := a.ready()
	if err != nil {
		return nil, err
	}
	return s.ListChapters(storyID)
}

func (a *App) GetChapter(id int64) (store.Chapter, error) {
	s, err := a.ready()
	if err != nil {
		return store.Chapter{}, err
	}
	return s.GetChapter(id)
}

func (a *App) CreateChapter(storyID int64, in store.ChapterInput) (store.IDResult, error) {
	s, err := a.ready()
	if err != nil {
		return store.IDResult{}, err
	}
	return s.CreateChapter(storyID, in)
}

func (a *App) UpdateChapter(id int64, in store.ChapterInput) (store.UpdatedResult, error) {
	s, err := a.ready()
	if err != nil {
		return store.UpdatedResult{}, err
	}
	return s.UpdateChapter(id, in)
}

func (a *App) DeleteChapter(id int64) (store.DeletedResult, error) {
	s, err := a.ready()
	if err != nil {
		return store.DeletedResult{}, err
	}
	return s.DeleteChapter(id)
}

func (a *App) ListStoryDocs(storyID int64) ([]store.StoryDoc, error) {
	s, err := a.ready()
	if err != nil {
		return nil, err
	}
	return s.ListStoryDocs(storyID)
}

func (a *App) CreateStoryDoc(storyID int64, in store.DocInput) (store.IDResult, error) {
	s, err := a.ready()
	if err != nil {
		return store.IDResult{}, err
	}
	return s.CreateStoryDoc(storyID, in)
}

func (a *App) ListCharacters(storyID int64) ([]store.CharacterProfile, error) {
	s, err := a.ready()
	if err != nil {
		return nil, err
	}
	return s.ListCharacters(storyID)
}

func (a *App) GetCharacter(id int64) (store.CharacterProfile, error) {
	s, err := a.ready()
	if err != nil {
		return store.CharacterProfile{}, err
	}
	return s.GetCharacter(id)
}

func (a *App) ListDocumentTypes() ([]store.DocumentType, error) {
	s, err := a.ready()
	if err != nil {
		return nil, err
	}
	return s.ListDocumentTypes()
}

func (a *App) ListSeriesDocs(seriesID int64) ([]store.SeriesDoc, error) {
	s, err := a.ready()
	if err != nil {
		return nil, err
	}
	return s.ListSeriesDocs(seriesID)
}

func (a *App) GetSeriesDoc(id int64) (store.BibleDoc, error) {
	s, err := a.ready()
	if err != nil {
		return store.BibleDoc{}, err
	}
	return s.GetSeriesDoc(id)
}

func (a *App) CreateSeriesDoc(seriesID int64, in store.SeriesDocInput) (store.IDResult, error) {
	s, err := a.ready()
	if err != nil {
		return store.IDResult{}, err
	}
	return s.CreateSeriesDoc(seriesID, in)
}

func (a *App) UpdateSeriesDoc(id int64, in store.SeriesDocInput) (store.UpdatedResult, error) {
	s, err := a.ready()
	if err != nil {
		return store.UpdatedResult{}, err
	}
	return s.UpdateSeriesDoc(id, in)
}

func (a *App) DeleteSeriesDoc(id int64) (store.DeletedResult, error) {
	s, err := a.ready()
	if err != nil {
		return store.DeletedResult{}, err
	}
	return s.DeleteSeriesDoc(id)
}

func (a *App) ListActs(storyID int64) ([]store.StoryAct, error) {
	s, err := a.ready()
	if err != nil {
		return nil, err
	}
	return s.ListActs(storyID)
}

func (a *App) GetAct(id int64) (store.StoryAct, error) {
	s, err := a.ready()
	if err != nil {
		return store.StoryAct{}, err
	}
	return s.GetAct(id)
}

func (a *App) CreateAct(storyID int64, in store.ActInput) (store.IDResult, error) {
	s, err := a.ready()
	if err != nil {
		return store.IDResult{}, err
	}
	return s.CreateAct(storyID, in)
}

func (a *App) UpdateAct(id int64, in store.ActInput) (store.UpdatedResult, error) {
	s, err := a.ready()
	if err != nil {
		return store.UpdatedResult{}, err
	}
	return s.UpdateAct(id, in)
}

func (a *App) DeleteAct(id int64) (store.DeletedResult, error) {
	s, err := a.ready()
	if err != nil {
		return store.DeletedResult{}, err
	}
	return s.DeleteAct(id)
}

func (a *App) ListSeriesCharacters(seriesID int64) ([]store.CharacterProfile, error) {
	s, err := a.ready()
	if err != nil {
		return nil, err
	}
	return s.ListSeriesCharacters(seriesID)
}

func (a *App) CreateCharacter(in store.CharacterInput) (store.IDResult, error) {
	s, err := a.ready()
	if err != nil {
		return store.IDResult{}, err
	}
	return s.CreateCharacter(in)
}

func (a *App) UpdateCharacter(id int64, in store.CharacterInput) (store.UpdatedResult, error) {
	s, err := a.ready()
	if err != nil {
		return store.UpdatedResult{}, err
	}
	return s.UpdateCharacter(id, in)
}

func (a *App) DeleteCharacter(id int64) (store.DeletedResult, error) {
	s, err := a.ready()
	if err != nil {
		return store.DeletedResult{}, err
	}
	return s.DeleteCharacter(id)
}

func (a *App) GetSetting(key string) (store.SettingValue, error) {
	s, err := a.ready()
	if err != nil {
		return store.SettingValue{}, err
	}
	return s.GetSetting(key)
}

func (a *App) PutSetting(key string, value string) (store.SettingValue, error) {
	s, err := a.ready()
	if err != nil {
		return store.SettingValue{}, err
	}
	out, err := s.PutSetting(key, value)
	if err == nil && key == "writing_root" {
		a.startFolderWatch()
		a.emitFoldersChanged(value)
	}
	return out, err
}

func (a *App) ReadTextFile(path string) (string, error) {
	return readTextFile(path)
}

func (a *App) PickImportFolder() (string, error) {
	if a.ctx == nil {
		return "", fmt.Errorf("app not ready")
	}
	return runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Choose folder",
	})
}

func (a *App) ListFolderLinks(scope string, ownerID int64) ([]store.FolderLink, error) {
	s, err := a.ready()
	if err != nil {
		return nil, err
	}
	return s.ListFolderLinks(scope, ownerID)
}

func (a *App) SetFolderLink(in store.FolderLinkInput) (store.FolderLink, error) {
	s, err := a.ready()
	if err != nil {
		return store.FolderLink{}, err
	}
	link, err := s.SetFolderLink(in)
	if err != nil {
		return store.FolderLink{}, err
	}
	a.startFolderWatch()
	a.emitFolderSync(nil)
	return link, nil
}

func (a *App) ClearFolderLink(scope string, ownerID int64, kind string) (store.DeletedResult, error) {
	s, err := a.ready()
	if err != nil {
		return store.DeletedResult{}, err
	}
	out, err := s.ClearFolderLink(scope, ownerID, kind)
	if err != nil {
		return store.DeletedResult{}, err
	}
	a.startFolderWatch()
	return out, nil
}

func (a *App) SyncAllFolders() ([]store.FolderChange, error) {
	s, err := a.ready()
	if err != nil {
		return nil, err
	}
	changes, err := s.SyncAllFolders()
	if err != nil {
		return nil, err
	}
	a.emitFolderSync(changes)
	return changes, nil
}

func (a *App) emitFolderSync(changes []store.FolderChange) {
	a.emitFoldersChanged("")
	_ = changes
}

func (a *App) ReadImportFiles(paths []string) ([]ImportedFile, error) {
	out := make([]ImportedFile, 0)
	for _, path := range paths {
		if strings.TrimSpace(path) == "" {
			continue
		}
		files, err := collectImportFiles(path)
		if err != nil {
			return nil, err
		}
		out = append(out, files...)
	}
	return out, nil
}

func (a *App) ImportDiskFiles(dir string, paths []string) (store.DiskFile, error) {
	files, err := a.ReadImportFiles(paths)
	if err != nil {
		return store.DiskFile{}, err
	}
	incoming := make([]store.IncomingFile, 0, len(files))
	for _, f := range files {
		incoming = append(incoming, store.IncomingFile{Name: f.Name, Text: f.Text})
	}
	return a.CreateDiskFiles(dir, incoming)
}

func collectImportFiles(root string) ([]ImportedFile, error) {
	info, err := os.Stat(root)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		if !importableName(info.Name()) {
			return nil, nil
		}
		f, err := readImported(root)
		if err != nil {
			return nil, err
		}
		f.Rel = filepath.Base(root)
		return []ImportedFile{f}, nil
	}
	var out []ImportedFile
	err = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		name := d.Name()
		if skipImportName(name) && path != root {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if d.IsDir() {
			return nil
		}
		if !importableName(name) {
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		f, err := readImported(path)
		if err != nil {
			return err
		}
		f.Rel = filepath.ToSlash(filepath.Join(filepath.Base(root), rel))
		out = append(out, f)
		return nil
	})
	return out, err
}

func skipImportName(name string) bool {
	if name == "." || name == ".." {
		return false
	}
	if strings.HasPrefix(name, ".") {
		return true
	}
	switch strings.ToLower(name) {
	case "node_modules", "__pycache__", "dist", "build":
		return true
	default:
		return false
	}
}

func importableName(name string) bool {
	return store.ImportableFileName(name)
}

func readImported(path string) (ImportedFile, error) {
	text, err := readTextFile(path)
	if err != nil {
		return ImportedFile{}, err
	}
	return ImportedFile{Name: filepath.Base(path), Text: text, Rel: filepath.Base(path)}, nil
}

func readTextFile(path string) (string, error) {
	info, err := os.Stat(path)
	if err != nil {
		return "", err
	}
	if !info.Mode().IsRegular() {
		return "", fmt.Errorf("not a file")
	}
	if info.Size() > 8<<20 {
		return "", fmt.Errorf("file too large")
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func (a *App) PlaceStoryDoc(id int64, in store.PlaceInput) (store.UpdatedResult, error) {
	s, err := a.ready()
	if err != nil {
		return store.UpdatedResult{}, err
	}
	return s.PlaceStoryDoc(id, in)
}

func (a *App) PlaceSeriesDoc(id int64, in store.PlaceInput) (store.UpdatedResult, error) {
	s, err := a.ready()
	if err != nil {
		return store.UpdatedResult{}, err
	}
	return s.PlaceSeriesDoc(id, in)
}

func (a *App) ReorderActs(storyID int64, in store.IDList) (store.UpdatedResult, error) {
	s, err := a.ready()
	if err != nil {
		return store.UpdatedResult{}, err
	}
	return s.ReorderActs(storyID, in)
}

func (a *App) ReorderStories(seriesID int64, in store.IDList) (store.UpdatedResult, error) {
	s, err := a.ready()
	if err != nil {
		return store.UpdatedResult{}, err
	}
	return s.ReorderStories(seriesID, in)
}

func (a *App) AdminListTables() (store.AdminTables, error) {
	s, err := a.ready()
	if err != nil {
		return store.AdminTables{}, err
	}
	return s.AdminListTables()
}

func (a *App) AdminTableSchema(table string) (store.AdminSchema, error) {
	s, err := a.ready()
	if err != nil {
		return store.AdminSchema{}, err
	}
	return s.AdminTableSchema(table)
}

func (a *App) AdminQueryTable(table string, limit int, offset int) (store.AdminTableRows, error) {
	s, err := a.ready()
	if err != nil {
		return store.AdminTableRows{}, err
	}
	return s.AdminQueryTable(table, limit, offset)
}

func (a *App) AdminDeleteRow(table string, id int64) (store.DeletedResult, error) {
	s, err := a.ready()
	if err != nil {
		return store.DeletedResult{}, err
	}
	return s.AdminDeleteRow(table, id)
}

func (a *App) AdminExecSQL(query string) (store.AdminSQLResult, error) {
	s, err := a.ready()
	if err != nil {
		return store.AdminSQLResult{}, err
	}
	return s.AdminExecSQL(query)
}
