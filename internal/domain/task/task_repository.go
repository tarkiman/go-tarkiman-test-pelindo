package task

import (
	"database/sql"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog/log"
	"github.com/tarkiman/go/infras"
	"github.com/tarkiman/go/shared/failure"
	"github.com/tarkiman/go/shared/logger"
)

var (
	queries = struct {
		selectData           string
		selectDataWithFilter string
		insertData           string
		updateData           string
		deleteData           string
		softDeleteData       string
	}{
		selectData: `
			SELECT 
				id, 
				title, 
				description, 
				status,
				created_at, 
				created_by, 
				updated_at, 
				updated_by, 
				deleted_at, 
				deleted_by
			FROM temp.tasks;`,
		selectDataWithFilter: `
			SELECT
				id, 
				title, 
				description, 
				status,
				COUNT(id) OVER() as count
			FROM temp.tasks;`,
		insertData: `
			INSERT INTO temp.tasks (
			    id,
				title, 
				description,
				status,
				created_by
			) VALUES (
			    :id,
				:title, 
				:description,
				:status,
				:created_by);`,
		updateData: `
				UPDATE temp.tasks SET
					title=:title, 
					description=:description, 
					status=:status,
					updated_by=:updated_by
				WHERE id=:id;`,
		deleteData: `DELETE FROM temp.tasks WHERE id = ?;`,
		softDeleteData: `
				UPDATE temp.tasks SET
					deleted_at=:deleted_at, 
					deleted_by=:deleted_by
				WHERE id=:id;`,
	}
)

// TaskRepository is the repository for Task data.
type TaskRepository interface {
	ResolveByID(id uuid.UUID) (task Task, exist bool, err error)
	Create(task Task) (err error)
	ResolveByFilter(filter TaskFilter) (tasks []TaskFilterQueryData, err error)
	Update(task Task) (err error)
	SoftDelete(task Task) (err error)
}

// TaskRepositoryOracle is the MySQL-backed implementation of TaskRepository.
type TaskRepositoryOracle struct {
	DB *infras.OracleConn
}

// ProvideTaskRepositoryOracle is the provider for this repository.
func ProvideTaskRepositoryOracle(db *infras.OracleConn) *TaskRepositoryOracle {
	s := new(TaskRepositoryOracle)
	s.DB = db
	return s
}

// ResolveByID resolves a Task by its ID
func (r *TaskRepositoryOracle) ResolveByID(id uuid.UUID) (task Task, exist bool, err error) {
	err = r.DB.Read.Get(
		&task,
		queries.selectData+" WHERE id = ? AND deleted_at IS NULL ",
		id)
	switch {
	case err == sql.ErrNoRows:
		return task, false, nil
	case err != nil:
		return task, false, err
	}
	return task, true, err
}

func (r *TaskRepositoryOracle) ResolveByFilter(filter TaskFilter) (tasks []TaskFilterQueryData, err error) {
	clauses, args, err := filterClause(filter)
	if err != nil {
		return
	}
	query := queries.selectDataWithFilter
	// checks if there's another args besides sort and pagination
	if len(args) > 0 {
		query += " WHERE deleted_at IS NULL AND " + clauses
	} else {
		query += " WHERE deleted_at IS NULL " + clauses
	}

	query, args, err = sqlx.In(query, args...)
	if err != nil {
		logger.ErrorWithStack(err)
		return
	}

	query = "SELECT id, title, description, status FROM anonymous.tasks"

	err = r.DB.Read.Select(&tasks, query, args...)
	if err != nil {
		log.Err(err)
		// logger.ErrorWithStack(err)
	}
	// use write DB in case the data hasn't been replicated to read DB
	if len(tasks) == 0 {
		log.Warn().Interface("Filter", filter).Msg("Task Filter not found with Read DB, trying using Write DB")
		err = r.DB.Write.Select(&tasks, query, args...)
		if err != nil {
			logger.ErrorWithStack(err)
		}
	}
	return
}

func filterClause(filter TaskFilter) (string, []interface{}, error) {
	args := make([]interface{}, 0)
	clause := make([]string, 0)

	if len(filter.Keyword) > 0 {
		clause = append(clause, "title LIKE ?")
		args = append(args, "%"+filter.Keyword+"%")
	}

	// if len(clause) == 0 {
	// 	err := errors.New(EmptyFilterError)
	// 	return "", args, err
	// }
	whereClause := strings.Join(clause, " AND ")

	whereClause += " ORDER BY " + filter.Sort.Field + " " + filter.Sort.Order

	limit := strconv.Itoa(filter.Pagination.PageSize)
	offset := strconv.Itoa((filter.Pagination.Page - 1) * filter.Pagination.PageSize)
	whereClause += " LIMIT " + limit + " OFFSET " + offset
	return whereClause, args, nil
}

// Create creates a new Task.
func (r *TaskRepositoryOracle) Create(task Task) (err error) {
	_, exists, err := r.ResolveByID(task.ID)
	if err != nil {
		logger.ErrorWithStack(err)
		return
	}

	if exists {
		err = failure.Conflict("create", "Task", "already exists")
		logger.ErrorWithStack(err)
		return
	}

	return r.DB.WithTransaction(func(tx *sqlx.Tx, e chan error) {
		if err := r.txCreate(tx, task); err != nil {
			e <- err
			return
		}

		e <- nil
	})
}

// txCreate creates a Task transactionally given the *sqlx.Tx param.
func (r *TaskRepositoryOracle) txCreate(tx *sqlx.Tx, task Task) (err error) {
	stmt, err := tx.PrepareNamed(queries.insertData)
	if err != nil {
		logger.ErrorWithStack(err)
		return
	}
	defer stmt.Close()

	_, err = stmt.Exec(task)
	if err != nil {
		logger.ErrorWithStack(err)
	}

	return
}

// Update updates a Task.
func (r *TaskRepositoryOracle) Update(task Task) (err error) {
	_, exists, err := r.ResolveByID(task.ID)
	if err != nil {
		logger.ErrorWithStack(err)
		return
	}

	if !exists {
		err = failure.NotFound("Task")
		logger.ErrorWithStack(err)
		return
	}

	return r.DB.WithTransaction(func(tx *sqlx.Tx, e chan error) {

		if err := r.txUpdate(tx, task); err != nil {
			e <- err
			return
		}

		e <- nil
	})
}

// txUpdate updates a Task transactionally, given the *sqlx.Tx param.
func (r *TaskRepositoryOracle) txUpdate(tx *sqlx.Tx, task Task) (err error) {
	stmt, err := tx.PrepareNamed(queries.updateData)
	if err != nil {
		logger.ErrorWithStack(err)
		return
	}
	defer stmt.Close()

	_, err = stmt.Exec(task)
	if err != nil {
		logger.ErrorWithStack(err)
	}

	return
}

// Delete a Task.
func (r *TaskRepositoryOracle) SoftDelete(task Task) (err error) {
	_, exists, err := r.ResolveByID(task.ID)
	if err != nil {
		logger.ErrorWithStack(err)
		return
	}

	if !exists {
		err = failure.NotFound("Task")
		logger.ErrorWithStack(err)
		return
	}

	return r.DB.WithTransaction(func(tx *sqlx.Tx, e chan error) {

		if err := r.txSoftDelete(tx, task); err != nil {
			e <- err
			return
		}

		e <- nil
	})
}

// txSoftDelete updates a Task transactionally, given the *sqlx.Tx param.
func (r *TaskRepositoryOracle) txSoftDelete(tx *sqlx.Tx, task Task) (err error) {
	stmt, err := tx.PrepareNamed(queries.softDeleteData)
	if err != nil {
		logger.ErrorWithStack(err)
		return
	}
	defer stmt.Close()

	_, err = stmt.Exec(task)
	if err != nil {
		logger.ErrorWithStack(err)
	}

	return
}

func (r *TaskRepositoryOracle) txDelete(tx *sqlx.Tx, id string) (err error) {
	_, err = tx.Exec(queries.deleteData, id)
	if err != nil {
		logger.ErrorWithStack(err)
	}
	return
}
