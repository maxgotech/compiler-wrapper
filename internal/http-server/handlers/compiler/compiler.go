package compiler

import (
	"bytes"
	resp "compiler-wrapper/internal/lib/api/response"
	"compiler-wrapper/internal/lib/logger/sl"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
)

type Request struct {
	Language string `json:"language" validate:"required"`
	Code     string `json:"code" validate:"required"`
}

type Response struct {
	Success int    `json:"success"`
	Result  string `json:"result"`
}

type ApiResponse struct {
	TimeStamp int    `json:"timeStamp"`
	Status    int    `json:"status"`
	Output    string `json:"output"`
	Err       string `json:"error"`
	Lang      string `json:"language"`
	Info      string `json:"info"`
}

type Compiler interface {
	Compile(code, lang string) (string, error)
}

// func New(log *slog.Logger, comp Compiler, alias_length int) http.HandlerFunc {
func New(log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.compiler.New"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		var req Request

		err := render.DecodeJSON(r.Body, &req)
		if err != nil {
			log.Error("failed to decode request body", sl.Err(err))

			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, resp.Error("failed to decode request"))

			return
		}

		log.Info("request body decoded", slog.Any("request", req))

		if err := validator.New().Struct(req); err != nil {
			validateErr := err.(validator.ValidationErrors)

			log.Error("invalid request", sl.Err(err))

			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, resp.ValidationError(validateErr))

			return
		}

		api := os.Getenv("COMPILER_API")

		bodyValues := map[string]string{"code": req.Code, "language": req.Language}
		jsonValue, _ := json.Marshal(bodyValues)
		body := []byte(jsonValue)

		res, err := http.Post(api, "application/json", bytes.NewBuffer(body))
		if err != nil {
			log.Error("failed to send request", sl.Err(err))

			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, resp.Error("failed to send request"))

			return
		}

		log.Info("request to api send")

		var apiRes ApiResponse

		err = render.DecodeJSON(res.Body, &apiRes)
		if err != nil {
			log.Error("failed to decode api response body", sl.Err(err))

			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, resp.Error("failed to decode api response"))

			return
		}

		log.Info("request api body decoded", slog.Any("request", apiRes))

		render.Status(r, http.StatusOK)
		if apiRes.Err == "" {
			responseOK(w, r, apiRes.Output)
		} else {
			responseNotOK(w, r, apiRes.Err)
		}
	}
}

func responseOK(w http.ResponseWriter, r *http.Request, result string) {
	render.JSON(w, r, Response{
		Success: 1,
		Result:  result,
	})
}

func responseNotOK(w http.ResponseWriter, r *http.Request, result string) {
	render.JSON(w, r, Response{
		Success: 0,
		Result:  result,
	})
}
