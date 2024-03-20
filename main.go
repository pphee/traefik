package main

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
)

func fetchWhitelist(orgID string) ([]string, error) {
	apiURL := fmt.Sprintf("http://localhost:8080/whitelist/%s", orgID)
	fmt.Println("apiURL", apiURL)
	resp, err := http.Get(apiURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusNoContent {
		return nil, nil
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request error: %s", resp.Status)
	}

	var whitelist []string
	if err := json.NewDecoder(resp.Body).Decode(&whitelist); err != nil {
		return nil, err
	}

	return whitelist, nil
}

func ipWhitelistMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		cfConnectingIP := c.GetHeader("Cf-Connecting-Ip")
		orgID := c.Param("orgID")
		fmt.Println("cfConnectingIP", cfConnectingIP)
		fmt.Println("orgID", orgID)

		whitelist, err := fetchWhitelist(orgID)
		fmt.Println("whitelist", whitelist)
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		if whitelist == nil {
			c.Next()
			return
		}

		isIPAllowed := false
		for _, ip := range whitelist {
			if cfConnectingIP == ip {
				isIPAllowed = true
				break
			}
		}

		if isIPAllowed {
			c.Next()
		} else {
			c.AbortWithStatus(http.StatusForbidden)
		}
	}
}

func main() {
	router := gin.Default()

	router.Use(ipWhitelistMiddleware())

	router.GET("/someEndpoint/:orgID", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Access granted"})
	})

	router.Run(":8000")
}
