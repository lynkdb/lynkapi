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

func Test_Field_UniqueKey(t *testing.T) {
	type Container struct {
		Table []*lynkapi.DataDict `json:"table" x_attrs:"rows"`
	}

	ctn := &Container{
		Table: []*lynkapi.DataDict{
			{
				Name: "unikey-1",
			},
		},
	}

	inst, err := oneobject.NewInstance("test", ctn)
	if err != nil {
		t.Fatal(err)
	}

	if err := inst.TableSetup("table"); err != nil {
		t.Fatal(err)
	}

	insert := &lynkapi.DataInsert{
		InstanceName: "test",
		TableName:    "table",
	}
	insert.SetField("name", "unikey-1")

	if _, err := inst.Insert(insert); err != nil {
		if ss := lynkapi.ParseError(err); ss.Code != lynkapi.StatusCode_Conflict {
			t.Fatal(err)
		}
	} else {
		t.Fatal("unique-key")
	}
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

	inst, err := oneobject.NewInstance("test", cfg)
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
		upsert := &lynkapi.DataInsert{
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
		upsert := &lynkapi.DataInsert{
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
