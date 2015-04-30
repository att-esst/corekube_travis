package util

import (
	"reflect"
	"runtime"
)

func GetFunctionName(i interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
}

func ExtractIPFromStackDetails(details StackDetails, key string) string {
	overlordIP := ""

	for _, i := range details.Stack.Outputs {
		if i.OutputKey == key {
			overlordIP = i.OutputValue.(string)
		}
	}

	return overlordIP
}

func ExtractArrayIPs(details StackDetails, key string) []string {
	ips := []string{}

	for _, i := range details.Stack.Outputs {
		if i.OutputKey == key {
			v := i.OutputValue
			switch t := v.(type) {
			case []interface{}:
				for _, ip := range t {
					ipStr := ip.(string)
					ips = append(ips, ipStr)
				}
			}
		}
	}

	return ips
}
