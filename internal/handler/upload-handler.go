package handler

import (
	"context"
	"fmt"
	"github.com/souravbiswassanto/concurrent-file-server/internal/util"
	"github.com/souravbiswassanto/concurrent-file-server/protocol/client/tcp"
)

func GetHandler(ctx context.Context, uc util.UploadConfig) (util.ConnHandler, error) {

	switch uc.Protocol {
	case "tcp":
		th, err := tcp.NewUploadHandler(ctx, uc)
		if err != nil {
			return nil, err
		}
		return th, nil
	}
	return nil, fmt.Errorf("no connection type matched")
}

func HandleUpload(uc util.UploadConfig) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	h, err := GetHandler(ctx, uc)
	if err != nil {
		return err
	}

	return h.HandleConn()
}
