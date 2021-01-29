package handlers

import (
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"net/http"
	"os"
	"strconv"

	"github.com/jiuzhou-zhao/go-fundamental/loge"
	"github.com/sbasestarter/file-center/internal/config"
	"github.com/sgostarter/libfs"
	_ "golang.org/x/image/bmp"
	_ "golang.org/x/image/tiff"
	_ "golang.org/x/image/webp"
)

type fsListResponse struct {
	ErrCode int      `json:"err_code"`
	ErrMsg  string   `json:"err_msg,omitempty"`
	Files   []string `json:"files"`
}

func HandleFsList(cfg *config.Config) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		err := handleFsList(w, r, cfg)
		if err != nil {
			err = doJSONObjectResponse(r.Context(), w, &fsListResponse{
				ErrCode: -1,
				ErrMsg:  err.Error(),
			})
			if err != nil {
				loge.Error(r.Context(), err)
			}
		}
	}
}

func checkImageExt(path string) (ext string, err error) {
	file, err := os.Open(path)
	if err != nil {
		return
	}
	defer func() {
		_ = file.Close()
	}()
	_, ext, err = image.Decode(file)
	return
}

func handleFsList(w http.ResponseWriter, r *http.Request, cfg *config.Config) error {
	kvs := r.URL.Query()
	lastfileid := kvs.Get("last_fileid")
	forward, err := strconv.ParseBool(kvs.Get("forward"))
	if err != nil {
		forward = true
	}
	count, err := strconv.Atoi(kvs.Get("count"))
	if err != nil {
		count = -1
	}
	err, files := libfs.GetFileList(lastfileid, cfg.StgRoot, forward, count)
	if err != nil {
		loge.Error(r.Context(), err)
		return err
	}
	checkImage, err := strconv.ParseBool(kvs.Get("checkImage"))
	if err != nil {
		checkImage = false
	}
	if checkImage {
		for idx := 0; idx < len(files); idx++ {
			item, err := libfs.NewSFSItemFromFileID(files[idx], cfg.StgRoot, cfg.StgTmpRoot)
			if err != nil {
				continue
			}
			dataFile, err := item.GetDataFile()
			if err != nil {
				continue
			}
			ext, err := checkImageExt(dataFile)
			if err != nil {
				continue
			}
			if ext != "" {
				files[idx] += "." + ext
			}
		}
	}

	return doJSONObjectResponse(r.Context(), w, &fsListResponse{
		Files: files,
	})
}
