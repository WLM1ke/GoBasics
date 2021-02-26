package main

import (
	"encoding/json"
	"encoding/xml"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"sort"
	"strconv"
	"strings"
	"testing"
)

type XMLUser struct {
	Id        int    `xml:"id"`
	FirstName string `xml:"first_name"`
	LastName  string `xml:"last_name"`
	Age       int    `xml:"age"`
	About     string `xml:"about"`
	Gender    string `xml:"gender"`
}

func GetData() []User {
	file, _ := ioutil.ReadFile("dataset.xml")

	var rows struct {
		Rows []XMLUser `xml:"row"`
	}
	xml.Unmarshal(file, &rows)

	users := make([]User, len(rows.Rows))

	for n, row := range rows.Rows {
		users[n].Id = row.Id
		users[n].Name = row.FirstName + row.LastName
		users[n].Age = row.Age
		users[n].About = row.About
		users[n].Gender = row.Gender
	}
	return users
}

var users []User = GetData()

type SearchServer struct {
	AccessToken string
	Status      int
}

func (svr *SearchServer) SearchServer(w http.ResponseWriter, r *http.Request) {

	if svr.AccessToken != r.Header.Get("AccessToken") {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	if svr.Status != http.StatusOK {
		w.WriteHeader(svr.Status)
		return
	}

	resUsers := make([]User, 0)

	query := r.URL.Query().Get("query")

	for _, user := range users {
		if strings.Contains(user.Name, query) || strings.Contains(user.About, query) {
			resUsers = append(resUsers, user)
		}
	}

	orderField := r.URL.Query().Get("order_field")
	if orderField == "" {
		orderField = "Name"
	}
	if !(orderField == "Id" || orderField == "Age" || orderField == "Name") {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	orderBy, _ := strconv.Atoi(r.URL.Query().Get("order_by"))

	if orderBy == OrderByAsc {
		switch orderField {
		case "Id":
			sort.Slice(resUsers, func(i, j int) bool { return resUsers[i].Id < resUsers[j].Id })
		case "Age":
			sort.Slice(resUsers, func(i, j int) bool { return resUsers[i].Age < resUsers[j].Age })
		case "Name":
			sort.Slice(resUsers, func(i, j int) bool { return resUsers[i].Name < resUsers[j].Name })

		}
	}

	if orderBy == OrderByDesc {
		switch orderField {
		case "Id":
			sort.Slice(resUsers, func(i, j int) bool { return resUsers[i].Id > resUsers[j].Id })
		case "Age":
			sort.Slice(resUsers, func(i, j int) bool { return resUsers[i].Age > resUsers[j].Age })
		case "Name":
			sort.Slice(resUsers, func(i, j int) bool { return resUsers[i].Name > resUsers[j].Name })

		}
	}

	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	resUsers = resUsers[offset:]

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit > len(resUsers) {
		limit = len(resUsers)
	}
	resUsers = resUsers[:limit]

	body, _ := json.Marshal(resUsers)

	w.Write(body)
}

func TestBadReq(t *testing.T) {
	badReqs := []SearchRequest{
		SearchRequest{Limit: -1},
		SearchRequest{Offset: -1},
	}

	client := SearchClient{}
	for _, req := range badReqs {
		rez, err := client.FindUsers(req)

		if rez != nil || err == nil {
			t.Errorf("Нет ошибки на некоректный запрос.")
		}
	}
}

func TestLongReq(t *testing.T) {
	server := SearchServer{"", http.StatusOK}
	ts := httptest.NewServer(http.HandlerFunc(server.SearchServer))

	client := SearchClient{"", ts.URL}

	rez, err := client.FindUsers(SearchRequest{Limit: 50})
	if !(len(rez.Users) == 25 && err == nil) {
		t.Errorf("Нет ошибки на некоректный токен.")
	}
}

func TestBadToken(t *testing.T) {
	server := SearchServer{"Some_token", http.StatusOK}
	ts := httptest.NewServer(http.HandlerFunc(server.SearchServer))

	client := SearchClient{"Wrong_token", ts.URL}

	rez, err := client.FindUsers(SearchRequest{})
	if rez != nil || err == nil {
		t.Errorf("Нет ошибки на некоректный токен.")
	}

}

func TestInternalServerError(t *testing.T) {
	server := SearchServer{"", http.StatusInternalServerError}
	ts := httptest.NewServer(http.HandlerFunc(server.SearchServer))

	client := SearchClient{"", ts.URL}

	rez, err := client.FindUsers(SearchRequest{})
	if rez != nil || err == nil {
		t.Errorf("Нет ошибки на внутреннюю ошибку сервера.")
	}

}
