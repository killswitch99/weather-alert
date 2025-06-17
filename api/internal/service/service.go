package service

import (
	"workflow-code-test/api/internal/execution"
	"workflow-code-test/api/internal/handler"
	"workflow-code-test/api/internal/repository"
	"workflow-code-test/api/internal/workflow"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgxpool"
)



type Service struct {
	DB      *pgxpool.Pool
	Handler *handler.WorkflowHandler
}

func NewService(dbPool *pgxpool.Pool, engine *execution.Engine) (*Service, error) {
	repo := repository.NewWorkflowRepository(dbPool)
	
	workflowService := workflow.NewWorkflowService(repo)
	workflowService.SetEngine(engine)
	handler := handler.NewWorkflowHandler(workflowService)
	
	return &Service{
		DB: dbPool,
		Handler: handler,
	}, nil
}


func (s *Service) LoadRoutes(parentRouter *mux.Router, isProduction bool) {
	router := parentRouter.PathPrefix("/workflows").Subrouter()
	router.StrictSlash(false)
	
	router.HandleFunc("/{id}", s.Handler.HandleGetWorkflow).Methods("GET")
	router.HandleFunc("/{id}/execute", s.Handler.HandleExecuteWorkflow).Methods("POST")
}
