package oneobject_test

import (
	"encoding/json"
	"testing"

	"github.com/lynkdb/lynkapi/go/lynkapi"
	"github.com/lynkdb/lynkapi/go/oneobject"
)

type ConfigItem struct {
	Name  string `json:"name" x_attrs:"primary_key"`
	Value string `json:"value"`
}

type ConfigObject struct {
	Name    string        `json:"name"`
	Options []*ConfigItem `json:"options"`
}

func Test_Instance(t *testing.T) {

	cfg := &ConfigObject{
		Name: "test",
		Options: []*ConfigItem{
			{
				Name:  "name-1",
				Value: "value-1",
			},
			{
				Name:  "name-2",
				Value: "value-2",
			},
		},
	}

	inst, err := oneobject.NewInstance(cfg)
	if err != nil {
		t.Fatal(err)
	}

	if err := inst.TableSetup("options"); err != nil {
		t.Fatal(err)
	}

	{ // query
		q := &lynkapi.DataQuery{
			TableName: "options",
			Filter:    &lynkapi.DataQuery_Filter{},
			Limit:     10,
		}
		q.Filter.And("name", "name-2")
		rs, err := inst.Query(q)
		if err != nil {
			t.Fatal(err)
		}
		if len(rs.Rows) != 1 {
			t.Fatalf("invalid hits %d", len(rs.Rows))
		}
		for i, row := range rs.Rows {
			js, _ := json.Marshal(row.Fields)
			t.Logf("  #%d  %v", i, string(js))
		}
	}

	{ // insert
		upsert := &lynkapi.DataUpsert{
			TableName: "options",
		}
		upsert.SetField("name", "name-3")
		upsert.SetField("value", "value-3")

		_, err := inst.Upsert(upsert)
		if err != nil {
			t.Fatal(err)
		}

		q := &lynkapi.DataQuery{
			TableName: "options",
			Filter:    &lynkapi.DataQuery_Filter{},
			Limit:     10,
		}
		q.Filter.And("name", "name-3")
		rs, err := inst.Query(q)
		if err != nil {
			t.Fatal(err)
		}
		if len(rs.Rows) != 1 {
			t.Fatalf("invalid hits %d", len(rs.Rows))
		}
		for i, row := range rs.Rows {
			js, _ := json.Marshal(row.Fields)
			t.Logf("  #%d  %v", i, string(js))
		}
	}

	{ // update
		upsert := &lynkapi.DataUpsert{
			TableName: "options",
		}
		upsert.SetField("name", "name-3")
		upsert.SetField("value", "value-3-3")

		_, err := inst.Upsert(upsert)
		if err != nil {
			t.Fatal(err)
		}

		q := &lynkapi.DataQuery{
			TableName: "options",
			Filter:    &lynkapi.DataQuery_Filter{},
			Limit:     10,
		}
		q.Filter.And("name", "name-3")
		rs, err := inst.Query(q)
		if err != nil {
			t.Fatal(err)
		}
		if len(rs.Rows) != 1 {
			t.Fatalf("invalid hits %d", len(rs.Rows))
		}
		for i, row := range rs.Rows {
			js, _ := json.Marshal(row.Fields)
			t.Logf("  #%d  %v", i, string(js))
		}
	}

	{ // delete
		del := &lynkapi.DataDelete{
			TableName: "options",
			Filter:    &lynkapi.DataQuery_Filter{},
		}
		del.Filter.And("name", "name-2")

		_, err := inst.Delete(del)
		if err != nil {
			t.Fatal(err)
		}

		q := &lynkapi.DataQuery{
			TableName: "options",
			Filter:    &lynkapi.DataQuery_Filter{},
			Limit:     10,
		}
		rs, err := inst.Query(q)
		if err != nil {
			t.Fatal(err)
		}
		if len(rs.Rows) != 2 {
			t.Fatalf("invalid hits %d", len(rs.Rows))
		}
		for i, row := range rs.Rows {
			js, _ := json.Marshal(row.Fields)
			t.Logf("  #%d  %v", i, string(js))
		}
	}
}
