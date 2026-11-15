package transport

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/leetun2k2/go-bedrock/logx"
	"github.com/leetun2k2/go-product/config"
	"github.com/leetun2k2/go-product/internal/handler"
	"github.com/leetun2k2/go-product/internal/repository"
	"github.com/leetun2k2/go-product/internal/service"
)

type Application struct {
	Config         *config.Config
	Logger         *logx.Logger
	CatalogHandler *handler.CatalogHandler
	AdminHandler   *handler.AdminHandler
	HealthHandler  *handler.HealthHandler
}

func NewApplication(cfg *config.Config, logger *logx.Logger, db *pgxpool.Pool) (*Application, error) {
	repo := repository.New(db, logger)
	catalogService := service.NewCatalogService(repo, logger)
	adminService := service.NewAdminService(repo, logger)

	return &Application{
		Config:         cfg,
		Logger:         logger,
		CatalogHandler: handler.NewCatalogHandler(catalogService, adminService),
		AdminHandler:   handler.NewAdminHandler(adminService),
		HealthHandler:  handler.NewHealthHandler(db),
	}, nil
}
