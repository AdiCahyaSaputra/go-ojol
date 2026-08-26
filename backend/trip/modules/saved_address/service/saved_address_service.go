package service

import (
	"context"
	"errors"
	"strconv"

	"github.com/AdiCahyaSaputra/go-ojol/backend/trip/database/entities"
	"github.com/AdiCahyaSaputra/go-ojol/backend/trip/modules/saved_address/dto"
	"github.com/AdiCahyaSaputra/go-ojol/backend/trip/modules/saved_address/repository"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

var (
	ErrNotFound         = dto.ErrSavedAddressNotFound
	ErrInvalidLatLong   = dto.ErrInvalidLatLong
	ErrCustomerNotInCtx = dto.ErrCustomerNotInCtx
)

type SavedAddressService interface {
	List(ctx context.Context) ([]dto.SavedAddressResponse, error)
	GetByID(ctx context.Context, id uuid.UUID) (dto.SavedAddressResponse, error)
	Create(ctx context.Context, req dto.SavedAddressCreateRequest) (dto.SavedAddressResponse, error)
	Update(ctx context.Context, id uuid.UUID, req dto.SavedAddressUpdateRequest) (dto.SavedAddressResponse, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type savedAddressService struct {
	savedAddressRepository repository.SavedAddressRepository
	db                     *gorm.DB
}

func NewSavedAddressService(
	savedAddressRepo repository.SavedAddressRepository,
	db *gorm.DB,
) SavedAddressService {
	return &savedAddressService{
		savedAddressRepository: savedAddressRepo,
		db:                     db,
	}
}

func (s *savedAddressService) List(ctx context.Context) ([]dto.SavedAddressResponse, error) {
	customer, err := customerFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	addresses, err := s.savedAddressRepository.ListByCustomer(ctx, nil, customer.ID)
	if err != nil {
		return nil, err
	}

	out := make([]dto.SavedAddressResponse, 0, len(addresses))
	for _, address := range addresses {
		out = append(out, toResponse(address))
	}
	return out, nil
}

func (s *savedAddressService) GetByID(ctx context.Context, id uuid.UUID) (dto.SavedAddressResponse, error) {
	customer, err := customerFromCtx(ctx)
	if err != nil {
		return dto.SavedAddressResponse{}, err
	}

	address, err := s.savedAddressRepository.GetByIDAndCustomer(ctx, nil, id, customer.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return dto.SavedAddressResponse{}, ErrNotFound
		}
		return dto.SavedAddressResponse{}, err
	}
	return toResponse(address), nil
}

func (s *savedAddressService) Create(ctx context.Context, req dto.SavedAddressCreateRequest) (dto.SavedAddressResponse, error) {
	customer, err := customerFromCtx(ctx)
	if err != nil {
		return dto.SavedAddressResponse{}, err
	}

	latLong, err := normalizeLatLong(req.LatLong)
	if err != nil {
		return dto.SavedAddressResponse{}, err
	}

	var created entities.SavedAddress
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if req.IsDefaultPickup {
			if err := s.savedAddressRepository.ClearDefaultPickup(ctx, tx, customer.ID, uuid.Nil); err != nil {
				return err
			}
		}

		address := entities.SavedAddress{
			CustomerID:      customer.ID,
			Name:            req.Name,
			LatLong:         latLong,
			IsDefaultPickup: req.IsDefaultPickup,
		}
		var createErr error
		created, createErr = s.savedAddressRepository.Create(ctx, tx, address)
		return createErr
	})
	if err != nil {
		return dto.SavedAddressResponse{}, err
	}

	return toResponse(created), nil
}

func (s *savedAddressService) Update(ctx context.Context, id uuid.UUID, req dto.SavedAddressUpdateRequest) (dto.SavedAddressResponse, error) {
	customer, err := customerFromCtx(ctx)
	if err != nil {
		return dto.SavedAddressResponse{}, err
	}

	latLong, err := normalizeLatLong(req.LatLong)
	if err != nil {
		return dto.SavedAddressResponse{}, err
	}

	var updated entities.SavedAddress
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		existing, getErr := s.savedAddressRepository.GetByIDAndCustomer(ctx, tx, id, customer.ID)
		if getErr != nil {
			if errors.Is(getErr, gorm.ErrRecordNotFound) {
				return ErrNotFound
			}
			return getErr
		}

		if req.IsDefaultPickup {
			if clearErr := s.savedAddressRepository.ClearDefaultPickup(ctx, tx, customer.ID, existing.ID); clearErr != nil {
				return clearErr
			}
		}

		existing.Name = req.Name
		existing.LatLong = latLong
		existing.IsDefaultPickup = req.IsDefaultPickup

		var updateErr error
		updated, updateErr = s.savedAddressRepository.Update(ctx, tx, existing)
		return updateErr
	})
	if err != nil {
		return dto.SavedAddressResponse{}, err
	}

	return toResponse(updated), nil
}

func (s *savedAddressService) Delete(ctx context.Context, id uuid.UUID) error {
	customer, err := customerFromCtx(ctx)
	if err != nil {
		return err
	}

	if err := s.savedAddressRepository.Delete(ctx, nil, id, customer.ID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrNotFound
		}
		return err
	}
	return nil
}

func customerFromCtx(ctx context.Context) (entities.Customer, error) {
	customer, ok := ctx.Value("customer").(entities.Customer)
	if !ok {
		return entities.Customer{}, ErrCustomerNotInCtx
	}
	return customer, nil
}

func normalizeLatLong(pair [2]string) (pq.StringArray, error) {
	if _, err := strconv.ParseFloat(pair[0], 64); err != nil {
		return nil, ErrInvalidLatLong
	}
	if _, err := strconv.ParseFloat(pair[1], 64); err != nil {
		return nil, ErrInvalidLatLong
	}
	return pq.StringArray{pair[0], pair[1]}, nil
}

func toResponse(address entities.SavedAddress) dto.SavedAddressResponse {
	latLong := [2]string{}
	if len(address.LatLong) >= 2 {
		latLong[0] = address.LatLong[0]
		latLong[1] = address.LatLong[1]
	}
	return dto.SavedAddressResponse{
		ID:              address.ID,
		Name:            address.Name,
		LatLong:         latLong,
		IsDefaultPickup: address.IsDefaultPickup,
		CreatedAt:       address.CreatedAt,
		UpdatedAt:       address.UpdatedAt,
	}
}
