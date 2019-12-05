package rdb_test

import (
	"database/sql"
	"testing"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/indebted-modules/rdb"
	"github.com/indebted-modules/uuid"
	"github.com/jmoiron/modl"
	"github.com/stretchr/testify/suite"
)

type RepoSuite struct {
	suite.Suite
	repo *rdb.Repo
}

// See schema/sample.sql
type EntitySample struct {
	ID      string
	Enabled bool
	Created time.Time
}

func (e *EntitySample) GetID() string {
	return e.ID
}

func (e *EntitySample) SetID(id string) {
	e.ID = id
}

func (e *EntitySample) PreInsert(s modl.SqlExecutor) error {
	e.Created = time.Now().Truncate(time.Microsecond).UTC()
	return nil
}

func TestRepoSuite(t *testing.T) {
	rdb.Register(EntitySample{})
	suite.Run(t, &RepoSuite{
		repo: rdb.NewRepo(),
	})
}

func (s *RepoSuite) TestCreateAndGet() {
	t1 := time.Now().Unix()
	newEntity := &EntitySample{}
	err := s.repo.Create(newEntity)
	t2 := time.Now().Unix()
	s.Nil(err)

	entity := &EntitySample{}
	err = s.repo.GetByID(entity, newEntity.ID)
	s.Equal(newEntity, entity)
	s.True(entity.Created.Unix() >= t1)
	s.True(entity.Created.Unix() <= t2)
	s.Nil(err)

	t1 = time.Now().Unix()
	entity.Enabled = true
	err = s.repo.Update(entity)
	t2 = time.Now().Unix()
	s.Nil(err)
}

func (s *RepoSuite) TestCreateGivenID() {
	id := uuid.NewID()
	newEntity := &EntitySample{ID: id}
	err := s.repo.Create(newEntity)
	s.Nil(err)

	entity := &EntitySample{}
	err = s.repo.GetByID(entity, id)
	s.Equal(newEntity, entity)
}

func (s *RepoSuite) TestGetNotFound() {
	entity := &EntitySample{}
	err := s.repo.GetByID(entity, uuid.NewID())
	s.Equal(sql.ErrNoRows, err)
	s.Empty(entity.ID)
}

func (s *RepoSuite) TestUpdateNone() {
	entity := &EntitySample{ID: uuid.NewID()}
	err := s.repo.Update(entity)
	s.Equal("Unexpected update count: 0", err.Error())
}

func (s *RepoSuite) TestGetCount() {
	err := s.repo.Create(&EntitySample{})
	s.Nil(err)
	var count int64
	sb := sq.
		Select("count(id)").
		From("entity_sample")
	err = s.repo.Get(&count, sb)
	s.True(count >= 1)
	s.Nil(err)
}

func (s *RepoSuite) TestFind() {
	e1 := &EntitySample{}
	err := s.repo.Create(e1)
	s.Nil(err)

	e2 := &EntitySample{}
	err = s.repo.Create(e2)
	s.Nil(err)

	sb := sq.
		Select("*").
		From("entity_sample")
	entities := []*EntitySample{}
	err = s.repo.Find(&entities, sb)
	s.Nil(err)

	s.True(len(entities) >= 2)
	s.Subset(entities, []*EntitySample{e1, e2})
}

func (s *RepoSuite) TestTxCommit() {
	newEntity := &EntitySample{}
	err := s.repo.Create(newEntity)
	s.Nil(err)

	tx, err := s.repo.Begin()
	s.Nil(err)

	entity := &EntitySample{}
	err = tx.GetByID(entity, newEntity.ID)
	s.Equal(newEntity, entity)
	s.False(entity.Enabled)
	s.Nil(err)

	entity.Enabled = true
	err = tx.Update(entity)
	s.Nil(err)

	updatedEntity := &EntitySample{}
	err = tx.GetByID(updatedEntity, newEntity.ID)
	s.True(updatedEntity.Enabled)
	s.Nil(err)

	err = tx.Commit()
	s.Nil(err)

	sb := sq.
		Select("*").
		From("entity_sample")
	entities := []*EntitySample{}
	err = s.repo.Find(&entities, sb)
	s.Nil(err)

	s.True(len(entities) >= 1)
	s.Contains(entities, updatedEntity)
	s.True(updatedEntity.Enabled)
}

func (s *RepoSuite) TestTxRollback() {
	newEntity := &EntitySample{}
	err := s.repo.Create(newEntity)
	s.Nil(err)

	tx, err := s.repo.Begin()
	s.Nil(err)

	entity := &EntitySample{}
	err = tx.GetByID(entity, newEntity.ID)
	s.Equal(newEntity, entity)
	s.False(entity.Enabled)
	s.Nil(err)

	entity.Enabled = true
	err = tx.Update(entity)
	s.Nil(err)

	updatedEntity := &EntitySample{}
	err = tx.GetByID(updatedEntity, newEntity.ID)
	s.True(updatedEntity.Enabled)
	s.Nil(err)

	err = tx.Rollback()
	s.Nil(err)

	sb := sq.
		Select("*").
		From("entity_sample")
	entities := []*EntitySample{}
	err = s.repo.Find(&entities, sb)
	s.Nil(err)

	s.True(len(entities) >= 1)
	s.NotContains(entities, updatedEntity)
	s.True(updatedEntity.Enabled)
	s.Contains(entities, newEntity)
	s.False(newEntity.Enabled)
}
