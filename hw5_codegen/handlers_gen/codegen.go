package main

import (
	"encoding/json"
	"reflect"

	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"strings"
)
type jsonInfo struct{
	URL string `json:"url"`
	Auth bool `json:"auth"`
	Method string `json:"method"`
}
type receiver struct{
	structTypeName string
	Methods []method
}
type method struct{
	URL string
	methodName string
}

func isIntField(field *ast.Field) bool {
	var typeName string
	switch t := field.Type.(type) {
	case *ast.StarExpr:
		if x, ok := t.X.(*ast.Ident); ok {
			typeName = x.Name
		}
	case *ast.Ident:
		typeName = t.Name
	}
	if typeName=="int"{
		return true
}
	return false
}
func getFuncResult(fn *ast.FuncDecl) string {
	var result string
	if fn.Type.Results.List != nil {
		for _, p := range fn.Type.Results.List {
			switch t := p.Type.(type) {
			case *ast.Ident:
				result = t.Name
			case *ast.StarExpr:
				if x, ok := t.X.(*ast.Ident); ok {
					result = x.Name
				}
			}
			if result != "error" {
				break
			}
		}
	}
	return result
}

func getFuncReceiver(fn *ast.FuncDecl) string {
	var result string
	if fn.Recv != nil {
		switch t := fn.Recv.List[0].Type.(type) {
		case *ast.StarExpr:
			if x, ok := t.X.(*ast.Ident); ok {
				result = x.Name
			}
		case *ast.Ident:
			result = t.Name
		}
	}
	return result
}

func getFuncParams(fn *ast.FuncDecl) string{
	var result string
	if fn.Type.Params.List != nil {
		for _, p := range fn.Type.Params.List {
			switch t := p.Type.(type) {
			case *ast.Ident:
				result = t.Name
			case *ast.SelectorExpr:
				result = t.Sel.Name
			}
			if result == "Context" {
				continue
			}
		}
	}
	return result
}

func getTagValues(tag string) map[string]string {
	values := make(map[string]string, 6)
	arr := strings.Split(tag, ",")
	for _, value := range arr {
		arr := strings.Split(value, "=")
		if len(arr) == 1 {
			values[arr[0]] = ""
		} else {
			values[arr[0]] = arr[1]
		}
	}
	return values
}

func fieldValidation(field *ast.Field,tagValues map[string]string, out *os.File, info jsonInfo){
	typ, ok := field.Type.(*ast.Ident)
	if !ok {
		return
	}
	fieldTypeName := typ.Name
	for key, value:=range tagValues{
		switch key {
		case "required" :
			fmt.Fprintln(out, "\tif params."+field.Names[0].Name+"==\"\"{\n\t\terr:=errors.New(\""+strings.ToLower(field.Names[0].Name)+" must me not empty\")\n\t\tnewApiError:=ApiError{HTTPStatus: http.StatusBadRequest, Err: err}\n\t\tw.WriteHeader(newApiError.HTTPStatus)\n\t\tresMap:=map[string]interface{}{\"error\": err.Error()}\n\t\tdata, _:=json.Marshal(resMap)\n\t\tw.Write(data)\n\t\treturn\n\t}")
		case "min":
			switch fieldTypeName{
			case "int": fmt.Fprintln(out, "\tif params."+field.Names[0].Name+"<"+value+"{\n\t\terr := errors.New(\""+strings.ToLower(field.Names[0].Name)+" must be >= 0\")\n\t\tnewApiError := ApiError{HTTPStatus: http.StatusBadRequest, Err: err}\n\t\tw.WriteHeader(newApiError.HTTPStatus)\n\t\tresMap:=map[string]interface{}{\"error\": newApiError.Err.Error()}\n\t\tdata,_:=json.Marshal(resMap)\n\t\tw.Write(data)\n\t\treturn\n\t}")
			case "string": fmt.Fprintln(out, "\tif len(params."+field.Names[0].Name+")<"+value+"{\n\t\terr := errors.New(\""+strings.ToLower(field.Names[0].Name)+" len must be >= 10\")\n\t\tnewApiError := ApiError{HTTPStatus: http.StatusBadRequest, Err: err}\n\t\tw.WriteHeader(newApiError.HTTPStatus)\n\t\tresMap:=map[string]interface{}{\"error\": newApiError.Err.Error()}\n\t\tdata,_:=json.Marshal(resMap)\n\t\tw.Write(data)\n\t\treturn\n\t}")
			}
		case "enum":
			enumSl:=strings.Split(value, "|")
			fmt.Fprint(out, "\tif params."+field.Names[0].Name+" != \""+enumSl[0]+"\"")
			for i:=1; i<len(enumSl);i++{
				fmt.Fprint(out," && params."+field.Names[0].Name+" !=\""+enumSl[i]+"\"")
			}
			fmt.Fprint(out, "{\n")
			fmt.Fprintln(out, "\t\tif params."+field.Names[0].Name+" == \"\"{")
			fmt.Fprintln(out, "\t\t\tparams."+field.Names[0].Name+" = \""+tagValues["default"]+"\"")
			fmt.Fprint(out, "\t\t} else {\n\t\t\terr := errors.New(\""+strings.ToLower(field.Names[0].Name) +" must be one of ["+enumSl[0])
			for i:=1; i<len(enumSl);i++{
				fmt.Fprint(out, ", "+enumSl[i])
			}
			fmt.Fprint(out, "]\")\n\t\t\tnewApiError := ApiError{HTTPStatus: http.StatusBadRequest, Err: err}\n\t\t\tw.WriteHeader(newApiError.HTTPStatus)\n\t\t\tresMap:=map[string]interface{}{\"error\": newApiError.Err.Error()}\n\t\t\tdata,_:=json.Marshal(resMap)\n\t\t\tw.Write(data)\n\t\t\treturn\n\t\t}\n\t}\n")
		case "max":
			fmt.Fprintln(out, "\tif params."+field.Names[0].Name+" > "+value+"{\n\t\terr := errors.New(\"age must be <= 128\")\n\t\tnewApiError := ApiError{HTTPStatus: http.StatusBadRequest, Err: err}\n\t\tw.WriteHeader(newApiError.HTTPStatus)\n\t\tresMap:=map[string]interface{}{\"error\": newApiError.Err.Error()}\n\t\tdata,_:=json.Marshal(resMap)\n\t\tw.Write(data)\n\t\treturn\n\t}")
			}
		}
}



func main() {
	receivers:=[]receiver{receiver{}}

	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, os.Args[1], nil, parser.ParseComments)
	if err != nil {
		log.Fatal(err)
	}

	out, _ := os.Create(os.Args[2])

	fmt.Fprintln(out, `package `+node.Name.Name)
	fmt.Fprintln(out) // empty line
	fmt.Fprintln(out, `import "encoding/json"`)
	fmt.Fprintln(out, `import "errors"`)
	fmt.Fprintln(out, `import "fmt"`)
	fmt.Fprintln(out, `import "net/http"`)
	fmt.Fprintln(out, `import "strconv"`)
	fmt.Fprintln(out) // empty line

	for _, f := range node.Decls {

		g, ok := f.(*ast.FuncDecl)
		if !ok {
			continue
		}

		if g.Recv == nil {
			continue
		}

		if g.Doc == nil {
			continue
		}

		for _, comment := range g.Doc.List {
			if strings.HasPrefix(comment.Text, "// apigen:api ") {

				dataInfo := []byte(strings.Trim(comment.Text, "// apigen:api "))

				var info jsonInfo

				json.Unmarshal(dataInfo, &info)
				currMethod := method{URL: info.URL, methodName: g.Name.Name}

				receiverTypeName := getFuncReceiver(g)
				paramsTypeName := getFuncParams(g)
				resultTypeName := getFuncResult(g)
				have := false
				for i, val := range receivers {
					if val.structTypeName == receiverTypeName {
						receivers[i].Methods= append(val.Methods, currMethod)
						have = true
						break
					}
				}
				if !have {
					newReceiver := receiver{structTypeName: receiverTypeName, Methods: []method{currMethod}}
					receivers = append(receivers, newReceiver)
				}

				var paramsStruct, resultStruct *ast.StructType
				for _, f := range node.Decls {
					g, ok := f.(*ast.GenDecl)
					if !ok {
						continue
					}
					for _, spec := range g.Specs {
						currType, ok := spec.(*ast.TypeSpec)
						if !ok {
							continue
						}
						currStruct, ok := currType.Type.(*ast.StructType)
						if !ok {
							continue
						}
						if currType.Name.Name == paramsTypeName {
							paramsStruct = currStruct
						}
						if currType.Name.Name == resultTypeName {
							resultStruct = currStruct
						}
					}
				}

				fmt.Fprintln(out, "func (h *"+receiverTypeName+") wrapper"+g.Name.Name+"(w http.ResponseWriter, req *http.Request) {")

				if info.Auth == true {
					fmt.Fprintln(out, "\tif req.Header.Get(\"X-Auth\") != \"100500\" {\n\t\terr := errors.New(\"unauthorized\")\n\t\tnewApiError := ApiError{HTTPStatus: http.StatusForbidden, Err: err}\n\t\tw.WriteHeader(newApiError.HTTPStatus)\n\t\tresMap:=map[string]interface{}{\"error\": err.Error()}\n\t\tdata, _:=json.Marshal(resMap)\n\t\tw.Write(data)\n\t\treturn\n\t}")
				}

				fmt.Fprintln(out)

				fmt.Fprintln(out, "\tparams:="+paramsTypeName+"{}")
				fmt.Fprintln(out)

				if info.URL == "/user/profile" {
					fmt.Fprintln(out, "\tif req.Method==http.MethodPost{")

					fmt.Fprintln(out, "\t\terr := req.ParseForm()\n\t\tif err != nil {\n\t\t\tfmt.Println(\"parsing error\")\n\t\t}")

					for _, field := range paramsStruct.Fields.List {
						if isIntField(field){
							if field.Tag != nil {
								tag := reflect.StructTag(field.Tag.Value[1 : len(field.Tag.Value)-1])
								tagVal:=tag.Get("apivalidator")
								values := getTagValues(tagVal)
								value, ok := values["paramname"]
								if ok {
									fmt.Fprintln(out, "\t\tparams."+field.Names[0].Name+", err=strconv.Atoi(req.Form.Get(\""+value+"\"))")
									fmt.Fprintln(out,"\t\tif err!=nil{\n\t\t\terr:=errors.New(\""+strings.ToLower(field.Names[0].Name)+ " must be int\")\n\t\t\tnewApiError:=ApiError{HTTPStatus: http.StatusBadRequest, Err: err}\n\t\t\tw.WriteHeader(newApiError.HTTPStatus)\n\t\t\tresMap:=map[string]interface{}{\"error\": newApiError.Err.Error()}\n\t\t\tdata,_:=json.Marshal(resMap)\n\t\t\tw.Write(data)\n\t\t\treturn\n\t\t}")
								} else {
									fmt.Fprintln(out, "\t\tparams."+field.Names[0].Name+", err=strconv.Atoi(req.Form.Get(\""+strings.ToLower(field.Names[0].Name)+"\"))")
									fmt.Fprintln(out,"\t\tif err!=nil{\n\t\t\terr:=errors.New(\""+strings.ToLower(field.Names[0].Name)+ " must be int\")\n\t\t\tnewApiError:=ApiError{HTTPStatus: http.StatusBadRequest, Err: err}\n\t\t\tw.WriteHeader(newApiError.HTTPStatus)\n\t\t\tresMap:=map[string]interface{}{\"error\": newApiError.Err.Error()}\n\t\t\tdata,_:=json.Marshal(resMap)\n\t\t\tw.Write(data)\n\t\t\treturn\n\t\t}")
								}
								continue
							}
							fmt.Fprintln(out, "\t\tparams."+field.Names[0].Name+", err:=strconv.Atoi(req.Form.Get(\""+strings.ToLower(field.Names[0].Name)+"\"))")
							fmt.Fprintln(out,"\t\tif err!=nil{\n\t\t\terr:=errors.New(\""+strings.ToLower(field.Names[0].Name)+ " must be int\")\n\t\t\tnewApiError:=ApiError{HTTPStatus: http.StatusBadRequest, Err: err}\n\t\t\tw.WriteHeader(newApiError.HTTPStatus)\n\t\t\tresMap:=map[string]interface{}{\"error\": newApiError.Err.Error()}\n\t\t\tdata,_:=json.Marshal(resMap)\n\t\t\tw.Write(data)\n\t\t\treturn\n\t\t}")
							continue
						}
						if field.Tag != nil {
							tag := reflect.StructTag(field.Tag.Value[1 : len(field.Tag.Value)-1])
							tagVal:=tag.Get("apivalidator")
							values := getTagValues(tagVal)
							value, ok := values["paramname"]
							if ok {
								fmt.Fprintln(out, "\t\tparams."+field.Names[0].Name+"=req.Form.Get(\""+value+"\")")
							} else {
								fmt.Fprintln(out, "\t\tparams."+field.Names[0].Name+"=req.Form.Get(\""+strings.ToLower(field.Names[0].Name)+"\")")
							}
							continue
						}
						fmt.Fprintln(out, "\t\tparams."+field.Names[0].Name+"=req.Form.Get(\""+strings.ToLower(field.Names[0].Name)+"\")")
					}
					fmt.Fprintln(out, "\t}else{")
					for _, field := range paramsStruct.Fields.List {
						if isIntField(field){
							if field.Tag != nil {
								tag := reflect.StructTag(field.Tag.Value[1 : len(field.Tag.Value)-1])
								tagVal:=tag.Get("apivalidator")
								values := getTagValues(tagVal)
								value, ok := values["paramname"]
								if ok {
									fmt.Fprintln(out, "\t\tparams."+field.Names[0].Name+", err=strconv.Atoi(req.URL.Query().Get(\""+value+"\"))")
									fmt.Fprintln(out,"\t\tif err!=nil{\n\t\t\terr:=errors.New(\""+strings.ToLower(field.Names[0].Name)+ " must be int\")\n\t\t\tnewApiError:=ApiError{HTTPStatus: http.StatusBadRequest, Err: err}\n\t\t\tw.WriteHeader(newApiError.HTTPStatus)\n\t\t\tresMap:=map[string]interface{}{\"error\": newApiError.Err.Error()}\n\t\t\tdata,_:=json.Marshal(resMap)\n\t\t\tw.Write(data)\n\t\t\treturn\n\t\t}")

								} else {
									fmt.Fprintln(out, "\t\tparams."+field.Names[0].Name+",err=strconv.Atoi(req.URL.Query().Get(\""+strings.ToLower(field.Names[0].Name)+"\"))")
									fmt.Fprintln(out,"\t\tif err!=nil{\n\t\t\terr:=errors.New(\""+strings.ToLower(field.Names[0].Name)+ " must be int\")\n\t\t\tnewApiError:=ApiError{HTTPStatus: http.StatusBadRequest, Err: err}\n\t\t\tw.WriteHeader(newApiError.HTTPStatus)\n\t\t\tresMap:=map[string]interface{}{\"error\": newApiError.Err.Error()}\n\t\t\tdata,_:=json.Marshal(resMap)\n\t\t\tw.Write(data)\n\t\t\treturn\n\t\t}")

								}
								continue
							}
							fmt.Fprintln(out, "\t\tparams."+field.Names[0].Name+", err=strconv.Atoi(req.URL.Query().Get(\""+strings.ToLower(field.Names[0].Name)+"\")")
							fmt.Fprintln(out,"\t\tif err!=nil{\n\t\t\terr:=errors.New(\""+strings.ToLower(field.Names[0].Name)+ " must be int\")\n\t\t\tnewApiError:=ApiError{HTTPStatus: http.StatusBadRequest, Err: err}\n\t\t\tw.WriteHeader(newApiError.HTTPStatus)\n\t\t\tresMap:=map[string]interface{}{\"error\": newApiError.Err.Error()}\n\t\t\tdata,_:=json.Marshal(resMap)\n\t\t\tw.Write(data)\n\t\t\treturn\n\t\t}")
							continue
						}
						if field.Tag != nil {
							tag := reflect.StructTag(field.Tag.Value[1 : len(field.Tag.Value)-1])
							tagVal:=tag.Get("apivalidator")
							values := getTagValues(tagVal)
							value, ok := values["paramname"]
							if ok {
								fmt.Fprintln(out, "\t\tparams."+field.Names[0].Name+"=req.URL.Query().Get(\""+value+"\")")

							} else {
								fmt.Fprintln(out, "\t\tparams."+field.Names[0].Name+"=req.URL.Query().Get(\""+strings.ToLower(field.Names[0].Name)+"\")")

							}
							continue
						}
						fmt.Fprintln(out, "\t\tparams."+field.Names[0].Name+"=req.URL.Query().Get(\""+strings.ToLower(field.Names[0].Name)+"\")")
					}
					fmt.Fprintln(out, "\t}")
				}
				if info.URL == "/user/create" {
					fmt.Fprintln(out, "\tif req.Method==http.MethodGet{")

					fmt.Fprintln(out, "\t\terr := errors.New(\"bad method\")\n\t\tnewApiError := ApiError{HTTPStatus: http.StatusNotAcceptable, Err: err}\n\t\tw.WriteHeader(newApiError.HTTPStatus)\n\t\tresMap:=map[string]interface{}{\"error\": newApiError.Err.Error()}\n\t\tdata,_:=json.Marshal(resMap)\n\t\tw.Write(data)\n\t\treturn\n\t\t\t} else {\n\t\terr := req.ParseForm()\n\t\tif err != nil {\n\t\t\tfmt.Println(\"parsing error\")\n\t\t}")
					for _, field:=range paramsStruct.Fields.List{
						if isIntField(field){
							if field.Tag != nil {
								tag := reflect.StructTag(field.Tag.Value[1 : len(field.Tag.Value)-1])
								tagVal:=tag.Get("apivalidator")
								values := getTagValues(tagVal)
								value, ok := values["paramname"]
								if ok {
									fmt.Fprintln(out, "\t\tparams."+field.Names[0].Name+", err=strconv.Atoi(req.Form.Get(\""+value+"\"))")
									fmt.Fprintln(out,"\t\tif err!=nil{\n\t\t\terr:=errors.New(\""+strings.ToLower(field.Names[0].Name)+ " must be int\")\n\t\t\tnewApiError:=ApiError{HTTPStatus: http.StatusBadRequest, Err: err}\n\t\t\tw.WriteHeader(newApiError.HTTPStatus)\n\t\t\tresMap:=map[string]interface{}{\"error\": newApiError.Err.Error()}\n\t\t\tdata,_:=json.Marshal(resMap)\n\t\t\tw.Write(data)\n\t\t\treturn\n\t\t}")

								} else {
									fmt.Fprintln(out, "\t\tparams."+field.Names[0].Name+", err=strconv.Atoi(req.Form.Get(\""+strings.ToLower(field.Names[0].Name)+"\"))")
									fmt.Fprintln(out,"\t\tif err!=nil{\n\t\t\terr:=errors.New(\""+strings.ToLower(field.Names[0].Name)+ " must be int\")\n\t\t\tnewApiError:=ApiError{HTTPStatus: http.StatusBadRequest, Err: err}\n\t\t\tw.WriteHeader(newApiError.HTTPStatus)\n\t\t\tresMap:=map[string]interface{}{\"error\": newApiError.Err.Error()}\n\t\t\tdata,_:=json.Marshal(resMap)\n\t\t\tw.Write(data)\n\t\t\treturn\n\t\t}")

								}
								continue
							}
							fmt.Fprintln(out, "\t\tparams."+field.Names[0].Name+", err=strconv.Atoi(req.Form.Get(\""+strings.ToLower(field.Names[0].Name)+"\"))")
							fmt.Fprintln(out,"\t\tif err!=nil{\n\t\t\terr:=errors.New(\""+strings.ToLower(field.Names[0].Name)+ " must be int\")\n\t\t\tnewApiError:=ApiError{HTTPStatus: http.StatusBadRequest, Err: err}\n\t\t\tw.WriteHeader(newApiError.HTTPStatus)\n\t\t\tresMap:=map[string]interface{}{\"error\": newApiError.Err.Error()}\n\t\t\tdata,_:=json.Marshal(resMap)\n\t\t\tw.Write(data)\n\t\t\treturn\n\t\t}")
							continue
						}
						if field.Tag != nil {
							tag := reflect.StructTag(field.Tag.Value[1 : len(field.Tag.Value)-1])
							tagVal:=tag.Get("apivalidator")
							values := getTagValues(tagVal)
							value, ok := values["paramname"]
							if ok {
								fmt.Fprintln(out, "\t\tparams."+field.Names[0].Name+"=req.Form.Get(\""+value+"\")")
							} else {
								fmt.Fprintln(out, "\t\tparams."+field.Names[0].Name+"=req.Form.Get(\""+strings.ToLower(field.Names[0].Name)+"\")")
							}
							continue
						}
						fmt.Fprintln(out, "\t\tparams."+field.Names[0].Name+"=req.Form.Get(\""+strings.ToLower(field.Names[0].Name)+"\")")
					}
					fmt.Fprintln(out,"\t}")
				}
				fmt.Fprintln(out)

				for _,field:=range paramsStruct.Fields.List{
					if field.Tag!=nil{
						tag := reflect.StructTag(field.Tag.Value[1 : len(field.Tag.Value)-1])
						tagVal:=tag.Get("apivalidator")
						values:=getTagValues(tagVal)
						fieldValidation(field, values,out,info)
					}
				}
				fmt.Fprintln(out)

				fmt.Fprintln(out, "\tctx:=req.Context()")
				fmt.Fprintln(out)
				fmt.Fprintln(out, "\tres, err:=h."+g.Name.Name+"(ctx, params)\n\tif err != nil {\n\t\tnewApiError, ok:=err.(ApiError)\n\t\tif !ok{\n\t\t\tw.WriteHeader(http.StatusInternalServerError)\n\t\t\tresMap:=map[string]interface{}{\"error\": err.Error()}\n\t\t\tdata, _:=json.Marshal(resMap)\n\t\t\tw.Write(data)\n\t\t\treturn\n\t\t}\n\t\tw.WriteHeader(newApiError.HTTPStatus)\n\t\tresMap:=map[string]interface{}{\"error\": newApiError.Err.Error()}\n\t\tdata, _:=json.Marshal(resMap)\n\t\tw.Write(data)\n\t\treturn\n\t}")
				fmt.Fprintln(out)
				fmt.Fprintln(out, "\tresMap:=map[string]interface{}{\n\t\t\"error\": \"\",\n\t\t\"response\": map[string]interface{}{")

				for _, field := range resultStruct.Fields.List {
					tag := reflect.StructTag(field.Tag.Value[1 : len(field.Tag.Value)-1])
					fmt.Fprintln(out, "\t\t\t\""+tag.Get("json")+"\": res."+field.Names[0].Name+",")
				}
				fmt.Fprintln(out, "\t\t},\n\t}")

				fmt.Fprintln(out)
				fmt.Fprintln(out, "\tdata,_:=json.Marshal(resMap)\n\tw.WriteHeader(http.StatusOK)\n\tw.Write(data)\n}")
			}
		}
		fmt.Fprintln(out)
	}

	for _, receiver:=range receivers{
		if receiver.structTypeName==""{
			continue
		}
		fmt.Fprintln(out, "func(h *"+receiver.structTypeName+") ServeHTTP(w http.ResponseWriter, r *http.Request){")
		fmt.Fprintln(out,"\tswitch r.URL.Path{")

		for _,method:=range receiver.Methods{
			fmt.Fprintln(out, "\tcase \""+method.URL+"\": h.wrapper"+method.methodName+"(w, r)")
		}
		fmt.Fprintln(out, "\tdefault:\n\t\terr:=errors.New(\"unknown method\")\n\t\tnewApiError:=ApiError{HTTPStatus: http.StatusNotFound, Err: err}\n\t\tw.WriteHeader(newApiError.HTTPStatus)\n\t\tresMap:=map[string]interface{}{\"error\":newApiError.Err.Error()}\n\t\tdata,_:=json.Marshal(resMap)\n\t\tw.Write(data)\n\t\treturn\n\t}\n}")
	}
}
