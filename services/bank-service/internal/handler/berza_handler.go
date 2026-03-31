package handler

import (
	"context"
	"errors"

	pb "banka-backend/proto/banka"
	"banka-backend/services/bank-service/internal/domain"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

// ListExchanges vraća listu berzi sa opcionim filterima.
// Mapped to: GET /bank/exchanges
func (h *BankHandler) ListExchanges(ctx context.Context, req *pb.ListExchangesRequest) (*pb.ListExchangesResponse, error) {
	filter := domain.ListExchangesFilter{
		Polity: req.GetPolity(),
		Search: req.GetSearch(),
	}
	exchanges, err := h.berzaService.ListExchanges(ctx, filter)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "greška pri dohvatanju berzi: %v", err)
	}
	resp := &pb.ListExchangesResponse{
		Exchanges: make([]*pb.Exchange, 0, len(exchanges)),
	}
	for _, e := range exchanges {
		resp.Exchanges = append(resp.Exchanges, exchangeToProto(e))
	}
	return resp, nil
}

// GetExchange dohvata jednu berzu po ID-u ili MIC kodu.
// Mapped to: GET /bank/exchanges/{id}  |  GET /bank/exchanges/mic/{mic_code}
func (h *BankHandler) GetExchange(ctx context.Context, req *pb.GetExchangeRequest) (*pb.Exchange, error) {
	if req.GetId() == 0 && req.GetMicCode() == "" {
		return nil, status.Error(codes.InvalidArgument, "mora biti prosleđen id ili mic_code")
	}
	e, err := h.berzaService.GetExchange(ctx, req.GetId(), req.GetMicCode())
	if err != nil {
		if errors.Is(err, domain.ErrExchangeNotFound) {
			return nil, status.Error(codes.NotFound, "berza nije pronađena")
		}
		return nil, status.Errorf(codes.Internal, "greška pri dohvatanju berze: %v", err)
	}
	return exchangeToProto(*e), nil
}

// ToggleMarketTestMode uključuje ili isključuje bypass radnog vremena berzi.
// Mapped to: POST /bank/admin/exchanges/test-mode
// Pristup: samo EMPLOYEE/Admin.
func (h *BankHandler) ToggleMarketTestMode(ctx context.Context, req *pb.ToggleMarketTestModeRequest) (*emptypb.Empty, error) {
	if _, err := extractEmployeeID(ctx); err != nil {
		return nil, err
	}
	if err := h.berzaService.ToggleMarketTestMode(ctx, req.GetEnabled()); err != nil {
		return nil, status.Errorf(codes.Internal, "greška pri podešavanju market test mode: %v", err)
	}
	return &emptypb.Empty{}, nil
}

// exchangeToProto mapira domenski Exchange u proto poruku.
func exchangeToProto(e domain.Exchange) *pb.Exchange {
	return &pb.Exchange{
		Id:         e.ID,
		Name:       e.Name,
		Acronym:    e.Acronym,
		MicCode:    e.MICCode,
		Polity:     e.Polity,
		CurrencyId: e.CurrencyID,
		Timezone:   e.Timezone,
	}
}
