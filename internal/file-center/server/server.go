package server

import (
	"bytes"
	"context"
	"strings"

	"github.com/gorilla/mux"
	uuid "github.com/satori/go.uuid"
	"github.com/sbasestarter/file-center/internal/config"
	"github.com/sbasestarter/file-center/internal/file-center/handlers"
	filepb "github.com/sbasestarter/proto-repo/gen/protorepo-file-go"
	sharepb "github.com/sbasestarter/proto-repo/gen/protorepo-share-go"
	"github.com/sgostarter/i/l"
	"github.com/sgostarter/libfs"
	"github.com/sgostarter/libservicetoolset/dbtoolset"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	APIVer = "/r/v1"
)

type Server struct {
	cfg       *config.Config
	dbToolset *dbtoolset.Toolset
	logger    l.WrapperWithContext
}

func (s *Server) GetKV(ctx context.Context, request *filepb.GetKVRequest) (*filepb.GetKVResponse, error) {
	redisCli := s.dbToolset.GetRedisByName("data")
	if redisCli == nil {
		return nil, status.Error(codes.Unimplemented, "")
	}

	v, err := redisCli.Get(ctx, request.GetKey()).Result()
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &filepb.GetKVResponse{Value: v}, nil
}

func (s *Server) SetKV(ctx context.Context, request *filepb.SetKVRequest) (*sharepb.Empty, error) {
	redisCli := s.dbToolset.GetRedisByName("data")
	if redisCli == nil {
		return nil, status.Error(codes.Unimplemented, "")
	}

	err := redisCli.Set(ctx, request.GetKey(), request.GetValue(), 0).Err()
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &sharepb.Empty{}, nil
}

func NewServer(ctx context.Context, cfg *config.Config, dbToolset *dbtoolset.Toolset) *Server {
	if cfg.Logger == nil {
		cfg.Logger = l.NewNopLoggerWrapper()
	}

	return &Server{
		cfg:       cfg,
		dbToolset: dbToolset,
		logger:    cfg.Logger.WithFields(l.StringField(l.ClsKey, "Server")).GetWrapperWithContext(),
	}
}

func (s *Server) saveFile(fileName string, content []byte) (fileID string, err error) {
	item, err := libfs.NewSFSItem(fileName, s.cfg.StgRoot, s.cfg.StgTmpRoot)
	if err != nil {
		return
	}

	err = item.WriteFile(bytes.NewBuffer(content))
	if err != nil {
		return
	}

	err = item.WriteFileRecord()
	if err != nil {
		return
	}

	return item.GetFileID()
}

func (s *Server) UpdateFile(ctx context.Context, r *filepb.UpdateFileRequest) (*filepb.UpdateFileResponse, error) {
	if r.FileName == "" || strings.Contains(r.FileName, "\\/") {
		s.logger.Warnf(ctx, "null or invalid file name: %v", r.FileName)
		r.FileName = uuid.NewV4().String()
	}

	fileID, err := s.saveFile(r.FileName, r.Content)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &filepb.UpdateFileResponse{
		FileUrl: fileID,
	}, nil
}

func (s *Server) HTTPRegister(r *mux.Router) {
	r.HandleFunc(APIVer+"/fs/query", handlers.HandleFsQuery(s.cfg))
	r.HandleFunc(APIVer+"/fs/upload", handlers.HandleFsUpload(s.cfg))
	r.HandleFunc(APIVer+"/fs/download/{file_id}", handlers.HandleFsDownload(s.cfg))
	r.HandleFunc(APIVer+"/fs/transfer", handlers.HandleFsTransfer(s.cfg))
	r.HandleFunc(APIVer+"/fs/list", handlers.HandleFsList(s.cfg))
	r.HandleFunc("/download/{file_id}", handlers.HandleFsDownloadV1(s.cfg))
	r.HandleFunc("/thumbnail/{width}/{height}/{file_id}", handlers.HandleFsImageCacheV1(s.cfg))
}
