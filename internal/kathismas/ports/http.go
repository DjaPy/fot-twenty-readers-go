package ports

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/DjaPy/fot-twenty-readers-go/internal/kathismas/adapters/excel"
	"github.com/DjaPy/fot-twenty-readers-go/internal/kathismas/app"
	"github.com/DjaPy/fot-twenty-readers-go/internal/kathismas/app/command"
	"github.com/DjaPy/fot-twenty-readers-go/internal/kathismas/app/query"
	"github.com/DjaPy/fot-twenty-readers-go/internal/kathismas/config"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/go-pkgz/rest"
	"github.com/gofrs/uuid/v5"
)

type Server struct {
	Version       string
	Conf          config.Config
	TemplLocation string
	App           *app.Application

	httpServer *http.Server
	templates  *template.Template
}

func (s *Server) Run(ctx context.Context, port int) {
	slog.Info("starting server", "port", port)

	serverLock := sync.Mutex{}

	go func() {
		<-ctx.Done()
		serverLock.Lock()
		defer serverLock.Unlock()
		if s.httpServer != nil {
			if clsErr := s.httpServer.Close(); clsErr != nil {
				slog.Error("failed to close proxy http server", "error", clsErr)
			}
		}
	}()

	if s.TemplLocation == "" {
		s.TemplLocation = "internal/kathismas/ports/templates/*"
	}
	slog.Debug("loading templates", "location", s.TemplLocation)

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
		WriteTimeout:      120 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	serverLock.Unlock()
	err := s.httpServer.ListenAndServe()
	slog.Warn("http server terminated", "error", err)
}

func (s *Server) router() *chi.Mux {
	router := chi.NewRouter()
	router.Use(rest.AppInfo("for-twenty-readers", "DjaPy", s.Version), rest.Ping)

	router.Get("/", s.groupsPage)

	router.Get("/calendar", s.calendar)
	router.Post("/calendar", s.createCalendar)

	router.Get("/groups", s.groupsPage)
	router.Get("/groups/list", s.listGroupsPartial)
	router.Post("/groups", s.createGroup)
	router.Get("/groups/{id}", s.getGroupPage)
	router.Put("/groups/{id}", s.updateGroup)
	router.Delete("/groups/{id}", s.deleteGroup)
	router.Post("/groups/{id}/readers", s.addReaderToGroup)
	router.Delete("/groups/{id}/readers/{readerId}", s.removeReaderFromGroup)
	router.Post("/groups/{id}/generate", s.generateCalendarForGroup)
	router.Post("/groups/{id}/regenerate", s.regenerateCalendarForGroup)
	router.Get("/groups/{id}/current-kathisma", s.getCurrentKathisma)

	return router
}

// Groups page - main page with HTMX
func (s *Server) groupsPage(w http.ResponseWriter, r *http.Request) {
	data := struct {
		Title           string
		ContentTemplate string
	}{
		Title:           "Группы чтецов",
		ContentTemplate: "groups-content",
	}

	if err := s.templates.ExecuteTemplate(w, "layout.gohtml", data); err != nil {
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

	data := struct {
		Groups      []query.ReaderGroupDTO
		CurrentYear int
	}{
		Groups:      groups,
		CurrentYear: time.Now().Year(),
	}

	if err := s.templates.ExecuteTemplate(w, "group-list-item.gohtml", data); err != nil {
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

		data := struct {
			Groups      []query.ReaderGroupDTO
			CurrentYear int
		}{
			Groups: []query.ReaderGroupDTO{{
				ID:             group.ID,
				Name:           group.Name,
				StartOffset:    group.StartOffset,
				ReadersCount:   len(group.Readers),
				CalendarsCount: 0,
				CreatedAt:      group.CreatedAt,
			}},
			CurrentYear: time.Now().Year(),
		}

		if err := s.templates.ExecuteTemplate(w, "group-list-item.gohtml", data); err != nil {
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

	data := struct {
		Title           string
		ContentTemplate string
		*query.ReaderGroupDetailDTO
	}{
		Title:                group.Name,
		ContentTemplate:      "group-detail-content",
		ReaderGroupDetailDTO: group,
	}

	if err := s.templates.ExecuteTemplate(w, "layout.gohtml", data); err != nil {
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

	if err := r.ParseMultipartForm(32 << 20); err != nil {
		// If it's not multipart, try ParseForm as fallback
		if err := r.ParseForm(); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}

	readerNumber := atoi(r.FormValue("reader_number"))
	if readerNumber < 1 || readerNumber > 20 {
		http.Error(w, "reader number must be between 1 and 20", http.StatusBadRequest)
		return
	}

	username := r.FormValue("username")
	if username == "" {
		http.Error(w, "username is required", http.StatusBadRequest)
		return
	}

	cmd := command.AddReaderToGroup{
		GroupID:      groupID,
		ReaderNumber: int8(readerNumber), //nolint:gosec // check up
		Username:     username,
		TelegramID:   int64(atoi(r.FormValue("telegram_id"))),
		Phone:        r.FormValue("phone"),
	}

	if err := s.App.Commands.AddReaderToGroup.Handle(r.Context(), cmd); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/groups/%s", idStr), http.StatusSeeOther)
}

func (s *Server) removeReaderFromGroup(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	groupID, err := uuid.FromString(idStr)
	if err != nil {
		http.Error(w, "invalid group id", http.StatusBadRequest)
		return
	}

	readerIDStr := chi.URLParam(r, "readerId")
	readerID, err := uuid.FromString(readerIDStr)
	if err != nil {
		http.Error(w, "invalid reader id", http.StatusBadRequest)
		return
	}

	cmd := command.RemoveReaderFromGroup{
		GroupID:  groupID,
		ReaderID: readerID,
	}

	if err := s.App.Commands.RemoveReaderFromGroup.Handle(r.Context(), cmd); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if r.Header.Get("HX-Request") == "true" {
		w.WriteHeader(http.StatusOK)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/groups/%s", idStr), http.StatusSeeOther)
}

func (s *Server) handleCalendarGeneration(w http.ResponseWriter, r *http.Request, isRegenerate bool) {
	idStr := chi.URLParam(r, "id")
	groupID, err := uuid.FromString(idStr)
	if err != nil {
		http.Error(w, "invalid group id", http.StatusBadRequest)
		return
	}

	if errParse := r.ParseForm(); errParse != nil {
		http.Error(w, "failed to parse form data", http.StatusBadRequest)
		return
	}
	year := atoi(r.FormValue("year"))
	currentYear := time.Now().Year()

	if year != 0 && (year < currentYear || year > 2045) {
		http.Error(w, fmt.Sprintf("year must be between %d and 2045", currentYear), http.StatusBadRequest)
		return
	}

	group, err := s.App.Queries.GetReaderGroup.Handle(r.Context(), query.GetReaderGroup{ID: groupID})
	if err != nil {
		http.Error(w, "group not found", http.StatusNotFound)
		return
	}

	var buffer *bytes.Buffer
	action := "generation"
	if isRegenerate {
		action = "regeneration"
		cmd := command.RegenerateCalendarForGroup{
			GroupID: groupID,
			Year:    year,
		}
		slog.Info("starting calendar "+action, "group_id", groupID, "year", year)
		startTime := time.Now()
		buffer, err = s.App.Commands.RegenerateCalendarForGroup.Handle(r.Context(), cmd)
		duration := time.Since(startTime)
		slog.Info("calendar "+action+" completed", "duration", duration)
	} else {
		cmd := command.GenerateCalendarForGroup{
			GroupID: groupID,
			Year:    year,
		}
		slog.Info("starting calendar "+action, "group_id", groupID, "year", year)
		startTime := time.Now()
		buffer, err = s.App.Commands.GenerateCalendarForGroup.Handle(r.Context(), cmd)
		duration := time.Since(startTime)
		slog.Info("calendar "+action+" completed", "duration", duration)
	}

	if err != nil {
		slog.Error("failed to "+action[0:len(action)-2]+" calendar", "error", err)
		http.Error(w, "failed to "+action[0:len(action)-2]+" calendar", http.StatusInternalServerError)
		return
	}
	if year == 0 {
		year = currentYear
	}

	filename := fmt.Sprintf("calendar_%s_%d.xlsx", sanitizeFilename(group.Name), year)
	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%q\"", filename))
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(buffer.Bytes()); err != nil {
		slog.Error("failed to write response", "error", err)
	}
}

func (s *Server) generateCalendarForGroup(w http.ResponseWriter, r *http.Request) {
	s.handleCalendarGeneration(w, r, false)
}

func (s *Server) updateGroup(w http.ResponseWriter, r *http.Request) {
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

	name := r.FormValue("name")
	startOffsetStr := r.FormValue("start_offset")

	var namePtr *string
	var startOffsetPtr *int

	if name != "" {
		namePtr = &name
	}

	if startOffsetStr != "" {
		startOffset := atoi(startOffsetStr)
		startOffsetPtr = &startOffset
	}

	cmd := command.UpdateReaderGroup{
		GroupID:     groupID,
		Name:        namePtr,
		StartOffset: startOffsetPtr,
	}

	if err := s.App.Commands.UpdateReaderGroup.Handle(r.Context(), cmd); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if r.Header.Get("HX-Request") == "true" {
		http.Redirect(w, r, fmt.Sprintf("/groups/%s", idStr), http.StatusSeeOther)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (s *Server) deleteGroup(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	groupID, err := uuid.FromString(idStr)
	if err != nil {
		http.Error(w, "invalid group id", http.StatusBadRequest)
		return
	}

	cmd := command.DeleteReaderGroup{
		GroupID: groupID,
	}

	if err := s.App.Commands.DeleteReaderGroup.Handle(r.Context(), cmd); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if r.Header.Get("HX-Request") == "true" {
		w.WriteHeader(http.StatusOK)
		return
	}

	http.Redirect(w, r, "/groups", http.StatusSeeOther)
}

func (s *Server) regenerateCalendarForGroup(w http.ResponseWriter, r *http.Request) {
	s.handleCalendarGeneration(w, r, true)
}

func sanitizeFilename(name string) string {
	result := ""
	for _, r := range name {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' || r == '-' {
			result += string(r)
		} else if r == ' ' {
			result += "_"
		}
	}
	if result == "" {
		return "calendar"
	}
	return result
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

	file, err := excel.CreateXlSCalendar(date, entry.StartKathisma, entry.Year)
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

func (s *Server) getCurrentKathisma(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	groupID, err := uuid.FromString(idStr)
	if err != nil {
		http.Error(w, "invalid group id", http.StatusBadRequest)
		return
	}

	readerNumberStr := r.URL.Query().Get("reader_number")
	if readerNumberStr == "" {
		http.Error(w, "reader_number query parameter is required", http.StatusBadRequest)
		return
	}

	readerNumber := atoi(readerNumberStr)
	if readerNumber < 1 || readerNumber > 20 {
		http.Error(w, "reader number must be between 1 and 20", http.StatusBadRequest)
		return
	}

	result, err := s.App.Queries.GetCurrentKathisma.Handle(r.Context(), query.GetCurrentKathisma{
		GroupID:      groupID,
		ReaderNumber: readerNumber,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if r.Header.Get("HX-Request") == "true" {
		if err := s.templates.ExecuteTemplate(w, "current-kathisma.gohtml", result); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	render.JSON(w, r, result)
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
