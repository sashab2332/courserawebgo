package main

import (
	"bufio"
	"encoding/json"
	"sync"
	"fmt"

	easyjson "github.com/mailru/easyjson"
	jlexer "github.com/mailru/easyjson/jlexer"
	jwriter "github.com/mailru/easyjson/jwriter"
	"io"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
)
const filePath1 string = "./data/users.txt"
var pool *sync.Pool

func initPool(){
	pool=&sync.Pool{
		New: func()interface{}{
			//return new(userStruct)
			return &userStruct{}
		},
	}
}
var (
	_ *json.RawMessage
	_ *jlexer.Lexer
	_ *jwriter.Writer
	_ easyjson.Marshaler
)
func (v *userStruct) UnmarshalJSON(data []byte) error {
	r := jlexer.Lexer{Data: data}
	easyjsonE4fc97f6DecodeMmain(&r, v)
	return r.Error()
}

func easyjsonE4fc97f6DecodeMmain(in *jlexer.Lexer, out *userStruct) {
	isTopLevel := in.IsStart()
	if in.IsNull() {
		if isTopLevel {
			in.Consumed()
		}
		in.Skip()
		return
	}
	in.Delim('{')
	for !in.IsDelim('}') {
		key := in.UnsafeFieldName(false)
		in.WantColon()
		if in.IsNull() {
			in.Skip()
			in.WantComma()
			continue
		}
		switch key {
		case "browsers":
			if in.IsNull() {
				in.Skip()
				out.Browsers = nil
			} else {
				in.Delim('[')
				if out.Browsers == nil {
					if !in.IsDelim(']') {
						out.Browsers = make([]string, 0, 4)
					} else {
						out.Browsers = []string{}
					}
				} else {
					out.Browsers = (out.Browsers)[:0]
				}
				for !in.IsDelim(']') {
					var v1 string
					v1 = string(in.String())
					out.Browsers = append(out.Browsers, v1)
					in.WantComma()
				}
				in.Delim(']')
			}
		case "company":
			out.Company = string(in.String())
		case "country":
			out.Country = string(in.String())
		case "email":
			out.Email = string(in.String())
		case "job":
			out.Job = string(in.String())
		case "name":
			out.Name = string(in.String())
		case "phone":
			out.Phone = string(in.String())
		default:
			in.SkipRecursive()
		}
		in.WantComma()
	}
	in.Delim('}')
	if isTopLevel {
		in.Consumed()
	}
}

func main(){
	FastSearch(ioutil.Discard)
}

type userStruct struct{
	Browsers []string `json:"browsers"`
	Company string `json:"company,-"`
	Country string `json:"country,-"`
	Email string `json:"email"`
	Job string `json:"job,-"`
	Name string `json:"name"`
	Phone string `json:"phone,-"`
}


func FastSearch(out io.Writer) {
	fmt.Fprintln(out, "found users:")
	file, err := os.Open(filePath1)
	if err != nil {
		fmt.Println("error while opening file")
		return
	}
	fileScanner:=bufio.NewScanner(file)
	uniqueBrowsers := 0
	seBrowsers:=make(map[string]bool)
	initPool()
	user:=pool.Get().(*userStruct)
	for i:=0;fileScanner.Scan() ; i++{
		err:=user.UnmarshalJSON(fileScanner.Bytes())
		if err != nil {
			panic(err)
		}


		isAndroid := false
		isMSIE := false

		for _, browser := range user.Browsers {

			var notSeenBefore bool

			if strings.Contains(browser, "Android"){
				isAndroid = true

				notSeenBefore=true

			}
			if strings.Contains(browser, "MSIE") {
				isMSIE = true

				notSeenBefore=true
			}
			if _, ok:=seBrowsers[browser]; !ok && notSeenBefore{
				seBrowsers[browser]=true
				uniqueBrowsers++
			}

		}

		if !(isAndroid && isMSIE) {
			continue
		}


		fmt.Fprintln(out, "["+strconv.Itoa(i)+"] "+user.Name+" <"+strings.ReplaceAll(user.Email, "@", " [at] ")+">")
		pool.Put(user)

	}

	fmt.Fprintln(out, "\nTotal unique browsers", uniqueBrowsers)
}
