package main

import (
	"encoding/json"
	"encoding/xml"
	"time"

	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"sort"
	"strconv"
	"strings"
	"testing"
)

type Root struct{
	XMLName xml.Name `xml:"root"`
	Users []Row `xml:"row"`
}

type Row struct{
	XMLName xml.Name `xml:"row"`
	Id int `xml:"id"`
	FName string `xml:"first_name"`
	LName string `xml:"last_name"`
	Age int `xml:"age"`
	About string `xml:"about"`
	Gender string `xml:"gender"`
}


type wrongUser struct{
	WrongUsers []User
	IsWrong bool

}


func SearchServer(w http.ResponseWriter, r *http.Request){
	if r.Header.Get("AccessToken")!="4fer4f6er4686er"&& r.Header.Get("AccessToken")!="forTimeoutToken"&&
		r.Header.Get("AccessToken")!="forTemporaryError"&& r.Header.Get("AccessToken")!="forFatalServerError" && r.Header.Get("AccessToken")!="forJsonError" &&
		r.Header.Get("AccessToken")!="forJsonError2"{
		w.WriteHeader(http.StatusUnauthorized)
	}
	if r.Header.Get("AccessToken")=="forTimeoutToken"{
		time.Sleep(2 * time.Second)
	}
	if r.Header.Get("AccessToken")=="forTemporaryError"{
		http.Error(w,"temporary error", http.StatusFound)
		return
	}


	limit,err:=strconv.Atoi(r.FormValue("limit"))
	if err!=nil{
		w.WriteHeader(http.StatusInternalServerError)
	}
	offset, err:=strconv.Atoi(r.FormValue("offset"))
	if err!=nil{
		w.WriteHeader(http.StatusInternalServerError)
	}
	query:=r.FormValue("query")
	orderField:=r.FormValue("order_field")
	orderBy, err:=strconv.Atoi(r.FormValue("order_by"))
	if err!=nil{
		w.WriteHeader(http.StatusInternalServerError)
	}

	if orderBy!=OrderByAsc && orderBy!=OrderByAsIs && orderBy!=OrderByDesc{
		w.WriteHeader(http.StatusBadRequest)
		if r.Header.Get("AccessToken")=="forJsonError2"{

			w.Write([]byte("suka"))
			return
		}else {
			a := SearchErrorResponse{Error: "ErrorBadOrderBy"}
			b, err := json.Marshal(a)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
			}
			w.Write(b)
			return
		}
	}
	if r.Header.Get("AccessToken")=="forFatalServerError"{
		w.WriteHeader(http.StatusInternalServerError)
		return
	}



	fileData, err:=ioutil.ReadFile("dataset.xml")
	if err!=nil{
		panic(err)
	}
	var v Root
	var result []User
	err=xml.Unmarshal(fileData, &v)
	if err!=nil{
		panic(err)
	}


	for _, user:=range v.Users{
		if query!="" {
			if strings.Contains(user.FName, query) || strings.Contains(user.LName, query) || strings.Contains(user.About, query) {
				result = append(result, User{
					Id: user.Id,
					Name: user.FName+" "+user.LName,
					Age: user.Age,
					About: user.About,
					Gender: user.Gender,
				})
			}
		}
	}
	if orderField!="Name" && orderField!="Id" && orderField!="Age" && orderField!="About" && orderField!="Gender"{
		w.WriteHeader(http.StatusBadRequest)
		a:=SearchErrorResponse{Error: "ErrorBadOrderField"}
		b,err:=json.Marshal(a)
		if err!=nil{
			w.WriteHeader(http.StatusInternalServerError)
		}
		w.Write(b)
		return
	}

	if orderBy==OrderByAsc {
		switch orderField {
		case "Name":
			sort.SliceStable(result, func(i, j int) bool { return result[i].Name < result[j].Name })
		case "Id":
			sort.SliceStable(result, func(i, j int) bool { return result[i].Id < result[j].Id })
		case "Age":
			sort.SliceStable(result, func(i, j int) bool { return result[i].Age < result[j].Age })
		case "About":
			sort.SliceStable(result, func(i, j int) bool { return result[i].About < result[j].About })
		case "Gender":
			sort.SliceStable(result, func(i, j int) bool { return result[i].Gender < result[j].Gender })

		}
	}

	if orderBy==OrderByDesc {
		switch orderField {
		case "Name":
			sort.SliceStable(result, func(i, j int) bool { return result[i].Name > result[j].Name })
		case "Id":
			sort.SliceStable(result, func(i, j int) bool { return result[i].Id > result[j].Id })
		case "Age":
			sort.SliceStable(result, func(i, j int) bool { return result[i].Age > result[j].Age })
		case "About":
			sort.SliceStable(result, func(i, j int) bool { return result[i].About > result[j].About })
		case "Gender":
			sort.SliceStable(result, func(i, j int) bool { return result[i].Gender > result[j].Gender })

		}
	}

	if offset>0 && len(result)>0{
		result = result[offset:]
	}
	if len(result)>limit{
		result=result[:limit]
	}
	if r.Header.Get("AccessToken")=="forJsonError"{
		wrUsers:=wrongUser{
			WrongUsers: result,
			IsWrong: true,
		}
		data, err:=json.Marshal(wrUsers)
		if err!=nil{
			w.WriteHeader(http.StatusInternalServerError)
		}
		w.Write(data)

	}else {
		//result=result[:limit]
		data, err := json.Marshal(result)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}

		w.Write(data)
	}
}


type testCase struct{
	Value interface{}
	Result interface{}
	Error bool
}




//////тсеты
func TestRequestLimit(t *testing.T) {

	casesLimit:=[]testCase{
		testCase{
			Value: -1,
			Result: nil,
			Error: true,
		},
		testCase{
			Value: 26,
			Result: &SearchResponse{
				Users: []User{
					User{
						Id: 1,
						Name: "Hylda Mayer",
						Age: 21,
						About: "Sit commodo consectetur minim amet ex. Elit aute mollit fugiat labore sint ipsum dolor cupidatat qui reprehenderit. Eu nisi in exercitation culpa sint aliqua nulla nulla proident eu. Nisi reprehenderit anim cupidatat dolor incididunt laboris mollit magna commodo ex. Cupidatat sit id aliqua amet nisi et voluptate voluptate commodo ex eiusmod et nulla velit.\n",
						Gender: "female",
					},
					User{
						Id: 6,
						Name: "Jennings Mays",
						Age: 39,
						About: "Veniam consectetur non non aliquip exercitation quis qui. Aliquip duis ut ad commodo consequat ipsum cupidatat id anim voluptate deserunt enim laboris. Sunt nostrud voluptate do est tempor esse anim pariatur. Ea do amet Lorem in mollit ipsum irure Lorem exercitation. Exercitation deserunt adipisicing nulla aute ex amet sint tempor incididunt magna. Quis et consectetur dolor nulla reprehenderit culpa laboris voluptate ut mollit. Qui ipsum nisi ullamco sit exercitation nisi magna fugiat anim consectetur officia.\n",
						Gender: "male",
					},
				},
				NextPage: false,
			},
			Error: false,
		},
		testCase{
			Value: 1,
			Result: &SearchResponse{
				Users: []User{
					User{
						Id: 1,
						Name: "Hylda Mayer",
						Age: 21,
						About: "Sit commodo consectetur minim amet ex. Elit aute mollit fugiat labore sint ipsum dolor cupidatat qui reprehenderit. Eu nisi in exercitation culpa sint aliqua nulla nulla proident eu. Nisi reprehenderit anim cupidatat dolor incididunt laboris mollit magna commodo ex. Cupidatat sit id aliqua amet nisi et voluptate voluptate commodo ex eiusmod et nulla velit.\n",
						Gender: "female",
					},
				},
				NextPage: true,
			},
			Error: false,
		},
	}
	ts:=httptest.NewServer(http.HandlerFunc(SearchServer))
	for caseNum, caseLimit:=range casesLimit{
		req:=SearchRequest{
			Limit: caseLimit.Value.(int),
			Offset: 0,
			Query: "commodo co",
			OrderField: "Name",
			OrderBy: 0,
		}
		client:=SearchClient{AccessToken: "4fer4f6er4686er", URL: ts.URL+"/client=0"}
		result, err:=client.FindUsers(req)
		if err!=nil && !caseLimit.Error{
			t.Errorf("[%d] unexprcted error: %#v", caseNum, err)
		}
		if err==nil && caseLimit.Error{
			t.Errorf("[%d] exprcted error, got nil", caseNum)
		}
		if caseLimit.Result==result{//!reflect.DeepEqual(caseLimit.Result, result){
			t.Errorf("[%d] wrong result, exprcted %#v, got %#v", caseNum, caseLimit.Result, result)
		}
	}
	ts.Close()
}

func TestAccessToken(t *testing.T) {
	casesToken := []testCase{
		testCase{
			Value:  "fvvdvdfvdfv",
			Result: nil,
			Error:  true,
		},
		testCase{
			Value: "4fer4f6er4686er",
			Result: &SearchResponse{
				Users: []User{
					User{
						Id:     1,
						Name:   "Hylda Mayer",
						Age:    21,
						About:  "Sit commodo consectetur minim amet ex. Elit aute mollit fugiat labore sint ipsum dolor cupidatat qui reprehenderit. Eu nisi in exercitation culpa sint aliqua nulla nulla proident eu. Nisi reprehenderit anim cupidatat dolor incididunt laboris mollit magna commodo ex. Cupidatat sit id aliqua amet nisi et voluptate voluptate commodo ex eiusmod et nulla velit.\n",
						Gender: "female",
					},
				},
			},
			Error: false,
		},
	}
	ts:=httptest.NewServer(http.HandlerFunc(SearchServer))
	for caseNum, caseToken:=range casesToken{
		req:=SearchRequest{
			Limit: 1,
			Offset: 0,
			Query: "commodo co",
			OrderField: "Name",
			OrderBy: 0,
		}
		client:=SearchClient{AccessToken: caseToken.Value.(string), URL: ts.URL+"/client=0"}
		result, err:=client.FindUsers(req)
		if err!=nil && !caseToken.Error{
			t.Errorf("[%d] unexprcted error: %#v", caseNum, err)
		}
		if err==nil && caseToken.Error{
			t.Errorf("[%d] exprcted error, got nil", caseNum)
		}
		if caseToken.Result==result{//!reflect.DeepEqual(caseLimit.Result, result){
			t.Errorf("[%d] wrong result, exprcted %#v, got %#v", caseNum, caseToken.Result, result)
		}
	}
	ts.Close()
}

func TestRequestOffset(t *testing.T){
	casesOffset:=[]testCase{
		testCase{
			Value: -1,
			Result: nil,
			Error: true,
		},
		testCase{
			Value: 0,
			Result: &SearchResponse{
				Users: []User{
					User{
						Id: 1,
						Name: "Hylda Mayer",
						Age: 21,
						About: "Sit commodo consectetur minim amet ex. Elit aute mollit fugiat labore sint ipsum dolor cupidatat qui reprehenderit. Eu nisi in exercitation culpa sint aliqua nulla nulla proident eu. Nisi reprehenderit anim cupidatat dolor incididunt laboris mollit magna commodo ex. Cupidatat sit id aliqua amet nisi et voluptate voluptate commodo ex eiusmod et nulla velit.\n",
						Gender: "female",
					},
				},
				NextPage: true,
			},
			Error: false,
		},
	}
	ts:=httptest.NewServer(http.HandlerFunc(SearchServer))
	for caseNum, caseOffset:=range casesOffset{
		req:=SearchRequest{
			Limit: 1,
			Offset: caseOffset.Value.(int),
			Query: "commodo co",
			OrderField: "Name",
			OrderBy: 0,
		}
		client:=SearchClient{AccessToken: "4fer4f6er4686er", URL: ts.URL+"/client=0"}
		result, err:=client.FindUsers(req)
		if err!=nil && !caseOffset.Error{
			t.Errorf("[%d] unexprcted error: %#v", caseNum, err)
		}
		if err==nil && caseOffset.Error{
			t.Errorf("[%d] exprcted error, got nil", caseNum)
		}
		if caseOffset.Result==result{//!reflect.DeepEqual(caseLimit.Result, result){
			t.Errorf("[%d] wrong result, exprcted %#v, got %#v", caseNum, caseOffset.Result, result)
		}
	}
	ts.Close()
}
func TestRequestQuery(t *testing.T){
	casesQuery:=[]testCase{
		testCase{
			Value: "commodo co",
			Result: &SearchResponse{
				Users: []User{
					User{
						Id: 1,
						Name: "Hylda Mayer",
						Age: 21,
						About: "Sit commodo consectetur minim amet ex. Elit aute mollit fugiat labore sint ipsum dolor cupidatat qui reprehenderit. Eu nisi in exercitation culpa sint aliqua nulla nulla proident eu. Nisi reprehenderit anim cupidatat dolor incididunt laboris mollit magna commodo ex. Cupidatat sit id aliqua amet nisi et voluptate voluptate commodo ex eiusmod et nulla velit.\n",
						Gender: "female",
					},
				},
				NextPage: true,
			},
			Error: false,
		},
		testCase{
			Value: "000",
			Result: &SearchResponse{
				Users: []User{},
				NextPage: false},
			Error: false,
		},
	}
	ts:=httptest.NewServer(http.HandlerFunc(SearchServer))
	for caseNum, caseQuery:=range casesQuery{
		req:=SearchRequest{
			Limit: 1,
			Offset: 0,
			Query: caseQuery.Value.(string),
			OrderField: "Name",
			OrderBy: 0,
		}
		client:=SearchClient{AccessToken: "4fer4f6er4686er", URL: ts.URL+"/client=0"}
		result, err:=client.FindUsers(req)
		if err!=nil && !caseQuery.Error{
			t.Errorf("[%d] unexprcted error: %#v", caseNum, err)
		}
		if err==nil && caseQuery.Error{
			t.Errorf("[%d] exprcted error, got nil", caseNum)
		}
		if caseQuery.Result==result{//!reflect.DeepEqual(caseLimit.Result, result){
			t.Errorf("[%d] wrong result, exprcted %#v, got %#v", caseNum, caseQuery.Result, result)
		}
	}
	ts.Close()
}

func TestRequestOrderField(t *testing.T){
	casesOrderField:=[]testCase{
		testCase{
			Value: "Id",
			Result: &SearchResponse{
				Users: []User{
					User{
						Id: 1,
						Name: "Hylda Mayer",
						Age: 21,
						About: "Sit commodo consectetur minim amet ex. Elit aute mollit fugiat labore sint ipsum dolor cupidatat qui reprehenderit. Eu nisi in exercitation culpa sint aliqua nulla nulla proident eu. Nisi reprehenderit anim cupidatat dolor incididunt laboris mollit magna commodo ex. Cupidatat sit id aliqua amet nisi et voluptate voluptate commodo ex eiusmod et nulla velit.\n",
						Gender: "female",
					},
					User{
						Id: 6,
						Name: "Jennings Mays",
						Age: 39,
						About: "Veniam consectetur non non aliquip exercitation quis qui. Aliquip duis ut ad commodo consequat ipsum cupidatat id anim voluptate deserunt enim laboris. Sunt nostrud voluptate do est tempor esse anim pariatur. Ea do amet Lorem in mollit ipsum irure Lorem exercitation. Exercitation deserunt adipisicing nulla aute ex amet sint tempor incididunt magna. Quis et consectetur dolor nulla reprehenderit culpa laboris voluptate ut mollit. Qui ipsum nisi ullamco sit exercitation nisi magna fugiat anim consectetur officia.\n",
						Gender: "male",
					},
				},
				NextPage: false,
			},
			Error: false,
		},
		testCase{
			Value: "Name",
			Result: &SearchResponse{
				Users: []User{
					User{
						Id: 1,
						Name: "Hylda Mayer",
						Age: 21,
						About: "Sit commodo consectetur minim amet ex. Elit aute mollit fugiat labore sint ipsum dolor cupidatat qui reprehenderit. Eu nisi in exercitation culpa sint aliqua nulla nulla proident eu. Nisi reprehenderit anim cupidatat dolor incididunt laboris mollit magna commodo ex. Cupidatat sit id aliqua amet nisi et voluptate voluptate commodo ex eiusmod et nulla velit.\n",
						Gender: "female",
					},
					User{
						Id: 6,
						Name: "Jennings Mays",
						Age: 39,
						About: "Veniam consectetur non non aliquip exercitation quis qui. Aliquip duis ut ad commodo consequat ipsum cupidatat id anim voluptate deserunt enim laboris. Sunt nostrud voluptate do est tempor esse anim pariatur. Ea do amet Lorem in mollit ipsum irure Lorem exercitation. Exercitation deserunt adipisicing nulla aute ex amet sint tempor incididunt magna. Quis et consectetur dolor nulla reprehenderit culpa laboris voluptate ut mollit. Qui ipsum nisi ullamco sit exercitation nisi magna fugiat anim consectetur officia.\n",
						Gender: "male",
					},
				},
				NextPage: false,
			},
			Error: false,
		},
		testCase{
			Value: "Age",
			Result: &SearchResponse{
				Users: []User{
					User{
						Id: 1,
						Name: "Hylda Mayer",
						Age: 21,
						About: "Sit commodo consectetur minim amet ex. Elit aute mollit fugiat labore sint ipsum dolor cupidatat qui reprehenderit. Eu nisi in exercitation culpa sint aliqua nulla nulla proident eu. Nisi reprehenderit anim cupidatat dolor incididunt laboris mollit magna commodo ex. Cupidatat sit id aliqua amet nisi et voluptate voluptate commodo ex eiusmod et nulla velit.\n",
						Gender: "female",
					},
					User{
						Id: 6,
						Name: "Jennings Mays",
						Age: 39,
						About: "Veniam consectetur non non aliquip exercitation quis qui. Aliquip duis ut ad commodo consequat ipsum cupidatat id anim voluptate deserunt enim laboris. Sunt nostrud voluptate do est tempor esse anim pariatur. Ea do amet Lorem in mollit ipsum irure Lorem exercitation. Exercitation deserunt adipisicing nulla aute ex amet sint tempor incididunt magna. Quis et consectetur dolor nulla reprehenderit culpa laboris voluptate ut mollit. Qui ipsum nisi ullamco sit exercitation nisi magna fugiat anim consectetur officia.\n",
						Gender: "male",
					},
				},
				NextPage: false,
			},
			Error: false,
		},
		testCase{
			Value: "About",
			Result: &SearchResponse{
				Users: []User{
					User{
						Id: 1,
						Name: "Hylda Mayer",
						Age: 21,
						About: "Sit commodo consectetur minim amet ex. Elit aute mollit fugiat labore sint ipsum dolor cupidatat qui reprehenderit. Eu nisi in exercitation culpa sint aliqua nulla nulla proident eu. Nisi reprehenderit anim cupidatat dolor incididunt laboris mollit magna commodo ex. Cupidatat sit id aliqua amet nisi et voluptate voluptate commodo ex eiusmod et nulla velit.\n",
						Gender: "female",
					},
					User{
						Id: 6,
						Name: "Jennings Mays",
						Age: 39,
						About: "Veniam consectetur non non aliquip exercitation quis qui. Aliquip duis ut ad commodo consequat ipsum cupidatat id anim voluptate deserunt enim laboris. Sunt nostrud voluptate do est tempor esse anim pariatur. Ea do amet Lorem in mollit ipsum irure Lorem exercitation. Exercitation deserunt adipisicing nulla aute ex amet sint tempor incididunt magna. Quis et consectetur dolor nulla reprehenderit culpa laboris voluptate ut mollit. Qui ipsum nisi ullamco sit exercitation nisi magna fugiat anim consectetur officia.\n",
						Gender: "male",
					},
				},
				NextPage: false,
			},
			Error: false,
		},
		testCase{
			Value: "Gender",
			Result: &SearchResponse{
				Users: []User{
					User{
						Id: 1,
						Name: "Hylda Mayer",
						Age: 21,
						About: "Sit commodo consectetur minim amet ex. Elit aute mollit fugiat labore sint ipsum dolor cupidatat qui reprehenderit. Eu nisi in exercitation culpa sint aliqua nulla nulla proident eu. Nisi reprehenderit anim cupidatat dolor incididunt laboris mollit magna commodo ex. Cupidatat sit id aliqua amet nisi et voluptate voluptate commodo ex eiusmod et nulla velit.\n",
						Gender: "female",
					},
					User{
						Id: 6,
						Name: "Jennings Mays",
						Age: 39,
						About: "Veniam consectetur non non aliquip exercitation quis qui. Aliquip duis ut ad commodo consequat ipsum cupidatat id anim voluptate deserunt enim laboris. Sunt nostrud voluptate do est tempor esse anim pariatur. Ea do amet Lorem in mollit ipsum irure Lorem exercitation. Exercitation deserunt adipisicing nulla aute ex amet sint tempor incididunt magna. Quis et consectetur dolor nulla reprehenderit culpa laboris voluptate ut mollit. Qui ipsum nisi ullamco sit exercitation nisi magna fugiat anim consectetur officia.\n",
						Gender: "male",
					},
				},
				NextPage: false,
			},
			Error: false,
		},
		testCase{
			Value: "Country",
			Result: nil,
			Error: true,
		},
	}
	ts:=httptest.NewServer(http.HandlerFunc(SearchServer))
	for caseNum, caseOrderField:=range casesOrderField{
		req:=SearchRequest{
			Limit: 5,
			Offset: 0,
			Query: "commodo co",
			OrderField: caseOrderField.Value.(string),
			OrderBy: 0,
		}
		client:=SearchClient{AccessToken: "4fer4f6er4686er", URL: ts.URL+"/client=0"}
		result, err:=client.FindUsers(req)
		if err!=nil && !caseOrderField.Error{
			t.Errorf("[%d] unexprcted error: %#v", caseNum, err)
		}
		if err==nil && caseOrderField.Error{
			t.Errorf("[%d] exprcted error, got nil", caseNum)
		}
		if caseOrderField.Result==result{//!reflect.DeepEqual(caseLimit.Result, result){
			t.Errorf("[%d] wrong result, exprcted %#v, got %#v", caseNum, caseOrderField.Result, result)
		}
	}
	ts.Close()

}

func TestRequestOrderBy(t *testing.T){
	casesOrderBy:=[]testCase{
		testCase{
			Value: 4,
			Result: nil,
			Error: true,
		},
		testCase{
			Value: -1,
			Result: &SearchResponse{
				Users: []User{
					User{
						Id: 6,
						Name: "Jennings Mays",
						Age: 39,
						About: "Veniam consectetur non non aliquip exercitation quis qui. Aliquip duis ut ad commodo consequat ipsum cupidatat id anim voluptate deserunt enim laboris. Sunt nostrud voluptate do est tempor esse anim pariatur. Ea do amet Lorem in mollit ipsum irure Lorem exercitation. Exercitation deserunt adipisicing nulla aute ex amet sint tempor incididunt magna. Quis et consectetur dolor nulla reprehenderit culpa laboris voluptate ut mollit. Qui ipsum nisi ullamco sit exercitation nisi magna fugiat anim consectetur officia.\n",
						Gender: "male",
					},
					User{
						Id: 1,
						Name: "Hylda Mayer",
						Age: 21,
						About: "Sit commodo consectetur minim amet ex. Elit aute mollit fugiat labore sint ipsum dolor cupidatat qui reprehenderit. Eu nisi in exercitation culpa sint aliqua nulla nulla proident eu. Nisi reprehenderit anim cupidatat dolor incididunt laboris mollit magna commodo ex. Cupidatat sit id aliqua amet nisi et voluptate voluptate commodo ex eiusmod et nulla velit.\n",
						Gender: "female",
					},
				},
				NextPage: false,
			},
			Error: false,
		},
		testCase{
			Value: 0,
			Result: &SearchResponse{
				Users: []User{
					User{
						Id: 1,
						Name: "Hylda Mayer",
						Age: 21,
						About: "Sit commodo consectetur minim amet ex. Elit aute mollit fugiat labore sint ipsum dolor cupidatat qui reprehenderit. Eu nisi in exercitation culpa sint aliqua nulla nulla proident eu. Nisi reprehenderit anim cupidatat dolor incididunt laboris mollit magna commodo ex. Cupidatat sit id aliqua amet nisi et voluptate voluptate commodo ex eiusmod et nulla velit.\n",
						Gender: "female",
					},
					User{
						Id: 6,
						Name: "Jennings Mays",
						Age: 39,
						About: "Veniam consectetur non non aliquip exercitation quis qui. Aliquip duis ut ad commodo consequat ipsum cupidatat id anim voluptate deserunt enim laboris. Sunt nostrud voluptate do est tempor esse anim pariatur. Ea do amet Lorem in mollit ipsum irure Lorem exercitation. Exercitation deserunt adipisicing nulla aute ex amet sint tempor incididunt magna. Quis et consectetur dolor nulla reprehenderit culpa laboris voluptate ut mollit. Qui ipsum nisi ullamco sit exercitation nisi magna fugiat anim consectetur officia.\n",
						Gender: "male",
					},
				},
				NextPage: false,
			},
			Error: false,
		},
		testCase{
			Value: 1,
			Result: &SearchResponse{
				Users: []User{
					User{
						Id: 1,
						Name: "Hylda Mayer",
						Age: 21,
						About: "Sit commodo consectetur minim amet ex. Elit aute mollit fugiat labore sint ipsum dolor cupidatat qui reprehenderit. Eu nisi in exercitation culpa sint aliqua nulla nulla proident eu. Nisi reprehenderit anim cupidatat dolor incididunt laboris mollit magna commodo ex. Cupidatat sit id aliqua amet nisi et voluptate voluptate commodo ex eiusmod et nulla velit.\n",
						Gender: "female",
					},
					User{
						Id: 6,
						Name: "Jennings Mays",
						Age: 39,
						About: "Veniam consectetur non non aliquip exercitation quis qui. Aliquip duis ut ad commodo consequat ipsum cupidatat id anim voluptate deserunt enim laboris. Sunt nostrud voluptate do est tempor esse anim pariatur. Ea do amet Lorem in mollit ipsum irure Lorem exercitation. Exercitation deserunt adipisicing nulla aute ex amet sint tempor incididunt magna. Quis et consectetur dolor nulla reprehenderit culpa laboris voluptate ut mollit. Qui ipsum nisi ullamco sit exercitation nisi magna fugiat anim consectetur officia.\n",
						Gender: "male",
					},
				},
				NextPage: false,
			},
			Error: false,
		},

	}
	ts:=httptest.NewServer(http.HandlerFunc(SearchServer))
	for caseNum, caseOrderBy:=range casesOrderBy{
		req:=SearchRequest{
			Limit: 5,
			Offset: 0,
			Query: "commodo co",
			OrderField: "Name",
			OrderBy: caseOrderBy.Value.(int),
		}
		client:=SearchClient{AccessToken: "4fer4f6er4686er", URL: ts.URL+"/client=0"}
		result, err:=client.FindUsers(req)
		if err!=nil && !caseOrderBy.Error{
			t.Errorf("[%d] unexprcted error: %#v", caseNum, err)
		}
		if err==nil && caseOrderBy.Error{
			t.Errorf("[%d] exprcted error, got nil", caseNum)
		}
		if caseOrderBy.Result==result{//!reflect.DeepEqual(caseLimit.Result, result){
			t.Errorf("[%d] wrong result, exprcted %#v, got %#v", caseNum, caseOrderBy.Result, result)
		}
	}
	ts.Close()
}

func TestClientError(t *testing.T){
	casesTimeout:=[]testCase{
		testCase{
			Value: "forTimeoutToken",
			Result: nil,
			Error: true,
		},
		testCase{
			Value: "forTemporaryError",
			Result: nil,
			Error: true,
		},
		testCase{
			Value: "forFatalServerError",
			Result: nil,
			Error: true,
		},
		//testCase{
		//	Value: "forJsonError2",
		//	Result: nil,
		//	Error: true,
		//},
	}
	ts:=httptest.NewServer(http.HandlerFunc(SearchServer))
	for caseNum, caseTimeout:=range casesTimeout{
		req:=SearchRequest{
			Limit: 5,
			Offset: 0,
			Query: "commodo co",
			OrderField: "Name",
			OrderBy: 0,
		}
		client:=SearchClient{AccessToken: caseTimeout.Value.(string), URL: ts.URL+"/client=0"}
		result, err:=client.FindUsers(req)
		if err!=nil && !caseTimeout.Error{
			t.Errorf("[%d] unexprcted error: %#v", caseNum, err)
		}
		if err==nil && caseTimeout.Error{
			t.Errorf("[%d] exprcted error, got nil", caseNum)
		}
		if caseTimeout.Result==result{//!reflect.DeepEqual(caseLimit.Result, result){
			t.Errorf("[%d] wrong result, exprcted %#v, got %#v", caseNum, caseTimeout.Result, result)
		}
	}
	ts.Close()
}

func TestWrongJson(t *testing.T){
	caseJson:=testCase{
		Value:"forJsonError",
		Result: nil,
		Error: true,
	}
	ts:=httptest.NewServer(http.HandlerFunc(SearchServer))

		req:=SearchRequest{
			Limit: 5,
			Offset: 0,
			Query: "commodo co",
			OrderField: "Name",
			OrderBy: 0,
		}
		client:=SearchClient{AccessToken: caseJson.Value.(string), URL: ts.URL+"/client=0"}
		result, err:=client.FindUsers(req)
		if err!=nil && !caseJson.Error{
			t.Errorf("[%d] unexprcted error: %#v", 0, err)
		}
		if err==nil && caseJson.Error{
			t.Errorf("[%d] exprcted error, got nil", 0)
		}
		if caseJson.Result==result{//!reflect.DeepEqual(caseLimit.Result, result){
			t.Errorf("[%d] wrong result, exprcted %#v, got %#v", 0, caseJson.Result, result)
		}

	ts.Close()
}

func TestWrongJson2(t *testing.T){
	case2:=testCase{
		Value: "forJsonError2",
		Result: nil,
		Error: true,
	}
	ts:=httptest.NewServer(http.HandlerFunc(SearchServer))

	req:=SearchRequest{
		Limit: 5,
		Offset: 0,
		Query: "commodo co",
		OrderField: "Name",
		OrderBy: 8,
	}
	client:=SearchClient{AccessToken: case2.Value.(string), URL: ts.URL+"/client=0"}
	result, err:=client.FindUsers(req)
	if err!=nil && !case2.Error{
		t.Errorf("[%d] unexprcted error: %#v", 0, err)
	}
	if err==nil && case2.Error{
		t.Errorf("[%d] exprcted error, got nil", 0)
	}
	if case2.Result==result{//!reflect.DeepEqual(caseLimit.Result, result){
		t.Errorf("[%d] wrong result, exprcted %#v, got %#v", 0, case2.Result, result)
	}

	ts.Close()
}

