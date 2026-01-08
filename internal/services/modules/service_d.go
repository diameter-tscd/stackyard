package modules

import (
	"context"
	"strconv"

	"test-go/pkg/infrastructure"
	"test-go/pkg/logger"
	"test-go/pkg/response"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

type Task struct {
	gorm.Model
	Title       string `json:"title"`
	Description string `json:"description"`
	Completed   bool   `json:"completed"`
}

type ServiceD struct {
	db      *infrastructure.PostgresManager
	logger  *logger.Logger
	enabled bool
}

func NewServiceD(db *infrastructure.PostgresManager, enabled bool, logger *logger.Logger) *ServiceD {
	if enabled && db != nil && db.ORM != nil {
		// Auto-migrate the schema
		if err := db.ORM.AutoMigrate(&Task{}); err != nil {
			logger.Error("Error migrating Task model", err)
		}
	}
	return &ServiceD{
		db:      db,
		logger:  logger,
		enabled: enabled,
	}
}

func (s *ServiceD) Name() string { return "Service D (Tasks - GORM)" }

func (s *ServiceD) Enabled() bool {
	// Service is enabled only if configured AND DB is available
	return s.enabled && s.db != nil && s.db.ORM != nil
}

func (s *ServiceD) Endpoints() []string { return []string{"/tasks"} }

func (s *ServiceD) RegisterRoutes(g *echo.Group) {
	sub := g.Group("/tasks")
	sub.GET("", s.listTasks)
	sub.POST("", s.createTask)
	sub.PUT("/:id", s.updateTask)
	sub.DELETE("/:id", s.deleteTask)
}

func (s *ServiceD) listTasks(c echo.Context) error {
	var tasks []Task

	// Use async GORM operation to avoid blocking main thread
	result := s.db.GORMFindAsync(context.Background(), &tasks)

	// Wait for the async operation to complete
	_, err := result.Wait()
	if err != nil {
		return response.InternalServerError(c, err.Error())
	}

	return response.Success(c, tasks)
}

func (s *ServiceD) createTask(c echo.Context) error {
	task := new(Task)
	if err := c.Bind(task); err != nil {
		return response.BadRequest(c, "Invalid input")
	}

	// Use async GORM operation to avoid blocking main thread
	result := s.db.GORMCreateAsync(context.Background(), task)

	// Wait for the async operation to complete
	_, err := result.Wait()
	if err != nil {
		return response.InternalServerError(c, err.Error())
	}

	return response.Created(c, task)
}

func (s *ServiceD) updateTask(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	var task Task

	// First check if task exists using async operation
	findResult := s.db.GORMFirstAsync(context.Background(), &task, id)
	_, err := findResult.Wait()
	if err != nil {
		return response.NotFound(c, "Task not found")
	}

	if err := c.Bind(&task); err != nil {
		return response.BadRequest(c, "Invalid input")
	}

	// Use async GORM update operation
	updateResult := s.db.GORMUpdateAsync(context.Background(), &task, task, "id = ?", id)
	_, err = updateResult.Wait()
	if err != nil {
		return response.InternalServerError(c, err.Error())
	}

	return response.Success(c, task)
}

func (s *ServiceD) deleteTask(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	var task Task

	// Use async GORM delete operation
	result := s.db.GORMDeleteAsync(context.Background(), &task, "id = ?", id)

	// Wait for the async operation to complete
	_, err := result.Wait()
	if err != nil {
		return response.InternalServerError(c, err.Error())
	}

	return response.Success(c, nil, "Task deleted")
}
