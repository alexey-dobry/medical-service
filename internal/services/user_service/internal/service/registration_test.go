package service

import (
	"bytes"
	"context"
	"errors"
	"testing"
	"time"

	logmocks "github.com/alexey-dobry/medical-service/internal/pkg/logger/mocks"
	pkgmodel "github.com/alexey-dobry/medical-service/internal/pkg/model"
	"github.com/alexey-dobry/medical-service/internal/services/user_service/internal/client"
	clientmocks "github.com/alexey-dobry/medical-service/internal/services/user_service/internal/client/mocks"
	"github.com/alexey-dobry/medical-service/internal/services/user_service/internal/domain/model"
	"github.com/alexey-dobry/medical-service/internal/services/user_service/internal/store"
	storemocks "github.com/alexey-dobry/medical-service/internal/services/user_service/internal/store/mocks"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

var fastRetry = RetryPolicy{Attempts: 3, InitialDelay: time.Millisecond, MaxDelay: 4 * time.Millisecond, Timeout: time.Second}

type env struct {
	registration *Registration
	auth         *clientmocks.MockClient
	users        *storemocks.MockUserRepository
	meta         *storemocks.MockMetaRepository
	search       *storemocks.MockSearchRepository
	photos       *storemocks.MockPhotosRepository
	logger       *logmocks.MockLogger
}

func newEnv(t *testing.T) *env {
	t.Helper()
	ctrl := gomock.NewController(t)
	e := &env{
		auth:   clientmocks.NewMockClient(ctrl),
		users:  storemocks.NewMockUserRepository(ctrl),
		meta:   storemocks.NewMockMetaRepository(ctrl),
		search: storemocks.NewMockSearchRepository(ctrl),
		photos: storemocks.NewMockPhotosRepository(ctrl),
		logger: logmocks.NewMockLogger(ctrl),
	}

	st := storemocks.NewMockStore(ctrl)
	st.EXPECT().User().Return(e.users).AnyTimes()
	st.EXPECT().Meta().Return(e.meta).AnyTimes()
	st.EXPECT().Search().Return(e.search).AnyTimes()
	st.EXPECT().Photos().Return(e.photos).AnyTimes()

	e.logger.EXPECT().WithFields(gomock.Any()).Return(e.logger).AnyTimes()

	e.registration = NewRegistration(e.auth, st, e.logger, fastRetry)
	return e
}

func patient() PatientRegistration {
	return PatientRegistration{
		User: model.User{
			ID:        uuid.NewString(),
			FirstName: "Ivan", MiddleName: "Ivanovich", LastName: "Ivanov",
			Phone: "+79990000000", Email: "ivan@example.com", Sex: "M",
			BirthDate: time.Date(1990, 1, 2, 0, 0, 0, 0, time.UTC),
		},
		Password: "secret-password",
	}
}

func tokens() client.RegisterPatientResponse {
	return client.RegisterPatientResponse{AccessToken: "access", RefreshToken: "refresh"}
}

func TestRegisterPatient_Success(t *testing.T) {
	e := newEnv(t)
	req := patient()

	e.auth.EXPECT().RegisterPatient(gomock.Any(), client.RegisterRequest{
		UserID: req.User.ID, Email: req.User.Email, Role: string(pkgmodel.PatientRole), Password: req.Password,
	}).Return(tokens(), nil)
	e.users.EXPECT().AddPatient(gomock.Any(), req.User).Return(nil)
	e.auth.EXPECT().DeleteUser(gomock.Any(), gomock.Any()).Times(0)

	resp, err := e.registration.RegisterPatient(context.Background(), req)
	require.NoError(t, err)
	assert.Equal(t, tokens(), resp)
}

func TestRegisterPatient_ProfileFailureDeletesCredentials(t *testing.T) {
	e := newEnv(t)
	req := patient()

	e.auth.EXPECT().RegisterPatient(gomock.Any(), gomock.Any()).Return(tokens(), nil)
	e.users.EXPECT().AddPatient(gomock.Any(), gomock.Any()).Return(errors.New("db is down"))
	e.auth.EXPECT().DeleteUser(gomock.Any(), client.DeleteRequest{UserID: req.User.ID}).Return(nil)
	e.logger.EXPECT().Warnf(gomock.Any(), gomock.Any()).AnyTimes()

	resp, err := e.registration.RegisterPatient(context.Background(), req)
	assert.Error(t, err)
	assert.Empty(t, resp.AccessToken, "tokens must not be returned for a rolled back registration")
}

func TestRegisterPatient_CompensationRetriesWithBackoff(t *testing.T) {
	e := newEnv(t)

	e.auth.EXPECT().RegisterPatient(gomock.Any(), gomock.Any()).Return(tokens(), nil)
	e.users.EXPECT().AddPatient(gomock.Any(), gomock.Any()).Return(errors.New("db is down"))
	gomock.InOrder(
		e.auth.EXPECT().DeleteUser(gomock.Any(), gomock.Any()).Return(errors.New("unavailable")),
		e.auth.EXPECT().DeleteUser(gomock.Any(), gomock.Any()).Return(errors.New("unavailable")),
		e.auth.EXPECT().DeleteUser(gomock.Any(), gomock.Any()).Return(nil),
	)
	e.logger.EXPECT().Warnf(gomock.Any(), gomock.Any()).AnyTimes()
	e.logger.EXPECT().Errorf(gomock.Any(), gomock.Any()).Times(0)

	_, err := e.registration.RegisterPatient(context.Background(), patient())
	assert.Error(t, err)
}

func TestRegisterPatient_CompensationFailureIsLogged(t *testing.T) {
	e := newEnv(t)
	req := patient()

	e.auth.EXPECT().RegisterPatient(gomock.Any(), gomock.Any()).Return(tokens(), nil)
	e.users.EXPECT().AddPatient(gomock.Any(), gomock.Any()).Return(errors.New("db is down"))
	e.auth.EXPECT().DeleteUser(gomock.Any(), gomock.Any()).Return(errors.New("unavailable")).Times(fastRetry.Attempts)

	var logged string
	e.logger.EXPECT().Errorf(gomock.Any(), gomock.Any()).DoAndReturn(func(format string, args ...any) {
		logged = format
		assert.Contains(t, args, req.User.ID, "the orphaned user id must be in the log")
	})

	_, err := e.registration.RegisterPatient(context.Background(), req)
	assert.Error(t, err)
	assert.Contains(t, logged, "COMPENSATION FAILED")
}

func TestRegisterPatient_CompensationSurvivesCancelledRequest(t *testing.T) {
	e := newEnv(t)
	ctx, cancel := context.WithCancel(context.Background())

	e.auth.EXPECT().RegisterPatient(gomock.Any(), gomock.Any()).Return(tokens(), nil)
	e.users.EXPECT().AddPatient(gomock.Any(), gomock.Any()).DoAndReturn(func(context.Context, model.User) error {
		cancel()
		return context.Canceled
	})
	e.auth.EXPECT().DeleteUser(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, _ client.DeleteRequest) error {
		assert.NoError(t, ctx.Err(), "compensation must not inherit the request cancellation")
		return nil
	})
	e.logger.EXPECT().Warnf(gomock.Any(), gomock.Any()).AnyTimes()

	_, err := e.registration.RegisterPatient(ctx, patient())
	assert.Error(t, err)
}

func TestRegisterPatient_AuthFailureNeedsNoCompensation(t *testing.T) {
	e := newEnv(t)
	e.auth.EXPECT().RegisterPatient(gomock.Any(), gomock.Any()).Return(client.RegisterPatientResponse{}, errors.New("already exists"))
	e.users.EXPECT().AddPatient(gomock.Any(), gomock.Any()).Times(0)
	e.auth.EXPECT().DeleteUser(gomock.Any(), gomock.Any()).Times(0)

	_, err := e.registration.RegisterPatient(context.Background(), patient())
	assert.Error(t, err)
}

func TestRegisterPatient_WithPhoto(t *testing.T) {
	e := newEnv(t)
	req := patient()
	req.Photo = &PhotoUpload{Name: "me.png", MimeType: "image/png", Size: 3, Content: bytes.NewReader([]byte("png"))}

	var photo model.Photo
	e.auth.EXPECT().RegisterPatient(gomock.Any(), gomock.Any()).Return(tokens(), nil)
	e.photos.EXPECT().Put(gomock.Any(), gomock.Any(), int64(3), "image/png").Return(nil)
	e.users.EXPECT().AddPatient(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, u model.User) error {
		assert.NotEmpty(t, u.PhotoID, "profile must reference the uploaded photo")
		return nil
	})
	e.meta.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, p model.Photo) error {
		photo = p
		return nil
	})

	_, err := e.registration.RegisterPatient(context.Background(), req)
	require.NoError(t, err)
	assert.Equal(t, req.User.ID, photo.UserID)
	assert.Equal(t, "photos/"+photo.ID, photo.StorageKey)
}

func TestRegisterPatient_InvalidPhotoIsRejectedBeforeAnything(t *testing.T) {
	e := newEnv(t)
	req := patient()
	req.Photo = &PhotoUpload{Name: "x.exe", MimeType: "application/octet-stream", Size: 10, Content: bytes.NewReader(nil)}
	e.auth.EXPECT().RegisterPatient(gomock.Any(), gomock.Any()).Times(0)

	_, err := e.registration.RegisterPatient(context.Background(), req)
	assert.ErrorIs(t, err, ErrInvalidPhoto)
}

func TestRegisterPatient_PhotoMetaFailureCleansUp(t *testing.T) {
	e := newEnv(t)
	req := patient()
	req.Photo = &PhotoUpload{Name: "me.png", MimeType: "image/png", Size: 3, Content: bytes.NewReader([]byte("png"))}

	e.auth.EXPECT().RegisterPatient(gomock.Any(), gomock.Any()).Return(tokens(), nil)
	e.photos.EXPECT().Put(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
	e.users.EXPECT().AddPatient(gomock.Any(), gomock.Any()).Return(nil)
	e.meta.EXPECT().Create(gomock.Any(), gomock.Any()).Return(errors.New("db is down"))
	e.users.EXPECT().DeleteUser(gomock.Any(), uuid.MustParse(req.User.ID)).Return(nil)
	e.photos.EXPECT().Delete(gomock.Any()).Return(nil)
	e.auth.EXPECT().DeleteUser(gomock.Any(), client.DeleteRequest{UserID: req.User.ID}).Return(nil)
	e.logger.EXPECT().Warnf(gomock.Any(), gomock.Any()).AnyTimes()

	_, err := e.registration.RegisterPatient(context.Background(), req)
	assert.Error(t, err)
}

func TestRegisterDoctor_IndexFailureRollsBackEverything(t *testing.T) {
	e := newEnv(t)
	p := patient()
	req := DoctorRegistration{
		User:     p.User,
		Doctor:   model.DoctorAdditionalData{UserID: p.User.ID, Specialty: "surgeon", WorkExperience: 5, Description: "d", Services: []string{"x"}},
		Password: "secret-password",
		Photo:    &PhotoUpload{Name: "doc.jpg", MimeType: "image/jpeg", Size: 3, Content: bytes.NewReader([]byte("jpg"))},
	}

	e.auth.EXPECT().RegisterDoctor(gomock.Any(), client.RegisterRequest{
		UserID: p.User.ID, Email: p.User.Email, Role: string(pkgmodel.DoctorRole), Password: req.Password,
	}).Return(nil)
	e.photos.EXPECT().Put(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
	e.users.EXPECT().AddDoctor(gomock.Any(), gomock.Any(), req.Doctor).Return(nil)
	e.meta.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil)
	e.search.EXPECT().AddDoctor(gomock.Any(), gomock.Any()).Return(errors.New("elasticsearch is down"))
	e.users.EXPECT().DeleteUser(gomock.Any(), uuid.MustParse(p.User.ID)).Return(nil)
	e.photos.EXPECT().Delete(gomock.Any()).Return(nil)
	e.auth.EXPECT().DeleteUser(gomock.Any(), client.DeleteRequest{UserID: p.User.ID}).Return(nil)
	e.logger.EXPECT().Warnf(gomock.Any(), gomock.Any()).AnyTimes()

	err := e.registration.RegisterDoctor(context.Background(), req)
	assert.Error(t, err)
}

func TestRegisterDoctor_DuplicateEmailInProfile(t *testing.T) {
	e := newEnv(t)
	p := patient()
	req := DoctorRegistration{
		User:   p.User,
		Doctor: model.DoctorAdditionalData{UserID: p.User.ID},
		Photo:  &PhotoUpload{Name: "doc.jpg", MimeType: "image/jpeg", Size: 3, Content: bytes.NewReader([]byte("jpg"))},
	}

	e.auth.EXPECT().RegisterDoctor(gomock.Any(), gomock.Any()).Return(nil)
	e.photos.EXPECT().Put(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
	e.users.EXPECT().AddDoctor(gomock.Any(), gomock.Any(), gomock.Any()).Return(store.ErrAlreadyExists)
	e.photos.EXPECT().Delete(gomock.Any()).Return(nil)
	e.auth.EXPECT().DeleteUser(gomock.Any(), gomock.Any()).Return(nil)
	e.logger.EXPECT().Warnf(gomock.Any(), gomock.Any()).AnyTimes()

	err := e.registration.RegisterDoctor(context.Background(), req)
	assert.ErrorIs(t, err, ErrAlreadyExists)
}

func TestRetryBackoffGrowsAndIsCapped(t *testing.T) {
	var calls []time.Time
	policy := RetryPolicy{Attempts: 4, InitialDelay: 10 * time.Millisecond, MaxDelay: 20 * time.Millisecond}

	err := retry(context.Background(), policy, func(context.Context) error {
		calls = append(calls, time.Now())
		return errors.New("fail")
	})
	require.Error(t, err)
	require.Len(t, calls, 4)

	gaps := []time.Duration{calls[1].Sub(calls[0]), calls[2].Sub(calls[1]), calls[3].Sub(calls[2])}
	assert.GreaterOrEqual(t, gaps[0], 10*time.Millisecond)
	assert.GreaterOrEqual(t, gaps[1], 20*time.Millisecond, "delay doubles")
	assert.Less(t, gaps[2], 40*time.Millisecond, "delay is capped by MaxDelay")
}
