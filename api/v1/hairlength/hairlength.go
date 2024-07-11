package hairlength

import (
	"encoding/json"
	"log"
	"fmt"
	"net/http"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	validation "github.com/go-ozzo/ozzo-validation"
)

type HairLength struct {
	EntryID          int64 `db:"entry_id"`
	HairLengthTypeID int64 `db:"hairlength_type_id"`
}

func (h *HairLength) Validate() error {
	return validation.ValidateStruct(h,
		validation.Field(&h.EntryID, validation.Required),
		validation.Field(&h.HairLengthTypeID, validation.Required),
	)
}

type HairLengthsJson struct {
	HairLengths []HairLength `json:"hairlengths"`
}

func (h *HairLengthsJson) Validate() error {
	return validation.ValidateStruct(h,
		validation.Field(&h.HairLengths, validation.Required),
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
		var hairLengthsJson HairLengthsJson
		query := `
			INSERT INTO hairlength (
				entry_id,
				hairlength_type_id
			) VALUES (
				:entry_id,
				:hairlength_type_id
			)
		`
		if err := json.NewDecoder(r.Body).Decode(&hairLengthsJson); err != nil {
			log.Printf(fmt.Sprintf("json decode error: %v body:%v", err, r.Body))
			http.Error(w, "json decode error", http.StatusBadRequest)
		}
		// jsonバリデーション
		err := hairLengthsJson.Validate()
		if err != nil {
			log.Printf(fmt.Sprintf("validation error: %v", err))
			http.Error(w, "validation error", http.StatusUnprocessableEntity)
		}
		for _, hl := range hairLengthsJson.HairLengths {
			// jsonのバリデーションを通過したデータをDBに登録
			err = hl.Validate()
			if err != nil {
				log.Printf(fmt.Sprintf("validation error: %v", err))
				http.Error(w, "validation error", http.StatusUnprocessableEntity)
			}
			if _, err := db.NamedExecContext(r.Context(), query, hl); err != nil {
				log.Printf(fmt.Sprintf("db error: %v", err))
				http.Error(w, "db error", http.StatusInternalServerError)
			}
		}
		// jsonを返す
		err = json.NewEncoder(w).Encode(&hairLengthsJson)
		if err != nil {
			log.Printf(fmt.Sprintf("json encode error: %v", err))
			http.Error(w, "json encode error", http.StatusInternalServerError)
		}
	case http.MethodPut:
	case http.MethodDelete:
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}
