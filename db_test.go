package db_test

import (
	"time"
	"testing"
	"github.com/fobilow/gitdb"
	"encoding/json"
	"os"
	"path/filepath"
	"io/ioutil"
	"strings"
	"errors"
	"fmt"
)

var testDb *db.Gitdb

func setup(){
	testDb = db.Start(getConfig())
}

func getConfig() *db.Config {
	return &db.Config{
		DbPath:         "/tmp/data",
		OnlineRemote:   "",
		SyncInterval:   time.Second * 120,
		EncryptionKey:  "",
		GitDriver: &db.GitBinary{},
		Factory: dbFactory,
		User: db.NewUser("Tester", "tester@gitdb.io"),
		Verbose: db.LOGLEVEL_TEST,
	}
}

func getTestMessage(rand bool) *Message {

	m := &Message{
		From: "alice@example.com",
		To: "bob@example.com",
		Body: "Hello",
	}

	if !rand {
		date := time.Date(2019,2,1,1,1,1,1, time.UTC)
		m.SetCreatedDate(date)
		m.SetUpdatedDate(date)
	}

	return m
}

func dbFactory(name string) db.Model {
	switch name {
	case "Message":
	default:
		return &Message{}
	}

	return &Message{}
}

type Message struct {
	db.BaseModel
	From string
	To string
	Body string
}

func (m *Message) GetSchema() *db.Schema {
	//Name of schema
	name := func() string {
		return "Message"
	}

	//Block of schema
	//block := func() string {
	//	return m.CreatedAt.Format("200601")
	//}

	block := db.NewAutoBlock(m, 2e8, 100)

	//Record of schema
	record := func() string {
		return m.CreatedAt.Format("20060102150405.999999999")
	}

	//Indexes speed up searching
	indexes := func() map[string]interface{} {
		indexes := make(map[string]interface{})

		indexes["From"] = m.From
		return indexes
	}

	return db.NewID(name, block, record, indexes)
}

func doInsert(benchmark bool, rand bool) error {

	m := getTestMessage(rand)
	if err := testDb.Insert(m); err != nil {
		return err
	}

	if !benchmark {
		//check that block file exist
		idParser := db.NewIDParser(m.Id())
		cfg := getConfig()
		blockFile := filepath.Join(cfg.DbPath, "data", idParser.BlockId()+"."+string(m.GetDataFormat()))
		if _, err := os.Stat(blockFile); err != nil {
			return err
		} else {
			b, err := ioutil.ReadFile(blockFile)
			if err != nil {
				return err
			}

			rep := strings.NewReplacer("\n", "", "\\", "", "\t", "", "\"{", "{", "}\"", "}", " ", "")
			got := rep.Replace(string(b))

			w := map[string]*Message{
				idParser.RecordId(): m,
			}

			x, _ := json.Marshal(w)
			want := string(x)


			want = want[1:len(want)-1]

			if !strings.Contains(got, want) {
				return errors.New(fmt.Sprintf("Want: %s, Got: %s", want, got))
			}
		}
	}

	return nil
}

func TestInsert(t *testing.T){
	getConfig().OnlineRemote  = "Blaa"
	setup()
	err := doInsert(false, false)
	if err != nil {
		t.Errorf(err.Error())
	}
}

func BenchmarkInsert(b *testing.B) {
	b.ReportAllocs()
	setup()

	for i :=0; i <= b.N; i++ {
		time.Sleep(300*time.Microsecond)
		doInsert(true, true)

	}
}

func TestGet(t *testing.T) {
	setup()
	m := getTestMessage(false)

	result := &Message{}
	err := testDb.Get(m.GetSchema().RecordId(), result)
	if err != nil {
		t.Error(err.Error())
	}

	if result.String() != m.GetSchema().RecordId() {
		t.Errorf("Want: %v, Got: %v", m.GetSchema().RecordId(), result.String())
	}
}

type MessageResult struct {
	Results []*Message
}

func (m *MessageResult) Add(msg db.Model){
	m.Results = append(m.Results, msg.(*Message))
}


func TestFetch(t *testing.T) {
	setup()
	messages, err := testDb.Fetch("Message")
	if err != nil {
		t.Error(err.Error())
	}

	if len(messages) != 1 {
		t.Errorf("Want: %d, Got: %d", 1, len(messages))
	}

	mr := &MessageResult{}
	mr.Add(messages[0])

	println(messages[0].(*Message).From)
}

func TestFetchMultithreaded(t *testing.T) {
	setup()
	messages, err := testDb.Fetch2("Message")
	if err != nil {
		t.Error(err.Error())
	}

	if len(messages) != 1 {
		t.Errorf("Want: %d, Got: %d", 1, len(messages))
	}
}

func BenchmarkFetch(b *testing.B) {
	setup()
	b.ReportAllocs()
	for i :=0; i <= b.N; i++ {
		testDb.Fetch("Message")
	}
}

func BenchmarkFetchMultithreaded(b *testing.B) {
	setup()
	b.ReportAllocs()
	for i :=0; i <= b.N; i++ {
		testDb.Fetch2("Message")
	}
}

func TestDelete(t *testing.T) {
	setup()

	m := getTestMessage(false)
	deleted, err := testDb.Delete(m.GetSchema().RecordId())
	if !deleted || err != nil {
		t.Errorf("Deleted: %v, Error: %s", deleted, err.Error())
	}
}

func TestDeleteOrFail(t *testing.T) {
	setup()
	deleted, err := testDb.DeleteOrFail("non_existent_id")
	if deleted || err == nil {
		t.Errorf("Deleted: %v, Error: %s", deleted, err.Error())
	}
}

func TestGetModel(t *testing.T) {
	m := getModel()
	println(m)
}

func getModel() *Message {

	in := map[string]string{
	"From": "alice@example.com",
	"To": "bob@example.com",
	"Body": "How are you?",
	}

	m := &Message{}
	testDb.GetModel(in, m)
	return m
}

func BenchmarkGetModel(b *testing.B) {
	b.ReportAllocs()
	for i :=0; i <= b.N; i++ {
      getModel()
	}
}

func TestParseId(t *testing.T){
	testId := "DatasetName/Block/RecordId"
	ds, block, recordId, err := testDb.ParseId(testId)

	passed := ds == "DatasetName" && block == "Block" && recordId == "RecordId" && err == nil
	if !passed {
		t.Errorf("want: DatasetName|Block|RecordId, Got:%s|%s|%s", ds,block,recordId)
	}
}

func BenchmarkParseId(b *testing.B) {
	b.ReportAllocs()
	for i :=0; i <= b.N; i++ {
		testDb.ParseId("DatasetName/Block/RecordId")
	}
}

func BenchmarkIDParserParseId(b *testing.B) {
	b.ReportAllocs()
	for i :=0; i <= b.N; i++ {
		db.NewIDParser("DatasetName/Block/RecordId")
	}
}