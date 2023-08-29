package register

type jsClass struct {
	Constructor  QJSFnType
	MountMethods map[string]QJSFnType
}

var jsClassList = map[string]*jsClass{}

var jsClassNameList = []string{}

func SetConstructorClass(className string, constructor QJSFnType) {
	if jsClassList[className] == nil {
		jsClassNameList = append(jsClassNameList, className)
		jsClassList[className] = &jsClass{Constructor: constructor}
	} else {
		jsClassList[className].Constructor = constructor
	}
}

func SetClassMethod(className, method string, qjsFn QJSFnType) {
	if jsClassList[className] == nil {
		jsClassNameList = append(jsClassNameList, className)
		jsClassList[className] = &jsClass{
			MountMethods: map[string]QJSFnType{},
		}
	} else if jsClassList[className].MountMethods == nil {
		jsClassList[className].MountMethods = map[string]QJSFnType{}
	}
	jsClassList[className].MountMethods[method] = qjsFn
}
