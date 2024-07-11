package hekiradarchart

import (
	"encoding/json"
	"log"
	"fmt"
	"net/http"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	validation "github.com/go-ozzo/ozzo-validation"
)

type HekiRadarChart struct {
	EntryID int64 `db:"entry_id"`
	AI      int64 `db:"ai"`
	NU      int64 `db:"nu"`
}

func (h *HekiRadarChart) Validate() error {
	return validation.ValidateStruct(h,
		validation.Field(&h.EntryID, validation.Required),
		validation.Field(&h.AI, validation.Required),
		validation.Field(&h.NU, validation.Required),
	)
}

type HekiRadarChartsJson struct {
	HekiRadarCharts []HekiRadarChart `json:"heki_radar_charts"`
}

func (h *HekiRadarChartsJson) Validate() error {
	return validation.ValidateStruct(h,
		validation.Field(&h.HekiRadarCharts, validation.Required),
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
		var hekiRadarChartsJson HekiRadarChartsJson
		query := `
			SELECT
				entry_id,
				ai,
				nu
			FROM
				heki_radar_chart
			WHERE
				entry_id IN (?)
		`
		queryIDs, ok := r.URL.Query()["entry_id"]
		if !ok {
			query = `
				SELECT
					entry_id,
					ai,
					nu
				FROM
					heki_radar_chart
			`
			err := db.SelectContext(r.Context(), &hekiRadarChartsJson.HekiRadarCharts, query)
			if err != nil {
				log.Println(fmt.Sprintf("db error: %v", err))
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			// json書き込み
			err = json.NewEncoder(w).Encode(&hekiRadarChartsJson)
			if err != nil {
				log.Println(fmt.Sprintf("json encode error: %v", err))
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			return
		} else if len(queryIDs) == 1 {
			query = `
				SELECT
					entry_id,
					ai,
					nu
				FROM
					heki_radar_chart
				WHERE
					entry_id = $1
			`
			err := db.SelectContext(r.Context(), &hekiRadarChartsJson.HekiRadarCharts, query, queryIDs[0])
			if err != nil {
				log.Println(fmt.Sprintf("db error: %v", err))
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			// json書き込み
			err = json.NewEncoder(w).Encode(&hekiRadarChartsJson)
			if err != nil {
				log.Println(fmt.Sprintf("json encode error: %v", err))
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			return
		}
		// idの数だけ置換文字を作成
		query, args, err := sqlx.In(query, queryIDs)
		if err != nil {
			log.Println(fmt.Sprintf("db error: %v", err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		// Postgresの場合は置換文字を$1, $2, ...とする必要がある
		query = sqlx.Rebind(len(queryIDs), query)
		err = db.SelectContext(r.Context(), &hekiRadarChartsJson.HekiRadarCharts, query, args...)
		if err != nil {
			log.Println(fmt.Sprintf("db error: %v", err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		// json書き込み
		err = json.NewEncoder(w).Encode(&hekiRadarChartsJson)
		if err != nil {
			log.Println(fmt.Sprintf("json encode error: %v", err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	case http.MethodPost:
		var HekiRadarChartsJson HekiRadarChartsJson
		query := `
			INSERT INTO heki_radar_chart (
				entry_id,
				ai,
				nu
			) VALUES (
				:entry_id,
				:ai,
				:nu
			)
		`
		err := json.NewDecoder(r.Body).Decode(&HekiRadarChartsJson)
		if err != nil {
			log.Println(fmt.Sprintf("json decode error: %v body:%v", err, r.Body))
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		// jsonバリデーション
		err = HekiRadarChartsJson.Validate()
		if err != nil {
			log.Println(fmt.Sprintf("json validate error: %v", err))
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
			return
		}
		for _, hrc := range HekiRadarChartsJson.HekiRadarCharts {
			// jsonバリデーション
			err = hrc.Validate()
			if err != nil {
				log.Println(fmt.Sprintf("json validate error: %v", err))
				http.Error(w, err.Error(), http.StatusUnprocessableEntity)
				return
			}
			_, err = db.NamedExecContext(r.Context(), query, hrc)
			if err != nil {
				log.Println(fmt.Sprintf("db error: %v", err))
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
		// json書き込み
		err = json.NewEncoder(w).Encode(&HekiRadarChartsJson)
		if err != nil {
			log.Println(fmt.Sprintf("json encode error: %v", err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	case http.MethodPut:
	case http.MethodDelete:
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}
