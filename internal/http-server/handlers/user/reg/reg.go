package reg

import (
	"compiler-wrapper/internal/lib/logger/sl"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
)

type Request struct {
	Name     string `json:"name" validate:"required"`
	Mail     string `json:"mail" validate:"required"`
	Password string `json:"pass" validate:"required"`
}

type Response struct {
	Result bool `json:"result"`
}

type Registrator interface {
	Reg(name, mail, pass string) error
}

func New(log *slog.Logger, registrator Registrator) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.url.save.New"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		var req Request

		err := render.DecodeJSON(r.Body, &req)
		if err != nil {
			log.Error("failed to decode request body", sl.Err(err))

			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, fmt.Errorf("failed to decode request"))

			return
		}

		log.Info("request body decoded", slog.Any("request", req))

		if err := validator.New().Struct(req); err != nil {
			validateErr := err.(validator.ValidationErrors)

			log.Error("invalid request", sl.Err(err))

			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, validateErr)

			return
		}

		err = registrator.Reg(req.Name, req.Mail, req.Password)
		if err != nil {
			log.Error("invalid request", sl.Err(err))

			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, err)

			return
		}

		render.Status(r, http.StatusCreated)
		render.JSON(w, r, Response{Result: true})
	}
}
