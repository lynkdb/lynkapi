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
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"sync"

	"github.com/hooto/hlog4g/hlog"
	"github.com/hooto/httpsrv"
	"google.golang.org/protobuf/types/known/structpb"
)

type LynkService struct {
	*httpsrv.Controller
	UnimplementedLynkServiceServer

	mu sync.RWMutex

	services    []*ServiceInstance
	mapServices map[string]*serviceInstance

	mapServiceMethods map[string]*serviceMethod

	dataProject *dataProjectManager
}

type serviceInstance struct {
	instance        *ServiceInstance
	refServiceValue reflect.Value
	mapMethods      map[string]*ServiceMethod
	mapTypes        map[string]reflect.Type
	preMethod       *reflect.Method
}

const (
	kServiceMethodType_Std  = 0
	kServiceMethodType_Grpc = 1
)

type serviceMethod struct {
	method          *ServiceMethod
	reqType         reflect.Type
	rspType         reflect.Type
	refServiceValue reflect.Value
	funType         int
	refMethod       reflect.Method
	refPreMethod    *reflect.Method
}

func NewService() *LynkService {
	return &LynkService{
		mapServices:       map[string]*serviceInstance{},
		mapServiceMethods: map[string]*serviceMethod{},
		dataProject:       newDataProjectManager(),
	}
}

func RefTypeKind(rt reflect.Type) string {
	pkgPath := rt.PkgPath()
	if pkgPath != "" {
		pkgPath += "."
	}
	return fmt.Sprintf("%s%s", pkgPath, rt.Name())
}

func (it *LynkService) RegisterService(st interface{}) error {

	if st == nil {
		return errors.New("invalid object")
	}

	if _, ok := st.(*LynkService); ok {
		// return fmt.Errorf("invalid object ref")
	}

	rt := reflect.TypeOf(st)
	if rt.Kind() != reflect.Pointer {
		return fmt.Errorf("invalid object type")
	}

	it.mu.Lock()
	defer it.mu.Unlock()

	srvName := rt.Elem().Name()

	srv, ok := it.mapServices[srvName]
	if !ok {
		srv = &serviceInstance{
			instance: &ServiceInstance{
				// Kind: RefTypeKind(rt.Elem()),
				Name: srvName,
			},
			mapMethods:      map[string]*ServiceMethod{},
			mapTypes:        map[string]reflect.Type{},
			refServiceValue: reflect.ValueOf(st),
		}
		it.mapServices[srvName] = srv
		it.services = append(it.services, srv.instance)

		hlog.Printf("info", "lynkapi service: init service instance %s", srv.instance.Name)
	}

	// common methods
	for i := 0; i < rt.NumMethod(); i++ {

		method := rt.Method(i)

		if method.Name == "PreMethod" && method.Type.NumIn() == 2 && method.Type.NumOut() == 1 {

			var (
				reqCtx = method.Type.In(1)
				rspErr = method.Type.Out(0)
			)

			if reqCtx.PkgPath() == "context" && reqCtx.Name() == "Context" &&
				rspErr.PkgPath() == "" && rspErr.Name() == "error" {
				srv.preMethod = &method
				break
			}
		}
	}

	// methods
	for i := 0; i < rt.NumMethod(); i++ {

		method := rt.Method(i)

		if !method.IsExported() || method.Name == "" ||
			method.Name[0] < 'A' || method.Name[0] > 'Z' {
			continue
		}

		if method.Type.NumIn() != 3 || method.Type.NumOut() != 2 {
			continue
		}

		//
		var funType = 0
		if reqCtx := method.Type.In(1); reqCtx.PkgPath() == "context" && reqCtx.Name() == "Context" {
			funType = kServiceMethodType_Grpc
		} else if reqCtx.PkgPath() == "github.com/lynkdb/lynkapi/go/lynkapi" && reqCtx.Name() == "Context" {
			funType = kServiceMethodType_Std
		} else {
			continue
		}

		//
		reqPtr := method.Type.In(2)
		if reqPtr.Kind() != reflect.Pointer {
			continue
		}
		reqTyp := reqPtr.Elem()
		if reqTyp.Kind() != reflect.Struct {
			continue
		}
		reqSpec, reqType, err := parseSpecByStructType(reqTyp)
		if err != nil {
			continue
		}

		if rspErr := method.Type.Out(1); rspErr.PkgPath() != "" || rspErr.Name() != "error" {
			continue
		}

		out0p := method.Type.Out(0)
		if out0p.Kind() != reflect.Pointer {
			continue
		}
		rspData := out0p.Elem()
		if rspData.Kind() != reflect.Struct {
			continue
		}

		rspSpec, rspType, err := parseSpecByStructType(rspData)
		if err != nil {
			continue
		}

		if _, ok := srv.mapTypes[reqSpec.Kind]; !ok {
			srv.mapTypes[reqSpec.Kind] = reqType
		}

		if _, ok := srv.mapTypes[rspSpec.Kind]; !ok {
			srv.mapTypes[rspSpec.Kind] = rspType
		}

		srvMethod, ok := srv.mapMethods[method.Name]
		if ok {
			continue
		}

		srvMethod = &ServiceMethod{
			Name:         method.Name,
			RequestSpec:  reqSpec,
			ResponseSpec: rspSpec,
		}
		srv.mapMethods[method.Name] = srvMethod
		srv.instance.Methods = append(srv.instance.Methods, srvMethod)

		hlog.Printf("info", "lynkapi init service instance %s, method %s",
			srv.instance.Name, method.Name)

		it.mapServiceMethods[srvName+"."+method.Name] = &serviceMethod{
			method:          srvMethod,
			reqType:         reqType,
			rspType:         rspType,
			funType:         funType,
			refMethod:       method,
			refServiceValue: srv.refServiceValue,
			refPreMethod:    srv.preMethod,
		}
	}

	return nil
}

func (it *LynkService) lookup(req *Request) *serviceMethod {
	it.mu.Lock()
	defer it.mu.Unlock()

	hit, ok := it.mapServiceMethods[req.ServiceName+"."+req.MethodName]
	if !ok {
		return nil
	}

	return hit
}

func (it *LynkService) RegisterDataService(ds DataService) error {
	return it.dataProject.RegisterService(ds)
}

func (it *LynkService) ApiList(
	ctx context.Context,
	req *ApiListRequest,
) (*ApiListResponse, error) {
	return &ApiListResponse{
		Services: it.services,
	}, nil
}

func (it *LynkService) ApiMethod(
	ctx context.Context,
	req *Request,
) (*ServiceMethod, error) {
	it.mu.Lock()
	defer it.mu.Unlock()
	serv, ok := it.mapServices[req.ServiceName]
	if !ok {
		return nil, fmt.Errorf("service (%s) not found", req.ServiceName)
	}
	method, ok := serv.mapMethods[req.MethodName]
	if !ok {
		return nil, fmt.Errorf("method (%s) not found", req.MethodName)
	}
	return method, nil
}

func (it *LynkService) Exec(
	ctx context.Context,
	req *Request,
) (*Response, error) {

	method := it.lookup(req)
	if method == nil {
		return NewResponseError(StatusCode_NotFound, "service/method not found"), nil
	}

	js, err := json.Marshal(req.Data)
	if err != nil {
		return NewResponseError(StatusCode_BadRequest, err.Error()), nil
	}

	reqData := reflect.New(method.reqType).Interface()
	if err := json.Unmarshal(js, reqData); err != nil {
		return NewResponseError(StatusCode_BadRequest, err.Error()), nil
	}

	if ctx == nil {
		ctx = context.TODO()
	}
	ctx = context.WithValue(ctx, RequestSpecNameInContext, method.method.RequestSpec)

	if method.refPreMethod != nil {
		prs := method.refPreMethod.Func.Call([]reflect.Value{
			method.refServiceValue,
			reflect.ValueOf(ctx),
		})
		if err := prs[0].Interface(); err != nil {
			return NewResponseError(StatusCode_BadRequest, err.(error).Error()), nil
		}
	}

	var rss []reflect.Value

	switch method.funType {
	case kServiceMethodType_Std:
		rss = method.refMethod.Func.Call([]reflect.Value{
			method.refServiceValue,
			reflect.ValueOf(&xContext{
				Context: ctx,
				spec:    method.method.RequestSpec,
			}),
			reflect.ValueOf(reqData),
		})

	case kServiceMethodType_Grpc:
		rss = method.refMethod.Func.Call([]reflect.Value{
			method.refServiceValue,
			reflect.ValueOf(ctx),
			reflect.ValueOf(reqData),
		})

	default:
		return NewResponseError(StatusCode_InternalServerError, "unspec method type"), nil
	}

	if err := rss[1].Interface(); err != nil {
		return nil, err.(error)
	}

	js, err = json.Marshal(rss[0].Interface())
	if err != nil {
		return nil, err
	}
	var data structpb.Struct
	if err = json.Unmarshal(js, &data); err != nil {
		return nil, err
	}

	return &Response{
		Kind: method.method.ResponseSpec.Kind,
		Data: &data,
	}, nil
}

func (it *LynkService) DataProject(
	ctx context.Context,
	req *DataProjectRequest,
) (*DataProjectResponse, error) {
	return &DataProjectResponse{
		Status:    NewServiceStatusOK(),
		Instances: it.dataProject.project.Instances,
	}, nil
}

func (it *LynkService) DataQuery(
	ctx context.Context,
	req *DataQuery,
) (*DataResult, error) {
	hlog.Printf("info", "data-query %s", jsonEncode(req))
	ds := it.dataProject.service(req.InstanceName)
	if ds == nil {
		return nil, NewNotFoundError("instance not found")
	}
	return ds.Query(req)
}

func (it *LynkService) DataUpsert(
	ctx context.Context,
	req *DataInsert,
) (*DataResult, error) {
	ds := it.dataProject.service(req.InstanceName)
	if ds == nil {
		return nil, NewNotFoundError("instance not found")
	}
	return ds.Upsert(req)
}

func (it *LynkService) DataIgsert(
	ctx context.Context,
	req *DataInsert,
) (*DataResult, error) {
	ds := it.dataProject.service(req.InstanceName)
	if ds == nil {
		return nil, NewNotFoundError("instance not found")
	}
	return ds.Igsert(req)
}

func (it *LynkService) DataDelete(
	ctx context.Context,
	req *DataDelete,
) (*DataResult, error) {
	ds := it.dataProject.service(req.InstanceName)
	if ds == nil {
		return nil, NewNotFoundError("instance not found")
	}
	return ds.Delete(req)
}

func (c LynkService) ApiListAction() {
	c.RenderJson(&ApiListResponse{
		Services: c.services,
	})
}

func (it *Response) OK() bool {
	if it.Status != nil {
		return it.Status.OK()
	}
	return false
}

func (it *Response) Err() error {
	if it.Status != nil {
		return it.Status.Err()
	}
	return NewError(StatusCode_InternalServerError, "unknown error")
}

func (it *Response) Decode(obj interface{}) error {
	js, err := json.Marshal(it.Data)
	if err != nil {
		return err
	}
	if err = json.Unmarshal(js, obj); err != nil {
		return err
	}
	return nil
}
