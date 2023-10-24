package main

import (
	"bootdev/api"
	"bootdev/database"
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
)

const port = "8080"

type apiConfig struct {
	fileServerHits int
}

func main() {
	err := database.NewDb()
	if err != nil {
		log.Fatal(err)
	}

	apiCfg := apiConfig{}

	router := chi.NewRouter()
	fsHandler := apiCfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir("public"))))
	// app
	router.Handle("/app/*", fsHandler)
	router.Handle("/app", fsHandler)

	// api
	apiRouter := chi.NewRouter()

	apiRouter.HandleFunc("/reset", apiCfg.resetMetrics)
	apiRouter.Get("/healthz", api.Healthz)

	apiRouter.Post("/chirps", api.CreateChirp)
	apiRouter.Get("/chirps/{id}", api.GetChirp)
	apiRouter.Get("/chirps", api.GetChrips)
	apiRouter.Delete("/chirps/{id}", api.DeleteChirp)

	apiRouter.Post("/users", api.CreateUser)
	apiRouter.Put("/users", api.UpdateUser)

	apiRouter.Post("/login", api.Login)
	apiRouter.Post("/refresh", api.RefreshToken)
	apiRouter.Post("/revoke", api.RevokeToken)

	apiRouter.Post("/polka/webhooks", api.UpgradeUser)

	// admin
	adminRouter := chi.NewRouter()
	adminRouter.Get("/metrics", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(fmt.Sprintf(adminTemplate, apiCfg.fileServerHits)))
	})

	router.Mount("/api", apiRouter)
	router.Mount("/admin", adminRouter)
	corsMux := middlewareCors(router)

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: corsMux,
	}

	fmt.Printf("Serving on %s\n", srv.Addr)
	log.Fatal(srv.ListenAndServe())
}
