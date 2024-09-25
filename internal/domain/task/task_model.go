package task

import (
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/guregu/null"
	"github.com/tarkiman/go/shared"
	"github.com/tarkiman/go/shared/failure"
)

const (
	EmptyFilterError string = "empty filter is not allowed"
)

// Task is a sample parent entity model.
type Task struct {
	ID          uuid.UUID   `db:"ID"`
	Title       string      `db:"TITLE" validate:"required"`
	Description null.String `db:"DESCRIPTION"`
	Status      null.String `db:"STATUS"`
	CreatedAt   null.Time   `db:"created_at"`
	CreatedBy   null.Int    `db:"created_by"`
	UpdatedAt   null.Time   `db:"updated_at"`
	UpdatedBy   null.Int    `db:"updated_by"`
	DeletedAt   null.Time   `db:"deleted_at"`
	DeletedBy   null.String `db:"deleted_by"`
}

// TaskRequestFormat represents a Task's standard formatting for JSON deserializing.
type TaskRequestFormat struct {
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
	Status      string `json:"status" validate:"oneof=pending completed"`
	ID          string `json:"-"`
	CreatedBy   int64  `json:"-"`
	UpdatedBy   int64  `json:"-"`
}

// TaskRequestFormat creates a new Task from its request format.
func (t Task) CreateRequestFormat(request TaskRequestFormat) (task Task, err error) {
	var createdBy int64
	createdBy = 1
	task = Task{
		ID:          uuid.New(),
		Title:       request.Title,
		Description: null.StringFrom(request.Description),
		Status:      null.StringFrom(request.Status),
		CreatedAt:   null.TimeFrom(time.Now()),
		CreatedBy:   null.IntFrom(createdBy), //to do get from token
	}
	return
}

// Update a Task.
func (u *Task) UpdateRequestFormat(request TaskRequestFormat) (err error) {
	u.Title = request.Title
	u.Description = null.StringFrom(request.Description)
	u.Status = null.StringFrom(request.Status)
	u.UpdatedAt = null.TimeFrom(time.Now())
	u.UpdatedBy = null.IntFrom(request.UpdatedBy)
	err = u.Validate()
	return
}

// TaskResponseFormat represents a Task's standard formatting for JSON serializing.
type TaskResponse struct {
	ID         string `json:"id"`
	Title      string `json:"title"`
	Desciption string `json:"description"`
	Status     string `json:"status"`
}

type TaskResponseFormat struct {
	Message string      `json:"message"`
	Task    interface{} `json:"task,omitempty"`
}

func (t Task) ToJSONResponseFormat(message string) (response TaskResponseFormat) {
	task := TaskResponse{
		ID:         t.ID.String(),
		Title:      t.Title,
		Desciption: t.Description.String,
		Status:     t.Status.String,
	}
	response.Message = message
	response.Task = task
	return
}

// SoftDelete marks a Task as deleted by set value of "deletedAt" and "deletedBy"
func (u *Task) SoftDelete(deletedBy string) (err error) {
	if u.DeletedAt.Valid {
		return failure.Conflict("softDelete", "Task", "already marked as deleted")
	}

	u.DeletedAt = null.TimeFrom(time.Now())
	u.DeletedBy = null.StringFrom(deletedBy)

	return
}

// Validate validates the entity.
func (u *Task) Validate() (err error) {
	validator := shared.GetValidator()
	return validator.Struct(u)
}

type TaskFilter struct {
	Keyword    string     `json:"keyword"`
	Sort       TaskSort   `json:"sort"`
	Pagination Pagination `json:"pagination"`
}

type TaskSort struct {
	Field string `json:"field" validate:"oneof=title status created_at"`
	Order string `json:"order" validate:"oneof=ASC DESC"`
}

type Pagination struct {
	Count     int `json:"count"`
	Page      int `json:"page"`
	PageSize  int `json:"pageSize"`
	TotalPage int `json:"totalPage"`
}

func (p *Pagination) SetDefaults() {
	if p.Page == 0 {
		p.Page = 1
	}
	if p.PageSize == 0 {
		p.PageSize = 10
	}
}

func (s *TaskSort) SetDefaults() (err error) {
	if s.Field == "" {
		s.Field = "created_at"
	}
	if s.Order == "" {
		s.Order = "DESC"
	}
	err = s.Validate()
	return
}

// Validate validates the entity.
func (s *TaskSort) Validate() (err error) {
	validator := shared.GetValidator()
	return validator.Struct(s)
}

type TaskFilterResponseFormat struct {
	Tasks      []TaskResponse `json:"tasks"`
	Pagination Pagination     `json:"pagination"`
	Sort       TaskSort       `json:"sort"`
}

type TaskFilterQueryData struct {
	Task
	FilterCount int `db:"count"`
}

func (t *TaskFilterResponseFormat) SetSortAndPagination(filter TaskFilter) {
	t.Sort = filter.Sort
	t.Pagination = filter.Pagination
	t.Pagination.TotalPage = int(math.Ceil(float64(t.Pagination.Count) / float64(t.Pagination.PageSize)))
}

func (t Task) ToResponseFormat() TaskResponse {
	resp := TaskResponse{
		ID:         t.ID.String(),
		Title:      t.Title,
		Desciption: t.Description.String,
		Status:     t.Status.String,
	}
	return resp
}
