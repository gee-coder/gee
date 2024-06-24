package orm

import (
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	geeLog "github.com/gee-coder/gee/log"
)

type GeeDb struct {
	db     *sql.DB
	logger *geeLog.Logger
	Prefix string
}

type GeeSession struct {
	geeDb       *GeeDb
	tx          *sql.Tx
	beginTx     bool
	tableName   string
	fieldName   []string
	placeHolder []string
	values      []any
	updateParam strings.Builder
	whereParam  strings.Builder
	whereValues []any
}

func Open(driverName string, source string) *GeeDb {
	db, err := sql.Open(driverName, source)
	if err != nil {
		panic(err)
	}
	// 最大空闲连接数，默认不配置，是2个最大空闲连接
	db.SetMaxIdleConns(5)
	// 最大连接数，默认不配置，是不限制最大连接数
	db.SetMaxOpenConns(100)
	// 连接最大存活时间
	db.SetConnMaxLifetime(time.Minute * 3)
	// 空闲连接最大存活时间
	db.SetConnMaxIdleTime(time.Minute * 1)
	GeeDb := &GeeDb{
		db:     db,
		logger: geeLog.Default(),
	}
	err = db.Ping()
	if err != nil {
		panic(err)
	}
	return GeeDb
}

func (geeDb *GeeDb) Close() error {
	return geeDb.db.Close()
}

func (db *GeeDb) New(data any) *GeeSession {
	m := &GeeSession{
		geeDb: db,
	}
	t := reflect.TypeOf(data)
	if t.Kind() != reflect.Pointer {
		panic(errors.New("data must be pointer"))
	}
	tVar := t.Elem()
	if m.tableName == "" {
		m.tableName = m.geeDb.Prefix + strings.ToLower(Name(tVar.Name()))
	}
	return m
}

func (s *GeeSession) Table(name string) *GeeSession {
	s.tableName = name
	return s
}

// 每一个操作是独立的 互不影响的 session
func (s *GeeSession) Insert(data any) (int64, int64, error) {
	// insert into table (xxx,xxx) values(?,?)
	s.fieldNames(data)
	query := fmt.Sprintf("insert into %s (%s) values (%s)", s.tableName, strings.Join(s.fieldName, ","), strings.Join(s.placeHolder, ","))
	s.geeDb.logger.Info(query)
	var stmt *sql.Stmt
	var err error
	if s.beginTx {
		stmt, err = s.tx.Prepare(query)
	} else {
		stmt, err = s.geeDb.db.Prepare(query)
	}
	if err != nil {
		return -1, -1, err
	}
	r, err := stmt.Exec(s.values...)
	if err != nil {
		return -1, -1, err
	}
	id, err := r.LastInsertId()
	if err != nil {
		return -1, -1, err
	}
	affected, err := r.RowsAffected()
	if err != nil {
		return -1, -1, err
	}
	return id, affected, nil
}

func (s *GeeSession) fieldNames(data any) {
	// 反射
	t := reflect.TypeOf(data)
	v := reflect.ValueOf(data)
	if t.Kind() != reflect.Pointer {
		panic(errors.New("data must be pointer"))
	}
	tVar := t.Elem()
	vVar := v.Elem()
	if s.tableName == "" {
		s.tableName = s.geeDb.Prefix + strings.ToLower(Name(tVar.Name()))
	}
	for i := 0; i < tVar.NumField(); i++ {
		fieldName := tVar.Field(i).Name
		tag := tVar.Field(i).Tag
		sqlTag := tag.Get("geeorm")
		if sqlTag == "" {
			sqlTag = strings.ToLower(Name(fieldName))
		} else {
			// 判定为自增主键id
			if strings.Contains(sqlTag, "auto_increment") {
				continue
			}
			if strings.Contains(sqlTag, ",") {
				// 这就要求数据库名必须在Tag的第一个value中
				sqlTag = sqlTag[:strings.Index(sqlTag, ",")]
			}
		}
		id := vVar.Field(i).Interface()
		// 判定为自增主键id
		if strings.ToLower(sqlTag) == "id" && IsAutoId(id) {
			continue
		}
		s.fieldName = append(s.fieldName, sqlTag)
		s.placeHolder = append(s.placeHolder, "?")
		s.values = append(s.values, vVar.Field(i).Interface())
	}
}

func (s *GeeSession) InsertBatch(data []any) (int64, int64, error) {
	// insert into table (xxx,xxx) values(?,?),(?,?)
	if len(data) == 0 {
		return -1, -1, errors.New("no data insert")
	}
	s.fieldNames(data[0])
	query := fmt.Sprintf("insert into %s (%s) values ", s.tableName, strings.Join(s.fieldName, ","))
	var sb strings.Builder
	sb.WriteString(query)
	for index, _ := range data {
		sb.WriteString("(")
		sb.WriteString(strings.Join(s.placeHolder, ","))
		sb.WriteString(")")
		if index < len(data)-1 {
			sb.WriteString(",")
		}
	}
	s.batchValues(data)
	s.geeDb.logger.Info(sb.String())
	var stmt *sql.Stmt
	var err error
	if s.beginTx {
		stmt, err = s.tx.Prepare(sb.String())
	} else {
		stmt, err = s.geeDb.db.Prepare(sb.String())
	}
	if err != nil {
		return -1, -1, err
	}
	r, err := stmt.Exec(s.values...)
	if err != nil {
		return -1, -1, err
	}
	id, err := r.LastInsertId()
	if err != nil {
		return -1, -1, err
	}
	affected, err := r.RowsAffected()
	if err != nil {
		return -1, -1, err
	}
	return id, affected, nil
}

func (s *GeeSession) UpdateParam(field string, value any) *GeeSession {
	if s.updateParam.String() != "" {
		s.updateParam.WriteString(",")
	}
	s.updateParam.WriteString(field)
	s.updateParam.WriteString(" = ? ")
	s.values = append(s.values, value)
	return s
}

func (s *GeeSession) UpdateMap(data map[string]any) *GeeSession {
	for k, v := range data {
		if s.updateParam.String() != "" {
			s.updateParam.WriteString(",")
		}
		s.updateParam.WriteString(k)
		s.updateParam.WriteString(" = ? ")
		s.values = append(s.values, v)
	}
	return s
}

func (s *GeeSession) Update(data ...any) (int64, int64, error) {
	// Update("age",1) or Update(user)
	if len(data) > 2 {
		return -1, -1, errors.New("param not valid")
	}
	if len(data) == 0 {
		query := fmt.Sprintf("update %s set %s", s.tableName, s.updateParam.String())
		var sb strings.Builder
		sb.WriteString(query)
		sb.WriteString(s.whereParam.String())
		s.geeDb.logger.Info(sb.String())
		var stmt *sql.Stmt
		var err error
		if s.beginTx {
			stmt, err = s.tx.Prepare(sb.String())
		} else {
			stmt, err = s.geeDb.db.Prepare(sb.String())
		}
		if err != nil {
			return -1, -1, err
		}
		s.values = append(s.values, s.whereValues...)
		r, err := stmt.Exec(s.values...)
		if err != nil {
			return -1, -1, err
		}
		id, err := r.LastInsertId()
		if err != nil {
			return -1, -1, err
		}
		affected, err := r.RowsAffected()
		if err != nil {
			return -1, -1, err
		}
		return id, affected, nil
	}
	single := true
	if len(data) == 2 {
		single = false
	}
	// update table set age=?,name=? where id=?
	if !single {
		if s.updateParam.String() != "" {
			s.updateParam.WriteString(",")
		}
		s.updateParam.WriteString(data[0].(string))
		s.updateParam.WriteString(" = ? ")
		s.values = append(s.values, data[1])
	} else {
		updateData := data[0]
		t := reflect.TypeOf(updateData)
		v := reflect.ValueOf(updateData)
		if t.Kind() != reflect.Pointer {
			panic(errors.New("updateData must be pointer"))
		}
		tVar := t.Elem()
		vVar := v.Elem()
		for i := 0; i < tVar.NumField(); i++ {
			fieldName := tVar.Field(i).Name
			tag := tVar.Field(i).Tag
			sqlTag := tag.Get("msorm")
			if sqlTag == "" {
				sqlTag = strings.ToLower(Name(fieldName))
			} else {
				if strings.Contains(sqlTag, "auto_increment") {
					// 自增长的主键id
					continue
				}
				if strings.Contains(sqlTag, ",") {
					sqlTag = sqlTag[:strings.Index(sqlTag, ",")]
				}
			}
			id := vVar.Field(i).Interface()
			if strings.ToLower(sqlTag) == "id" && IsAutoId(id) {
				continue
			}
			if s.updateParam.String() != "" {
				s.updateParam.WriteString(",")
			}
			s.updateParam.WriteString(sqlTag)
			s.updateParam.WriteString(" = ? ")
			s.values = append(s.values, vVar.Field(i).Interface())
		}
	}
	query := fmt.Sprintf("update %s set %s", s.tableName, s.updateParam.String())
	var sb strings.Builder
	sb.WriteString(query)
	sb.WriteString(s.whereParam.String())
	s.geeDb.logger.Info(sb.String())
	var stmt *sql.Stmt
	var err error
	if s.beginTx {
		stmt, err = s.tx.Prepare(sb.String())
	} else {
		stmt, err = s.geeDb.db.Prepare(sb.String())
	}
	if err != nil {
		return -1, -1, err
	}
	s.values = append(s.values, s.whereValues...)
	r, err := stmt.Exec(s.values...)
	if err != nil {
		return -1, -1, err
	}
	id, err := r.LastInsertId()
	if err != nil {
		return -1, -1, err
	}
	affected, err := r.RowsAffected()
	if err != nil {
		return -1, -1, err
	}
	return id, affected, nil
}

func (s *GeeSession) Delete() (int64, error) {
	// delete from table where id=?
	query := fmt.Sprintf("delete from %s ", s.tableName)
	var sb strings.Builder
	sb.WriteString(query)
	sb.WriteString(s.whereParam.String())
	s.geeDb.logger.Info(sb.String())

	var stmt *sql.Stmt
	var err error
	if s.beginTx {
		stmt, err = s.tx.Prepare(sb.String())
	} else {
		stmt, err = s.geeDb.db.Prepare(sb.String())
	}
	if err != nil {
		return 0, err
	}
	r, err := stmt.Exec(s.whereValues...)
	if err != nil {
		return 0, err
	}
	return r.RowsAffected()
}

func (s *GeeSession) Select(data any, fields ...string) ([]any, error) {
	t := reflect.TypeOf(data)
	if t.Kind() != reflect.Pointer {
		return nil, errors.New("data must be pointer")
	}
	fieldStr := "*"
	if len(fields) > 0 {
		fieldStr = strings.Join(fields, ",")
	}
	query := fmt.Sprintf("select %s from %s ", fieldStr, s.tableName)
	var sb strings.Builder
	sb.WriteString(query)
	sb.WriteString(s.whereParam.String())
	s.geeDb.logger.Info(sb.String())

	stmt, err := s.geeDb.db.Prepare(sb.String())
	if err != nil {
		return nil, err
	}
	rows, err := stmt.Query(s.whereValues...)
	if err != nil {
		return nil, err
	}
	// id user_name age
	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	result := make([]any, 0)
	for {
		if rows.Next() {
			// 由于 传进来的是一个指针地址 如果每次赋值，实际都是一个 result里面 值都一样
			// 每次查询的时候 data都重新换一个地址
			data := reflect.New(t.Elem()).Interface()
			values := make([]any, len(columns))
			fieldScan := make([]any, len(columns))
			for i := range fieldScan {
				fieldScan[i] = &values[i]
			}
			err := rows.Scan(fieldScan...)
			if err != nil {
				return nil, err
			}
			tVar := t.Elem()
			vVar := reflect.ValueOf(data).Elem()
			for i := 0; i < tVar.NumField(); i++ {
				name := tVar.Field(i).Name
				tag := tVar.Field(i).Tag
				// id,auto
				sqlTag := tag.Get("msorm")
				if sqlTag == "" {
					sqlTag = strings.ToLower(Name(name))
				} else {
					if strings.Contains(sqlTag, ",") {
						sqlTag = sqlTag[:strings.Index(sqlTag, ",")]
					}
				}

				for j, colName := range columns {
					if sqlTag == colName {
						target := values[j]
						targetValue := reflect.ValueOf(target)
						fieldType := tVar.Field(i).Type
						// 这样不行 类型不匹配 转换类型
						result := reflect.ValueOf(targetValue.Interface()).Convert(fieldType)
						vVar.Field(i).Set(result)
					}
				}

			}
			result = append(result, data)
		} else {
			break
		}
	}

	return result, nil
}

// select * from table where id=1000
func (s *GeeSession) SelectOne(data any, fields ...string) error {
	t := reflect.TypeOf(data)
	if t.Kind() != reflect.Pointer {
		return errors.New("data must be pointer")
	}
	fieldStr := "*"
	if len(fields) > 0 {
		fieldStr = strings.Join(fields, ",")
	}
	query := fmt.Sprintf("select %s from %s ", fieldStr, s.tableName)
	var sb strings.Builder
	sb.WriteString(query)
	sb.WriteString(s.whereParam.String())
	s.geeDb.logger.Info(sb.String())
	stmt, err := s.geeDb.db.Prepare(sb.String())
	if err != nil {
		return err
	}
	rows, err := stmt.Query(s.whereValues...)
	if err != nil {
		return err
	}
	// id user_name age
	columns, err := rows.Columns()
	if err != nil {
		return err
	}
	values := make([]any, len(columns))
	fieldScan := make([]any, len(columns))
	for i := range fieldScan {
		fieldScan[i] = &values[i]
	}
	if rows.Next() {
		err := rows.Scan(fieldScan...)
		if err != nil {
			return err
		}
		tVar := t.Elem()
		vVar := reflect.ValueOf(data).Elem()
		for i := 0; i < tVar.NumField(); i++ {
			name := tVar.Field(i).Name
			tag := tVar.Field(i).Tag
			// id,auto
			sqlTag := tag.Get("msorm")
			if sqlTag == "" {
				sqlTag = strings.ToLower(Name(name))
			} else {
				if strings.Contains(sqlTag, ",") {
					sqlTag = sqlTag[:strings.Index(sqlTag, ",")]
				}
			}

			for j, colName := range columns {
				if sqlTag == colName {
					target := values[j]
					targetValue := reflect.ValueOf(target)
					fieldType := tVar.Field(i).Type
					// 这样不行 类型不匹配 转换类型
					result := reflect.ValueOf(targetValue.Interface()).Convert(fieldType)
					vVar.Field(i).Set(result)
				}
			}

		}
	}
	return nil
}

func (s *GeeSession) Count() (int64, error) {
	return s.Aggregate("count", "*")
}

func (s *GeeSession) Aggregate(funcName string, field string) (int64, error) {
	var fieldSb strings.Builder
	fieldSb.WriteString(funcName)
	fieldSb.WriteString("(")
	fieldSb.WriteString(field)
	fieldSb.WriteString(")")
	query := fmt.Sprintf("select %s from %s ", fieldSb.String(), s.tableName)
	var sb strings.Builder
	sb.WriteString(query)
	sb.WriteString(s.whereParam.String())
	s.geeDb.logger.Info(sb.String())
	stmt, err := s.geeDb.db.Prepare(sb.String())
	if err != nil {
		return 0, err
	}
	row := stmt.QueryRow(s.whereValues...)
	if row.Err() != nil {
		return 0, err
	}
	var result int64
	err = row.Scan(&result)
	if err != nil {
		return 0, err
	}
	return result, nil
}

// 原生sql的支持
func (s *GeeSession) Exec(query string, values ...any) (int64, error) {
	var stmt *sql.Stmt
	var err error
	if s.beginTx {
		stmt, err = s.tx.Prepare(query)
	} else {
		stmt, err = s.geeDb.db.Prepare(query)
	}
	if err != nil {
		return 0, err
	}
	r, err := stmt.Exec(values)
	if err != nil {
		return 0, err
	}
	if strings.Contains(strings.ToLower(query), "insert") {
		return r.LastInsertId()
	}
	return r.RowsAffected()
}

func (s *GeeSession) QueryRow(sql string, data any, queryValues ...any) error {
	t := reflect.TypeOf(data)
	if t.Kind() != reflect.Pointer {
		return errors.New("data must be pointer")
	}
	stmt, err := s.geeDb.db.Prepare(sql)
	if err != nil {
		return err
	}
	rows, err := stmt.Query(queryValues...)
	if err != nil {
		return err
	}
	// id user_name age
	columns, err := rows.Columns()
	if err != nil {
		return err
	}
	values := make([]any, len(columns))
	fieldScan := make([]any, len(columns))
	for i := range fieldScan {
		fieldScan[i] = &values[i]
	}
	if rows.Next() {
		err := rows.Scan(fieldScan...)
		if err != nil {
			return err
		}
		tVar := t.Elem()
		vVar := reflect.ValueOf(data).Elem()
		for i := 0; i < tVar.NumField(); i++ {
			name := tVar.Field(i).Name
			tag := tVar.Field(i).Tag
			// id,auto
			sqlTag := tag.Get("msorm")
			if sqlTag == "" {
				sqlTag = strings.ToLower(Name(name))
			} else {
				if strings.Contains(sqlTag, ",") {
					sqlTag = sqlTag[:strings.Index(sqlTag, ",")]
				}
			}
			for j, colName := range columns {
				if sqlTag == colName {
					target := values[j]
					targetValue := reflect.ValueOf(target)
					fieldType := tVar.Field(i).Type
					// 这样不行 类型不匹配 转换类型
					result := reflect.ValueOf(targetValue.Interface()).Convert(fieldType)
					vVar.Field(i).Set(result)
				}
			}

		}
	}
	return nil
}

func (s *GeeSession) Begin() error {
	tx, err := s.geeDb.db.Begin()
	if err != nil {
		return err
	}
	s.tx = tx
	s.beginTx = true
	return nil
}

func (s *GeeSession) Commit() error {
	err := s.tx.Commit()
	if err != nil {
		return err
	}
	s.beginTx = false
	return nil
}

func (s *GeeSession) Rollback() error {
	err := s.tx.Rollback()
	if err != nil {
		return err
	}
	s.beginTx = false
	return nil
}

func (s *GeeSession) Where(field string, value any) *GeeSession {
	// id=1 and name=xx
	if s.whereParam.String() == "" {
		s.whereParam.WriteString(" where ")
	}
	s.whereParam.WriteString(field)
	s.whereParam.WriteString(" = ")
	s.whereParam.WriteString(" ? ")
	s.whereValues = append(s.whereValues, value)
	return s
}

func (s *GeeSession) Like(field string, value any) *GeeSession {
	// name like %s%
	if s.whereParam.String() == "" {
		s.whereParam.WriteString(" where ")
	}
	s.whereParam.WriteString(field)
	s.whereParam.WriteString(" like ")
	s.whereParam.WriteString(" ? ")
	s.whereValues = append(s.whereValues, "%"+value.(string)+"%")
	return s
}

func (s *GeeSession) LikeRight(field string, value any) *GeeSession {
	// name like %s%
	if s.whereParam.String() == "" {
		s.whereParam.WriteString(" where ")
	}
	s.whereParam.WriteString(field)
	s.whereParam.WriteString(" like ")
	s.whereParam.WriteString(" ? ")
	s.whereValues = append(s.whereValues, value.(string)+"%")
	return s
}

func (s *GeeSession) LikeLeft(field string, value any) *GeeSession {
	// name like %s%
	if s.whereParam.String() == "" {
		s.whereParam.WriteString(" where ")
	}
	s.whereParam.WriteString(field)
	s.whereParam.WriteString(" like ")
	s.whereParam.WriteString(" ? ")
	s.whereValues = append(s.whereValues, "%"+value.(string))
	return s
}

func (s *GeeSession) Group(field ...string) *GeeSession {
	// group by aa,bb
	s.whereParam.WriteString(" group by ")
	s.whereParam.WriteString(strings.Join(field, ","))
	return s
}

func (s *GeeSession) OrderDesc(field ...string) *GeeSession {
	// order by aa,bb desc
	s.whereParam.WriteString(" order by ")
	s.whereParam.WriteString(strings.Join(field, ","))
	s.whereParam.WriteString(" desc ")
	return s
}

func (s *GeeSession) OrderAsc(field ...string) *GeeSession {
	// order by aa,bb asc
	s.whereParam.WriteString(" order by ")
	s.whereParam.WriteString(strings.Join(field, ","))
	s.whereParam.WriteString(" asc ")
	return s
}

// Order order by aa desc, bb asc
// Order  Order("aa","desc","bb","asc)
func (s *GeeSession) Order(field ...string) *GeeSession {
	if len(field)%2 != 0 {
		panic("field num not true")
	}
	s.whereParam.WriteString(" order by ")
	for index, v := range field {
		s.whereParam.WriteString(v + " ")
		if index%2 != 0 && index < len(field)-1 {
			s.whereParam.WriteString(",")
		}
	}
	return s
}

func (s *GeeSession) And() *GeeSession {
	s.whereParam.WriteString(" and ")
	return s
}

func (s *GeeSession) Or() *GeeSession {
	s.whereParam.WriteString(" or ")
	return s
}

func (s *GeeSession) batchValues(data []any) {
	s.values = make([]any, 0)
	for _, v := range data {
		t := reflect.TypeOf(v)
		v := reflect.ValueOf(v)
		if t.Kind() != reflect.Pointer {
			panic(errors.New("data must be pointer"))
		}
		tVar := t.Elem()
		vVar := v.Elem()
		for i := 0; i < tVar.NumField(); i++ {
			fieldName := tVar.Field(i).Name
			tag := tVar.Field(i).Tag
			sqlTag := tag.Get("msorm")
			if sqlTag == "" {
				sqlTag = strings.ToLower(Name(fieldName))
			} else {
				if strings.Contains(sqlTag, "auto_increment") {
					// 自增长的主键id
					continue
				}
			}
			id := vVar.Field(i).Interface()
			if strings.ToLower(sqlTag) == "id" && IsAutoId(id) {
				continue
			}
			s.values = append(s.values, vVar.Field(i).Interface())
		}
	}
}

func IsAutoId(id any) bool {
	t := reflect.TypeOf(id)
	switch t.Kind() {
	case reflect.Int64:
		if id.(int64) <= 0 {
			return true
		}
	case reflect.Int32:
		if id.(int32) <= 0 {
			return true
		}
	case reflect.Int:
		if id.(int) <= 0 {
			return true
		}
	default:
		return false
	}
	return false
}

func Name(name string) string {
	var names = name[:]
	lastIndex := 0
	var sb strings.Builder
	for index, value := range names {
		if value >= 65 && value <= 90 {
			// 大写字母
			if index == 0 {
				continue
			}
			sb.WriteString(name[:index])
			sb.WriteString("_")
			lastIndex = index
		}
	}
	sb.WriteString(name[lastIndex:])
	return sb.String()
}
