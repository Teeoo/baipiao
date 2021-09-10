package Lua

import (
	"context"
	"errors"
	"fmt"
	"github.com/cosmotek/loguago"
	"github.com/rs/zerolog"
	lua "github.com/yuin/gopher-lua"
	"io/ioutil"
	luar "layeh.com/gopher-luar"
	"log"
	"os"
	"path"
	"runtime"
	"strings"
	"time"
)

// LuaVMRun ...
func LuaVMRun(parms []interface{}) {
	L := lua.NewState(lua.Options{SkipOpenLibs: true})
	//lua.Options{SkipOpenLibs: true}
	lua.OpenPackage(L) // Must be first
	lua.OpenBase(L)
	lua.OpenTable(L)
	lua.OpenIo(L)
	lua.OpenOs(L)
	lua.OpenString(L)
	lua.OpenMath(L)
	lua.OpenDebug(L)
	lua.OpenChannel(L)
	lua.OpenCoroutine(L)
	defer func() {
		L.Close()
		runtime.GC()
		if err := recover(); err != nil {
			log.Println(err)
		}

	}()
	zlogger := zerolog.New(os.Stdout)
	logger := loguago.NewLogger(zlogger.With().Str("unit", "baipiao").Logger())
	L.PreloadModule("Event", NewEventModule().Loader)
	L.PreloadModule("logger", logger.Loader)
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	L.SetContext(ctx)
	defer cancel()
	scripts, _ := GetLuaList("/scripts/", ".lua")
	for fName, scLua := range scripts {
		var err error
		if err = L.DoString(scLua); err != nil {
			panic(err)
			return
		}
		var Ret int
		var r lua.LValue
		if len(parms) == 1 {
			switch parms[0].(type) {
			case map[string]interface{}:
				fName = fmt.Sprintf("File %s when  call Process params [0] %v", fName, parms[0])
				r, err = CallGlobalValue(L, "Process", parms[0])
				break
			}
		}
		if err != nil {
			log.Printf("LuaVMRun CallGlobal err %v detail %v\n", err, fName)
			return

		}
		if n, ok := r.(lua.LNumber); ok {

			Ret = int(n)

		}
		if Ret == 1 || err != nil {
			continue
		}
		if Ret == 2 {
			return
		}

	}

}

func CallGlobalValue(L *lua.LState, fnName string, args ...interface{}) (r lua.LValue, err error) {
	fn := L.GetGlobal(fnName)
	if fn.Type() != lua.LTFunction {
		err = errors.New(fmt.Sprintf("Unknow Lua Function:%v", fnName))
		return
	}
	// 组合参数列表
	lpValues := []lua.LValue{}
	argsArr := []interface{}(args)
	for _, v := range argsArr {
		lpValues = append(lpValues, luar.New(L, v))
	}

	err = L.CallByParam(lua.P{
		Fn:      fn,
		NRet:    1,
		Protect: true,
	}, lpValues...)

	r = L.Get(-1)
	L.Pop(1)
	return
}

// GetAppPath 取文件所在路径
func GetAppPath() string {
	//file, _ := exec.LookPath(os.Args[0])
	//path, _ := filepath.Abs(file)
	//index := strings.LastIndex(path, string(os.PathSeparator))
	//return path[:index]
	_, filepath, _, _ := runtime.Caller(1)
	return path.Dir(filepath)
}

// GetLuaList 获取指定目录下的所有文件，不进入下一级目录搜索，可以匹配后缀过滤。
func GetLuaList(dirPth string, suffix string) (luascripts map[string]string, err error) {
	scripts := make(map[string]string)
	dir, err := ioutil.ReadDir(GetAppPath() + dirPth)
	if err != nil {
		return scripts, err
	}
	suffix = strings.ToUpper(suffix) //忽略后缀匹配的大小写
	for _, fi := range dir {
		if fi.IsDir() { // 忽略目录
			continue
		}
		if strings.HasSuffix(strings.ToUpper(fi.Name()), suffix) { //匹配文件
			userBuff := InFile(dirPth + fi.Name())
			scripts[fi.Name()] = string(userBuff)
		}
	}
	return scripts, nil
}

func InFile(name string) []byte {
	if contents, err := ioutil.ReadFile(GetAppPath() + name); err == nil {
		return contents
	}
	return nil
}
