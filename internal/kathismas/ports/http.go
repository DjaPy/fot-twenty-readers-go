package ports

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/DjaPy/fot-twenty-readers-go/internal/kathismas/config"
	"github.com/DjaPy/fot-twenty-readers-go/internal/kathismas/proc"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/go-pkgz/rest"
)

type Server struct {
	Version       string
	Conf          config.Conf
	TemplLocation string

	httpServer *http.Server
	templates  *template.Template
}

func (s *Server) Run(ctx context.Context, port int) {
	log.Printf("[INFO] starting server on port %d", port)

	serverLock := sync.Mutex{}

	go func() {
		<-ctx.Done()
		serverLock.Lock()
		if s.httpServer != nil {
			if clsErr := s.httpServer.Close(); clsErr != nil {
				log.Printf("[ERROR] failed to close proxy http server, %v", clsErr)
			}
		}
	}()

	if s.TemplLocation == "" {
		s.TemplLocation = "app/webapp/templates/*"
	}
	log.Printf("[DEBUG] loading templates from %s", s.TemplLocation)
	s.templates = template.Must(template.ParseGlob(s.TemplLocation))

	serverLock.Lock()
	s.httpServer = &http.Server{
		Addr:              fmt.Sprintf(":%d", port),
		Handler:           s.router(),
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       30 * time.Second,
	}
	serverLock.Unlock()
	err := s.httpServer.ListenAndServe()
	log.Printf("[WARN] http server terminated, %s", err)
}

func (s *Server) router() *chi.Mux {
	router := chi.NewRouter()
	router.Use(rest.AppInfo("for-twenty-readers", "djapy", s.Version), rest.Ping)

	router.Get("/calendar", s.calendar)
	router.Post("/calendar/create", s.createCalendar)
	return router
}

func (s *Server) calendar(w http.ResponseWriter, r *http.Request) {
	tmplData := struct {
		Title string
	}{
		Title: "Calendar",
	}
	err := s.templates.ExecuteTemplate(w, "base.gohtml", tmplData)
	if err != nil {
		s.renderErrorPage(w, r, err, 400)
		return
	}
}

func (s *Server) createCalendar(w http.ResponseWriter, r *http.Request) {

	err := r.ParseForm()
	if err != nil {
		s.renderErrorPage(w, r, err, 400)
		return
	}
	entry := CalendarEntry{
		StartDateKathisma: r.FormValue("start_date_kathisma"),
		StartKathisma:     atoi(r.FormValue("start_kathisma")),
		Year:              atoi(r.FormValue("year")),
	}
	date, err := time.Parse("2006-01-02", entry.StartDateKathisma)
	if err != nil {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, rest.JSON{"error": err.Error()})
		return
	}
	file, err := proc.CreateXlSCalendar(date, entry.StartKathisma, entry.Year)
	if err != nil {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, rest.JSON{"error": err.Error()})
		return
	}
	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition", "attachment; filename=\"file.xlsx\"")

	w.WriteHeader(http.StatusOK)
	_, err = w.Write(file.Bytes())
	if err != nil {
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, rest.JSON{"error": err.Error()})
		return
	}
	render.Status(r, http.StatusCreated)
	render.JSON(w, r, rest.JSON{"result": "All created"})
}

func (s *Server) renderErrorPage(w http.ResponseWriter, r *http.Request, err error, errCode int) { // nolint
	tmplData := struct {
		Status int
		Error  string
	}{Status: errCode, Error: err.Error()}

	if err := s.templates.ExecuteTemplate(w, "error.tmpl", &tmplData); err != nil {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, rest.JSON{"error": err.Error()})
		return
	}
}

type CalendarEntry struct {
	StartDateKathisma string `json:"start_date_kathisma"`
	StartKathisma     int    `json:"start_kathisma"`
	Year              int    `json:"year"`
}

func atoi(s string) int {
	i, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return i
}
