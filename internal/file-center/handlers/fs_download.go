package handlers

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/disintegration/imaging"
	"github.com/gorilla/mux"
	"github.com/sbasestarter/file-center/internal/config"
	"github.com/sgostarter/libeasygo/cuserror"
	"github.com/sgostarter/libeasygo/pathutils"
	"github.com/sgostarter/libfs"
)

// HandleFsDownload function
func HandleFsDownload(cfg *config.Config) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		err := handleFsDownload(w, r, cfg)
		if err != nil {
			cfg.ContextLogger.Warn(r.Context(), err)

			http.NotFound(w, r)
		}
	}
}

func HandleFsDownloadV1(cfg *config.Config) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		err := handleFsDownloadV1(w, r, cfg)
		if err != nil {
			cfg.ContextLogger.Warn(r.Context(), err)

			http.NotFound(w, r)
		}
	}
}

func HandleFsImageCacheV1(cfg *config.Config) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		err := handleFsImageCacheV1(w, r, cfg)
		if err != nil {
			cfg.ContextLogger.Warn(r.Context(), err)

			http.NotFound(w, r)
		}
	}
}

func cacheImage(fileid string, cfg *config.Config, expectW, expectH int64, thumbnailDir, thumbnailTmpFile,
	thumbnailFile string) error {
	item, err := libfs.NewSFSItemFromFileID(fileid, cfg.StgRoot, cfg.StgTmpRoot)
	if err != nil {
		return err
	}

	dataFile, err := item.GetDataFile()
	if err != nil {
		return err
	}

	exists, err := pathutils.IsFileExists(dataFile)
	if err != nil {
		return err
	}

	if !exists {
		return cuserror.NewWithErrorMsg(fmt.Sprintf("%s not exists", fileid))
	}

	src, err := loadImage(dataFile)
	if err != nil {
		return err
	}

	bound := src.Bounds()
	dx := bound.Dx()
	dy := bound.Dy()

	var thumbnailsWidth, thumbnailsHeight int
	if expectW != 0 && expectH == 0 {
		thumbnailsWidth = int(expectW)
		thumbnailsHeight = int(expectW) * dy / dx
	} else if expectH != 0 && expectW == 0 {
		thumbnailsHeight = int(expectH)
		thumbnailsWidth = int(expectH) * dx / dy
	} else if expectW == 0 && expectH == 0 {
		thumbnailsWidth = dx
		thumbnailsHeight = dy
	} else {
		thumbnailsWidth = int(expectW)
		thumbnailsHeight = int(expectH)
	}

	thumb := imaging.Thumbnail(src, thumbnailsWidth, thumbnailsHeight, imaging.CatmullRom)
	dst := imaging.New(thumbnailsWidth, thumbnailsHeight, color.NRGBA{
		R: 0,
		G: 0,
		B: 0,
		A: 0,
	})
	dst = imaging.Paste(dst, thumb, image.Pt(0, 0))

	err = pathutils.MustDirExists(thumbnailDir)
	if err != nil {
		return err
	}

	file, err := os.Create(thumbnailTmpFile)
	if err != nil {
		return err
	}

	defer func() {
		_ = file.Close()
	}()

	err = png.Encode(file, dst)
	if err != nil {
		return err
	}

	_ = file.Close()

	err = os.Rename(thumbnailTmpFile, thumbnailFile)
	if err != nil {
		return err
	}

	return nil
}

func handleFsImageCacheV1(w http.ResponseWriter, r *http.Request, cfg *config.Config) error {
	vars := mux.Vars(r)
	fileid := vars["file_id"]

	if len([]rune(fileid)) < 32 {
		return cuserror.NewWithErrorMsg("error:FileID incorrect")
	}

	expectW, err := strconv.ParseInt(vars["width"], 10, 64)
	if err != nil {
		return err
	}

	if expectW <= 0 {
		expectW = 0
	}

	expectH, err := strconv.ParseInt(vars["height"], 10, 64)
	if err != nil {
		return err
	}

	if expectH <= 0 {
		expectH = 0
	}

	expectWStr := strconv.FormatInt(expectW, 10)
	expectHStr := strconv.FormatInt(expectH, 10)

	imageCachePool := filepath.Join(cfg.StgRoot, "image_cache_pool")

	thumbnailDir := filepath.Join(imageCachePool, expectWStr, expectHStr)
	thumbnailFile := filepath.Join(thumbnailDir, fileid)
	thumbnailTmpFile := thumbnailFile + ".tmp"

	if exists, err := pathutils.IsFileExists(thumbnailFile); !exists || err != nil {
		err = cacheImage(fileid, cfg, expectW, expectH, thumbnailDir, thumbnailTmpFile, thumbnailFile)
		if err != nil {
			return err
		}
	}

	http.ServeFile(w, r, thumbnailFile)

	return nil
}

func loadImage(path string) (img image.Image, err error) {
	file, err := os.Open(path)
	if err != nil {
		return
	}

	defer func() {
		_ = file.Close()
	}()

	img, _, err = image.Decode(file)

	return
}

func processNewSizeImage(w http.ResponseWriter, r *http.Request, dataFile string) error {
	expectW, err := strconv.ParseInt(r.URL.Query().Get("w"), 10, 64)
	if err != nil {
		return err
	}

	if expectW <= 0 {
		return cuserror.NewWithErrorMsg("noImageX")
	}

	expectH, err := strconv.ParseInt(r.URL.Query().Get("h"), 10, 64)
	if err != nil || expectH <= 0 {
		expectH = 0
	}

	src, err := loadImage(dataFile)
	if err != nil {
		return cuserror.NewWithErrorMsg("loadImageFailed")
	}

	bound := src.Bounds()
	dx := bound.Dx()
	dy := bound.Dy()

	thumbnailsWidth := int(expectW)
	thumbnailsHeight := int(expectW) * dy / dx

	if expectH != 0 {
		thumbnailsHeight = int(expectH)
	}

	thumb := imaging.Thumbnail(src, thumbnailsWidth, thumbnailsHeight, imaging.CatmullRom)
	dst := imaging.New(thumbnailsWidth, thumbnailsHeight, color.NRGBA{R: 0, G: 0, B: 0, A: 0})
	dst = imaging.Paste(dst, thumb, image.Pt(0, 0))

	header := w.Header()
	header.Add("Content-Type", "image/jpeg")

	return png.Encode(w, dst)
}

func handleFsDownloadV1(w http.ResponseWriter, r *http.Request, cfg *config.Config) error {
	vars := mux.Vars(r)

	fileid := vars["file_id"]

	if strings.HasPrefix(fileid, "raw-") {
		img := filepath.Join(cfg.StgRoot, "raw", fileid[4:])
		if exists, err := pathutils.IsFileExists(img); err != nil || !exists {
			return cuserror.NewWithErrorMsg(fmt.Sprintf("%s not exists", fileid))
		}

		http.ServeFile(w, r, img)

		return nil
	}

	if len([]rune(fileid)) < 32 {
		return cuserror.NewWithErrorMsg("error:FileID incorrect")
	}

	item, err := libfs.NewSFSItemFromFileID(fileid, cfg.StgRoot, cfg.StgTmpRoot)
	if err != nil {
		return err
	}

	dataFile, err := item.GetDataFile()
	if err != nil {
		return err
	}

	exists, err := pathutils.IsFileExists(dataFile)
	if err != nil {
		return err
	}

	if !exists {
		return cuserror.NewWithErrorMsg(fmt.Sprintf("%s not exists", fileid))
	}

	tp := r.URL.Query().Get("type")
	if tp == "image" {
		err = processNewSizeImage(w, r, dataFile)
		if err == nil {
			return nil
		}
	}

	http.ServeFile(w, r, dataFile)

	return nil
}

func handleFsDownload(w http.ResponseWriter, r *http.Request, cfg *config.Config) error {
	vars := mux.Vars(r)

	fileid := vars["file_id"]

	if strings.HasPrefix(fileid, "raw-") {
		img := filepath.Join(cfg.StgRoot, "raw", fileid[4:])
		if exists, err := pathutils.IsFileExists(img); err != nil || !exists {
			return cuserror.NewWithErrorMsg(fmt.Sprintf("%s not exists", fileid))
		}

		http.ServeFile(w, r, img)

		return nil
	}

	item, err := libfs.NewSFSItemFromFileID(fileid, cfg.StgRoot, cfg.StgTmpRoot)
	if err != nil {
		return err
	}

	dataFile, err := item.GetDataFile()
	if err != nil {
		return err
	}

	exists, err := pathutils.IsFileExists(dataFile)
	if err != nil {
		return err
	}

	if !exists {
		return cuserror.NewWithErrorMsg(fmt.Sprintf("%s not exists", fileid))
	}

	http.ServeFile(w, r, dataFile)

	return nil
}
