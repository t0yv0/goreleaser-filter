package main

import (
	"flag"
	"log"
	"os"

	"gopkg.in/yaml.v2"
)

type filters struct {
	goos   string
	goarch string
}

func isMatchingBuild(filters filters, build *value) bool {
	if filters.goarch != "" {
		if !build.at("goarch").containsString(filters.goarch) {
			return false
		}
	}
	if filters.goos != "" {
		if !build.at("goos").containsString(filters.goos) {
			return false
		}
	}
	return true
}

func cleanupBuild(filters filters, build *value) {
	if filters.goarch != "" {
		build.at("goarch").set(arrayValue(toValue(filters.goarch)))
	}
	if filters.goos != "" {
		build.at("goos").set(arrayValue(toValue(filters.goos)))
	}
}

func isMatchingArchive(buildIds *value, archive *value) bool {
	for _, reqBuild := range archive.at("builds").elements() {
		if rB, ok := reqBuild.toInterface().(string); ok {
			if !buildIds.containsString(rB) {
				return false
			}
		}
	}
	return true
}

func buildIds(config *value) *value {
	var ids []*value
	for _, build := range config.at("builds").elements() {
		if id, ok := build.at("id").toInterface().(string); ok {
			ids = append(ids, toValue(id))
		}
	}
	return arrayValue(ids...)
}

func cleanupArchives(config *value) {
	matchingBuilds := buildIds(config)
	var matches []*value
	for _, archive := range config.at("archives").elements() {
		if isMatchingArchive(matchingBuilds, archive) {
			matches = append(matches, archive)
		}
	}
	config.at("archives").set(arrayValue(matches...))
}

func runYamlTransformer(main func(*value)) {
	var rawConfig interface{}
	if err := yaml.NewDecoder(os.Stdin).Decode(&rawConfig); err != nil {
		log.Fatalf("error: %v", err)
	}
	config := toValue(rawConfig)
	main(config)
	enc := yaml.NewEncoder(os.Stdout)
	defer func() {
		if err := enc.Close(); err != nil {
			log.Fatalf("error: %v", err)
		}
	}()
	if err := enc.Encode(config.toInterface()); err != nil {
		log.Fatalf("error: %v", err)
	}
}

func yamlTransformer(filters filters, config *value) {
	var matchingBuilds []*value
	for _, build := range config.at("builds").elements() {
		if isMatchingBuild(filters, build) {
			cleanupBuild(filters, build)
			matchingBuilds = append(matchingBuilds, build)
		}
	}
	config.at("builds").set(arrayValue(matchingBuilds...))
	cleanupArchives(config)
}

func main() {
	goos := flag.String("goos", "", "goos filter for builds")
	goarch := flag.String("goarch", "", "goarch filter for builds")

	flag.Parse()

	filters := filters{goos: *goos, goarch: *goarch}
	runYamlTransformer(func(config *value) {
		yamlTransformer(filters, config)
	})
}

// values

type value struct {
	inner    interface{}
	setInner func(interface{})
}

func toValue(x interface{}) *value {
	return &value{
		inner: x,
		setInner: func(interface{}) {
			panic("attempt to set read-only value")
		},
	}
}

func arrayValue(values ...*value) *value {
	x := []interface{}{}
	for _, v := range values {
		x = append(x, v.toInterface())
	}
	return toValue(x)
}

func (v *value) set(x *value) {
	v.setInner(x.inner)
}

func (v *value) toInterface() interface{} {
	return v.inner
}

func (v *value) at(key string) *value {
	if v == nil {
		return toValue(nil)
	}
	m, ok := v.toInterface().(map[interface{}]interface{})
	if !ok {
		return toValue(nil)
	}
	x, ok := m[key]
	if !ok {
		return toValue(nil)
	}
	return &value{
		inner: x,
		setInner: func(v interface{}) {
			m[key] = v
		},
	}
}

func (v *value) elements() []*value {
	var result []*value
	if slice, ok := v.toInterface().([]interface{}); ok {
		for i, s := range slice {
			result = append(result, &value{
				inner: s,
				setInner: func(v interface{}) {
					slice[i] = v
				},
			})
		}
	}
	return result
}

func (v *value) containsString(x string) bool {
	if v == nil {
		return false
	}
	elements, ok := v.toInterface().([]interface{})
	if !ok {
		return false
	}
	for _, e := range elements {
		if es, ok := e.(string); ok {
			if es == x {
				return true
			}
		}
	}
	return false
}
