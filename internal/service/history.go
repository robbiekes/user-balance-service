package service

import (
	"context"
	v1 "user-balance-api/internal/controller/http/v1"
	"user-balance-api/internal/entity"
)

type HistoryService struct {
	repo HistoryRepo
}

func NewHistoryService(repo HistoryRepo) *HistoryService {
	return &HistoryService{repo: repo}
}

func (h *HistoryService) ShowAll(ctx context.Context) ([]entity.History, error) {
	return h.repo.ShowAll(ctx)
}

func (h *HistoryService) ShowById(ctx context.Context, id int) ([]entity.History, error) {
	return h.repo.ShowById(ctx, id)
}

func (h *HistoryService) ShowSorted(ctx context.Context, srt v1.SortArgs) ([]entity.History, error) {
	return h.repo.ShowSorted(ctx, srt)
}

func (h *HistoryService) SaveHistory(ctx context.Context, input entity.History) (int, error) {
	return h.repo.SaveHistory(ctx, input)
}

func (h *HistoryService) Pagination(ctx context.Context, args v1.PaginationArgs) ([]entity.History, error) {
	return h.repo.Pagination(ctx, args)
}
