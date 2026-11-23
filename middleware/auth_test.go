package middleware

import (
	"database/sql"
	"database/sql/driver"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"go-expense-tracker-api/config"
	"go-expense-tracker-api/repositories"
	"go-expense-tracker-api/services"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// FAKE SQL DRIVER: EVERY QUERY RETURNS ZERO ROWS (USER NOT FOUND), NO REAL DB NEEDED
var emptyDriverQueries int64

type emptyDriver struct{}

func (emptyDriver) Open(string) (driver.Conn, error) { return emptyConn{}, nil }

type emptyConn struct{}

func (emptyConn) Prepare(string) (driver.Stmt, error) { return emptyStmt{}, nil }
func (emptyConn) Close() error                        { return nil }
func (emptyConn) Begin() (driver.Tx, error)           { return nil, driver.ErrSkip }

type emptyStmt struct{}

func (emptyStmt) Close() error  { return nil }
func (emptyStmt) NumInput() int { return -1 }
func (emptyStmt) Exec([]driver.Value) (driver.Result, error) {
	return driver.RowsAffected(0), nil
}
func (emptyStmt) Query([]driver.Value) (driver.Rows, error) {
	atomic.AddInt64(&emptyDriverQueries, 1)
	return emptyRows{}, nil
}

type emptyRows struct{}

func (emptyRows) Columns() []string         { return []string{"id"} }
func (emptyRows) Close() error              { return nil }
func (emptyRows) Next([]driver.Value) error { return io.EOF }

func init() {
	sql.Register("auth_test_empty", emptyDriver{})
}

func TestAuthMiddleware_UserNotFound_Returns401AndStopsChain(t *testing.T) {
	gin.SetMode(gin.TestMode)

	sqlDB, err := sql.Open("auth_test_empty", "")
	if err != nil {
		t.Fatalf("open fake sql db: %v", err)
	}
	defer sqlDB.Close()

	db, err := gorm.Open(postgres.New(postgres.Config{Conn: sqlDB}), &gorm.Config{
		DisableAutomaticPing: true,
		Logger:               logger.Discard,
	})
	if err != nil {
		t.Fatalf("open gorm: %v", err)
	}

	cfg := &config.Config{JWT: config.JWTConfig{
		Secret:             "test-secret",
		ExpireHours:        1,
		RefreshSecret:      "test-refresh-secret",
		RefreshExpireHours: 1,
	}}
	jwtService := services.NewJWTService(cfg)
	userRepo := repositories.NewUserRepository(db)

	token, err := jwtService.GenerateToken(9999, "ghost@example.com", "jti-1")
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}

	nextCalled := false
	r := gin.New()
	r.GET("/protected", AuthMiddleware(jwtService, userRepo), func(c *gin.Context) {
		nextCalled = true
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if atomic.LoadInt64(&emptyDriverQueries) == 0 {
		t.Fatal("user lookup never reached the database layer; test is not exercising the user-not-found branch")
	}
	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
	if nextCalled {
		t.Error("next handler was called after user-not-found; AuthMiddleware must abort")
	}
}
