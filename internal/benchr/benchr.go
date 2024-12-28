package benchr

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/Masterminds/semver"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/pkg/errors"
	"golang.org/x/mod/modfile"
)

type Benchr struct {
	path             string
	targetDependency string
	versions         []string

	title        string
	benchResults io.Writer
	nsChart      io.WriteCloser
	allocsChart  io.WriteCloser
}

func New(path string, dep string, versions []string) *Benchr {
	return &Benchr{
		path:             path,
		targetDependency: dep,
		versions:         versions,
		title:            dep,

		benchResults: os.Stdout,
	}
}

func (b *Benchr) Run(opts ...Option) error {
	for _, opt := range opts {
		opt(b)
	}

	benchmarkData := make(map[string]map[string]map[string]float64)
	for _, v := range b.versions {
		slog.Info("start benchmark", "ref", v)
		err := b.updateToRef(v)
		if err != nil {
			return err
		}

		rawData, err := b.runOnBaseline("go", "test", "-bench=.", "-benchmem")
		if err != nil {
			return err
		}

		benchmarkData[v] = parseBenchmarkData(rawData)
		_, err = io.Copy(b.benchResults, strings.NewReader(rawData))
		if err != nil {
			return errors.Wrap(err, "failed to copy bench results")
		}
	}

	if b.nsChart != nil && b.allocsChart != nil {
		return b.generateCharts(benchmarkData)
	}
	return nil
}

func parseBenchmarkData(data string) map[string]map[string]float64 {
	results := make(map[string]map[string]float64)
	scanner := bufio.NewScanner(strings.NewReader(data))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "Benchmark") {
			fields := strings.Fields(line)
			if len(fields) >= 6 {
				funcName := fields[0]
				nsPerOp := parseFloat(fields[2])
				allocsPerOp := parseFloat(fields[6])

				if results[funcName] == nil {
					results[funcName] = make(map[string]float64)
				}
				results[funcName]["ns/op"] = nsPerOp
				results[funcName]["allocs/op"] = allocsPerOp
			}
		}
	}
	return results
}

func parseFloat(value string) float64 {
	var result float64
	_, err := fmt.Sscanf(value, "%f", &result)
	if err != nil {
		log.Fatal("cannot parse float from", value, err)
	}
	return result
}

func (b *Benchr) generateCharts(data map[string]map[string]map[string]float64) error {
	var funcNames []string

	vs := make([]*semver.Version, len(data))
	i := 0
	for version := range data {
		v, err := semver.NewVersion(version)
		if err != nil {
			return errors.Wrap(err, "cannot parse version")
		}
		vs[i] = v
		i++
	}
	if len(vs) > 0 {
		for funcName := range data[vs[0].Original()] {
			funcNames = append(funcNames, funcName)
		}
	}

	sort.Sort(semver.Collection(vs))

	err := b.createChart("ns/op", vs, funcNames, data, b.nsChart)
	if err != nil {
		return errors.Wrap(err, "failed to create ns chart")
	}
	err = b.createChart("allocs/op", vs, funcNames, data, b.allocsChart)
	if err != nil {
		return errors.Wrap(err, "failed to create allocs chart")
	}

	return nil
}

func (b *Benchr) createChart(metric string,
	versions []*semver.Version, funcNames []string,
	data map[string]map[string]map[string]float64,
	w io.WriteCloser,
) error {
	lineChart := charts.NewLine()
	lineChart.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{Title: b.title}),
		charts.WithXAxisOpts(opts.XAxis{Name: "Version", Scale: opts.Bool(false)}),
		charts.WithYAxisOpts(opts.YAxis{Name: metric, Scale: opts.Bool(false)}),
	)

	lineChart.SetXAxis(versions)

	for _, funcName := range funcNames {
		var results []opts.LineData
		for _, version := range versions {
			results = append(results, opts.LineData{Value: data[version.Original()][funcName][metric]})
		}
		lineChart.AddSeries(funcName, results)
	}

	defer w.Close()
	return lineChart.Render(w)
}

func (b *Benchr) updateToRef(version string) error {
	goModPath := filepath.Join(b.path, "go.mod")
	data, err := os.ReadFile(goModPath)
	if err != nil {
		return errors.Wrap(err, "failed to read go.mod")
	}

	modFile, err := modfile.Parse(goModPath, data, nil)
	if err != nil {
		return errors.Wrap(err, "failed to parse go.mod")
	}

	err = modFile.AddRequire(b.targetDependency, version)
	if err != nil {
		return errors.Wrap(err, "failed to add replace directive")
	}

	newGoModData, err := modFile.Format()
	if err != nil {
		return errors.Wrap(err, "failed to format go.mod")
	}

	err = os.WriteFile(goModPath, newGoModData, 0o600)
	if err != nil {
		return errors.Wrap(err, "failed to save changes to go.mod")
	}

	_, err = b.runOnBaseline("go", "mod", "tidy")
	return err
}

func (b *Benchr) runOnBaseline(exe string, args ...string) (string, error) {
	cmd := exec.Command(exe, args...)
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	cmd.Dir = filepath.Join(wd, b.path)
	out, err := cmd.CombinedOutput()
	return string(out), err
}
