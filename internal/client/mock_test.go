package client

import (
	"context"
	"errors"
	"io"

	api "github.com/alexsey-popov/gkeep/proto"
	pbCard "github.com/alexsey-popov/gkeep/proto/card"
	pbFile "github.com/alexsey-popov/gkeep/proto/file"
	pbPassword "github.com/alexsey-popov/gkeep/proto/password"
	pbText "github.com/alexsey-popov/gkeep/proto/text"
	pb "github.com/alexsey-popov/gkeep/proto/user"
	"google.golang.org/grpc"
)

type mockUsersClient struct {
	registerFunc       func(ctx context.Context, in *pb.Request, opts ...grpc.CallOption) (*pb.Response, error)
	loginFunc          func(ctx context.Context, in *pb.Request, opts ...grpc.CallOption) (*pb.Response, error)
	passwordListFunc   func(ctx context.Context, in *api.ListRequest, opts ...grpc.CallOption) (*pbPassword.ListResponse, error)
	passwordCreateFunc func(ctx context.Context, in *pbPassword.CreateRequest, opts ...grpc.CallOption) (*pbPassword.Response, error)
	passwordShowFunc   func(ctx context.Context, in *api.IDRequest, opts ...grpc.CallOption) (*pbPassword.Response, error)
	passwordUpdateFunc func(ctx context.Context, in *pbPassword.UpdateRequest, opts ...grpc.CallOption) (*pbPassword.Response, error)
	passwordDeleteFunc func(ctx context.Context, in *api.IDRequest, opts ...grpc.CallOption) (*pbPassword.Response, error)
	cardListFunc       func(ctx context.Context, in *api.ListRequest, opts ...grpc.CallOption) (*pbCard.ListResponse, error)
	cardCreateFunc     func(ctx context.Context, in *pbCard.CreateRequest, opts ...grpc.CallOption) (*pbCard.Response, error)
	cardShowFunc       func(ctx context.Context, in *api.IDRequest, opts ...grpc.CallOption) (*pbCard.Response, error)
	cardUpdateFunc     func(ctx context.Context, in *pbCard.UpdateRequest, opts ...grpc.CallOption) (*pbCard.Response, error)
	cardDeleteFunc     func(ctx context.Context, in *api.IDRequest, opts ...grpc.CallOption) (*pbCard.Response, error)
	textListFunc       func(ctx context.Context, in *api.ListRequest, opts ...grpc.CallOption) (*pbText.ListResponse, error)
	textCreateFunc     func(ctx context.Context, in *pbText.CreateRequest, opts ...grpc.CallOption) (*pbText.Response, error)
	textShowFunc       func(ctx context.Context, in *api.IDRequest, opts ...grpc.CallOption) (*pbText.Response, error)
	textUpdateFunc     func(ctx context.Context, in *pbText.UpdateRequest, opts ...grpc.CallOption) (*pbText.Response, error)
	textDeleteFunc     func(ctx context.Context, in *api.IDRequest, opts ...grpc.CallOption) (*pbText.Response, error)
	fileListFunc       func(ctx context.Context, in *api.ListRequest, opts ...grpc.CallOption) (*pbFile.ListResponse, error)
	fileShowFunc       func(ctx context.Context, in *api.IDRequest, opts ...grpc.CallOption) (*pbFile.Response, error)
	fileDeleteFunc     func(ctx context.Context, in *api.IDRequest, opts ...grpc.CallOption) (*pbFile.Response, error)
	fileUploadFunc     func(ctx context.Context, opts ...grpc.CallOption) (grpc.ClientStreamingClient[pbFile.UploadFileRequest, pbFile.UploadFileResponse], error)
	fileDownloadFunc   func(ctx context.Context, in *pbFile.DownloadFileRequest, opts ...grpc.CallOption) (grpc.ServerStreamingClient[pbFile.DownloadFileResponse], error)
}

func (m *mockUsersClient) RegisterUser(ctx context.Context, in *pb.Request, opts ...grpc.CallOption) (*pb.Response, error) {
	if m.registerFunc != nil {
		return m.registerFunc(ctx, in, opts...)
	}
	return nil, errors.New("not implemented")
}

func (m *mockUsersClient) LoginUser(ctx context.Context, in *pb.Request, opts ...grpc.CallOption) (*pb.Response, error) {
	if m.loginFunc != nil {
		return m.loginFunc(ctx, in, opts...)
	}
	return nil, errors.New("not implemented")
}

func (m *mockUsersClient) PasswordList(ctx context.Context, in *api.ListRequest, opts ...grpc.CallOption) (*pbPassword.ListResponse, error) {
	if m.passwordListFunc != nil {
		return m.passwordListFunc(ctx, in, opts...)
	}
	return nil, errors.New("not implemented")
}

func (m *mockUsersClient) PasswordCreate(ctx context.Context, in *pbPassword.CreateRequest, opts ...grpc.CallOption) (*pbPassword.Response, error) {
	if m.passwordCreateFunc != nil {
		return m.passwordCreateFunc(ctx, in, opts...)
	}
	return nil, errors.New("not implemented")
}

func (m *mockUsersClient) PasswordShow(ctx context.Context, in *api.IDRequest, opts ...grpc.CallOption) (*pbPassword.Response, error) {
	if m.passwordShowFunc != nil {
		return m.passwordShowFunc(ctx, in, opts...)
	}
	return nil, errors.New("not implemented")
}

func (m *mockUsersClient) PasswordUpdate(ctx context.Context, in *pbPassword.UpdateRequest, opts ...grpc.CallOption) (*pbPassword.Response, error) {
	if m.passwordUpdateFunc != nil {
		return m.passwordUpdateFunc(ctx, in, opts...)
	}
	return nil, errors.New("not implemented")
}

func (m *mockUsersClient) PasswordDelete(ctx context.Context, in *api.IDRequest, opts ...grpc.CallOption) (*pbPassword.Response, error) {
	if m.passwordDeleteFunc != nil {
		return m.passwordDeleteFunc(ctx, in, opts...)
	}
	return nil, errors.New("not implemented")
}

func (m *mockUsersClient) CardList(ctx context.Context, in *api.ListRequest, opts ...grpc.CallOption) (*pbCard.ListResponse, error) {
	if m.cardListFunc != nil {
		return m.cardListFunc(ctx, in, opts...)
	}
	return nil, errors.New("not implemented")
}

func (m *mockUsersClient) CardCreate(ctx context.Context, in *pbCard.CreateRequest, opts ...grpc.CallOption) (*pbCard.Response, error) {
	if m.cardCreateFunc != nil {
		return m.cardCreateFunc(ctx, in, opts...)
	}
	return nil, errors.New("not implemented")
}

func (m *mockUsersClient) CardShow(ctx context.Context, in *api.IDRequest, opts ...grpc.CallOption) (*pbCard.Response, error) {
	if m.cardShowFunc != nil {
		return m.cardShowFunc(ctx, in, opts...)
	}
	return nil, errors.New("not implemented")
}

func (m *mockUsersClient) CardUpdate(ctx context.Context, in *pbCard.UpdateRequest, opts ...grpc.CallOption) (*pbCard.Response, error) {
	if m.cardUpdateFunc != nil {
		return m.cardUpdateFunc(ctx, in, opts...)
	}
	return nil, errors.New("not implemented")
}

func (m *mockUsersClient) CardDelete(ctx context.Context, in *api.IDRequest, opts ...grpc.CallOption) (*pbCard.Response, error) {
	if m.cardDeleteFunc != nil {
		return m.cardDeleteFunc(ctx, in, opts...)
	}
	return nil, errors.New("not implemented")
}

func (m *mockUsersClient) TextList(ctx context.Context, in *api.ListRequest, opts ...grpc.CallOption) (*pbText.ListResponse, error) {
	if m.textListFunc != nil {
		return m.textListFunc(ctx, in, opts...)
	}
	return nil, errors.New("not implemented")
}

func (m *mockUsersClient) TextCreate(ctx context.Context, in *pbText.CreateRequest, opts ...grpc.CallOption) (*pbText.Response, error) {
	if m.textCreateFunc != nil {
		return m.textCreateFunc(ctx, in, opts...)
	}
	return nil, errors.New("not implemented")
}

func (m *mockUsersClient) TextShow(ctx context.Context, in *api.IDRequest, opts ...grpc.CallOption) (*pbText.Response, error) {
	if m.textShowFunc != nil {
		return m.textShowFunc(ctx, in, opts...)
	}
	return nil, errors.New("not implemented")
}

func (m *mockUsersClient) TextUpdate(ctx context.Context, in *pbText.UpdateRequest, opts ...grpc.CallOption) (*pbText.Response, error) {
	if m.textUpdateFunc != nil {
		return m.textUpdateFunc(ctx, in, opts...)
	}
	return nil, errors.New("not implemented")
}

func (m *mockUsersClient) TextDelete(ctx context.Context, in *api.IDRequest, opts ...grpc.CallOption) (*pbText.Response, error) {
	if m.textDeleteFunc != nil {
		return m.textDeleteFunc(ctx, in, opts...)
	}
	return nil, errors.New("not implemented")
}

func (m *mockUsersClient) FileList(ctx context.Context, in *api.ListRequest, opts ...grpc.CallOption) (*pbFile.ListResponse, error) {
	if m.fileListFunc != nil {
		return m.fileListFunc(ctx, in, opts...)
	}
	return nil, errors.New("not implemented")
}

func (m *mockUsersClient) FileShow(ctx context.Context, in *api.IDRequest, opts ...grpc.CallOption) (*pbFile.Response, error) {
	if m.fileShowFunc != nil {
		return m.fileShowFunc(ctx, in, opts...)
	}
	return nil, errors.New("not implemented")
}

func (m *mockUsersClient) FileDelete(ctx context.Context, in *api.IDRequest, opts ...grpc.CallOption) (*pbFile.Response, error) {
	if m.fileDeleteFunc != nil {
		return m.fileDeleteFunc(ctx, in, opts...)
	}
	return nil, errors.New("not implemented")
}

func (m *mockUsersClient) FileUpload(ctx context.Context, opts ...grpc.CallOption) (grpc.ClientStreamingClient[pbFile.UploadFileRequest, pbFile.UploadFileResponse], error) {
	if m.fileUploadFunc != nil {
		return m.fileUploadFunc(ctx, opts...)
	}
	return nil, errors.New("not implemented")
}

func (m *mockUsersClient) FileDownload(ctx context.Context, in *pbFile.DownloadFileRequest, opts ...grpc.CallOption) (grpc.ServerStreamingClient[pbFile.DownloadFileResponse], error) {
	if m.fileDownloadFunc != nil {
		return m.fileDownloadFunc(ctx, in, opts...)
	}
	return nil, errors.New("not implemented")
}

type mockUploadStreamClient struct {
	grpc.ClientStreamingClient[pbFile.UploadFileRequest, pbFile.UploadFileResponse]
	sentReqs     []*pbFile.UploadFileRequest
	sendErr      error
	closeResp    *pbFile.UploadFileResponse
	closeRecvErr error
}

func (m *mockUploadStreamClient) Send(req *pbFile.UploadFileRequest) error {
	if m.sendErr != nil {
		return m.sendErr
	}
	m.sentReqs = append(m.sentReqs, req)
	return nil
}

func (m *mockUploadStreamClient) CloseAndRecv() (*pbFile.UploadFileResponse, error) {
	if m.closeRecvErr != nil {
		return nil, m.closeRecvErr
	}
	return m.closeResp, nil
}

type mockDownloadStreamClient struct {
	grpc.ServerStreamingClient[pbFile.DownloadFileResponse]
	msgs    []*pbFile.DownloadFileResponse
	recvIdx int
	recvErr error
}

func (m *mockDownloadStreamClient) Recv() (*pbFile.DownloadFileResponse, error) {
	if m.recvErr != nil {
		return nil, m.recvErr
	}
	if m.recvIdx >= len(m.msgs) {
		return nil, io.EOF
	}
	msg := m.msgs[m.recvIdx]
	m.recvIdx++
	return msg, nil
}
