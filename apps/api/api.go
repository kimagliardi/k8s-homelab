package main

import (
	"log/slog"
	"math/rand"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type Config struct {
	LatencyMS int
	FailRate  float64
	MemoryMB  int
}

var (
	config Config
	logger *slog.Logger
)

func init() {
	logger = slog.New(slog.NewJSONHandler(os.Stdout, nil))

	config.LatencyMS = getEnvInt("LATENCY_MS", 0)
	config.FailRate = getEnvFloat("FAIL_RATE", 0.0)
	config.MemoryMB = getEnvInt("MEMORY_MB", 0)

	logger.Info("config loaded",
		"latency_ms", config.LatencyMS,
		"fail_rate", config.FailRate,
		"memory_mb", config.MemoryMB,
	)
}

func getEnvInt(key string, defaultVal int) int {
	if val := os.Getenv(key); val != "" {
		if i, err := strconv.Atoi(val); err == nil {
			return i
		}
	}
	return defaultVal
}

func getEnvFloat(key string, defaultVal float64) float64 {
	if val := os.Getenv(key); val != "" {
		if f, err := strconv.ParseFloat(val, 64); err == nil {
			return f
		}
	}
	return defaultVal
}

func setupRouter() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	v1 := r.Group("/api/v1")
	{
		v1.GET("/ping", pingHandler)
		v1.GET("/healthz", healthzHandler)
		v1.GET("/work", workHandler)
		v1.POST("/echo", echoHandler)
	}

	return r
}

func pingHandler(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "pong",
	})
}

func healthzHandler(c *gin.Context) {
	c.JSON(200, gin.H{
		"status": "ok",
	})
}

func workHandler(c *gin.Context) {
	start := time.Now()

	if config.LatencyMS > 0 {
		time.Sleep(time.Duration(config.LatencyMS) * time.Millisecond)
	}

	var memoryHog []byte
	if config.MemoryMB > 0 {
		memoryHog = make([]byte, config.MemoryMB*1024*1024)
		for i := 0; i < len(memoryHog); i += 4096 {
			memoryHog[i] = 1
		}
	}

	if config.FailRate > 0 && rand.Float64() < config.FailRate {
		duration := time.Since(start)
		logger.Error("work failed",
			"path", c.Request.URL.Path,
			"method", c.Request.Method,
			"status", 500,
			"duration_ms", duration.Milliseconds(),
		)
		c.JSON(500, gin.H{
			"error":       "simulated failure",
			"duration_ms": duration.Milliseconds(),
		})
		return
	}

	duration := time.Since(start)
	logger.Info("work completed",
		"path", c.Request.URL.Path,
		"method", c.Request.Method,
		"status", 200,
		"duration_ms", duration.Milliseconds(),
	)

	_ = memoryHog

	c.JSON(200, gin.H{
		"status":      "completed",
		"duration_ms": duration.Milliseconds(),
		"config": gin.H{
			"latency_ms": config.LatencyMS,
			"fail_rate":  config.FailRate,
			"memory_mb":  config.MemoryMB,
		},
	})
}

func echoHandler(c *gin.Context) {
	var body map[string]interface{}

	if err := c.ShouldBindJSON(&body); err != nil {
		logger.Warn("invalid JSON", "error", err.Error())
		c.JSON(400, gin.H{
			"error": "invalid JSON body",
		})
		return
	}

	logger.Info("echo request",
		"path", c.Request.URL.Path,
		"method", c.Request.Method,
	)

	c.JSON(200, gin.H{
		"echo": body,
	})
}

func main() {
	logger.Info("starting server", "port", 8080)
	router := setupRouter()
	router.Run(":8080")
}
