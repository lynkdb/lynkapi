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

package cli

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/hooto/hauth/go/hauth/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"github.com/lynkdb/lynkx/datax"
)

const (
	grpcMsgByteMax = 12 << 20
)

var (
	rpcClientConns = map[string]*grpc.ClientConn{}
	rpcClientMu    sync.Mutex

	dbMut sync.Mutex

	dataxConns = map[string]*dataxClient{}
)

type ConfigService struct {
	Name      string           `toml:"name" json:"name"`
	Addr      string           `toml:"addr" json:"addr"`
	AccessKey *hauth.AccessKey `toml:"access_key" json:"access_key"`
}

type ConfigCommon struct {
	Services   []*ConfigService `toml:"services" json:"services"`
	LastActive string           `toml:"last_active" json:"last_active"`
}

type dataxClient struct {
	_ak     string
	cfg     *ConfigService
	rpcConn *grpc.ClientConn
	ac      datax.DataxServiceClient
	err     error
}

func (it *ConfigService) NewClient() (*dataxClient, error) {

	if it.AccessKey == nil {
		return nil, errors.New("access key not setup")
	}

	ak := fmt.Sprintf("%s.%s", it.Addr, it.AccessKey.Id)

	dbMut.Lock()
	defer dbMut.Unlock()

	if dataxConns == nil {
		dataxConns = map[string]*dataxClient{}
	}

	dataxConn, ok := dataxConns[ak]
	if !ok {

		conn, err := rpcClientConnect(it.Addr, it.AccessKey, false)
		if err != nil {
			return nil, err
		}

		dataxConn = &dataxClient{
			_ak:     ak,
			cfg:     it,
			rpcConn: conn,
			ac:      datax.NewDataxServiceClient(conn),
		}
		dataxConns[ak] = dataxConn
	}

	return dataxConn, nil
}

func (it *ConfigService) timeout() time.Duration {
	return time.Second * 60
}

func (it *dataxClient) ApiList(req *datax.ApiListRequest) *datax.ApiListResponse {

	ctx, fc := context.WithTimeout(context.Background(), it.cfg.timeout())
	defer fc()

	rs, err := it.ac.ApiList(ctx, req)
	if err != nil {
		return &datax.ApiListResponse{
			Status: datax.ParseError(err),
		}
	}
	return rs
}

func (it *dataxClient) Exec(req *datax.Request) *datax.Response {

	ctx, fc := context.WithTimeout(context.Background(), it.cfg.timeout())
	defer fc()

	rs, err := it.ac.Exec(ctx, req)
	if err != nil {
		return &datax.Response{
			Status: datax.ParseError(err),
		}
	}
	return rs
}

func rpcClientConnect(
	addr string,
	key *hauth.AccessKey,
	forceNew bool,
) (*grpc.ClientConn, error) {

	if key == nil {
		return nil, errors.New("not auth key setup")
	}

	ck := fmt.Sprintf("%s.%s", addr, key.Id)

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
		grpc.WithPerRPCCredentials(newAppCredential(key)),
		grpc.WithMaxMsgSize(grpcMsgByteMax * 2),
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(grpcMsgByteMax * 2)),
		grpc.WithDefaultCallOptions(grpc.MaxCallSendMsgSize(grpcMsgByteMax * 2)),
	}

	dialOptions = append(dialOptions, grpc.WithInsecure())

	c, err := grpc.Dial(addr, dialOptions...)
	if err != nil {
		return nil, err
	}

	rpcClientConns[ck] = c

	return c, nil
}

func newAppCredential(key *hauth.AccessKey) credentials.PerRPCCredentials {
	return hauth.NewGrpcAppCredential(key)
}
