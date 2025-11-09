// Copyright 2024 Eryx <evorui at gmail dot com>, All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package lynkapi

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/hooto/hauth/go/hauth/v1"
	hauth2 "github.com/hooto/hauth/v2/hauth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

var (
	rpcClientConns = map[string]*grpc.ClientConn{}
	rpcClientMu    sync.Mutex

	dbMut sync.Mutex

	clientConns = map[string]*clientImpl{}
)

type Client interface {
	ApiList(req *ApiListRequest) *ApiListResponse
	Exec(req *Request) *Response
	//
	DataProject(req *DataProjectRequest) *DataProjectResponse
	DataQuery(req *DataQuery) *DataResult
	DataUpsert(req *DataInsert) *DataResult
}

type ClientConfig struct {
	Name      string           `toml:"name,omitempty" json:"name,omitempty"`
	Addr      string           `toml:"addr" json:"addr"`
	AccessKey *hauth.AccessKey `toml:"access_key" json:"access_key"`
}

type clientImpl struct {
	ak        string
	cfg       *ClientConfig
	rpcConn   *grpc.ClientConn
	rpcClient LynkServiceClient
	err       error

	authConnector hauth2.AuthConnector
}

func (it *ClientConfig) NewClient() (*clientImpl, error) {

	// if it.AccessKey == nil {
	// 	return nil, errors.New("access key not setup")
	// }

	var ak string

	if it.AccessKey == nil {
		ak = it.Addr
	} else {
		ak = fmt.Sprintf("%s.%s", it.Addr, it.AccessKey.Id)
	}

	dbMut.Lock()
	defer dbMut.Unlock()

	if clientConns == nil {
		clientConns = map[string]*clientImpl{}
	}

	clientConn, ok := clientConns[ak]
	if !ok {

		ac := hauth2.NewAuthConnectorWithAccessKey(it.AccessKey)

		conn, err := rpcClientConnect(it.Addr, ac, false)
		if err != nil {
			return nil, err
		}

		clientConn = &clientImpl{
			ak:            ak,
			cfg:           it,
			rpcConn:       conn,
			rpcClient:     NewLynkServiceClient(conn),
			authConnector: ac,
		}
		clientConns[ak] = clientConn
	}

	return clientConn, nil
}

func (it *ClientConfig) timeout() time.Duration {
	return time.Second * 60
}

func (it *clientImpl) tryAuth(force bool) error {

	if !force && it.authConnector != nil &&
		it.authConnector.AccessToken() != "" {
		return nil
	}

	ctx, fc := context.WithTimeout(context.Background(), it.cfg.timeout())
	defer fc()

	resp, err := it.rpcClient.Auth(ctx, &AuthRequest{
		LoginToken: it.authConnector.LoginToken(),
	})
	if err != nil {
		return err
	}

	if err := it.authConnector.RefreshAccessToken(resp.AccessToken); err != nil {
		return err
	}

	return nil
}

func (it *clientImpl) ApiList(req *ApiListRequest) *ApiListResponse {

	ctx, fc := context.WithTimeout(context.Background(), it.cfg.timeout())
	defer fc()

	rs, err := it.rpcClient.ApiList(ctx, req)
	if err != nil {
		if status, ok := status.FromError(err); ok && len(status.Message()) > 5 {
			return &ApiListResponse{
				Status: ParseError(errors.New(status.Message())),
			}
		}
		return &ApiListResponse{
			Status: ParseError(err),
		}
	}
	if rs.Status == nil {
		rs.Status = NewServiceStatusOK()
	}
	return rs
}

func (it *clientImpl) Exec(req *Request) *Response {

	if err := it.tryAuth(false); err != nil {
		return &Response{
			Status: NewServiceStatus(StatusCode_UnAuth, err.Error()),
		}
	}

	call := func(req *Request) *Response {

		ctx, fc := context.WithTimeout(context.Background(), it.cfg.timeout())
		defer fc()

		rs, err := it.rpcClient.Exec(ctx, req)
		if err != nil {
			if status, ok := status.FromError(err); ok && len(status.Message()) > 5 {
				return &Response{
					Status: ParseError(errors.New(status.Message())),
				}
			}
			return &Response{
				Status: ParseError(err),
			}
		}

		if rs.Status == nil {
			rs.Status = NewServiceStatusOK()
		}

		return rs
	}

	rs := call(req)

	if rs.Status.Code == StatusCode_AuthExpired {

		if err := it.tryAuth(true); err != nil {
			return &Response{
				Status: NewServiceStatus(StatusCode_UnAuth, err.Error()),
			}
		}

		rs = call(req)
	}
	// fmt.Println("rs.Status", rs.jtatus)
	return rs
}

func (it *clientImpl) DataProject(req *DataProjectRequest) *DataProjectResponse {

	ctx, fc := context.WithTimeout(context.Background(), it.cfg.timeout())
	defer fc()

	rs, err := it.rpcClient.DataProject(ctx, req)
	if err != nil {
		if status, ok := status.FromError(err); ok && len(status.Message()) > 5 {
			return &DataProjectResponse{
				Status: ParseError(errors.New(status.Message())),
			}
		}
		return &DataProjectResponse{
			Status: ParseError(err),
		}
	}
	if rs.Status == nil {
		rs.Status = NewServiceStatusOK()
	}
	return rs
}

func (it *clientImpl) DataQuery(req *DataQuery) *DataResult {

	ctx, fc := context.WithTimeout(context.Background(), it.cfg.timeout())
	defer fc()

	rs, err := it.rpcClient.DataQuery(ctx, req)
	if err != nil {
		if status, ok := status.FromError(err); ok && len(status.Message()) > 5 {
			return &DataResult{
				Status: ParseError(errors.New(status.Message())),
			}
		}
		return &DataResult{
			Status: ParseError(err),
		}
	}
	if rs.Status == nil {
		rs.Status = NewServiceStatusOK()
	}
	return rs
}

func (it *clientImpl) DataUpsert(req *DataInsert) *DataResult {

	ctx, fc := context.WithTimeout(context.Background(), it.cfg.timeout())
	defer fc()

	rs, err := it.rpcClient.DataUpsert(ctx, req)
	if err != nil {
		if status, ok := status.FromError(err); ok && len(status.Message()) > 5 {
			return &DataResult{
				Status: ParseError(errors.New(status.Message())),
			}
		}
		return &DataResult{
			Status: ParseError(err),
		}
	}
	if rs.Status == nil {
		rs.Status = NewServiceStatusOK()
	}
	return rs
}

func rpcClientConnect(
	addr string,
	ac hauth2.AuthConnector,
	forceNew bool,
) (*grpc.ClientConn, error) {

	// if ac == nil {
	// 	return nil, errors.New("not auth key setup")
	// }

	var ck string

	if ac != nil {
		ck = fmt.Sprintf("%s.%s", addr, ac.AccessKey().Id)
	} else {
		ck = addr
	}

	rpcClientMu.Lock()
	defer rpcClientMu.Unlock()

	if c, ok := rpcClientConns[ck]; ok {
		if forceNew {
			c.Close()
			c = nil
			delete(rpcClientConns, ck)
		} else {
			return c, nil
		}
	}

	dialOptions := []grpc.DialOption{
		grpc.WithMaxMsgSize(grpcMsgByteMax * 2),
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(grpcMsgByteMax * 2)),
		grpc.WithDefaultCallOptions(grpc.MaxCallSendMsgSize(grpcMsgByteMax * 2)),
	}

	if ac != nil {
		dialOptions = append(dialOptions, grpc.WithPerRPCCredentials(hauth2.NewGrpcAppCredential(ac)))
	}

	dialOptions = append(dialOptions, grpc.WithInsecure())

	c, err := grpc.Dial(addr, dialOptions...)
	if err != nil {
		return nil, err
	}

	rpcClientConns[ck] = c

	return c, nil
}
