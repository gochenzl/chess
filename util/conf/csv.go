package conf

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"os"
	"reflect"
	"strconv"
)

func LoadCsvFromFile(confFile string, configData interface{}, skipHead bool) error {
	f, err := os.Open(confFile)
	if err != nil {
		return err
	}
	defer f.Close()

	return LoadCsv(f, configData, skipHead)
}

func LoadCsv(reader io.Reader, configData interface{}, skipHead bool) error {
	sliceValue := reflect.ValueOf(configData).Elem().Field(0)
	sliceValue.SetLen(0)
	rt := sliceValue.Type().Elem()

	csvReader := csv.NewReader(reader)
	csvReader.FieldsPerRecord = -1
	csvReader.TrimLeadingSpace = true
	records, err := csvReader.ReadAll()
	if err != nil {
		return err
	}

	for i := 0; i < len(records); i++ {
		if skipHead && i == 0 { // ignore head line
			continue
		}

		// ignore empty line
		if len(records[i]) == 1 && len(records[i][0]) == 0 {
			continue
		}

		structVal, err := buildCsvStruct(i, records[i], rt)
		if err != nil {
			return err
		}
		sliceValue = reflect.Append(sliceValue, structVal)
	}

	reflect.ValueOf(configData).Elem().Field(0).Set(sliceValue)

	return nil
}

func buildCsvStruct(lineno int, strs []string, rt reflect.Type) (reflect.Value, error) {
	ptr := reflect.New(rt)
	structVal := ptr.Elem()
	for i := 0; i < len(strs) && i < rt.NumField(); i++ {
		field := rt.Field(i)

		switch field.Type.Kind() {
		case reflect.Int:
			fallthrough
		case reflect.Int8:
			fallthrough
		case reflect.Int16:
			fallthrough
		case reflect.Int32:
			fallthrough
		case reflect.Int64:
			if x, err := strconv.ParseInt(strs[i], 10, 64); err != nil {
				return reflect.Zero(rt), errors.New(fmt.Sprintf("line %d field %d invalid integer", lineno, i))
			} else {
				structVal.Field(i).SetInt(x)
			}

		case reflect.Uint8:
			fallthrough
		case reflect.Uint16:
			fallthrough
		case reflect.Uint32:
			fallthrough
		case reflect.Uint64:
			if x, err := strconv.ParseUint(strs[i], 10, 64); err != nil {
				return reflect.Zero(rt), errors.New(fmt.Sprintf("line %d field %d invalid integer", lineno, i))
			} else {
				structVal.Field(i).SetUint(x)
			}

		case reflect.Float32:
			fallthrough
		case reflect.Float64:
			if x, err := strconv.ParseFloat(strs[i], 64); err != nil {
				return reflect.Zero(rt), errors.New(fmt.Sprintf("line %d field %d invalid float", lineno, i))
			} else {
				structVal.Field(i).SetFloat(x)
			}

		case reflect.Bool:
			if x, err := strconv.ParseBool(strs[i]); err != nil {
				return reflect.Zero(rt), errors.New(fmt.Sprintf("line %d field %d invalid bool", lineno, i))
			} else {
				structVal.Field(i).SetBool(x)
			}

		case reflect.String:
			structVal.Field(i).SetString(strs[i])

		default:
			return reflect.Zero(rt), errors.New(fmt.Sprintf("line %d field %d unknown type", lineno, i))

		}
	}

	return ptr.Elem(), nil
}
