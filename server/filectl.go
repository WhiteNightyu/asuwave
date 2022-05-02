package server

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/scutrobotlab/asuwave/fromelf"
	"github.com/scutrobotlab/asuwave/variable"
)

func removeWathcer() error {
	l := fromelf.Watcher.WatchList()
	for _, p := range l {
		err := fromelf.Watcher.Remove(p)
		if err != nil {
			return err
		}
	}
	return nil
}

// 上传elf或axf文件
func fileUploadCtrl(w http.ResponseWriter, r *http.Request) {
	defer variable.Refresh()
	defer r.Body.Close()
	w.Header().Set("Content-Type", "application/json")
	switch r.Method {
	case http.MethodPut:
		err := removeWathcer()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			io.WriteString(w, errorJson(err.Error()))
			return
		}

		r.ParseMultipartForm(32 << 20)
		file, _, err := r.FormFile("file")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			io.WriteString(w, errorJson(err.Error()))
			return
		}
		defer file.Close()

		tempFile, err := os.CreateTemp("", "elf")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			io.WriteString(w, errorJson(err.Error()))
			return
		}
		defer os.Remove(tempFile.Name())

		io.Copy(tempFile, file)

		f, err := fromelf.Check(tempFile)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			io.WriteString(w, errorJson(err.Error()))
			return
		}
		defer f.Close()

		err = fromelf.ReadVariable(&variable.ToProj, f)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			io.WriteString(w, errorJson(err.Error()))
			return
		}

		w.WriteHeader(http.StatusNoContent)
		io.WriteString(w, "")
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		io.WriteString(w, errorJson(http.StatusText(http.StatusMethodNotAllowed)))
	}
}

// 监控elf或axf文件
func filePathCtrl(w http.ResponseWriter, r *http.Request) {
	defer variable.Refresh()
	defer r.Body.Close()
	w.Header().Set("Content-Type", "application/json")
	switch r.Method {
	case http.MethodGet:
		j := fromelf.Watcher.WatchList()
		b, _ := json.Marshal(j)
		io.WriteString(w, string(b))

	case http.MethodPut:
		j := struct {
			Path string
		}{}
		data, _ := io.ReadAll(r.Body)
		err := json.Unmarshal(data, &j)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			io.WriteString(w, errorJson("Invaild json"))
			return
		}

		err = removeWathcer()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			io.WriteString(w, errorJson(err.Error()))
			return
		}

		err = fromelf.Watcher.Add(j.Path)
		if err != nil {
			log.Fatal(err)
		}

		file, err := os.Open(j.Path)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			io.WriteString(w, errorJson(err.Error()))
			return
		}

		f, err := fromelf.Check(file)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			io.WriteString(w, errorJson(err.Error()))
			return
		}
		defer f.Close()

		err = fromelf.ReadVariable(&variable.ToProj, f)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			io.WriteString(w, errorJson(err.Error()))
			return
		}

		w.WriteHeader(http.StatusNoContent)
		io.WriteString(w, "")

	case http.MethodDelete:
		err := removeWathcer()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			io.WriteString(w, errorJson(err.Error()))
			return
		}

		w.WriteHeader(http.StatusNoContent)
		io.WriteString(w, "")

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		io.WriteString(w, errorJson(http.StatusText(http.StatusMethodNotAllowed)))
	}
}