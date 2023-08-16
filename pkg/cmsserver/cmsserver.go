package cmsserver

import (
	"app-store-server/internal/constants"
	"app-store-server/internal/mongo"
	"app-store-server/pkg/api"
	servicev1 "app-store-server/pkg/cmsserver/service/v1"
	"net/http"

	restfulspec "github.com/emicklei/go-restful-openapi/v2"
	"github.com/emicklei/go-restful/v3"
	"github.com/go-openapi/spec"
	"github.com/golang/glog"
)

type CMSServer struct {
	Server *http.Server

	// RESTful Server
	container *restful.Container
}

func New() (*CMSServer, error) {
	as := &CMSServer{}

	server := &http.Server{
		Addr: constants.CMSServerListenAddress,
	}

	as.Server = server
	return as, nil
}

func (s *CMSServer) PrepareRun() error {
	s.container = restful.NewContainer()
	s.container.Filter(api.LogRequestAndResponse)
	s.container.Router(restful.CurlyRouter{})
	s.container.RecoverHandler(func(panicReason interface{}, httpWriter http.ResponseWriter) {
		api.LogStackOnRecover(panicReason, httpWriter)
	})

	s.installModuleAPI()
	s.installAPIDocs()

	for _, ws := range s.container.RegisteredWebServices() {
		glog.Infof("registered module: %s", ws.RootPath())
	}

	s.Server.Handler = s.container

	initMiddlewares()

	return nil
}

func initMiddlewares() {
	err := mongo.Init()
	if err != nil {
		glog.Fatalln(err)
	}
}

func (s *CMSServer) Run() error {
	return s.Server.ListenAndServe()
}

func (s *CMSServer) installAPIDocs() {
	config := restfulspec.Config{
		WebServices:                   s.container.RegisteredWebServices(), // you control what services are visible
		APIPath:                       "/app-store-admin-server/v1/apidocs.json",
		PostBuildSwaggerObjectHandler: enrichSwaggerObject}
	s.container.Add(restfulspec.NewOpenAPIService(config))

	cors := restful.CrossOriginResourceSharing{
		AllowedHeaders: []string{"Content-Type", "Accept"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE"},
		CookiesAllowed: false,
		Container:      restful.DefaultContainer}
	s.container.Filter(cors.Filter)
}

func (s *CMSServer) installModuleAPI() {
	servicev1.AddToContainer(s.container)
}

func enrichSwaggerObject(swo *spec.Swagger) {
	swo.Info = &spec.Info{
		InfoProps: spec.InfoProps{
			Title:       "App Store Admin Server",
			Description: "App Store Admin Server",
			Contact: &spec.ContactInfo{
				ContactInfoProps: spec.ContactInfoProps{
					Name:  "bytetrade",
					Email: "dev@bytetrade.io",
					URL:   "http://bytetrade.io",
				},
			},
			License: &spec.License{
				LicenseProps: spec.LicenseProps{
					Name: "Apache License 2.0",
					URL:  "http://www.apache.org/licenses/LICENSE-2.0",
				},
			},
			Version: "1.0.0",
		},
	}
	swo.Tags = []spec.Tag{{TagProps: spec.TagProps{
		Name:        "App Store Admin Server",
		Description: "App Store Admin Server"}}}
	swo.Schemes = []string{"http", "https"}
}
