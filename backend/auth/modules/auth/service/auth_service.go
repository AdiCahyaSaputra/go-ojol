package service

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/AdiCahyaSaputra/go-ojol/backend/auth/database/entities"
	"github.com/AdiCahyaSaputra/go-ojol/backend/auth/modules/auth/dto"
	authrepo "github.com/AdiCahyaSaputra/go-ojol/backend/auth/modules/auth/repository"
	casbinrepo "github.com/AdiCahyaSaputra/go-ojol/backend/auth/modules/casbin/repository"
	userDto "github.com/AdiCahyaSaputra/go-ojol/backend/auth/modules/user/dto"
	"github.com/AdiCahyaSaputra/go-ojol/backend/auth/modules/user/repository"
	pkgcasbin "github.com/AdiCahyaSaputra/go-ojol/backend/auth/pkg/casbin"
	"github.com/AdiCahyaSaputra/go-ojol/backend/auth/pkg/constants"
	"github.com/AdiCahyaSaputra/go-ojol/backend/auth/pkg/helpers"
	"github.com/AdiCahyaSaputra/go-ojol/backend/auth/pkg/uploadthing"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

const maxProfilePictureBytes = 2 * 1024 * 1024

var allowedProfilePictureMIME = map[string]struct{}{
	"image/jpeg": {},
	"image/png":  {},
	"image/webp": {},
}

type AuthService interface {
	Register(ctx context.Context, req userDto.UserCreateRequest) (userDto.UserResponse, error)
	Login(ctx context.Context, req userDto.UserLoginRequest, meta dto.LoginMeta) (dto.TokenResponse, error)
	Refresh(ctx context.Context, req dto.RefreshTokenRequest) (dto.TokenResponse, error)
	Logout(ctx context.Context, sessionID string) error
	LogoutAll(ctx context.Context, userID string) error
	ListSessions(ctx context.Context, userID string, currentSessionID string) ([]dto.SessionResponse, error)
	RevokeSession(ctx context.Context, userID string, sessionID string) error
}

type authService struct {
	userRepository    repository.UserRepository
	casbinRepository  casbinrepo.CasbinRepository
	sessionRepository authrepo.SessionRepository
	jwtService        JWTService
	enforcer          pkgcasbin.Enforcer
	uploadClient      uploadthing.Client
	db                *gorm.DB
}

func NewAuthService(
	userRepo repository.UserRepository,
	casbinRepo casbinrepo.CasbinRepository,
	sessionRepo authrepo.SessionRepository,
	jwtService JWTService,
	enforcer pkgcasbin.Enforcer,
	uploadClient uploadthing.Client,
	db *gorm.DB,
) AuthService {
	return &authService{
		userRepository:    userRepo,
		casbinRepository:  casbinRepo,
		sessionRepository: sessionRepo,
		jwtService:        jwtService,
		enforcer:          enforcer,
		uploadClient:      uploadClient,
		db:                db,
	}
}

func hashRefreshToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func optionalTrimmedString(value string) *string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func (s *authService) Register(ctx context.Context, req userDto.UserCreateRequest) (userDto.UserResponse, error) {
	if req.Role != constants.ENUM_ROLE_CUSTOMER && req.Role != constants.ENUM_ROLE_DRIVER {
		return userDto.UserResponse{}, userDto.ErrInvalidRole
	}

	if req.Role == constants.ENUM_ROLE_DRIVER {
		if strings.TrimSpace(req.VehicleName) == "" ||
			strings.TrimSpace(req.VehicleLicenseNumber) == "" ||
			req.VehicleMaxSize <= 0 ||
			(req.VehicleType != string(entities.VehicleTypeCar) && req.VehicleType != string(entities.VehicleTypeMotorcycle)) {
			return userDto.UserResponse{}, userDto.ErrVehicleRequired
		}
	}

	_, isExist, err := s.userRepository.CheckEmail(ctx, s.db, req.Email)
	if err != nil && err != gorm.ErrRecordNotFound {
		return userDto.UserResponse{}, err
	}

	if isExist {
		return userDto.UserResponse{}, userDto.ErrEmailAlreadyExists
	}

	var profilePictureURL *string
	if req.ProfilePicture != nil {
		url, err := s.uploadProfilePicture(ctx, req)
		if err != nil {
			return userDto.UserResponse{}, err
		}
		profilePictureURL = &url
	}

	// NOTE: We already hash the password in BeforeCreate hook
	// hashedPassword, err := helpers.HashPassword(req.Password)
	// if err != nil {
	// 	return userDto.UserResponse{}, err
	// }

	user := entities.User{
		ID:       uuid.New(),
		Email:    req.Email,
		Password: req.Password,
	}

	var (
		createdUser     entities.User
		createdCustomer *entities.Customer
		createdDriver   *entities.Driver
	)

	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		createdUser, err = s.userRepository.Register(ctx, tx, user)
		if err != nil {
			return err
		}

		if err := s.casbinRepository.AddGroupingPolicy(ctx, tx, createdUser.Email, req.Role); err != nil {
			return err
		}

		switch req.Role {
		case constants.ENUM_ROLE_CUSTOMER:
			customer, createErr := s.userRepository.CreateCustomer(ctx, tx, entities.Customer{
				UserID:            createdUser.ID,
				Name:              req.Name,
				PhoneNumber:       req.PhoneNumber,
				ProfilePictureUrl: profilePictureURL,
			})
			if createErr != nil {
				return createErr
			}
			createdCustomer = &customer
		case constants.ENUM_ROLE_DRIVER:
			licenseNumber := strings.TrimSpace(req.VehicleLicenseNumber)
			existing, findErr := s.userRepository.FindVehicleByLicenseNumber(ctx, tx, licenseNumber)
			if findErr != nil && findErr != gorm.ErrRecordNotFound {
				return findErr
			}
			if existing != nil {
				return userDto.ErrLicenseNumberExists
			}

			vehicle, createVehicleErr := s.userRepository.CreateVehicle(ctx, tx, entities.Vehicle{
				Name:          strings.TrimSpace(req.VehicleName),
				LicenseNumber: licenseNumber,
				MaxSize:       req.VehicleMaxSize,
				Type:          entities.VehicleType(req.VehicleType),
			})
			if createVehicleErr != nil {
				return createVehicleErr
			}

			driver, createErr := s.userRepository.CreateDriver(ctx, tx, entities.Driver{
				UserID:            createdUser.ID,
				VehicleID:         vehicle.ID,
				Name:              req.Name,
				PhoneNumber:       req.PhoneNumber,
				Address:           "",
				ProfilePictureUrl: profilePictureURL,
			})
			if createErr != nil {
				return createErr
			}
			createdDriver = &driver
		}

		return nil
	})
	if err != nil {
		return userDto.UserResponse{}, err
	}

	if err := s.enforcer.LoadPolicy(); err != nil {
		return userDto.UserResponse{}, err
	}

	resp := userDto.UserResponse{
		ID:        createdUser.ID.String(),
		Email:     createdUser.Email,
		Role:      req.Role,
		CreatedAt: createdUser.CreatedAt,
		UpdatedAt: createdUser.UpdatedAt,
	}

	if createdCustomer != nil {
		resp.Customer = &userDto.CustomerProfileResponse{
			ID:                createdCustomer.ID.String(),
			Name:              createdCustomer.Name,
			PhoneNumber:       createdCustomer.PhoneNumber,
			ProfilePictureUrl: createdCustomer.ProfilePictureUrl,
		}
	}

	if createdDriver != nil {
		// Intentionally omit vehicle from register response.
		resp.Driver = &userDto.DriverProfileResponse{
			ID:                createdDriver.ID.String(),
			Name:              createdDriver.Name,
			PhoneNumber:       createdDriver.PhoneNumber,
			Address:           createdDriver.Address,
			ProfilePictureUrl: createdDriver.ProfilePictureUrl,
		}
	}

	return resp, nil
}

func (s *authService) uploadProfilePicture(ctx context.Context, req userDto.UserCreateRequest) (string, error) {
	fileHeader := req.ProfilePicture
	if fileHeader.Size <= 0 || fileHeader.Size > maxProfilePictureBytes {
		return "", userDto.ErrInvalidProfilePicture
	}

	file, err := fileHeader.Open()
	if err != nil {
		return "", userDto.ErrInvalidProfilePicture
	}
	defer file.Close()

	head := make([]byte, 512)
	n, readErr := io.ReadFull(file, head)
	if readErr != nil && readErr != io.ErrUnexpectedEOF && readErr != io.EOF {
		return "", userDto.ErrInvalidProfilePicture
	}

	contentType := http.DetectContentType(head[:n])
	if idx := strings.Index(contentType, ";"); idx >= 0 {
		contentType = strings.TrimSpace(contentType[:idx])
	}
	if _, ok := allowedProfilePictureMIME[contentType]; !ok {
		return "", userDto.ErrInvalidProfilePicture
	}

	reader := io.MultiReader(bytes.NewReader(head[:n]), file)
	url, err := s.uploadClient.Upload(ctx, fileHeader.Filename, contentType, fileHeader.Size, reader)
	if err != nil {
		return "", userDto.ErrUploadProfilePicture
	}
	return url, nil
}

func (s *authService) issueTokensForUser(ctx context.Context, user entities.User, role string, meta dto.LoginMeta) (dto.TokenResponse, error) {
	refreshToken, expiresAt, err := s.jwtService.GenerateRefreshToken()
	if err != nil {
		return dto.TokenResponse{}, err
	}

	session := entities.Session{
		ID:               uuid.New(),
		UserID:           user.ID,
		RefreshTokenHash: hashRefreshToken(refreshToken),
		ExpiresAt:        expiresAt,
		UserAgent:        optionalTrimmedString(meta.UserAgent),
		IP:               optionalTrimmedString(meta.IP),
	}

	created, err := s.sessionRepository.Create(ctx, s.db, session)
	if err != nil {
		return dto.TokenResponse{}, err
	}

	accessToken, err := s.jwtService.GenerateAccessToken(user.ID.String(), user.Email, role, created.ID.String())
	if err != nil {
		return dto.TokenResponse{}, err
	}

	return dto.TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		Role:         role,
	}, nil
}

func (s *authService) Login(ctx context.Context, req userDto.UserLoginRequest, meta dto.LoginMeta) (dto.TokenResponse, error) {
	user, err := s.userRepository.GetUserByEmail(ctx, s.db, req.Email)
	if err != nil {
		return dto.TokenResponse{}, userDto.ErrEmailNotFound
	}

	isValid, err := helpers.CheckPassword(user.Password, []byte(req.Password))
	if err != nil || !isValid {
		return dto.TokenResponse{}, dto.ErrInvalidCredentials
	}

	roles, err := s.casbinRepository.GetRolesForUser(ctx, s.db, user.Email)
	if err != nil {
		return dto.TokenResponse{}, err
	}
	if len(roles) == 0 {
		return dto.TokenResponse{}, userDto.ErrRoleNotAssigned
	}

	return s.issueTokensForUser(ctx, user, roles[0], meta)
}

func (s *authService) Refresh(ctx context.Context, req dto.RefreshTokenRequest) (dto.TokenResponse, error) {
	hash := hashRefreshToken(req.RefreshToken)
	session, err := s.sessionRepository.FindByRefreshTokenHash(ctx, s.db, hash)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return dto.TokenResponse{}, dto.ErrRefreshTokenNotFound
		}
		return dto.TokenResponse{}, err
	}

	if session.RevokedAt != nil {
		return dto.TokenResponse{}, dto.ErrSessionRevoked
	}
	if !session.ExpiresAt.After(time.Now()) {
		_ = s.sessionRepository.RevokeByID(ctx, s.db, session.ID)
		return dto.TokenResponse{}, dto.ErrRefreshTokenExpired
	}

	user, err := s.userRepository.GetUserById(ctx, s.db, session.UserID.String())
	if err != nil {
		return dto.TokenResponse{}, err
	}

	roles, err := s.casbinRepository.GetRolesForUser(ctx, s.db, user.Email)
	if err != nil {
		return dto.TokenResponse{}, err
	}
	if len(roles) == 0 {
		return dto.TokenResponse{}, userDto.ErrRoleNotAssigned
	}
	role := roles[0]

	newRefresh, expiresAt, err := s.jwtService.GenerateRefreshToken()
	if err != nil {
		return dto.TokenResponse{}, err
	}

	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return s.sessionRepository.UpdateRefreshToken(ctx, tx, session.ID, hashRefreshToken(newRefresh), expiresAt)
	})
	if err != nil {
		return dto.TokenResponse{}, err
	}

	accessToken, err := s.jwtService.GenerateAccessToken(user.ID.String(), user.Email, role, session.ID.String())
	if err != nil {
		return dto.TokenResponse{}, err
	}

	return dto.TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: newRefresh,
		Role:         role,
	}, nil
}

func (s *authService) Logout(ctx context.Context, sessionID string) error {
	id, err := uuid.Parse(sessionID)
	if err != nil {
		return dto.ErrSessionNotFound
	}
	return s.sessionRepository.RevokeByID(ctx, s.db, id)
}

func (s *authService) LogoutAll(ctx context.Context, userID string) error {
	id, err := uuid.Parse(userID)
	if err != nil {
		return dto.ErrSessionNotFound
	}
	return s.sessionRepository.RevokeAllByUserID(ctx, s.db, id)
}

func (s *authService) ListSessions(ctx context.Context, userID string, currentSessionID string) ([]dto.SessionResponse, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, dto.ErrSessionNotFound
	}

	sessions, err := s.sessionRepository.ListByUserID(ctx, s.db, uid)
	if err != nil {
		return nil, err
	}

	result := make([]dto.SessionResponse, 0, len(sessions))
	for _, session := range sessions {
		result = append(result, dto.SessionResponse{
			ID:        session.ID.String(),
			UserAgent: session.UserAgent,
			IP:        session.IP,
			CreatedAt: session.CreatedAt,
			ExpiresAt: session.ExpiresAt,
			RevokedAt: session.RevokedAt,
			IsCurrent: session.ID.String() == currentSessionID,
		})
	}
	return result, nil
}

func (s *authService) RevokeSession(ctx context.Context, userID string, sessionID string) error {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return dto.ErrSessionNotFound
	}
	sid, err := uuid.Parse(sessionID)
	if err != nil {
		return dto.ErrSessionNotFound
	}

	session, err := s.sessionRepository.FindByID(ctx, s.db, sid)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return dto.ErrSessionNotFound
		}
		return err
	}
	if session.UserID != uid {
		return dto.ErrSessionNotFound
	}

	return s.sessionRepository.RevokeByID(ctx, s.db, sid)
}
