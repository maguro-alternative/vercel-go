package hairstyletype

import (
	"encoding/json"
	"log"
	"fmt"
	"net/http"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	validation "github.com/go-ozzo/ozzo-validation"
)

type HairStyleType struct {
	ID    int64  `db:"id"`
	Style string `db:"style"`
}

func (h *HairStyleType) Validate() error {
	return validation.ValidateStruct(h,
		validation.Field(&h.Style, validation.Required),
	)
}

type HairStyleTypesJson struct {
	HairStyleTypes []HairStyleType `json:"hairstyle_types"`
}

func (h *HairStyleTypesJson) Validate() error {
	return validation.ValidateStruct(h,
		validation.Field(&h.HairStyleTypes, validation.Required),
	)
}

type IDs struct {
	IDs []int64 `json:"ids"`
}

func (i *IDs) Validate() error {
	return validation.ValidateStruct(i,
		validation.Field(&i.IDs, validation.Required),
	)
}

func Handler(w http.ResponseWriter, r *http.Request) {
	db, err := sqlx.Open("postgres", "")
	if err != nil {
		log.Printf("sql.Open error %s", err)
	}
	defer db.Close()
	switch r.Method {
	case http.MethodGet:
	case http.MethodPost:
		var hairStyleTypesJson HairStyleTypesJson
		query := `
			INSERT INTO hairstyle_type (
				style
			) VALUES (
				:style
			)
		`
		err := json.NewDecoder(r.Body).Decode(&hairStyleTypesJson)
		if err != nil {
			log.Printf(fmt.Sprintf("json decode error: %v body:%v", err, r.Body))
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		// リクエストボディのバリデーション
		err = hairStyleTypesJson.Validate()
		if err != nil {
			log.Printf(fmt.Sprintf("json validate error: %v", err))
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
			return
		}
		for _, hairStyleType := range hairStyleTypesJson.HairStyleTypes {
			// リクエストボディのバリデーション
			err = hairStyleType.Validate()
			if err != nil {
				log.Printf(fmt.Sprintf("json validate error: %v", err))
				http.Error(w, err.Error(), http.StatusUnprocessableEntity)
				return
			}
			_, err = db.NamedExecContext(r.Context(), query, hairStyleType)
			if err != nil {
				log.Printf(fmt.Sprintf("db error: %v", err))
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
		// レスポンスの作成
		err = json.NewEncoder(w).Encode(&hairStyleTypesJson)
		if err != nil {
			log.Printf(fmt.Sprintf("json encode error: %v", err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	case http.MethodPut:
	case http.MethodDelete:
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}
