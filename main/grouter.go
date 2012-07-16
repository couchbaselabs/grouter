package main

import (
	"flag"
	"log"
	"strings"

	"github.com/steveyen/grouter"
)

// Available sources of requests.
var SourceFuncs = map[string]func(string, int, chan []grouter.Request){
	"memcached":       grouter.NetListenSourceFunc(&grouter.AsciiSource{}),
	"memcached-ascii": grouter.NetListenSourceFunc(&grouter.AsciiSource{}),
	"workload":        grouter.WorkLoad,
}

// Available targets of requests.
var TargetFuncs = map[string]func(string, chan []grouter.Request){
	"http":             grouter.CouchbaseTargetRun,
	"couchbase":        grouter.CouchbaseTargetRun,
	"memcached-ascii":  grouter.MemcachedAsciiTargetRun,
	"memcached-binary": grouter.MemcachedBinaryTargetRun,
	"memory":           grouter.MemoryStorageRun,
}

func MainStart(sourceSpec string, sourceMaxConns int,
	targetSpec string, targetChanSize int) {
	log.Printf("grouter")
	log.Printf("  source: %v", sourceSpec)
	log.Printf("    sourceMaxConns: %v", sourceMaxConns)
	log.Printf("  target: %v", targetSpec)
	log.Printf("    targetChanSize: %v", targetChanSize)

	sourceKind := strings.Split(sourceSpec, ":")[0]
	if sourceFunc, ok := SourceFuncs[sourceKind]; ok {
		targetKind := strings.Split(targetSpec, ":")[0]
		if targetFunc, ok := TargetFuncs[targetKind]; ok {
			targetChan := make(chan []grouter.Request, targetChanSize)
			go func() {
				targetFunc(targetSpec, targetChan)
			}()
			sourceFunc(sourceSpec, sourceMaxConns, targetChan)
		} else {
			log.Fatalf("error: unknown target kind: %s", targetSpec)
		}
	} else {
		log.Fatalf("error: unknown source kind: %s", sourceSpec)
	}
}

func main() {
	sourceSpec := flag.String("source", "memcached-ascii::11300",
		"source of requests\n" +
		"    which should follow a format of KIND[:PARAMS] like...\n" +
		"      memcached-ascii:LISTEN_INTERFACE:LISTEN_PORT\n" +
		"      workload")
	sourceMaxConns := flag.Int("source-max-conns", 3,
		"max conns allowed from source")

	targetSpec := flag.String("target", "memory",
		"target of requests\n" +
		"    which should follow a format of KIND[:PARAMS] like...\n" +
		"      http:\\\\COUCHBASE_HOST:COUCHBASE_PORT\n" +
		"      couchbase:\\\\COUCHBASE_HOST:COUCHBASE_PORT\n" +
		"      memcached:HOST:PORT\n" +
		"      memory")
	targetChanSize := flag.Int("target-chan-size", 5,
		"target chan size to control concurrency")

	flag.Parse()
	MainStart(*sourceSpec, *sourceMaxConns, *targetSpec, *targetChanSize)
}
