package store

import (
	"database/sql"
	"fmt"
	"regexp"
	"strings"
)

func (s *Store) GetSetting(key string) (SettingValue, error) {
	var value string
	err := s.DB.QueryRow(`SELECT value FROM user_settings WHERE user_id = ? AND key = ?`, s.UserID, key).Scan(&value)
	if err == sql.ErrNoRows {
		return SettingValue{}, fmt.Errorf("setting not found")
	}
	if err != nil {
		return SettingValue{}, err
	}
	return SettingValue{Key: key, Value: value}, nil
}

func (s *Store) PutSetting(key, value string) (SettingValue, error) {
	_, err := s.DB.Exec(`
		INSERT INTO user_settings (user_id, key, value, updated_at)
		VALUES (?, ?, ?, datetime('now'))
		ON CONFLICT(user_id, key) DO UPDATE SET value = excluded.value, updated_at = datetime('now')`,
		s.UserID, key, value)
	if err != nil {
		return SettingValue{}, err
	}
	return SettingValue{Key: key, Value: value}, nil
}

var identRE = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

func (s *Store) AdminListTables() (AdminTables, error) {
	rows, err := s.DB.Query(`SELECT name FROM sqlite_master WHERE type = 'table' AND name NOT LIKE 'sqlite_%' ORDER BY name`)
	if err != nil {
		return AdminTables{}, err
	}
	defer rows.Close()
	out := AdminTables{Tables: make([]string, 0)}
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return AdminTables{}, err
		}
		out.Tables = append(out.Tables, name)
	}
	return out, rows.Err()
}

func (s *Store) AdminTableSchema(table string) (AdminSchema, error) {
	if !identRE.MatchString(table) {
		return AdminSchema{}, fmt.Errorf("invalid table name")
	}
	rows, err := s.DB.Query(fmt.Sprintf(`PRAGMA table_info(%s)`, table))
	if err != nil {
		return AdminSchema{}, err
	}
	defer rows.Close()
	out := AdminSchema{Columns: make([]AdminColumn, 0)}
	for rows.Next() {
		var cid int
		var name, ctype string
		var notnull, pk int
		var dflt sql.NullString
		if err := rows.Scan(&cid, &name, &ctype, &notnull, &dflt, &pk); err != nil {
			return AdminSchema{}, err
		}
		nullable := "YES"
		if notnull != 0 {
			nullable = "NO"
		}
		col := AdminColumn{Name: name, Type: ctype, Nullable: nullable}
		if dflt.Valid {
			v := dflt.String
			col.Default = &v
		}
		out.Columns = append(out.Columns, col)
	}
	return out, rows.Err()
}

func (s *Store) AdminQueryTable(table string, limit, offset int) (AdminTableRows, error) {
	if !identRE.MatchString(table) {
		return AdminTableRows{}, fmt.Errorf("invalid table name")
	}
	if limit <= 0 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}
	var total int
	if err := s.DB.QueryRow(fmt.Sprintf(`SELECT COUNT(*) FROM %s`, table)).Scan(&total); err != nil {
		return AdminTableRows{}, err
	}
	rows, err := s.DB.Query(fmt.Sprintf(`SELECT * FROM %s LIMIT ? OFFSET ?`, table), limit, offset)
	if err != nil {
		return AdminTableRows{}, err
	}
	defer rows.Close()
	return scanMaps(rows, total)
}

func (s *Store) AdminDeleteRow(table string, id int64) (DeletedResult, error) {
	if !identRE.MatchString(table) {
		return DeletedResult{}, fmt.Errorf("invalid table name")
	}
	res, err := s.DB.Exec(fmt.Sprintf(`DELETE FROM %s WHERE id = ?`, table), id)
	if err != nil {
		return DeletedResult{}, err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return DeletedResult{}, err
	}
	if n == 0 {
		return DeletedResult{}, fmt.Errorf("row not found")
	}
	return DeletedResult{Deleted: true}, nil
}

func (s *Store) AdminExecSQL(query string) (AdminSQLResult, error) {
	trimmed := strings.TrimSpace(query)
	if trimmed == "" {
		return AdminSQLResult{}, fmt.Errorf("empty SQL")
	}
	upper := strings.ToUpper(trimmed)
	if strings.HasPrefix(upper, "SELECT") || strings.HasPrefix(upper, "PRAGMA") || strings.HasPrefix(upper, "WITH") {
		rows, err := s.DB.Query(trimmed)
		if err != nil {
			return AdminSQLResult{}, err
		}
		defer rows.Close()
		mapped, err := scanMaps(rows, 0)
		if err != nil {
			return AdminSQLResult{}, err
		}
		return AdminSQLResult{
			Columns:      mapped.Columns,
			Rows:         mapped.Rows,
			RowsAffected: int64(len(mapped.Rows)),
		}, nil
	}
	res, err := s.DB.Exec(trimmed)
	if err != nil {
		return AdminSQLResult{}, err
	}
	n, _ := res.RowsAffected()
	return AdminSQLResult{Message: "SQL executed successfully", RowsAffected: n}, nil
}

func scanMaps(rows *sql.Rows, total int) (AdminTableRows, error) {
	cols, err := rows.Columns()
	if err != nil {
		return AdminTableRows{}, err
	}
	out := AdminTableRows{Columns: cols, Rows: make([][]string, 0)}
	for rows.Next() {
		raw := make([]any, len(cols))
		ptrs := make([]any, len(cols))
		for i := range raw {
			ptrs[i] = &raw[i]
		}
		if err := rows.Scan(ptrs...); err != nil {
			return AdminTableRows{}, err
		}
		line := make([]string, len(cols))
		for i := range cols {
			line[i] = cellString(raw[i])
		}
		out.Rows = append(out.Rows, line)
	}
	if total == 0 {
		out.Total = len(out.Rows)
	} else {
		out.Total = total
	}
	return out, rows.Err()
}

func cellString(v any) string {
	switch t := v.(type) {
	case nil:
		return ""
	case []byte:
		return string(t)
	case string:
		return t
	default:
		return fmt.Sprint(t)
	}
}
