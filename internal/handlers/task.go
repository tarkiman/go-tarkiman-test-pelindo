package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/tarkiman/go/internal/domain/task"
	"github.com/tarkiman/go/shared"
	"github.com/tarkiman/go/shared/failure"
	"github.com/tarkiman/go/transport/http/middleware"
	"github.com/tarkiman/go/transport/http/response"
)

// TaskHandler is the HTTP handler for Task domain.
type TaskHandler struct {
	TaskService    task.TaskService
	AuthMiddleware *middleware.Authentication
}

// ProvideTaskHandler is the provider for this handler.
func ProvideTaskHandler(TaskService task.TaskService, authMiddleware *middleware.Authentication) TaskHandler {
	return TaskHandler{
		TaskService:    TaskService,
		AuthMiddleware: authMiddleware,
	}
}

// Router sets up the router for this domain.
func (h *TaskHandler) Router(r chi.Router) {
	r.Route("/tasks", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			// r.Use(h.AuthMiddleware.ClientCredential)
			r.Post("/", h.CreateTask)
			r.Get("/", h.ResolveTaskByFilter)
			r.Get("/{id}", h.ResolveTaskByID)
			r.Put("/{id}", h.UpdateTask)
			r.Delete("/{id}", h.SoftDeleteTask)
		})
	})
}

// CreateTask creates a new Task.
// @Summary Create a new Task.
// @Description This endpoint creates a new Task.
// @Tags Task
// @Security OauthToken
// @Param Task body task.TaskRequestFormat true "The Task to be created."
// @Produce json
// @Success 201 {object} response.Base{data=task.TaskResponseFormat}
// @Failure 400 {object} response.Base
// @Failure 409 {object} response.Base
// @Failure 500 {object} response.Base
// @Router /v1/tasks [post]
func (h *TaskHandler) CreateTask(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var requestFormat task.TaskRequestFormat
	err := decoder.Decode(&requestFormat)
	if err != nil {
		response.WithError(w, failure.BadRequest(err))
		return
	}

	err = shared.GetValidator().Struct(requestFormat)
	if err != nil {
		response.WithError(w, failure.BadRequest(err))
		return
	}

	// userID := r.Header.Get("x-userid")
	// createdBy, err := strconv.ParseInt(userID, 10, 64)
	// if err != nil {
	// 	response.WithError(w, failure.BadRequest(err))
	// 	return
	// }
	// requestFormat.CreatedBy = createdBy
	resp, err := h.TaskService.Create(requestFormat)
	if err != nil {
		response.WithError(w, err)
		return
	}

	response.WithJSON(w, http.StatusCreated, resp)
}

// ResolveTaskByFilter resolves a Task by filter.
// @Summary Resolve Task by filter
// @Description This endpoint resolves a Task by filter.
// @Tags Task
// @Param TaskFilter body task.TaskFilter true "The filter of task to be searched"
// @Produce json
// @Success 200 {object} response.Base{data=task.TaskFilterResponseFormat}
// @Failure 400 {object} response.Base
// @Failure 404 {object} response.Base
// @Failure 500 {object} response.Base
// @Router /v1/tasks  [get]
func (h *TaskHandler) ResolveTaskByFilter(w http.ResponseWriter, r *http.Request) {
	var taskFilter task.TaskFilter

	// err := json.NewDecoder(r.Body).Decode(&taskFilter)
	// if err != nil {
	// 	response.WithError(w, failure.BadRequest(err))
	// 	log.Info().Msg(err.Error())
	// 	return
	// }

	tasks, err := h.TaskService.ResolveByFilter(taskFilter)
	if err != nil {
		if err.Error() == task.EmptyFilterError {
			err = failure.BadRequest(err)
		}
		response.WithError(w, err)
		return
	}

	response.WithJSON(w, http.StatusOK, tasks)
}

// ResolveTaskByID resolves a Task by its ID.
// @Summary Resolve Task by ID
// @Description This endpoint resolves a Task by its ID.
// @Tags Task
// @Security OauthToken
// @Param id path string true "The Task's identifier."
// @Produce json
// @Success 200 {object} response.Base{data=task.TaskResponseFormat}
// @Failure 400 {object} response.Base
// @Failure 404 {object} response.Base
// @Failure 500 {object} response.Base
// @Router /v1/task/{id} [get]
func (h *TaskHandler) ResolveTaskByID(w http.ResponseWriter, r *http.Request) {
	idString := chi.URLParam(r, "id")

	id, err := uuid.Parse(idString)
	if err != nil {
		response.WithError(w, failure.BadRequest(err))
		log.Info().Msg(err.Error())
		return
	}

	task, err := h.TaskService.ResolveByID(id)
	if err != nil {
		response.WithError(w, err)
		return
	}

	response.WithJSON(w, http.StatusOK, task)
}

// UpdateTask updates a Task.
// @Summary Update a Task.
// @Description This endpoint updates an existing Task.
// @Tags Task
// @Security OauthToken
// @Param id path string true "The Task's identifier."
// @Param task body task.TaskRequestFormat true "The Task to be updated."
// @Produce json
// @Success 200 {object} task.TaskResponseFormat
// @Failure 400 {object} response.Base
// @Failure 409 {object} response.Base
// @Failure 500 {object} response.Base
// @Router /v1/tasks/{id} [put]
func (h *TaskHandler) UpdateTask(w http.ResponseWriter, r *http.Request) {
	idString := chi.URLParam(r, "id")

	id, err := uuid.Parse(idString)
	if err != nil {
		response.WithError(w, failure.BadRequest(err))
		log.Info().Msg(err.Error())
		return
	}

	decoder := json.NewDecoder(r.Body)
	var requestFormat task.TaskRequestFormat
	err = decoder.Decode(&requestFormat)
	if err != nil {
		response.WithError(w, failure.BadRequest(err))
		return
	}

	err = shared.GetValidator().Struct(requestFormat)
	if err != nil {
		response.WithError(w, failure.BadRequest(err))
		return
	}

	// userID := r.Header.Get("x-userid")
	// updatedBy, err := strconv.ParseInt(userID, 10, 64)
	// if err != nil {
	// 	response.WithError(w, failure.BadRequest(err))
	// 	return
	// }
	// requestFormat.UpdatedBy = updatedBy

	resp, err := h.TaskService.Update(id, requestFormat)
	if err != nil {
		response.WithError(w, err)
		return
	}

	response.WithJSON(w, http.StatusOK, resp)
}

// SoftDeleteTask marks a Task as deleted.
// @Summary Marks a Task as deleted.
// @Description This endpoint marks an existing Task as deleted. This is done by
// @Description set values of "deletedAt" and "deletedBy" properties of the Task.
// @Tags Task
// @Security OauthToken
// @Param id path string true "The Task's identifier."
// @Produce json
// @Success 200 {object} task.TaskResponse
// @Failure 400 {object} response.Base
// @Failure 409 {object} response.Base
// @Failure 500 {object} response.Base
// @Router /v1/tasks/{id} [delete]
func (h *TaskHandler) SoftDeleteTask(w http.ResponseWriter, r *http.Request) {
	idString := chi.URLParam(r, "id")
	id, err := uuid.Parse(idString)
	if err != nil {
		response.WithError(w, failure.BadRequest(err))
		log.Info().Msg(err.Error())
		return
	}

	deletedBy := r.Header.Get("x-userid")
	resp, err := h.TaskService.SoftDelete(id, deletedBy)
	if err != nil {
		response.WithError(w, err)
		return
	}

	response.WithMessage(w, http.StatusOK, resp.Message)
}
