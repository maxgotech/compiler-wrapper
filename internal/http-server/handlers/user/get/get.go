package get

import (
	"compiler-wrapper/internal/db/postgres"
	"compiler-wrapper/internal/lib/logger/sl"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
)

type Response struct {
	Result []postgres.User `json:"result"`
}

type Getter interface {
	GetUsers() ([]postgres.User, error)
}

func New(log *slog.Logger, getter Getter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.url.get.New"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		users, err := getter.GetUsers()
		if err != nil {
			log.Error("some went south", sl.Err(err))

			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, err)

			return
		}

		render.Status(r, http.StatusOK)
		render.JSON(w, r, Response{Result: users})
	}
}
