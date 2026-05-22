package handler

// import (
// 	"net/http"
// 	"net/http/httptest"
// 	"strconv"
// 	"strings"
// 	"testing"

// 	"github.com/gin-gonic/gin"
// 	"github.com/stretchr/testify/assert"
// )

// func setupRouter() *gin.Engine {
// 	gin.SetMode(gin.TestMode)

// 	r := gin.Default()

// 	r.GET("/health", func(c *gin.Context) {
// 		c.JSON(http.StatusOK, gin.H{
// 			"status": "ok",
// 		})
// 	})

// 	r.GET("/api/v1/tasks", func(c *gin.Context) {
// 		c.JSON(http.StatusOK, gin.H{
// 			"message": "get all tasks successfully",
// 		})
// 	})

// 	r.GET("/api/v1/tasks/:id", func(c *gin.Context) {
// 		taskID := c.Param("id")
// 		idInt, err := strconv.Atoi(taskID)

// 		if err != nil {
// 			c.JSON(http.StatusBadRequest, gin.H{
// 				"error": "Invalid task ID",
// 			})
// 			return
// 		}

// 		if idInt >= 3 {
// 			c.JSON(http.StatusNotFound, gin.H{
// 				"error": "Task not found",
// 			})
// 			return
// 		}

// 		if idInt == 1 || idInt == 2 {
// 			c.JSON(http.StatusOK, gin.H{
// 				"message": "Task found",
// 			})
// 			return
// 		}
// 	})
// 	r.POST("/api/v1/tasks", func(c *gin.Context) {
// 	var req struct {
// 	Title string `json:"title" binding:"required"` // để bắt lỗi nếu title bị bỏ trống (binding:"required")
// }

// 	err := c.ShouldBindJSON(&req)

// 	if err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{
// 			"error": "Invalid body",
// 		})
// 		return
// 	}

// 	c.JSON(http.StatusCreated, gin.H{
// 		"message": req.Title,
// 	})
// })
// r.PUT("/api/v1/tasks/:id", func(c *gin.Context) {
// 	taskID := c.Param("id")

// 	idInt, err := strconv.Atoi(taskID)
// 	if err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{
// 			"error": "Invalid task ID",
// 		})
// 		return
// 	}

// 	if idInt >= 3 {
// 		c.JSON(http.StatusNotFound, gin.H{
// 			"error": "Task not found",
// 		})
// 		return
// 	}

// 	var req struct {
// 		Title string `json:"title" binding:"required"`
// 	}

// 	if err := c.ShouldBindJSON(&req); err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{
// 			"error": "Invalid body",
// 		})
// 		return
// 	}

// 	c.JSON(http.StatusOK, gin.H{
// 		"message": "Task updated",
// 		"title":   req.Title,
// 	})
// })
// r.DELETE("/api/v1/tasks/:id", func(c *gin.Context) {
// 	taskID := c.Param("id")

// 	idInt, err := strconv.Atoi(taskID)
// 	if err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{
// 			"error": "Invalid task ID",
// 		})
// 		return
// 	}

// 	if idInt >= 3 {
// 		c.JSON(http.StatusNotFound, gin.H{
// 			"error": "Task not found",
// 		})
// 		return
// 	}

// 	c.JSON(http.StatusOK, gin.H{
// 		"message": "Task deleted",
// 	})
// })
// 	return r
// }

// func TestHealth_ShouldReturn200(t *testing.T) {

// 	// Arrange
// 	router := setupRouter()

// 	req, _ := http.NewRequest(
// 		http.MethodGet,
// 		"/health",
// 		nil,
// 	)

// 	w := httptest.NewRecorder()

// 	// Act
// 	router.ServeHTTP(w, req)

// 	// Assert
// 	assert.Equal(t, http.StatusOK, w.Code)
// 	assert.Contains(t, w.Body.String(), "ok")
// }
// func TestGetAllTasks_ShouldReturn200(t *testing.T) {
// 	// Arrange
// 	router := setupRouter()
// 	req, _ := http.NewRequest(
// 		http.MethodGet,
// 		"/api/v1/tasks",
// 		nil,
// 	)
// 	w := httptest.NewRecorder()
// 	// Act
// 	router.ServeHTTP(w, req)

// 	// Assert
// 	assert.Equal(t, http.StatusOK, w.Code)
// 	assert.Contains(t, w.Body.String(), "get all tasks successfully")
// }
// func TestGetTaskById_ShouldReturn200(t *testing.T) {
// 	// Arrange
// 	router := setupRouter()
// 	req, _ := http.NewRequest(
// 		http.MethodGet,
// 		"/api/v1/tasks/1",
// 		nil,
// 	)
// 	w := httptest.NewRecorder()
// 	// Act
// 	router.ServeHTTP(w, req)
// 	// Assert
// 	assert.Equal(t, http.StatusOK, w.Code)
// 	assert.Contains(t, w.Body.String(), "Task found")
// }
// func TestGetTaskById_ShouldReturn400(t *testing.T) {
// 	// Arrange
// 	router := setupRouter()
// 	req, _ := http.NewRequest(
// 		http.MethodGet,
// 		"/api/v1/tasks/abc",
// 		nil,
// 	)
// 	w := httptest.NewRecorder()
// 	// Act
// 	router.ServeHTTP(w, req)
// 	// Assert
// 	assert.Equal(t, http.StatusBadRequest, w.Code)
// 	assert.Contains(t, w.Body.String(), "Invalid task ID")
// }
// func TestGetTaskById_ShouldReturn404(t *testing.T) {
// 	// Arrange
// 	router := setupRouter()
// 	req, _ := http.NewRequest(
// 		http.MethodGet,
// 		"/api/v1/tasks/3",
// 		nil,
// 	)
// 	w := httptest.NewRecorder()
// 	// Act
// 	router.ServeHTTP(w, req)
// 	// Assert
// 	assert.Equal(t, http.StatusNotFound, w.Code)
// 	assert.Contains(t, w.Body.String(), "Task not found")
// }

// func TestPostTask_ShouldReturn201_WhenBodyValid(t *testing.T) {
// 	// Arrange
// 	router := setupRouter()

// 	body := strings.NewReader(`{
// 		"title": "Learn Go"
// 	}`)

// 	req, _ := http.NewRequest(
// 		http.MethodPost,
// 		"/api/v1/tasks",
// 		body,
// 	)

// 	req.Header.Set("Content-Type", "application/json")

// 	w := httptest.NewRecorder()

// 	// Act
// 	router.ServeHTTP(w, req)

// 	// Assert
// 	assert.Equal(t, http.StatusCreated, w.Code)
// 	assert.Contains(t, w.Body.String(), "Learn Go")
// }
// func TestPostTask_ShouldReturn400_WhenBodyInvalid(t *testing.T) {
// 	// Arrange
// 	router := setupRouter()

// 	body := strings.NewReader(`{
// 		"title": ""
// 	}`)

// 	req, _ := http.NewRequest(
// 		http.MethodPost,
// 		"/api/v1/tasks",
// 		body,
// 	)

// 	req.Header.Set("Content-Type", "application/json")

// 	w := httptest.NewRecorder()

// 	// Act
// 	router.ServeHTTP(w, req)

// 	// Assert
// 	assert.Equal(t, http.StatusBadRequest, w.Code)
// 	assert.Contains(t, w.Body.String(), "Invalid body")
// }

// func TestPutTask_ShouldReturn200(t *testing.T) {
// 	router := setupRouter()

// 	body := strings.NewReader(`{
// 		"title": "Learn Go"
// 	}`)

// 	req, _ := http.NewRequest(
// 		http.MethodPut,
// 		"/api/v1/tasks/1",
// 		body,
// 	)

// 	req.Header.Set("Content-Type", "application/json")

// 	w := httptest.NewRecorder()

// 	router.ServeHTTP(w, req)

// 	assert.Equal(t, http.StatusOK, w.Code)
// 	assert.Contains(t, w.Body.String(), "Task updated")
// 	assert.Contains(t, w.Body.String(), "Learn Go")
// }
// func TestPutTask_ShouldReturn400_WhenBodyInvalid(t *testing.T) {
// 	router := setupRouter()

// 	body := strings.NewReader(`{
// 		"title": ""
// 	}`)

// 	req, _ := http.NewRequest(
// 		http.MethodPut,
// 		"/api/v1/tasks/1",
// 		body,
// 	)

// 	req.Header.Set("Content-Type", "application/json")

// 	w := httptest.NewRecorder()

// 	router.ServeHTTP(w, req)

// 	assert.Equal(t, http.StatusBadRequest, w.Code)
// 	assert.Contains(t, w.Body.String(), "Invalid body")
// }	
// func TestPutTask_ShouldReturn400_WhenIdInvalid(t *testing.T) {
// 	router := setupRouter()
// 	body := strings.NewReader(`{
// 		"title": "Learn Go"
// 	}`)
// 	req, _ := http.NewRequest(
// 		http.MethodPut,
// 		"/api/v1/tasks/abc",
// 		body,
// 	)
// 	req.Header.Set("Content-Type", "application/json")
// 	w := httptest.NewRecorder()
// 	router.ServeHTTP(w, req)
// 	assert.Equal(t, http.StatusBadRequest, w.Code)
// 	assert.Contains(t, w.Body.String(), "Invalid task ID")
// }
// func TestPutTask_ShouldReturn404_WhenIdNotFound(t *testing.T) {
// 	router := setupRouter()
// 	body := strings.NewReader(`{
// 		"title": "Learn Go"
// 	}`)
// 	req, _ := http.NewRequest(
// 		http.MethodPut,
// 		"/api/v1/tasks/3",
// 		body,
// 	)
// 	req.Header.Set("Content-Type", "application/json")
// 	w := httptest.NewRecorder()
// 	router.ServeHTTP(w, req)
// 	assert.Equal(t, http.StatusNotFound, w.Code)
// 	assert.Contains(t, w.Body.String(), "Task not found")
// }
// func TestDeleteTask_ShouldReturn200(t *testing.T) {
// 	router := setupRouter()
// 	req, _ := http.NewRequest(
// 		http.MethodDelete,
// 		"/api/v1/tasks/1",
// 		nil,
// 	)
// 	w := httptest.NewRecorder()
// 	router.ServeHTTP(w, req)
// 	assert.Equal(t, http.StatusOK, w.Code)
// 	assert.Contains(t, w.Body.String(), "Task deleted")
// }
// func TestDeleteTask_ShouldReturn400_WhenIdInvalid(t *testing.T) {
// 	router := setupRouter()
// 	req, _ := http.NewRequest(
// 		http.MethodDelete,
// 		"/api/v1/tasks/abc",
// 		nil,
// 	)
// 	w := httptest.NewRecorder()
// 	router.ServeHTTP(w, req)
// 	assert.Equal(t, http.StatusBadRequest, w.Code)
// 	assert.Contains(t, w.Body.String(), "Invalid task ID")
// }
// func TestDeleteTask_ShouldReturn404_WhenIdNotFound(t *testing.T) {
// 	router := setupRouter()
// 	req, _ := http.NewRequest(
// 		http.MethodDelete,
// 		"/api/v1/tasks/3",
// 		nil,
// 	)
// 	w := httptest.NewRecorder()
// 	router.ServeHTTP(w, req)
// 	assert.Equal(t, http.StatusNotFound, w.Code)
// 	assert.Contains(t, w.Body.String(), "Task not found")
// }