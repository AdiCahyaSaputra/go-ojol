package middlewares

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/AdiCahyaSaputra/go-ojol/backend/auth/pkg/constants"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type stubEnforcer struct {
	allow bool
	err   error
	sub   string
	obj   string
	act   string
}

func (s *stubEnforcer) Enforce(rvals ...interface{}) (bool, error) {
	if len(rvals) >= 3 {
		s.sub, _ = rvals[0].(string)
		s.obj, _ = rvals[1].(string)
		s.act, _ = rvals[2].(string)
	}
	return s.allow, s.err
}

func (s *stubEnforcer) LoadPolicy() error {
	return nil
}

func TestAuthorize_AllowsWhenEnforcerAllows(t *testing.T) {
	gin.SetMode(gin.TestMode)
	enforcer := &stubEnforcer{allow: true}

	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/api/user/me", nil)
	ctx.Set("email", "ada@example.com")

	Authorize(enforcer, constants.ENUM_RESOURCE_USER, constants.ENUM_ACTION_READ)(ctx)

	assert.False(t, ctx.IsAborted())
	assert.Equal(t, "ada@example.com", enforcer.sub)
	assert.Equal(t, constants.ENUM_RESOURCE_USER, enforcer.obj)
	assert.Equal(t, constants.ENUM_ACTION_READ, enforcer.act)
}

func TestAuthorize_DeniesWhenEnforcerDenies(t *testing.T) {
	gin.SetMode(gin.TestMode)
	enforcer := &stubEnforcer{allow: false}

	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = httptest.NewRequest(http.MethodDelete, "/api/user/1", nil)
	ctx.Set("email", "ada@example.com")

	Authorize(enforcer, constants.ENUM_RESOURCE_USER, constants.ENUM_ACTION_DELETE)(ctx)

	assert.True(t, ctx.IsAborted())
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestAuthorize_DeniesWhenEmailMissing(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/api/user/me", nil)

	Authorize(&stubEnforcer{allow: true}, constants.ENUM_RESOURCE_USER, constants.ENUM_ACTION_READ)(ctx)

	assert.True(t, ctx.IsAborted())
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestAuthorize_DeniesWhenEnforcerErrors(t *testing.T) {
	gin.SetMode(gin.TestMode)
	enforcer := &stubEnforcer{err: assert.AnError}

	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/api/user/me", nil)
	ctx.Set("email", "ada@example.com")

	Authorize(enforcer, constants.ENUM_RESOURCE_USER, constants.ENUM_ACTION_READ)(ctx)

	assert.True(t, ctx.IsAborted())
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestAuthorize_UpdateAction(t *testing.T) {
	gin.SetMode(gin.TestMode)
	enforcer := &stubEnforcer{allow: true}

	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = httptest.NewRequest(http.MethodPut, "/api/user/1", nil)
	ctx.Set("email", "driver@example.com")

	Authorize(enforcer, constants.ENUM_RESOURCE_USER, constants.ENUM_ACTION_UPDATE)(ctx)

	require.False(t, ctx.IsAborted())
	assert.Equal(t, constants.ENUM_ACTION_UPDATE, enforcer.act)
}
