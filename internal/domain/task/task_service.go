package task

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/tarkiman/go/configs"
	"github.com/tarkiman/go/shared/failure"
)

// TaskService is the service interface for Task entities.
type TaskService interface {
	Create(requestFormat TaskRequestFormat) (response TaskResponseFormat, err error)
	ResolveByFilter(filter TaskFilter) (task TaskFilterResponseFormat, err error)
	ResolveByID(id uuid.UUID) (task Task, err error)
	Update(id uuid.UUID, requestFormat TaskRequestFormat) (response TaskResponseFormat, err error)
	SoftDelete(id uuid.UUID, deletedBy string) (response TaskResponseFormat, err error)
}

// TaskServiceImpl is the service implementation for Task entities.
type TaskServiceImpl struct {
	Config         *configs.Config
	TaskRepository TaskRepository
}

// ProvideTaskServiceImpl is the provider for this service.
func ProvideTaskServiceImpl(
	config *configs.Config,
	taskRepository TaskRepository) *TaskServiceImpl {
	s := new(TaskServiceImpl)
	s.Config = config
	s.TaskRepository = taskRepository

	return s
}

// Create creates a new Task.
func (s *TaskServiceImpl) Create(requestFormat TaskRequestFormat) (response TaskResponseFormat, err error) {
	task := Task{}
	task, err = task.CreateRequestFormat(requestFormat)
	if err != nil {
		return response, failure.BadRequest(err)
	}

	err = s.TaskRepository.Create(task)
	if err != nil {
		log.Err(err).Msg("[Create] error TaskRepository.Create")
		return response, failure.InternalError(err)
	}

	response = task.ToJSONResponseFormat(MessageSuccessCreatedData)
	return
}

// ResolveByFilter resolves tasks by filter
func (s *TaskServiceImpl) ResolveByFilter(filter TaskFilter) (taskResponse TaskFilterResponseFormat, err error) {
	err = filter.Sort.SetDefaults()
	if err != nil {
		return
	}
	filter.Pagination.SetDefaults()
	tasks, err := s.TaskRepository.ResolveByFilter(filter)
	if err != nil {
		return
	}

	for _, task := range tasks {
		taskResponse.Tasks = append(taskResponse.Tasks, task.ToResponseFormat())
	}
	if len(tasks) > 0 {
		filter.Pagination.Count = tasks[0].FilterCount
		taskResponse.SetSortAndPagination(filter)
	}
	return
}

// ResolveByID resolves a Task by its ID.
func (s *TaskServiceImpl) ResolveByID(id uuid.UUID) (task Task, err error) {
	task, exist, err := s.TaskRepository.ResolveByID(id)
	if err != nil {
		log.Err(err).Msg("[ResolveByID] error Task ResolveByID")
	}
	if !exist {
		return task, failure.NotFound("Task")
	}

	return
}

// Update updates a Task.
func (s *TaskServiceImpl) Update(id uuid.UUID, requestFormat TaskRequestFormat) (response TaskResponseFormat, err error) {
	task, exist, err := s.TaskRepository.ResolveByID(id)
	if err != nil {
		log.Err(err).Msg("[Update] error TaskRepository.ResolveByID")
		return
	}
	if !exist {
		err = failure.NotFound("Task")
		return
	}

	err = task.UpdateRequestFormat(requestFormat)
	if err != nil {
		return
	}

	err = s.TaskRepository.Update(task)
	if err != nil {
		log.Err(err).Msg("[Update] error TaskRepository.Update")
		return
	}

	response = task.ToJSONResponseFormat(MessageSuccessUpdatedData)
	return
}

// SoftDelete marks a Task as deleted by setting its `deletedAt` and `deletedBy` properties.
func (s *TaskServiceImpl) SoftDelete(id uuid.UUID, deletedBy string) (response TaskResponseFormat, err error) {
	task, exist, err := s.TaskRepository.ResolveByID(id)
	if err != nil {
		log.Err(err).Msg("[SoftDelete] error TaskRepository.ResolveByID")
		return
	}
	if !exist {
		err = failure.NotFound(fmt.Sprintf("TaskID %s", id.String()))
		return
	}

	err = task.SoftDelete(deletedBy)
	if err != nil {
		return
	}

	err = s.TaskRepository.SoftDelete(task)
	if err != nil {
		log.Err(err).Msg("[SoftDelete] error TaskRepository.SoftDelete")
		err = failure.InternalError(err)
		return
	}
	response.Message = MessageSuccessDeletedData
	return
}
