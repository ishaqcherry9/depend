package swagger

import (
	"fmt"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/swaggo/swag"

	"github.com/ishaqcherry9/depend/pkg/gofile"
)

func DefaultRouter(r *gin.Engine, jsonContent []byte) {
	registerSwagger("swagger", jsonContent)
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
}

func DefaultRouterByFile(r *gin.Engine, jsonFile string) {
	jsonContent, err := os.ReadFile(jsonFile)
	if err != nil {
		fmt.Printf("\nos.ReadFile error: %v\n\n", err)
		return
	}
	registerSwagger("swagger", jsonContent)
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
}

func CustomRouter(r *gin.Engine, name string, jsonContent []byte) {
	registerSwagger(name, jsonContent)
	r.GET(fmt.Sprintf("/%s/swagger/*any", name), ginSwagger.WrapHandler(swaggerFiles.NewHandler(), ginSwagger.InstanceName(name)))
}

func CustomRouterByFile(r *gin.Engine, jsonFile string) {
	jsonContent, err := os.ReadFile(jsonFile)
	if err != nil {
		fmt.Printf("\nos.ReadFile error: %v\n\n", err)
		return
	}

	filename := gofile.GetFilename(jsonFile)
	name := strings.Split(filename, ".")[0]
	registerSwagger(name, jsonContent)

	r.GET(fmt.Sprintf("/%s/swagger/*any", name), ginSwagger.WrapHandler(swaggerFiles.NewHandler(), ginSwagger.InstanceName(name)))
}

func registerSwagger(infoInstanceName string, jsonContent []byte) {
	swaggerInfo := &swag.Spec{
		Schemes:          []string{"http", "https"},
		InfoInstanceName: infoInstanceName,
		SwaggerTemplate:  string(jsonContent),
	}

	swag.Register(swaggerInfo.InstanceName(), swaggerInfo)
}
