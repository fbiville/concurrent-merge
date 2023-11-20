/*
 * Copyright (c) "Neo4j"
 * Neo4j Sweden AB [https://neo4j.com]
 *
 * This file is part of Neo4j.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      https://www.apache.org/licenses/LICENSE-2.0
 *
 *  Unless required by applicable law or agreed to in writing, software
 *  distributed under the License is distributed on an "AS IS" BASIS,
 *  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *  See the License for the specific language governing permissions and
 *  limitations under the License.
 */

package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"log"
	"os"
	"sync"
)

func main() {
	var waitForIndices bool
	flag.BoolVar(&waitForIndices, "wait-for-indices", false, "Wait for indices")
	var goroutineCount int
	flag.IntVar(&goroutineCount, "goroutine-count", 20, "Number of concurrent MERGE statements")
	var neo4jUri string
	flag.StringVar(&neo4jUri, "uri", "", "Neo4j URI")
	var neo4jUser string
	flag.StringVar(&neo4jUser, "user", "neo4j", "Neo4j username")
	var neo4jPassword string
	flag.StringVar(&neo4jPassword, "password", "", "Neo4j password")

	flag.Parse()

	failed := false
	if neo4jUri == "" {
		failed = true
		_, _ = fmt.Fprintln(os.Stderr, "Missing Neo4j URI")
	}
	if neo4jPassword == "" {
		failed = true
		_, _ = fmt.Fprintln(os.Stderr, "Missing Neo4j password")
	}
	if failed {
		log.Fatalf("Usage: %s -uri=<URI> -password=<PASSWORD> [-wait-for-indices] [-goroutine-count=<N>] [-user=<USER>] \n", os.Args[0])
	}

	ctx := context.Background()
	driver, err := neo4j.NewDriverWithContext(neo4jUri, neo4j.BasicAuth(neo4jUser, neo4jPassword, ""))
	oops(err)
	defer oopsClose(ctx, driver)

	log.Println("Creating node key constraint")
	oops(runQuery(ctx, driver, "CREATE CONSTRAINT unique_foobar IF NOT EXISTS FOR (foo:Foo) REQUIRE foo.bar IS NODE KEY", nil))
	if waitForIndices {
		log.Println("Waiting for index population")
		oops(runQuery(ctx, driver, "CALL db.awaitIndexes", nil))
	}

	var wait sync.WaitGroup
	wait.Add(goroutineCount)

	log.Printf("Running %d concurrent merge(s)\n", goroutineCount)
	for i := 0; i < goroutineCount; i++ {
		go func() {
			defer wait.Done()
			oops(runQuery(ctx, driver, "MERGE (foo:Foo {bar: $bar})", map[string]any{"bar": "fighters"}))
		}()
	}

	wait.Wait()
	log.Println("Done!")
}

func runQuery(ctx context.Context, driver neo4j.DriverWithContext, cypher string, params map[string]any) error {
	// avoiding ExecuteQuery on purpose, as I don't want causal chaining
	session := driver.NewSession(ctx, neo4j.SessionConfig{})
	defer oopsClose(ctx, session)
	result, err := session.Run(ctx, cypher, params)
	if err != nil {
		return err
	}
	_, err = result.Consume(ctx)
	return err
}

func oopsClose(ctx context.Context, closeable contextCloseable) {
	oops(closeable.Close(ctx))
}

func oops(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

type contextCloseable interface {
	Close(context.Context) error
}
