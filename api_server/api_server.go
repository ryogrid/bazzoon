package api_server

import (
	"fmt"
	"github.com/ant0ine/go-json-rest/rest"
	"github.com/ryogrid/buzzoon/core"
	"github.com/ryogrid/buzzoon/glo_val"
	"log"
	"net/http"
)

type PostEventReq struct {
	Content string
}

type UpdateProfileReq struct {
	Name    string
	About   string
	Picture string
}

type GeneralResp struct {
	Status string
}

type ApiServer struct {
	buzzPeer *core.BuzzPeer
}

func NewApiServer(peer *core.BuzzPeer) *ApiServer {
	return &ApiServer{peer}
}

func (s *ApiServer) postEvent(w rest.ResponseWriter, req *rest.Request) {
	input := PostEventReq{}
	err := req.DecodeJsonPayload(&input)

	if err != nil {
		fmt.Println(err)
		rest.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if input.Content == "" {
		rest.Error(w, "Content is required", 400)
		return
	}

	evt := s.buzzPeer.MessageMan.BrodcastOwnPost(input.Content)
	// display for myself
	s.buzzPeer.MessageMan.DataMan.DispPostAtStdout(evt)

	w.WriteJson(&GeneralResp{
		"SUCCESS",
	})
}

func (s *ApiServer) updateProfile(w rest.ResponseWriter, req *rest.Request) {
	input := UpdateProfileReq{}
	err := req.DecodeJsonPayload(&input)

	if err != nil {
		fmt.Println(err)
		rest.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if input.Name == "" {
		rest.Error(w, "Name is required", 400)
		return
	}

	prof := s.buzzPeer.MessageMan.BrodcastOwnProfile(&input.Name, &input.About, &input.Picture)
	// update local profile
	glo_val.ProfileMyOwn = prof

	w.WriteJson(&GeneralResp{
		"SUCCESS",
	})
}

func (s *ApiServer) LaunchAPIServer(addrStr string) {
	api := rest.NewApi()

	// the Middleware stack
	//api.Use(rest.DefaultDevStack...)
	api.Use(
		//&rest.AccessLogApacheMiddleware{},
		&rest.TimerMiddleware{},
		&rest.RecorderMiddleware{},
		&rest.PoweredByMiddleware{},
		&rest.RecoverMiddleware{
			EnableResponseStackTrace: true,
		},
		&rest.JsonIndentMiddleware{},
		&rest.ContentTypeCheckerMiddleware{},
	)
	api.Use(&rest.JsonpMiddleware{
		CallbackNameKey: "cb",
	})
	api.Use(&rest.CorsMiddleware{
		RejectNonCorsRequests: false,
		OriginValidator: func(origin string, request *rest.Request) bool {
			return true
		},
		AllowedMethods:                []string{"POST"},
		AllowedHeaders:                []string{"Accept", "content-type"},
		AccessControlAllowCredentials: true,
		AccessControlMaxAge:           3600,
	})

	router, err := rest.MakeRouter(
		&rest.Route{"POST", "/postEvent", s.postEvent},
		&rest.Route{"POST", "/updateProfile", s.updateProfile},
	)
	if err != nil {
		log.Fatal(err)
	}
	api.SetApp(router)

	log.Printf("Server started")
	log.Fatal(http.ListenAndServe(
		addrStr,
		api.MakeHandler(),
	))
}