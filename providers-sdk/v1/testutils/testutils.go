// Copyright (c) Mondoo, Inc.
// SPDX-License-Identifier: BUSL-1.1

package testutils

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mondoo.com/cnquery"
	"go.mondoo.com/cnquery/llx"
	"go.mondoo.com/cnquery/logger"
	"go.mondoo.com/cnquery/mql"
	"go.mondoo.com/cnquery/mqlc"
	"go.mondoo.com/cnquery/providers"
	"go.mondoo.com/cnquery/providers-sdk/v1/lr"
	"go.mondoo.com/cnquery/providers-sdk/v1/resources"
	"go.mondoo.com/cnquery/providers-sdk/v1/testutils/mockprovider"
	"go.mondoo.com/cnquery/providers/mock"
	networkconf "go.mondoo.com/cnquery/providers/network/config"
	networkprovider "go.mondoo.com/cnquery/providers/network/provider"
	osconf "go.mondoo.com/cnquery/providers/os/config"
	osprovider "go.mondoo.com/cnquery/providers/os/provider"
)

var (
	Features     cnquery.Features
	TestutilsDir string
)

func init() {
	logger.InitTestEnv()
	Features = getEnvFeatures()

	_, pathToFile, _, ok := runtime.Caller(0)
	if !ok {
		panic("unable to get runtime for testutils for cnquery providers")
	}
	TestutilsDir = path.Dir(pathToFile)
}

func getEnvFeatures() cnquery.Features {
	env := os.Getenv("FEATURES")
	if env == "" {
		return cnquery.Features{byte(cnquery.PiperCode)}
	}

	arr := strings.Split(env, ",")
	var fts cnquery.Features
	for i := range arr {
		v, ok := cnquery.FeaturesValue[arr[i]]
		if ok {
			fmt.Println("--> activate feature: " + arr[i])
			fts = append(Features, byte(v))
		} else {
			panic("cannot find requested feature: " + arr[i])
		}
	}
	return fts
}

func TomlMock(filepath string) llx.Runtime {
	trans, err := mock.NewFromTomlFile(filepath)
	if err != nil {
		panic(err.Error())
	}

	return trans
}

type tester struct {
	Runtime llx.Runtime
}

func InitTester(runtime llx.Runtime) *tester {
	return &tester{
		Runtime: runtime,
	}
}

func (ctx *tester) Compile(query string) (*llx.CodeBundle, error) {
	return mqlc.Compile(query, nil, mqlc.NewConfig(ctx.Runtime.Schema(), Features))
}

func (ctx *tester) ExecuteCode(bundle *llx.CodeBundle, props map[string]*llx.Primitive) (map[string]*llx.RawResult, error) {
	return mql.ExecuteCode(ctx.Runtime, bundle, props, Features)
}

func (ctx *tester) TestQueryP(t *testing.T, query string, props map[string]*llx.Primitive) []*llx.RawResult {
	t.Helper()
	bundle, err := mqlc.Compile(query, props, mqlc.NewConfig(ctx.Runtime.Schema(), Features))
	if err != nil {
		t.Fatal("failed to compile code: " + err.Error())
	}
	err = mqlc.Invariants.Check(bundle)
	require.NoError(t, err)
	return ctx.TestMqlc(t, bundle, props)
}

func (ctx *tester) TestQuery(t *testing.T, query string) []*llx.RawResult {
	return ctx.TestQueryP(t, query, nil)
}

func (ctx *tester) TestMqlc(t *testing.T, bundle *llx.CodeBundle, props map[string]*llx.Primitive) []*llx.RawResult {
	t.Helper()

	resultMap, err := mql.ExecuteCode(ctx.Runtime, bundle, props, Features)
	require.NoError(t, err)

	lastQueryResult := &llx.RawResult{}
	results := make([]*llx.RawResult, 0, len(resultMap)+1)

	refs := make([]uint64, 0, len(bundle.CodeV2.Checksums))
	for _, datapointArr := range [][]uint64{bundle.CodeV2.Datapoints(), bundle.CodeV2.Entrypoints()} {
		refs = append(refs, datapointArr...)
	}

	sort.Slice(refs, func(i, j int) bool {
		return refs[i] < refs[j]
	})

	for idx, ref := range refs {
		checksum := bundle.CodeV2.Checksums[ref]
		if d, ok := resultMap[checksum]; ok {
			results = append(results, d)
			if idx+1 == len(refs) {
				lastQueryResult.CodeID = d.CodeID
				if d.Data.Error != nil {
					lastQueryResult.Data = &llx.RawData{
						Error: d.Data.Error,
					}
				} else {
					success, valid := d.Data.IsSuccess()
					lastQueryResult.Data = llx.BoolData(success && valid)
				}
			}
		}
	}

	results = append(results, lastQueryResult)
	return results
}

func mustLoadSchema(provider string) *resources.Schema {
	var path string
	if provider == "mockprovider" {
		path = filepath.Join(TestutilsDir, "mockprovider/resources/mockprovider.lr")
	} else {
		path = filepath.Join(TestutilsDir, "../../../providers/"+provider+"/resources/"+provider+".lr")
	}

	res, err := lr.Resolve(path, func(path string) ([]byte, error) { return os.ReadFile(path) })
	if err != nil {
		panic(err.Error())
	}

	schema, err := lr.Schema(res)
	if err != nil {
		panic(err.Error())
	}

	return schema
}

func Local() llx.Runtime {
	osSchema := mustLoadSchema("os")
	coreSchema := mustLoadSchema("core")
	networkSchema := mustLoadSchema("network")
	mockSchema := mustLoadSchema("mockprovider")

	runtime := providers.Coordinator.NewRuntime()

	provider := &providers.RunningProvider{
		Name:   osconf.Config.Name,
		ID:     osconf.Config.ID,
		Plugin: osprovider.Init(),
		Schema: osSchema.Add(coreSchema),
	}
	runtime.Provider = &providers.ConnectedProvider{Instance: provider}
	runtime.AddConnectedProvider(runtime.Provider)

	provider = &providers.RunningProvider{
		Name:   networkconf.Config.Name,
		ID:     networkconf.Config.ID,
		Plugin: networkprovider.Init(),
		Schema: networkSchema,
	}
	runtime.AddConnectedProvider(&providers.ConnectedProvider{Instance: provider})

	provider = &providers.RunningProvider{
		Name:   mockprovider.Config.Name,
		ID:     mockprovider.Config.ID,
		Plugin: mockprovider.Init(),
		Schema: mockSchema,
	}
	runtime.AddConnectedProvider(&providers.ConnectedProvider{Instance: provider})

	return runtime
}

func mockRuntime(testdata string) llx.Runtime {
	return mockRuntimeAbs(filepath.Join(TestutilsDir, testdata))
}

func mockRuntimeAbs(testdata string) llx.Runtime {
	runtime := Local().(*providers.Runtime)

	abs, _ := filepath.Abs(testdata)
	recording, err := providers.LoadRecordingFile(abs)
	if err != nil {
		panic("failed to load recording: " + err.Error())
	}
	roRecording := recording.ReadOnly()

	err = runtime.SetMockRecording(roRecording, runtime.Provider.Instance.ID, true)
	if err != nil {
		panic("failed to set recording: " + err.Error())
	}
	err = runtime.SetMockRecording(roRecording, networkconf.Config.ID, true)
	if err != nil {
		panic("failed to set recording: " + err.Error())
	}
	err = runtime.SetMockRecording(roRecording, mockprovider.Config.ID, true)
	if err != nil {
		panic("failed to set recording: " + err.Error())
	}

	return runtime
}

func LinuxMock() llx.Runtime {
	return mockRuntime("testdata/arch.json")
}

func KubeletMock() llx.Runtime {
	return mockRuntime("testdata/kubelet.json")
}

func KubeletAKSMock() llx.Runtime {
	return mockRuntime("testdata/kubelet-aks.json")
}

func WindowsMock() llx.Runtime {
	return mockRuntime("testdata/windows.json")
}

func RecordingMock(absTestdataPath string) llx.Runtime {
	return mockRuntimeAbs(absTestdataPath)
}

func CustomMock(path string) llx.Runtime {
	return TomlMock(path)
}

type SimpleTest struct {
	Code        string
	ResultIndex int
	Expectation interface{}
}

func (ctx *tester) TestSimple(t *testing.T, tests []SimpleTest) {
	t.Helper()
	for i := range tests {
		cur := tests[i]
		t.Run(cur.Code, func(t *testing.T) {
			res := ctx.TestQuery(t, cur.Code)
			assert.NotEmpty(t, res)

			if len(res) <= cur.ResultIndex {
				t.Error("insufficient results, looking for result idx " + strconv.Itoa(cur.ResultIndex))
				return
			}

			data := res[cur.ResultIndex].Data
			require.NoError(t, data.Error)
			assert.Equal(t, cur.Expectation, data.Value)
		})
	}
}

func (ctx *tester) TestNoErrorsNonEmpty(t *testing.T, tests []SimpleTest) {
	t.Helper()
	for i := range tests {
		cur := tests[i]
		t.Run(cur.Code, func(t *testing.T) {
			res := ctx.TestQuery(t, cur.Code)
			assert.NotEmpty(t, res)
		})
	}
}

func (ctx *tester) TestSimpleErrors(t *testing.T, tests []SimpleTest) {
	for i := range tests {
		cur := tests[i]
		t.Run(cur.Code, func(t *testing.T) {
			res := ctx.TestQuery(t, cur.Code)
			assert.NotEmpty(t, res)
			assert.Equal(t, cur.Expectation, res[cur.ResultIndex].Result().Error)
			assert.Nil(t, res[cur.ResultIndex].Data.Value)
		})
	}
}

func TestNoResultErrors(t *testing.T, r []*llx.RawResult) bool {
	var found bool
	for i := range r {
		err := r[i].Data.Error
		if err != nil {
			t.Error("result has error: " + err.Error())
			found = true
		}
	}
	return found
}
