package haircolortype

import (
	"encoding/json"
	"log"
	"fmt"
	"net/http"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	validation "github.com/go-ozzo/ozzo-validation"
)

type HairColorType struct {
	ID    int64 `db:"id"`
	Color string `db:"color"`
}

func (h *HairColorType) Validate() error {
	return validation.ValidateStruct(h,
		validation.Field(&h.Color, validation.Required),
	)
}

type HairColorTypesJson struct {
	HairColorTypes []HairColorType `json:"haircolor_types"`
}

func (h *HairColorTypesJson) Validate() error {
	return validation.ValidateStruct(h,
		validation.Field(&h.HairColorTypes, validation.Required),
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
		var hairColorTypesJson HairColorTypesJson
		query := `
			INSERT INTO haircolor_type (
				color
			) VALUES (
				:color
			)
		`
		err := json.NewDecoder(r.Body).Decode(&hairColorTypesJson)
		if err != nil {
			log.Printf(fmt.Sprintf("json decode error: %v body:%v", err, r.Body))
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		// jsonバリデーション
		err = hairColorTypesJson.Validate()
		if err != nil {
			log.Printf(fmt.Sprintf("validation error: %v", err))
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
			return
		}
		for _, hairColorType := range hairColorTypesJson.HairColorTypes {
			// jsonのバリデーションを通過したデータをDBに登録
			err = hairColorType.Validate()
			if err != nil {
				log.Printf(fmt.Sprintf("validation error: %v", err))
				http.Error(w, err.Error(), http.StatusUnprocessableEntity)
				return
			}
			_, err = db.NamedExecContext(r.Context(), query, hairColorType)
			if err != nil {
				log.Printf(fmt.Sprintf("db error: %v", err))
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
		// レスポンスの作成
		err = json.NewEncoder(w).Encode(&hairColorTypesJson)
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
