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
	"crypto/tls"
	"net"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	_ "google.golang.org/grpc/encoding/gzip"

	hauth2 "github.com/hooto/hauth/v2/hauth"
	"github.com/hooto/hlog4g/hlog"
)

type ServerConfig struct {
	Bind    string         `toml:"bind" json:"bind"`
	TLSCert *TLSCertConfig `toml:"tls_cert" json:"tls_cert"`
}

type TLSCertConfig struct {
	KeyFile  string `toml:"key_file" json:"key_file"`
	KeyData  string `toml:"key_data" json:"key_data"`
	CertFile string `toml:"cert_file" json:"cert_file"`
	CertData string `toml:"cert_data" json:"cert_data"`
}

const grpcMsgByteMax = 16 << 20

type LynkServer struct {
	cfg        *ServerConfig
	listener   net.Listener
	Service    *LynkService
	GrpcServer *grpc.Server
}

func NewServer(cfg *ServerConfig) (*LynkServer, error) {

	srv := &LynkServer{
		cfg:     cfg,
		Service: NewService(),
	}

	srv.Service.RegisterService(srv.Service)

	if err := srv.grpcSetup(); err != nil {
		return nil, err
	}

	return srv, nil
}

func (it *LynkServer) Run() error {
	RegisterLynkServiceServer(it.GrpcServer, it.Service)
	go func() {
		hlog.Printf("info", "lynkapi/grpc-server run")
		if err := it.GrpcServer.Serve(it.listener); err != nil {
			hlog.Printf("info", "lynkapi/grpc-server run fail %s", err.Error())
		}
	}()
	return nil
}

func (it *LynkServer) SetupIdentityAuthService(s hauth2.IdentityAuthService) {
	it.Service.identityAuthService = s
}

func (it *LynkServer) grpcSetup() error {

	host, port, err := net.SplitHostPort(it.cfg.Bind)
	if err != nil {
		return err
	}

	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return err
	}
	hlog.Printf("info", "lynkapi/grpc-server bind %s:%s", host, port)

	it.listener = lis

	it.cfg.Bind = host + ":" + port

	serverOptions := []grpc.ServerOption{
		grpc.MaxMsgSize(grpcMsgByteMax * 2),
		grpc.MaxSendMsgSize(grpcMsgByteMax * 2),
		grpc.MaxRecvMsgSize(grpcMsgByteMax * 2),
		// testing
		grpc.ConnectionTimeout(61 * time.Second),
	}

	if it.cfg.TLSCert != nil {

		cert, err := tls.X509KeyPair(
			[]byte(it.cfg.TLSCert.CertData),
			[]byte(it.cfg.TLSCert.KeyData))
		if err != nil {
			return err
		}

		certs := credentials.NewServerTLSFromCert(&cert)

		serverOptions = append(serverOptions, grpc.Creds(certs))
	}

	it.GrpcServer = grpc.NewServer(serverOptions...)

	return nil
}
