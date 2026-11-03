package store

import (
	"database/sql"
	"encoding/json"
	"fmt"
)

func (s *Store) ListCharacters(storyID int64) ([]CharacterProfile, error) {
	if err := s.ownsStory(storyID); err != nil {
		return nil, err
	}
	return s.listCharacters(`cp.story_id = ?`, storyID)
}

func (s *Store) ListSeriesCharacters(seriesID int64) ([]CharacterProfile, error) {
	if err := s.ownsSeries(seriesID); err != nil {
		return nil, err
	}
	return s.listCharacters(`cp.series_id = ?`, seriesID)
}

func (s *Store) listCharacters(where string, arg int64) ([]CharacterProfile, error) {
	rows, err := s.DB.Query(`
		SELECT id, series_id, story_id, character_name, role_in_story, story_importance, pov_character,
		       character_want, character_need, arc_status_at_start, arc_status_at_end,
		       key_relationships, defining_trait, first_appearance_book,
		       primary_obstacle, internal_vs_external, core_wound,
		       age, thematic_resonance, character_voice_notes,
		       created_at, updated_at
		FROM character_profiles cp WHERE `+where+` ORDER BY character_name`, arg)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]CharacterProfile, 0)
	for rows.Next() {
		cp, err := scanCharacter(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, cp)
	}
	return out, rows.Err()
}

func (s *Store) GetCharacter(id int64) (CharacterProfile, error) {
	row := s.DB.QueryRow(`
		SELECT cp.id, cp.series_id, cp.story_id, cp.character_name, cp.role_in_story, cp.story_importance, cp.pov_character,
		       cp.character_want, cp.character_need, cp.arc_status_at_start, cp.arc_status_at_end,
		       cp.key_relationships, cp.defining_trait, cp.first_appearance_book,
		       cp.primary_obstacle, cp.internal_vs_external, cp.core_wound,
		       cp.age, cp.thematic_resonance, cp.character_voice_notes,
		       cp.created_at, cp.updated_at
		FROM character_profiles cp
		LEFT JOIN stories st ON st.id = cp.story_id
		LEFT JOIN series se ON se.id = cp.series_id
		WHERE cp.id = ? AND (st.user_id = ? OR se.user_id = ?)`, id, s.UserID, s.UserID)
	cp, err := scanCharacter(row)
	if err != nil {
		return CharacterProfile{}, fmt.Errorf("character not found")
	}
	return cp, nil
}

func (s *Store) CreateCharacter(in CharacterInput) (IDResult, error) {
	if in.CharacterName == "" {
		return IDResult{}, fmt.Errorf("character_name is required")
	}
	if (in.SeriesID == 0) == (in.StoryID == 0) {
		return IDResult{}, fmt.Errorf("set series_id or story_id")
	}
	if in.StoryID != 0 {
		if err := s.ownsStory(in.StoryID); err != nil {
			return IDResult{}, err
		}
	} else if err := s.ownsSeries(in.SeriesID); err != nil {
		return IDResult{}, err
	}
	rels, _ := json.Marshal(in.KeyRelationships)
	if in.KeyRelationships == nil {
		rels = []byte("[]")
	}
	res, err := s.DB.Exec(`
		INSERT INTO character_profiles (
			series_id, story_id, character_name, role_in_story, story_importance, pov_character,
			character_want, character_need, arc_status_at_start, arc_status_at_end,
			key_relationships, defining_trait, first_appearance_book,
			primary_obstacle, internal_vs_external, core_wound,
			age, thematic_resonance, character_voice_notes, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, datetime('now'), datetime('now'))`,
		zeroToNilInt64(in.SeriesID), zeroToNilInt64(in.StoryID), in.CharacterName, emptyToNil(in.RoleInStory), emptyToNil(in.StoryImportance), boolToInt(in.POVCharacter),
		emptyToNil(in.CharacterWant), emptyToNil(in.CharacterNeed), emptyToNil(in.ArcStatusAtStart), emptyToNil(in.ArcStatusAtEnd),
		string(rels), emptyToNil(in.DefiningTrait), zeroToNil(in.FirstAppearanceBook),
		emptyToNil(in.PrimaryObstacle), emptyToNil(in.InternalVsExternal), emptyToNil(in.CoreWound),
		zeroToNil(in.Age), emptyToNil(in.ThematicResonance), emptyToNil(in.CharacterVoiceNotes),
	)
	if err != nil {
		return IDResult{}, err
	}
	id, err := res.LastInsertId()
	if err == nil {
		scope, owner := characterScope(in.SeriesID, in.StoryID)
		s.mirrorWrite(scope, owner, "character", "", in.CharacterName, in.CharacterVoiceNotes)
	}
	return IDResult{ID: id}, err
}

func (s *Store) UpdateCharacter(id int64, in CharacterInput) (UpdatedResult, error) {
	var seriesID, storyID sql.NullInt64
	var oldName string
	var oldNotes sql.NullString
	_ = s.DB.QueryRow(`SELECT series_id, story_id, character_name, character_voice_notes FROM character_profiles WHERE id = ?`, id).
		Scan(&seriesID, &storyID, &oldName, &oldNotes)
	rels, _ := json.Marshal(in.KeyRelationships)
	if in.KeyRelationships == nil {
		rels = nil
	}
	res, err := s.DB.Exec(`
		UPDATE character_profiles SET
			character_name = CASE WHEN ? != '' THEN ? ELSE character_name END,
			role_in_story = COALESCE(?, role_in_story),
			story_importance = COALESCE(?, story_importance),
			pov_character = ?,
			character_want = COALESCE(?, character_want),
			character_need = COALESCE(?, character_need),
			arc_status_at_start = COALESCE(?, arc_status_at_start),
			arc_status_at_end = COALESCE(?, arc_status_at_end),
			key_relationships = COALESCE(?, key_relationships),
			defining_trait = COALESCE(?, defining_trait),
			first_appearance_book = COALESCE(?, first_appearance_book),
			primary_obstacle = COALESCE(?, primary_obstacle),
			internal_vs_external = COALESCE(?, internal_vs_external),
			core_wound = COALESCE(?, core_wound),
			age = COALESCE(?, age),
			thematic_resonance = COALESCE(?, thematic_resonance),
			character_voice_notes = COALESCE(?, character_voice_notes),
			updated_at = datetime('now')
		WHERE id = ? AND (
			story_id IN (SELECT id FROM stories WHERE user_id = ?)
			OR series_id IN (SELECT id FROM series WHERE user_id = ?)
		)`,
		in.CharacterName, in.CharacterName,
		emptyToNil(in.RoleInStory), emptyToNil(in.StoryImportance), boolToInt(in.POVCharacter),
		emptyToNil(in.CharacterWant), emptyToNil(in.CharacterNeed), emptyToNil(in.ArcStatusAtStart), emptyToNil(in.ArcStatusAtEnd),
		nullableJSON(rels), emptyToNil(in.DefiningTrait), zeroToNil(in.FirstAppearanceBook),
		emptyToNil(in.PrimaryObstacle), emptyToNil(in.InternalVsExternal), emptyToNil(in.CoreWound),
		zeroToNil(in.Age), emptyToNil(in.ThematicResonance), emptyToNil(in.CharacterVoiceNotes),
		id, s.UserID, s.UserID,
	)
	if err != nil {
		return UpdatedResult{}, err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return UpdatedResult{}, err
	}
	if n == 0 {
		return UpdatedResult{}, fmt.Errorf("character not found")
	}
	name := in.CharacterName
	if name == "" {
		name = oldName
	}
	notes := in.CharacterVoiceNotes
	if notes == "" && oldNotes.Valid {
		notes = oldNotes.String
	}
	scope, owner := characterScope(nullInt64Val(seriesID), nullInt64Val(storyID))
	s.mirrorWrite(scope, owner, "character", oldName, name, notes)
	return UpdatedResult{Updated: true}, nil
}

func (s *Store) DeleteCharacter(id int64) (DeletedResult, error) {
	var seriesID, storyID sql.NullInt64
	var name string
	_ = s.DB.QueryRow(`SELECT series_id, story_id, character_name FROM character_profiles WHERE id = ?`, id).
		Scan(&seriesID, &storyID, &name)
	res, err := s.DB.Exec(`
		DELETE FROM character_profiles
		WHERE id = ? AND (
			story_id IN (SELECT id FROM stories WHERE user_id = ?)
			OR series_id IN (SELECT id FROM series WHERE user_id = ?)
		)`, id, s.UserID, s.UserID)
	if err != nil {
		return DeletedResult{}, err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return DeletedResult{}, err
	}
	if n == 0 {
		return DeletedResult{}, fmt.Errorf("character not found")
	}
	scope, owner := characterScope(nullInt64Val(seriesID), nullInt64Val(storyID))
	s.mirrorRemove(scope, owner, "character", name)
	return DeletedResult{Deleted: true}, nil
}

func characterScope(seriesID, storyID int64) (string, int64) {
	if storyID != 0 {
		return "story", storyID
	}
	return "series", seriesID
}

func nullInt64Val(n sql.NullInt64) int64 {
	if !n.Valid {
		return 0
	}
	return n.Int64
}

func scanCharacter(row scanner) (CharacterProfile, error) {
	var cp CharacterProfile
	var seriesID, storyID sql.NullInt64
	var role, imp, want, need, start, end, rels, trait, obst, ive, wound, theme, voice sql.NullString
	var book, age sql.NullInt64
	var pov int
	err := row.Scan(
		&cp.ID, &seriesID, &storyID, &cp.CharacterName, &role, &imp, &pov,
		&want, &need, &start, &end, &rels, &trait, &book, &obst, &ive, &wound, &age, &theme, &voice,
		&cp.CreatedAt, &cp.UpdatedAt,
	)
	cp.SeriesID = int64Ptr(seriesID)
	cp.StoryID = int64Ptr(storyID)
	cp.RoleInStory = strPtr(role)
	cp.StoryImportance = strPtr(imp)
	cp.POVCharacter = pov != 0
	cp.CharacterWant = strPtr(want)
	cp.CharacterNeed = strPtr(need)
	cp.ArcStatusAtStart = strPtr(start)
	cp.ArcStatusAtEnd = strPtr(end)
	cp.DefiningTrait = strPtr(trait)
	cp.FirstAppearanceBook = intPtr(book)
	cp.PrimaryObstacle = strPtr(obst)
	cp.InternalVsExternal = strPtr(ive)
	cp.CoreWound = strPtr(wound)
	cp.Age = intPtr(age)
	cp.ThematicResonance = strPtr(theme)
	cp.CharacterVoiceNotes = strPtr(voice)
	if rels.Valid && rels.String != "" {
		_ = json.Unmarshal([]byte(rels.String), &cp.KeyRelationships)
	}
	return cp, err
}

func boolToInt(v bool) int {
	if v {
		return 1
	}
	return 0
}

func zeroToNil(n int) any {
	if n == 0 {
		return nil
	}
	return n
}

func nullableJSON(b []byte) any {
	if b == nil {
		return nil
	}
	return string(b)
}
