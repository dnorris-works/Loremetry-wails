package store

import "database/sql"

type Store struct {
	DB     *sql.DB
	UserID int64
}

type IDResult struct {
	ID int64 `json:"id"`
}

type PathResult struct {
	Path string `json:"path"`
}

type RenameCount struct {
	Section string `json:"section"`
	Renamed int    `json:"renamed"`
}

type UpdatedResult struct {
	Updated bool `json:"updated"`
}

type DeletedResult struct {
	Deleted bool `json:"deleted"`
}

type AuthSession struct {
	Authenticated bool   `json:"authenticated"`
	ID            string `json:"id,omitempty"`
	Email         string `json:"email,omitempty"`
	FirstName     string `json:"firstName,omitempty"`
	LastName      string `json:"lastName,omitempty"`
	IsAdmin       bool   `json:"isAdmin,omitempty"`
	BreakGlass    bool   `json:"breakGlass,omitempty"`
	Reason        string `json:"reason,omitempty"`
}

type Series struct {
	ID          int64   `json:"id"`
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
	Summary     *string `json:"summary,omitempty"`
	PenName     string  `json:"pen_name"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
}

type SeriesInput struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Summary     string `json:"summary"`
	PenName     string `json:"pen_name"`
}

type Story struct {
	ID              int64   `json:"id"`
	SeriesID        *int64  `json:"series_id,omitempty"`
	SeriesSortOrder *int    `json:"series_sort_order,omitempty"`
	Name            string  `json:"name"`
	Description     *string `json:"description,omitempty"`
	PenName         string  `json:"pen_name"`
	CreatedAt       string  `json:"created_at"`
	UpdatedAt       string  `json:"updated_at"`
}

type StoryInput struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	SeriesID    int64  `json:"series_id"`
	PenName     string `json:"pen_name"`
}

type Pen struct {
	Name string `json:"name"`
}

type WritingProjectInput struct {
	PenName    string `json:"pen_name"`
	PenPath    string `json:"pen_path"`
	Name       string `json:"name"`
	SeriesPath string `json:"series_path"`
}

type Chapter struct {
	ID          int64   `json:"id"`
	StoryID     int64   `json:"story_id"`
	FileName    string  `json:"file_name"`
	TextContent *string `json:"text_content,omitempty"`
	SortOrder   *int    `json:"sort_order,omitempty"`
	ActID       *int64  `json:"act_id,omitempty"`
	CreatedAt   string  `json:"created_at"`
}

type StoryDoc struct {
	ID        int64  `json:"id"`
	StoryID   int64  `json:"story_id"`
	Kind      string `json:"kind"`
	FileName  string `json:"file_name"`
	SortOrder *int   `json:"sort_order,omitempty"`
	ActID     *int64 `json:"act_id,omitempty"`
	CreatedAt string `json:"created_at"`
}

type ChapterInput struct {
	FileName    *string `json:"file_name"`
	TextContent *string `json:"text_content"`
	SortOrder   *int    `json:"sort_order"`
	ActID       *int64  `json:"act_id"`
}

type DocInput struct {
	Kind        string `json:"kind"`
	FileName    string `json:"file_name"`
	TextContent string `json:"text_content"`
}

type SeriesDoc struct {
	ID        int64  `json:"id"`
	SeriesID  int64  `json:"series_id"`
	Category  string `json:"category"`
	FileName  string `json:"file_name"`
	CreatedAt string `json:"created_at"`
}

type SeriesDocInput struct {
	Category    string `json:"category"`
	FileName    string `json:"file_name"`
	TextContent string `json:"text_content"`
}

type BibleDoc struct {
	ID          int64  `json:"id"`
	SeriesID    int64  `json:"series_id"`
	FileName    string `json:"file_name"`
	TextContent string `json:"text_content"`
}

type StoryAct struct {
	ID        int64   `json:"id"`
	StoryID   int64   `json:"story_id"`
	Title     string  `json:"title"`
	Summary   *string `json:"summary,omitempty"`
	SortOrder *int    `json:"sort_order,omitempty"`
	CreatedAt string  `json:"created_at"`
	UpdatedAt string  `json:"updated_at"`
}

type ActInput struct {
	Title   string `json:"title"`
	Summary string `json:"summary"`
}

type DocumentType struct {
	Code            string `json:"code"`
	DisplayName     string `json:"display_name"`
	SortOrder       int    `json:"sort_order"`
	AppliesToStory  bool   `json:"applies_to_story"`
	AppliesToSeries bool   `json:"applies_to_series"`
	Active          bool   `json:"active"`
	Editor          string `json:"editor"`
	Storage         string `json:"storage"`
}

type CharacterProfile struct {
	ID                  int64    `json:"id"`
	SeriesID            *int64   `json:"series_id,omitempty"`
	StoryID             *int64   `json:"story_id,omitempty"`
	CharacterName       string   `json:"character_name"`
	RoleInStory         *string  `json:"role_in_story,omitempty"`
	StoryImportance     *string  `json:"story_importance,omitempty"`
	POVCharacter        bool     `json:"pov_character"`
	CharacterWant       *string  `json:"character_want,omitempty"`
	CharacterNeed       *string  `json:"character_need,omitempty"`
	ArcStatusAtStart    *string  `json:"arc_status_at_start,omitempty"`
	ArcStatusAtEnd      *string  `json:"arc_status_at_end,omitempty"`
	KeyRelationships    []string `json:"key_relationships,omitempty"`
	DefiningTrait       *string  `json:"defining_trait,omitempty"`
	FirstAppearanceBook *int     `json:"first_appearance_book,omitempty"`
	PrimaryObstacle     *string  `json:"primary_obstacle,omitempty"`
	InternalVsExternal  *string  `json:"internal_vs_external,omitempty"`
	CoreWound           *string  `json:"core_wound,omitempty"`
	Age                 *int     `json:"age,omitempty"`
	ThematicResonance   *string  `json:"thematic_resonance,omitempty"`
	CharacterVoiceNotes *string  `json:"character_voice_notes,omitempty"`
	CreatedAt           string   `json:"created_at"`
	UpdatedAt           string   `json:"updated_at"`
}

type CharacterInput struct {
	SeriesID            int64    `json:"series_id"`
	StoryID             int64    `json:"story_id"`
	CharacterName       string   `json:"character_name"`
	RoleInStory         string   `json:"role_in_story"`
	StoryImportance     string   `json:"story_importance"`
	POVCharacter        bool     `json:"pov_character"`
	CharacterWant       string   `json:"character_want"`
	CharacterNeed       string   `json:"character_need"`
	ArcStatusAtStart    string   `json:"arc_status_at_start"`
	ArcStatusAtEnd      string   `json:"arc_status_at_end"`
	KeyRelationships    []string `json:"key_relationships"`
	DefiningTrait       string   `json:"defining_trait"`
	FirstAppearanceBook int      `json:"first_appearance_book"`
	PrimaryObstacle     string   `json:"primary_obstacle"`
	InternalVsExternal  string   `json:"internal_vs_external"`
	CoreWound           string   `json:"core_wound"`
	Age                 int      `json:"age"`
	ThematicResonance   string   `json:"thematic_resonance"`
	CharacterVoiceNotes string   `json:"character_voice_notes"`
}

type FolderLink struct {
	Scope   string `json:"scope"`
	OwnerID int64  `json:"owner_id"`
	Kind    string `json:"kind"`
	Path    string `json:"path"`
}

type FolderLinkInput struct {
	Scope   string `json:"scope"`
	OwnerID int64  `json:"owner_id"`
	Kind    string `json:"kind"`
	Path    string `json:"path"`
}

type FolderChange struct {
	Type string `json:"type"`
	ID   int64  `json:"id"`
}

type SettingValue struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type AdminColumn struct {
	Name     string  `json:"name"`
	Type     string  `json:"type"`
	Nullable string  `json:"nullable"`
	Default  *string `json:"default"`
}

type AdminTables struct {
	Tables []string `json:"tables"`
}

type AdminSchema struct {
	Columns []AdminColumn `json:"columns"`
}

type AdminTableRows struct {
	Columns []string   `json:"columns"`
	Rows    [][]string `json:"rows"`
	Total   int        `json:"total"`
}

type AdminSQLResult struct {
	Columns      []string   `json:"columns"`
	Rows         [][]string `json:"rows"`
	RowsAffected int64      `json:"rows_affected"`
	Message      string     `json:"message"`
}

func nullStr(s *string) any {
	if s == nil {
		return nil
	}
	return *s
}

func nullInt(n *int) any {
	if n == nil {
		return nil
	}
	return *n
}

func emptyToNil(s string) any {
	if s == "" {
		return nil
	}
	return s
}

func nullInt64(n *int64) any {
	if n == nil {
		return nil
	}
	return *n
}

func zeroToNilInt64(n int64) any {
	if n == 0 {
		return nil
	}
	return n
}

func strPtr(ns sql.NullString) *string {
	if !ns.Valid {
		return nil
	}
	v := ns.String
	return &v
}

func intPtr(ni sql.NullInt64) *int {
	if !ni.Valid {
		return nil
	}
	v := int(ni.Int64)
	return &v
}

func int64Ptr(ni sql.NullInt64) *int64 {
	if !ni.Valid {
		return nil
	}
	v := ni.Int64
	return &v
}
