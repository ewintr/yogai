package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"ewintr.nl/yogai/model"
	"ewintr.nl/yogai/storage"
	"golang.org/x/exp/slog"
)

type VideoAPI struct {
	videoRepo storage.VideoRepository
	logger    *slog.Logger
}

func NewVideoAPI(videoRepo storage.VideoRepository, logger *slog.Logger) *VideoAPI {
	return &VideoAPI{
		videoRepo: videoRepo,
		logger:    logger,
	}
}

func (v *VideoAPI) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	videoID, _ := ShiftPath(r.URL.Path)

	switch {
	case r.Method == http.MethodGet && videoID == "":
		v.List(w, r)
	default:
		Error(w, http.StatusNotFound, "not found", fmt.Errorf("method %s with subpath %q was not registered in the repository api", r.Method, videoID))
	}
}

func (v *VideoAPI) List(w http.ResponseWriter, r *http.Request) {
	video, err := v.videoRepo.FindByStatus(model.StatusReady)
	if err != nil {
		v.returnErr(r.Context(), w, http.StatusInternalServerError, "could not list repositories", err)
		return
	}

	type respVideo struct {
		YoutubeID string `json:"youtube_url"`
		Title     string `json:"title"`
		Summary   string `json:"summary"`
	}
	var resp []respVideo
	for _, v := range video {
		resp = append(resp, respVideo{
			YoutubeID: string(v.YoutubeID),
			Title:     v.YoutubeTitle,
			Summary:   v.Summary,
		})
	}

	jsonBody, err := json.Marshal(resp)
	if err != nil {
		v.returnErr(r.Context(), w, http.StatusInternalServerError, "could not marshal response", err)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, string(jsonBody))
}

func (v *VideoAPI) returnErr(_ context.Context, w http.ResponseWriter, status int, message string, err error, details ...any) {
	v.logger.Error(message, slog.String("err", err.Error()), slog.String("details", fmt.Sprintf("%+v", details)))
	Error(w, status, message, err, details...)
}
