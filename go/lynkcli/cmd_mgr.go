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

package lynkcli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/chzyer/readline"
	"github.com/olekukonko/tablewriter"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/lynkdb/lynkapi/go/lynkapi"
)

// Management Commands
type mgrService struct {
	lowerName  string
	name       string
	methods    []*lynkapi.ServiceMethod
	idxMethods map[string]*lynkapi.ServiceMethod
}

var (
	arrMgrServices []*mgrService
	idxMgrServices = map[string]*mgrService{}
	idxMgrRenders  = map[string]methodRender{}
)

type methodRender func(data *structpb.Struct) (string, error)

func mgrSetup() error {
	if len(arrMgrServices) > 0 {
		return nil
	}

	rs := client.ApiList(&lynkapi.ApiListRequest{})
	if err := rs.Status.Err(); err != nil {
		return err
	}
	if len(rs.Services) == 0 {
		return nil
		return fmt.Errorf("service not ready 0")
	}
	for _, srv := range rs.Services {

		if len(srv.Methods) == 0 {
			return fmt.Errorf("service (%s), no method found", srv.Name)
		}

		ms := &mgrService{
			name:       srv.Name,
			lowerName:  lowerName(srv.Name),
			methods:    srv.Methods,
			idxMethods: map[string]*lynkapi.ServiceMethod{},
		}

		if strings.HasSuffix(ms.lowerName, "-service") {
			ms.lowerName = ms.lowerName[:len(ms.lowerName)-len("-service")]
		}

		for _, m := range srv.Methods {
			ms.idxMethods[lowerName(m.Name)] = m
		}

		if _, ok := idxMgrServices[ms.lowerName]; !ok {
			idxMgrServices[ms.lowerName] = ms
			arrMgrServices = append(arrMgrServices, ms)

			RegisterCommonCommand(ms)
		}
	}

	return nil
}

func RegisterRender(serviceName, methodName string, fn func(data *structpb.Struct) (string, error)) {
	idxMgrRenders[fmt.Sprintf("%s.%s", serviceName, methodName)] = methodRender(fn)
}

func (it *mgrService) help(fg FlagSet) string {

	var (
		tbuf  bytes.Buffer
		table = tablewriter.NewWriter(&tbuf)
	)

	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetAutoWrapText(false)

	if fg.Has("show-request-spec") || fg.Has("show-response-spec") {

		headers := []string{"Command"}
		if fg.Has("show-request-spec") {
			headers = append(headers, "Request Spec")
		}
		if fg.Has("show-response-spec") {
			headers = append(headers, "Response Spec")
		}

		table.SetRowLine(true)
		table.SetHeader(headers)

	} else {

		table.SetRowLine(false)
		table.SetColumnSeparator("")
		table.SetHeaderLine(false)
		table.SetBorder(false)
		table.EnableBorder(false)
		table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	}

	for _, method := range it.methods {

		vals := []string{it.lowerName + " " + lowerName(method.Name)}

		if fg.Has("show-request-spec") {
			req, _ := json.MarshalIndent(method.RequestSpec, "", "  ")
			vals = append(vals, string(req))
		}

		if fg.Has("show-response-spec") {
			rsp, _ := json.MarshalIndent(method.ResponseSpec, "", "  ")
			vals = append(vals, string(rsp))
		}

		table.Append(vals)
	}

	table.Render()

	return fmt.Sprintf("\n%s Commands:\n\n%s\n", it.name, tbuf.String())
}

func (it *mgrService) Spec() BaseCommandSpec {
	return BaseCommandSpec{
		Mgr:  true,
		Path: it.lowerName,
	}
}

type cliContext struct {
	dataCreate bool
	dataUpdate bool
}

func (it *mgrService) Action(fg FlagSet, l *readline.Instance) (string, error) {

	if len(fg.VarArgs) == 0 {
		return it.help(fg), nil
	}

	m := it.idxMethods[fg.VarArgs[0]]
	if m == nil {
		return it.help(fg), nil
	}

	ctx := &cliContext{}

	if strings.HasSuffix(fg.VarArgs[0], "-create") {
		ctx.dataCreate = true
	}

	if strings.HasSuffix(fg.VarArgs[0], "-update") {
		ctx.dataUpdate = true
	}

	reqData := &structpb.Struct{
		Fields: map[string]*structpb.Value{},
	}
	if err := scanInput(ctx, 1, reqData, &lynkapi.FieldSpec{
		Fields: m.RequestSpec.Fields,
	}, l); err != nil {
		return "", err
	}
	req := &lynkapi.Request{
		ServiceName: it.name,
		MethodName:  m.Name,
		Data:        reqData,
	}
	rs := client.Exec(req)
	if err := rs.Status.Err(); err != nil {
		return "", err
	}

	if render, ok := idxMgrRenders[it.lowerName+"."+fg.VarArgs[0]]; ok {
		return render(rs.Data)
	}

	str, err := iterOutput(rs.Data, m.ResponseSpec)
	if err != nil {
		return "", err
	}

	if str == "" {
		str = "DONE"
	}

	return str, nil
}

func iterOutput(data *structpb.Struct, spec *lynkapi.TypeSpec) (string, error) {

	if data == nil {
		return "", nil
	}

	specField, specData := spec.Rows(data)

	if specData != nil && len(specField.Fields) > 0 && len(specData.Values) > 0 {

		var (
			mapFields  = map[string]int{}
			tbuf       bytes.Buffer
			table      = tablewriter.NewWriter(&tbuf)
			fieldNames []string
		)

		for _, f := range specField.Fields {
			mapFields[f.TagName] = len(mapFields)
			fieldNames = append(fieldNames, f.TagName)
		}

		if len(fieldNames) != len(mapFields) {
			return "", fmt.Errorf("invalid spec defines")
		}

		table.SetHeader(fieldNames)
		table.SetRowLine(true)
		table.SetAutoWrapText(false)
		table.SetAlignment(tablewriter.ALIGN_LEFT)

		for _, row := range specData.Values {

			structValue := row.GetStructValue()
			// fmt.Println(structValue)
			if structValue == nil || len(structValue.Fields) == 0 {
				continue
			}

			fieldValues := make([]string, len(fieldNames))
			for name, value := range structValue.Fields {
				if value == nil {
					continue
				}
				idx, ok := mapFields[name]
				if !ok {
					continue
				}
				switch value.Kind.(type) {
				case *structpb.Value_StringValue:
					fieldValues[idx] = value.GetStringValue()

				case *structpb.Value_NumberValue:
					if value.GetNumberValue() != 0 {
						if fd := specField.Field(name); fd != nil {
							switch fd.Type {
							case lynkapi.FieldSpec_Int:
								fieldValues[idx] = fmt.Sprintf("%d", int64(value.GetNumberValue()))

							case lynkapi.FieldSpec_Uint:
								fieldValues[idx] = fmt.Sprintf("%d", uint64(value.GetNumberValue()))

							case lynkapi.FieldSpec_Float:
								fieldValues[idx] = fmt.Sprintf("%f", value.GetNumberValue())
							}
						}
					}

				case *structpb.Value_BoolValue:
					fieldValues[idx] = fmt.Sprintf("%b", value.GetBoolValue())

				default:
					js, _ := json.MarshalIndent(value, "", "  ")
					fieldValues[idx] = string(js)
				}
			}

			//
			table.Append(fieldValues)
		}

		table.Render()

		return tbuf.String(), nil
	}

	if data != nil && len(data.Fields) > 0 {

		var (
			tbuf  bytes.Buffer
			table = tablewriter.NewWriter(&tbuf)
		)

		table.SetHeader([]string{"Field", "Value"})

		table.SetRowLine(true)
		table.SetAutoWrapText(false)
		table.SetAlignment(tablewriter.ALIGN_LEFT)

		for name, value := range data.Fields {

			if value == nil {
				continue
			}

			switch value.Kind.(type) {
			case *structpb.Value_StringValue:
				table.Append([]string{name, value.GetStringValue()})

			case *structpb.Value_NumberValue:
				if value.GetNumberValue() != 0 {
					if specField := spec.Field(name); specField != nil {
						switch specField.Type {
						case lynkapi.FieldSpec_Int:
							table.Append([]string{name, fmt.Sprintf("%d", int64(value.GetNumberValue()))})

						case lynkapi.FieldSpec_Uint:
							table.Append([]string{name, fmt.Sprintf("%d", uint64(value.GetNumberValue()))})
						case lynkapi.FieldSpec_Float:
							table.Append([]string{name, fmt.Sprintf("%f", value.GetNumberValue())})
						}
					}
				}

			case *structpb.Value_BoolValue:
				table.Append([]string{name, fmt.Sprintf("%b", value.GetBoolValue())})

			default:
				js, _ := json.MarshalIndent(value, "", "  ")
				table.Append([]string{name, string(js)})
			}
		}

		table.Render()

		return tbuf.String(), nil
	}

	return "", fmt.Errorf("no response")
}

func scanInput(ctx *cliContext, depth int, data *structpb.Struct, specField *lynkapi.FieldSpec, l *readline.Instance) error {

	for _, field := range specField.Fields {

		prompt := fmt.Sprintf("%s%s", strings.Repeat(" ", depth*2), field.Name)

		if field.Type == lynkapi.FieldSpec_Struct {

			prompt += fmt.Sprintf(" (type `yes` to edit this sub-object, `no` to skip)")

			yes := false

			if (ctx.dataCreate && slices.Contains(field.Attrs, "create_required")) ||
				(ctx.dataUpdate && slices.Contains(field.Attrs, "update_required")) {
				prompt += fmt.Sprintf(" (required) [yes]")
				yes = true
			} else {
				prompt += fmt.Sprintf(" (optional) [no]")
			}

			l.SetPrompt(prompt + " : ")

			v, err := l.Readline()
			if err != nil {
				return err
			}
			v = strings.ToLower(strings.TrimSpace(v))
			if v == "yes" || yes {
				structValue := &structpb.Struct{
					Fields: map[string]*structpb.Value{},
				}
				data.Fields[field.TagName] = structpb.NewStructValue(structValue)
				if err := scanInput(ctx, depth+1, structValue, field, l); err != nil {
					return err
				}
			}
			continue
		}

		if len(field.Enums) > 0 {
			prompt += fmt.Sprintf(" [enums: %s]", strings.Join(field.Enums, ","))
		}

		if slices.Contains(field.Attrs, "primary_key") ||
			(ctx.dataCreate && slices.Contains(field.Attrs, "create_required")) ||
			(ctx.dataUpdate && slices.Contains(field.Attrs, "update_required")) {
			prompt += fmt.Sprintf(" (required)")
		} else {
			prompt += fmt.Sprintf(" (optional)")
		}

		l.SetPrompt(prompt + " : ")

		v, err := l.Readline()
		if err != nil {
			return err
		}
		v = strings.TrimSpace(v)

		if v == "" {
			continue
		}

		switch field.Type {
		case lynkapi.FieldSpec_String:
			data.Fields[field.TagName] = structpb.NewStringValue(v)

		case lynkapi.FieldSpec_Int:
			if num, err := strconv.ParseInt(v, 10, 64); err != nil {
				return err
			} else {
				data.Fields[field.TagName] = structpb.NewNumberValue(float64(num))
			}

		case lynkapi.FieldSpec_Uint:
			if num, err := strconv.ParseUint(v, 10, 64); err != nil {
				return err
			} else {
				data.Fields[field.TagName] = structpb.NewNumberValue(float64(num))
			}

		case lynkapi.FieldSpec_Float:
			if num, err := strconv.ParseFloat(v, 64); err != nil {
				return err
			} else {
				data.Fields[field.TagName] = structpb.NewNumberValue(float64(num))
			}

		case lynkapi.FieldSpec_Bool:
			if b, err := strconv.ParseBool(v); err != nil {
				return err
			} else {
				data.Fields[field.TagName] = structpb.NewBoolValue(b)
			}

		case lynkapi.FieldSpec_Struct:
			structValue := &structpb.Struct{
				Fields: map[string]*structpb.Value{},
			}
			data.Fields[field.TagName] = structpb.NewStructValue(structValue)
			if err := scanInput(ctx, depth+1, structValue, field, l); err != nil {
				return err
			}

		default:
		}
	}

	return nil
}
