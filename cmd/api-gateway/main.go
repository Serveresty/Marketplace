package main

import (
	"io"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header missing"})
			c.Abort()
			return
		}

		isAuthorized, err := checkAuthService(token)
		if err != nil || !isAuthorized {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}

		c.Next()
	}
}

func checkAuthService(token string) (bool, error) {
	resp, err := http.Get("http://localhost:8081/validate?token=" + token)
	if err != nil {
		log.Println("Failed to connect to auth-service:", err)
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		return true, nil
	}
	return false, nil
}

func main() {
	router := gin.Default()

	authService := router.Group("/auth") // Требует проверки на стороне сервиса на присутствие/отсутствие авторизации
	{
		servicePath := "/auth"
		authService.Any("/*proxyPath", func(c *gin.Context) {
			proxy(c, "http://localhost:8081", servicePath)
		})
	}

	productService := router.Group("/product") // Требует авторизацию на некоторые эндпоинты
	{
		servicePath := "/product"
		productService.Any("/*proxyPath", func(c *gin.Context) {
			proxy(c, "http://localhost:8082", servicePath)
		})
	}

	authorizedOnly := router.Group("") // Требуют авторизации
	authorizedOnly.Use(AuthMiddleware())
	{
		user := authorizedOnly.Group("/user")
		{
			servicePath := "/user"
			user.Any("/*proxyPath", func(c *gin.Context) {
				proxy(c, "http://localhost:8083", servicePath)
			})
		}

		cart := authorizedOnly.Group("/cart")
		{
			servicePath := "/cart"
			cart.Any("/*proxyPath", func(c *gin.Context) {
				proxy(c, "http://localhost:8084", servicePath)
			})
		}

		order := authorizedOnly.Group("/order")
		{
			servicePath := "/order"
			order.Any("/*proxyPath", func(c *gin.Context) {
				proxy(c, "http://localhost:8085", servicePath)
			})
		}

		payment := authorizedOnly.Group("/payment")
		{
			servicePath := "/payment"
			payment.Any("/*proxyPath", func(c *gin.Context) {
				proxy(c, "http://localhost:8086", servicePath)
			})
		}

		shipping := authorizedOnly.Group("/shipping")
		{
			servicePath := "/shipping"
			shipping.Any("/*proxyPath", func(c *gin.Context) {
				proxy(c, "http://localhost:8087", servicePath)
			})
		}

		notification := authorizedOnly.Group("/notification")
		{
			servicePath := "/notification"
			notification.Any("/*proxyPath", func(c *gin.Context) {
				proxy(c, "http://localhost:8088", servicePath)
			})
		}

		analytics := authorizedOnly.Group("/analytics")
		{
			servicePath := "/analytics"
			analytics.Any("/*proxyPath", func(c *gin.Context) {
				proxy(c, "http://localhost:8089", servicePath)
			})
		}

	}

	reviewService := router.Group("/review") // Требует авторизацию на некоторые эндпоинты
	{
		servicePath := "/review"
		reviewService.Any("/*proxyPath", func(c *gin.Context) {
			proxy(c, "http://localhost:8090", servicePath)
		})
	}

	router.Run(":8080")
}

func proxy(c *gin.Context, target, servicePath string) {
	proxyPath := c.Param("proxyPath")
	targetURL := target + servicePath + proxyPath
	log.Printf("Proxying request to: %s", targetURL)

	req, err := http.NewRequest(c.Request.Method, targetURL, c.Request.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create request"})
		return
	}

	for key, values := range c.Request.Header {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to proxy request"})
		return
	}
	defer resp.Body.Close()

	c.Status(resp.StatusCode)
	c.Header("Content-Type", resp.Header.Get("Content-Type"))
	c.Stream(func(w io.Writer) bool {
		io.Copy(w, resp.Body)
		return false
	})
}
