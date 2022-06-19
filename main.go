package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"golang.org/x/exp/slices"
)

type Arguments map[string]string
type Record struct {
	Id    string `json:"id"`
	Email string `json:"email"`
	Age   int8   `json:"age"`
}

var pOperation = "operation"
var pFileName = "fileName"
var pItem = "item"
var pId = "id"

var opAdd = "add"
var opList = "list"
var opFind = "findById"
var opRemove = "remove"

var operationsIndex = []string{opAdd, opList, opFind, opRemove}

type opRequirements []string

var requiredFlag = pOperation
var additionalRequiredFlags = map[string]opRequirements{
	opList:   []string{pFileName},
	opAdd:    []string{pFileName, pItem},
	opFind:   []string{pFileName, pId},
	opRemove: []string{pFileName, pId},
}

func Perform(args Arguments, writer io.Writer) error {
	err0 := validateConsequently(args, requiredFlag, additionalRequiredFlags)

	if err0 != nil {
		return err0
	}

	var buf []byte
	var err error

	switch args[pOperation] {
	case opList:
		buf, err = read(args)
	case opAdd:
		buf, err = add(args)
	case opFind:
		buf, err = find(args)
	case opRemove:
		buf, err = remove(args)
	}

	if err != nil {
		fmt.Println("err?", err, args)
		return err
	}

	writer.Write(buf)

	return nil
}

// func validateOpList(args Arguments) error {
// 	fileName := args[pFileName]

// 	if _, err := os.Stat(fileName); errors.Is(err, os.ErrNotExist) {
// 		return err
// 	}

// 	return nil
// }

// func doList(args Arguments) ([]byte, error) {
// 	// 1. Validate all entered params have values
// 	err0 := validateParamEntered(args, []string{pOperation, pFileName})

// 	if err0 != nil {
// 		return nil, err0
// 	}

// 	// 2. Validate that operation requested is allowed
// 	err1 := validateOpAllowed(args[pOperation])

// 	if err1 != nil {
// 		return nil, err1
// 	}

// 	var buf []byte
// 	var err error

// 	if err != nil {
// 		fmt.Println("err?", err, args)
// 		return nil, err
// 	}

// 	return buf, nil
// }

func read(args Arguments) ([]byte, error) {
	var r, err = os.OpenFile(args[pFileName], os.O_RDONLY|os.O_CREATE, 0444)

	defer r.Close()

	if err != nil {
		return nil, err
	}

	buf, err := ioutil.ReadAll(r)

	if err != nil {
		return nil, err
	}

	// m := []Record{}

	// err2 := json.Unmarshal(buf, &m)

	// if err2 != nil {
	// 	return nil, err2
	// }

	return buf, nil
}

func add(args Arguments) ([]byte, error) {
	var rw, err = os.OpenFile(args[pFileName], os.O_RDWR|os.O_CREATE, 0644)

	defer rw.Close()

	if err != nil {
		return nil, err
	}

	buf, err := ioutil.ReadAll(rw)

	m := []Record{}

	if len(buf) != 0 {
		err2 := json.Unmarshal(buf, &m)
		if err2 != nil {
			return nil, err2
		}
	}

	n := Record{}
	item := args[pItem]

	err3 := json.Unmarshal([]byte(item), &n)

	if err3 != nil {
		return nil, err3
	}

	for _, rec := range m {
		if rec.Id == n.Id {
			return []byte("Item with id " + rec.Id + " already exists"), nil
		}
	}

	m = append(m, n)
	marshalled, errLast := json.Marshal(m)

	if errLast == nil {
		if err := ioutil.WriteFile(args[pFileName], marshalled, 0660); err != nil {
			return nil, err
		}
	}

	return marshalled, errLast
}

func find(args Arguments) ([]byte, error) {
	var rw, err = os.OpenFile(args[pFileName], os.O_RDONLY|os.O_CREATE, 0644)

	defer rw.Close()

	if err != nil {
		return nil, err
	}

	buf, err := ioutil.ReadAll(rw)

	m := []Record{}

	if len(buf) == 0 {
		return nil, errors.New("No data to search for a record")
	}

	err2 := json.Unmarshal(buf, &m)
	if err2 != nil {
		return nil, err2
	}

	for _, rec := range m {
		if rec.Id == args[pId] {
			marshalled, err := json.Marshal(rec)
			return marshalled, err
		}
	}

	return nil, nil
}

func remove(args Arguments) ([]byte, error) {
	var rw, err = os.OpenFile(args[pFileName], os.O_RDONLY|os.O_CREATE, 0644)

	defer rw.Close()

	if err != nil {
		return nil, err
	}

	buf, err := ioutil.ReadAll(rw)

	m := []Record{}

	if len(buf) == 0 {
		return nil, errors.New("No data to search for a record")
	}

	err2 := json.Unmarshal(buf, &m)
	if err2 != nil {
		return nil, err2
	}

	var foundId string

	for ind, rec := range m {
		if rec.Id == args[pId] {
			foundId = rec.Id
			m = append(m[:ind], m[ind+1:]...)
		}
	}

	if foundId == "" {
		return []byte("Item with id " + args[pId] + " not found"), nil
	}

	marshalled, errLast := json.Marshal(m)

	if errLast == nil {
		if err := ioutil.WriteFile(args[pFileName], marshalled, 0660); err != nil {
			return nil, err
		}
	}

	return marshalled, errLast
}

func validateConsequently(args Arguments, reqFlag string, params map[string]opRequirements) error {
	errOperationSpecified := validateParamEntered(args, []string{reqFlag})

	if errOperationSpecified != nil {
		return errOperationSpecified
	}

	errAllowed := validateOpAllowed(args[reqFlag])

	if errAllowed != nil {
		return errAllowed
	}

	op := args[reqFlag]
	errEntered := validateParamEntered(args, params[op])

	if errEntered != nil {
		return errEntered
	}

	return nil
}

func validateOpAllowed(operation string) error {
	if !slices.Contains(operationsIndex, operation) {
		return errors.New("Operation " + operation + " not allowed!")
	}

	return nil
}

func validateParamEntered(args Arguments, params []string) error {
	for _, param := range params {
		value := args[param]
		if len(value) == 0 {
			return errors.New("-" + param + " flag has to be specified")
		}
	}

	return nil
}

func parseArgs() Arguments {
	var flagNames = [4]string{pOperation, pFileName, pItem, pId}
	var flagValues = [4]string{"", "", "", ""}

	flag.StringVar(&flagValues[0], flagNames[0], "", "Possible values are 'list', 'add', 'findById', 'remove'")
	flag.StringVar(&flagValues[1], flagNames[1], "", "Path to a DB file")
	flag.StringVar(&flagValues[2], flagNames[2], "", "Item to add")
	flag.StringVar(&flagValues[3], flagNames[3], "", "ID to search for")

	flag.Parse()

	var args = make(Arguments)

	for indx, flagName := range flagNames {
		fmt.Println(indx, flagName, flagValues[indx])
		if len(flagValues[indx]) > 0 {
			args[flagName] = flagValues[indx]
		}
	}

	fmt.Println("!!args", args)

	return args
}

func main() {
	err := Perform(parseArgs(), os.Stdout)
	if err != nil {
		panic(err)
	}
}
