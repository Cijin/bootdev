package main

import (
	"bootdev/database"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
)

const (
	port   = "8080"
	dbFile = "db.json"
)

type apiConfig struct {
	fileServerHits int
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileServerHits++

		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) printMetrics(w http.ResponseWriter, r *http.Request) {
	_, err := w.Write([]byte(fmt.Sprintf("Hits: %d", cfg.fileServerHits)))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (cfg *apiConfig) resetMetrics(w http.ResponseWriter, r *http.Request) {
	cfg.fileServerHits = 0

	w.WriteHeader(http.StatusOK)
}

func middlewareCors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "*")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func respondWithError(w http.ResponseWriter, code int, msg string) {
	type errRes struct {
		Error string `json:"error,omitempty"`
	}

	r := errRes{}

	r.Error = msg
	d, err := json.Marshal(r)
	if err != nil {
		log.Printf("Marshal error: %s", err.Error())
	}

	w.WriteHeader(code)
	w.Write(d)
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	d, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Marshal error: %s", err.Error())
	}

	w.WriteHeader(code)
	w.Write(d)
}

func main() {
	apiCfg := new(apiConfig)
	db, err := database.NewDb(dbFile)
	if err != nil {
		log.Panic(err)
	}

	router := chi.NewRouter()
	fsHandler := apiCfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir("public"))))
	// app
	router.Handle("/app/*", fsHandler)
	router.Handle("/app", fsHandler)

	// api
	apiRouter := chi.NewRouter()

	apiRouter.HandleFunc("/reset", apiCfg.resetMetrics)

	apiRouter.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)

		_, err := w.Write([]byte("OK"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	apiRouter.Post("/chirps", func(w http.ResponseWriter, r *http.Request) {
		bannedWords := []string{"kerfuffle", "sharbert", "fornax"}
		censor := "****"

		decoder := json.NewDecoder(r.Body)

		c := &database.Chirp{}

		err := decoder.Decode(&c)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Something went wrong")
			return
		}

		if len(c.Body) > 140 {
			respondWithError(w, http.StatusBadRequest, "Chirp is too long")
			return
		}

		words := strings.Split(c.Body, " ")
		for i, w := range words {
			for _, bw := range bannedWords {
				if strings.EqualFold(w, bw) {
					words[i] = censor
				}
			}
		}
		censoredText := strings.Join(words, " ")

		chirp, err := db.CreateChirp(censoredText)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Something went wrong")
			return
		}

		respondWithJSON(w, http.StatusCreated, chirp)
		return
	})

	apiRouter.Get("/chirps/{id}", func(w http.ResponseWriter, r *http.Request) {
		idParam := chi.URLParam(r, "id")
		id, err := strconv.Atoi(idParam)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "Not a valid id")
			return
		}

		c, err := db.GetChirp(id)
		if err != nil {
			if errors.Is(err, database.ErrNotFound) {
				respondWithError(w, http.StatusNotFound, "Chirp does not exist")
				return
			}
			respondWithError(w, http.StatusInternalServerError, "Something went wrong")
			return
		}

		respondWithJSON(w, http.StatusOK, c)
	})

	apiRouter.Get("/chirps", func(w http.ResponseWriter, r *http.Request) {
		chirps, err := db.GetChirps()
		if err != nil {
			log.Print(err)
			respondWithError(w, http.StatusInternalServerError, "Something went wrong")
			return
		}

		respondWithJSON(w, http.StatusOK, chirps)
	})

	apiRouter.Post("/users", func(w http.ResponseWriter, r *http.Request) {
		decoder := json.NewDecoder(r.Body)

		u := &database.User{}

		err := decoder.Decode(&u)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Something went wrong")
			return
		}

		user, err := db.CreateUser(u.Email, u.Password)
		if err != nil {
			if errors.Is(err, database.ErrDuplicateEmail) {
				respondWithError(w, http.StatusUnauthorized, "User with email already exists")
				return
			}
			respondWithError(w, http.StatusInternalServerError, "Something went wrong")
			return
		}

		respondWithJSON(w, http.StatusCreated, user)
		return
	})

	apiRouter.Post("/login", func(w http.ResponseWriter, r *http.Request) {
		decoder := json.NewDecoder(r.Body)

		u := &database.User{}

		err := decoder.Decode(&u)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Something went wrong")
			return
		}

		user, err := db.Login(u.Email, u.Password)
		if err != nil {
			respondWithError(w, http.StatusUnauthorized, "Unauthorized")
			return
		}

		respondWithJSON(w, http.StatusOK, user)
		return
	})

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

	fmt.Printf("Serving on port: %s\n", port)
	log.Fatal(srv.ListenAndServe())
}
