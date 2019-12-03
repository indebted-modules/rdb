package rdb

import (
	"errors"
	"fmt"
	"reflect"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/iancoleman/strcase"
	"github.com/indebted-modules/cfg"
	"github.com/indebted-modules/uuid"
	"github.com/jmoiron/modl"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/rs/zerolog/log"

	// postgres driver
	_ "github.com/lib/pq"
)

var registry []interface{}

// Entity interface
type Entity interface {
	GetID() string
	SetID(string)
}

// Repo implements the repository pattern for entity management
type Repo struct {
	dbmap *modl.DbMap
}

// Tx holds a database transaction
type Tx struct {
	transaction *modl.Transaction
	dbmap       *modl.DbMap
}

// ErrEntityAlreadyExists is returned when created an existing entity
type ErrEntityAlreadyExists struct {
	OriginalError error
}

// Error .
func (e *ErrEntityAlreadyExists) Error() string {
	return e.OriginalError.Error()
}

// NewRepo creates a repository for entity types in the registry
func NewRepo() *Repo {
	dbmap := newDBMap()
	for _, entity := range registry {
		dbmap.AddTable(entity).SetKeys(false, "id")
	}
	return &Repo{
		dbmap: dbmap,
	}
}

// Register an entity type for persistence
func Register(i interface{}) {
	registry = append(registry, i)
}

func newDBMap() *modl.DbMap {
	db, err := sqlx.Connect("postgres", cfg.DatabaseURL())
	if err != nil {
		log.
			Fatal().
			Err(err).
			Msg("Failed connecting to the database")
	}
	db.SetConnMaxLifetime(time.Hour)
	db.SetMaxIdleConns(1)
	db.SetMaxOpenConns(1)
	modl.TableNameMapper = strcase.ToSnake
	sqlx.NameMapper = strcase.ToSnake
	return modl.NewDbMap(db.DB, modl.PostgresDialect{})
}

// Create executes an insert statement
func (r *Repo) Create(entity Entity) error {
	return create(entity, r.dbmap)
}

// Create executes an insert statement within a transaction
func (t *Tx) Create(entity Entity) error {
	return create(entity, t.transaction)
}

// Update executes an update statement matching by primary key
func (r *Repo) Update(entity Entity) error {
	return update(entity, r.dbmap)
}

// Update executes an update statement within a transaction matching by primary key
func (t *Tx) Update(entity Entity) error {
	return update(entity, t.transaction)
}

// Get executes a select statement scanning the first row into dest
func (r *Repo) Get(dest interface{}, sb sq.SelectBuilder) error {
	return get(dest, sb, r.dbmap)
}

// Get executes a select statement within a transaction scanning the first row into dest
func (t *Tx) Get(dest interface{}, sb sq.SelectBuilder) error {
	return get(dest, sb, t.transaction)
}

// GetByID executes a select statement by ID scanning the matching row into dest
func (r *Repo) GetByID(dest interface{}, id string) error {
	return getByID(dest, id, r.dbmap, r.dbmap)
}

// GetByID executes a select statement by ID within a transaction scanning the matching row into dest
func (t *Tx) GetByID(dest interface{}, id string) error {
	return getByID(dest, id, t.transaction, t.dbmap)
}

// Find executes a select statement scanning all rows into dest
func (r *Repo) Find(dest interface{}, sb sq.SelectBuilder) error {
	return find(dest, sb, r.dbmap)
}

// Find executes a select statement within a transaction scanning all rows into dest
func (t *Tx) Find(dest interface{}, sb sq.SelectBuilder) error {
	return find(dest, sb, t.transaction)
}

// Begin begins a transaction
func (r *Repo) Begin() (*Tx, error) {
	t, err := r.dbmap.Begin()
	if err != nil {
		return nil, err
	}
	return &Tx{
		transaction: t,
		dbmap:       r.dbmap,
	}, nil
}

// Commit commits a transaction
func (t *Tx) Commit() error {
	err := t.transaction.Commit()
	if err != nil {
		return err
	}
	return nil
}

// Rollback rolls back a transaction
func (t *Tx) Rollback() error {
	err := t.transaction.Rollback()
	if err != nil {
		return err
	}
	return nil
}

func create(entity Entity, se modl.SqlExecutor) error {
	if len(entity.GetID()) == 0 {
		entity.SetID(uuid.NewID())
	}
	err := se.Insert(entity)
	if err != nil {
		var pqError *pq.Error
		if errors.As(err, &pqError) && pqError.Code.Name() == "unique_violation" {
			return &ErrEntityAlreadyExists{OriginalError: pqError}
		}

		return err
	}
	return nil
}

func update(entity Entity, se modl.SqlExecutor) error {
	count, err := se.Update(entity)
	if err == nil && count == 0 {
		err = fmt.Errorf("Unexpected update count: %d", count)
	}
	if err != nil {
		return err
	}
	return nil
}

func get(dest interface{}, sb sq.SelectBuilder, se modl.SqlExecutor) error {
	query, args, err := sb.PlaceholderFormat(sq.Dollar).ToSql()
	if err != nil {
		return err
	}
	err = se.SelectOne(dest, query, args...)
	if err != nil {
		return err
	}
	return nil
}

func getByID(dest interface{}, id string, se modl.SqlExecutor, dbmap *modl.DbMap) error {
	tmap := dbmap.TableFor(dest)
	if tmap == nil {
		return fmt.Errorf("Unknown table for type: %s", reflect.TypeOf(dest).Name())
	}
	sb := sq.
		Select("*").
		From(tmap.TableName).
		Where("id = ?", id)
	return get(dest, sb, se)
}

func find(dest interface{}, sb sq.SelectBuilder, se modl.SqlExecutor) error {
	query, args, err := sb.PlaceholderFormat(sq.Dollar).ToSql()
	if err != nil {
		return err
	}
	err = se.Select(dest, query, args...)
	if err != nil {
		return err
	}
	return nil
}
