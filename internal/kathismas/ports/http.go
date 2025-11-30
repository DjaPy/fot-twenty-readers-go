package ports

import (
	"context"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/DjaPy/fot-twenty-readers-go/internal/kathismas/app"
	"github.com/DjaPy/fot-twenty-readers-go/internal/kathismas/app/command"
	"github.com/DjaPy/fot-twenty-readers-go/internal/kathismas/app/query"
	"github.com/DjaPy/fot-twenty-readers-go/internal/kathismas/config"
	"github.com/DjaPy/fot-twenty-readers-go/internal/kathismas/service"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/go-pkgz/rest"
	"github.com/gofrs/uuid/v5"
)

type Server struct {
	Version       string
	Conf          config.Conf
	TemplLocation string
	App           *app.Application

	httpServer *http.Server
	templates  *template.Template
}

func (s *Server) Run(ctx context.Context, port int) {
	slog.Info(fmt.Sprintf("starting server on port %d", port))

	serverLock := sync.Mutex{}

	go func() {
		<-ctx.Done()
		serverLock.Lock()
		if s.httpServer != nil {
			if clsErr := s.httpServer.Close(); clsErr != nil {
				slog.Error(fmt.Sprintf("failed to close proxy http server, %v", clsErr))
			}
		}
	}()

	if s.TemplLocation == "" {
		s.TemplLocation = "internal/kathismas/ports/templates/*"
	}
	slog.Debug(fmt.Sprintf("loading templates from %s", s.TemplLocation))

	funcMap := template.FuncMap{
		"add": func(a, b int) int {
			return a + b
		},
	}
	s.templates = template.Must(template.New("").Funcs(funcMap).ParseGlob(s.TemplLocation))

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
	slog.Warn(fmt.Sprintf("http server terminated, %s", err))
}

func (s *Server) router() *chi.Mux {
	router := chi.NewRouter()
	router.Use(rest.AppInfo("for-twenty-readers", "djapy", s.Version), rest.Ping)

	// Redirect root to groups
	router.Get("/", s.groupsPage)

	// Old calendar endpoints
	router.Get("/calendar", s.calendar)
	router.Post("/calendar", s.createCalendar)

	// New groups endpoints
	router.Get("/groups", s.groupsPage)
	router.Get("/groups/list", s.listGroupsPartial)
	router.Post("/groups", s.createGroup)
	router.Get("/groups/{id}", s.getGroupPage)
	router.Post("/groups/{id}/readers", s.addReaderToGroup)
	router.Post("/groups/{id}/generate", s.generateCalendarForGroup)

	return router
}

// Groups page - main page with HTMX
func (s *Server) groupsPage(w http.ResponseWriter, r *http.Request) {
	data := struct {
		Title string
	}{
		Title: "Группы чтецов",
	}

	if err := s.templates.ExecuteTemplate(w, "layout.gohtml", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Need to execute the content template
	if err := s.templates.ExecuteTemplate(w, "groups.gohtml", nil); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// List groups partial for HTMX
func (s *Server) listGroupsPartial(w http.ResponseWriter, r *http.Request) {
	groups, err := s.App.Queries.ListReaderGroups.Handle(r.Context(), query.ListReaderGroups{})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := s.templates.ExecuteTemplate(w, "group-list-item.gohtml", groups); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *Server) createGroup(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	cmd := command.CreateReaderGroup{
		Name:        r.FormValue("name"),
		StartOffset: atoi(r.FormValue("start_offset")),
	}

	groupID, err := s.App.Commands.CreateReaderGroup.Handle(r.Context(), cmd)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if r.Header.Get("HX-Request") == "true" {
		group, err := s.App.Queries.GetReaderGroup.Handle(r.Context(), query.GetReaderGroup{ID: groupID})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		groups := []query.ReaderGroupDTO{{
			ID:             group.ID,
			Name:           group.Name,
			StartOffset:    group.StartOffset,
			ReadersCount:   len(group.Readers),
			CalendarsCount: 0,
			CreatedAt:      group.CreatedAt,
		}}

		if err := s.templates.ExecuteTemplate(w, "group-list-item.gohtml", groups); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	render.Status(r, http.StatusCreated)
	render.JSON(w, r, rest.JSON{"id": groupID.String()})
}

func (s *Server) getGroupPage(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.FromString(idStr)
	if err != nil {
		http.Error(w, "invalid group id", http.StatusBadRequest)
		return
	}

	group, err := s.App.Queries.GetReaderGroup.Handle(r.Context(), query.GetReaderGroup{ID: id})
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	if r.Header.Get("Accept") == "application/json" {
		render.JSON(w, r, group)
		return
	}

	if err := s.templates.ExecuteTemplate(w, "layout.gohtml", struct{ Title string }{Title: group.Name}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := s.templates.ExecuteTemplate(w, "group-detail.gohtml", group); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *Server) addReaderToGroup(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	groupID, err := uuid.FromString(idStr)
	if err != nil {
		http.Error(w, "invalid group id", http.StatusBadRequest)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	cmd := command.AddReaderToGroup{
		GroupID:    groupID,
		Username:   r.FormValue("username"),
		TelegramID: int64(atoi(r.FormValue("telegram_id"))),
		Phone:      r.FormValue("phone"),
	}

	if err := s.App.Commands.AddReaderToGroup.Handle(r.Context(), cmd); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/groups/%s", idStr), http.StatusSeeOther)
}

func (s *Server) generateCalendarForGroup(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	groupID, err := uuid.FromString(idStr)
	if err != nil {
		http.Error(w, "invalid group id", http.StatusBadRequest)
		return
	}

	if errParse := r.ParseForm(); errParse != nil {
		http.Error(w, errParse.Error(), http.StatusBadRequest)
		return
	}

	cmd := command.GenerateCalendarForGroup{
		GroupID: groupID,
		Year:    atoi(r.FormValue("year")),
	}

	buffer, errGC := s.App.Commands.GenerateCalendarForGroup.Handle(r.Context(), cmd)
	if errGC != nil {
		http.Error(w, errGC.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition", "attachment; filename=\"calendar.xlsx\"")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(buffer.Bytes()); err != nil {
		slog.Error(fmt.Sprintf("failed to write response: %v", err))
	}
}

func (s *Server) calendar(w http.ResponseWriter, r *http.Request) {
	tmplData := struct {
		Title string
	}{
		Title: "Calendar",
	}
	if err := s.templates.ExecuteTemplate(w, "base.gohtml", tmplData); err != nil {
		s.renderErrorPage(w, r, err, 400)
	}
}

func (s *Server) createCalendar(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
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

	file, err := service.CreateXlSCalendar(date, entry.StartKathisma, entry.Year)
	if err != nil {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, rest.JSON{"error": err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition", "attachment; filename=\"file.xlsx\"")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(file.Bytes()); err != nil {
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, rest.JSON{"error": err.Error()})
	}
}

func (s *Server) renderErrorPage(w http.ResponseWriter, r *http.Request, err error, errCode int) { // nolint
	tmplData := struct {
		Status int
		Error  string
	}{Status: errCode, Error: err.Error()}

	if err := s.templates.ExecuteTemplate(w, "error.tmpl", &tmplData); err != nil {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, rest.JSON{"error": err.Error()})
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
