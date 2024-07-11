package personality

import (
	"encoding/json"
	"log"
	"fmt"
	"net/http"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	validation "github.com/go-ozzo/ozzo-validation"
)

type Personality struct {
	EntryID int64 `db:"entry_id"`
	TypeID  int64 `db:"type_id"`
}

func (p *Personality) Validate() error {
	return validation.ValidateStruct(p,
		validation.Field(&p.EntryID, validation.Required),
		validation.Field(&p.TypeID, validation.Required),
	)
}

type PersonalitiesJson struct {
	Personalities []Personality `json:"personalities"`
}

func (p *PersonalitiesJson) Validate() error {
	return validation.ValidateStruct(p,
		validation.Field(&p.Personalities, validation.Required),
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
		var personalitiesJson PersonalitiesJson
		query := `
			SELECT
				entry_id,
				type_id
			FROM
				personality
			WHERE
				entry_id IN (?)
		`
		queryIDs, ok := r.URL.Query()["entry_id"]
		if !ok {
			query = `
				SELECT
					entry_id,
					type_id
				FROM
					personality
			`
			err := db.SelectContext(r.Context(), &personalitiesJson.Personalities, query)
			if err != nil {
				log.Printf(fmt.Sprintf("select error: %v", err))
				http.Error(w, err.Error(), http.StatusBadRequest)
			}
			// json encode
			err = json.NewEncoder(w).Encode(&personalitiesJson)
			if err != nil {
				log.Printf(fmt.Sprintf("json encode error: %v", err))
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		} else if len(queryIDs) == 1 {
			query = `
				SELECT
					entry_id,
					type_id
				FROM
					personality
				WHERE
					entry_id = $1
			`
			err := db.SelectContext(r.Context(), &personalitiesJson.Personalities, query, queryIDs[0])
			if err != nil {
				log.Printf(fmt.Sprintf("select error: %v", err))
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			// json encode
			err = json.NewEncoder(w).Encode(&personalitiesJson)
			if err != nil {
				log.Printf(fmt.Sprintf("json encode error: %v", err))
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}
		// idの数だけ置換文字を作成
		query, args, err := sqlx.In(query, queryIDs)
		if err != nil {
			log.Printf(fmt.Sprintf("in query error: %v", err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		// Postgresの場合は置換文字を$1, $2, ...とする必要がある
		query = sqlx.Rebind(len(queryIDs), query)
		err = db.SelectContext(r.Context(), &personalitiesJson.Personalities, query, args...)
		if err != nil {
			log.Printf(fmt.Sprintf("select error: %v", err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		// json encode
		err = json.NewEncoder(w).Encode(&personalitiesJson)
		if err != nil {
			log.Printf(fmt.Sprintf("json encode error: %v", err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	case http.MethodPost:
		var personalitiesJson PersonalitiesJson
		query := `
			INSERT INTO personality (
				entry_id,
				type_id
			) VALUES (
				:entry_id,
				:type_id
			)
		`
		// json decode
		err := json.NewDecoder(r.Body).Decode(&personalitiesJson)
		if err != nil {
			log.Printf(fmt.Sprintf("json decode error: %v body:%v", err, r.Body))
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		// validation
		err = personalitiesJson.Validate()
		if err != nil {
			log.Printf(fmt.Sprintf("validation error: %v", err))
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		}
		for _, personality := range personalitiesJson.Personalities {
			// validation
			err = personality.Validate()
			if err != nil {
				log.Printf(fmt.Sprintf("validation error: %v", err))
				http.Error(w, err.Error(), http.StatusUnprocessableEntity)
			}
			_, err = db.NamedExecContext(r.Context(), query, personality)
			if err != nil {
				log.Printf(fmt.Sprintf("insert error: %v", err))
				http.Error(w, err.Error(), http.StatusBadRequest)
			}
		}
		// json encode
		err = json.NewEncoder(w).Encode(&personalitiesJson)
		if err != nil {
			log.Printf(fmt.Sprintf("json encode error: %v", err))
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
	case http.MethodPut:
		var personalitiesJson PersonalitiesJson
		query := `
			UPDATE
				personality
			SET
				type_id = :type_id
			WHERE
				entry_id = :entry_id
		`
		err := json.NewDecoder(r.Body).Decode(&personalitiesJson)
		if err != nil {
			log.Printf(fmt.Sprintf("json decode error: %v body:%v", err, r.Body))
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		// validation
		err = personalitiesJson.Validate()
		if err != nil {
			log.Printf(fmt.Sprintf("validation error: %v", err))
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		}
		for _, personality := range personalitiesJson.Personalities {
			// validation
			err = personality.Validate()
			if err != nil {
				log.Printf(fmt.Sprintf("validation error: %v", err))
				http.Error(w, err.Error(), http.StatusUnprocessableEntity)
			}
			_, err = db.NamedExecContext(r.Context(), query, personality)
			if err != nil {
				log.Printf(fmt.Sprintf("update error: %v", err))
				http.Error(w, err.Error(), http.StatusBadRequest)
			}
		}
		// json encode
		err = json.NewEncoder(w).Encode(&personalitiesJson)
		if err != nil {
			log.Printf(fmt.Sprintf("json encode error: %v", err))
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
	case http.MethodDelete:
		var delIDs IDs
		query := `
			DELETE FROM
				personality
			WHERE
				entry_id IN (?)
		`
		err := json.NewDecoder(r.Body).Decode(&delIDs)
		if err != nil {
			log.Printf(fmt.Sprintf("json decode error: %v body:%v", err, r.Body))
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		// validation
		err = delIDs.Validate()
		if err != nil {
			log.Printf(fmt.Sprintf("validation error: %v", err))
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		}
		if len(delIDs.IDs) == 0 {
			return
		} else if len(delIDs.IDs) == 1 {
			query = `
				DELETE FROM
					personality
				WHERE
					entry_id = $1
			`
			_, err = db.ExecContext(r.Context(), query, delIDs.IDs[0])
			if err != nil {
				log.Printf(fmt.Sprintf("delete error: %v", err))
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			// json encode
			err = json.NewEncoder(w).Encode(&delIDs)
			if err != nil {
				log.Printf(fmt.Sprintf("json encode error: %v", err))
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}
		// idの数だけ置換文字を作成
		query, args, err := sqlx.In(query, delIDs.IDs)
		if err != nil {
			log.Printf(fmt.Sprintf("in query error: %v", err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		// Postgresの場合は置換文字を$1, $2, ...とする必要がある
		query = sqlx.Rebind(len(delIDs.IDs), query)
		_, err = db.ExecContext(r.Context(), query, args...)
		if err != nil {
			log.Printf(fmt.Sprintf("delete error: %v", err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		// json encode
		err = json.NewEncoder(w).Encode(&delIDs)
		if err != nil {
			log.Printf(fmt.Sprintf("json encode error: %v", err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}
