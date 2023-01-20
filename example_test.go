// Copyright 2022 Tyler Yahn (MrAlias)
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package collex_test

import (
	"context"
	"log"

	"github.com/MrAlias/collex"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/exporter"
	"go.opentelemetry.io/collector/exporter/loggingexporter"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/sdk/trace"
	"go.uber.org/zap"
)

func Example() {
	settings := &exporter.CreateSettings{
		TelemetrySettings: component.TelemetrySettings{
			Logger:         zap.NewExample(), // Log to STDOUT for example.
			TracerProvider: otel.GetTracerProvider(),
		},
	}
	factory, err := collex.NewFactory(loggingexporter.NewFactory(), settings)
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()
	exp, err := factory.SpanExporter(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}

	provider := trace.NewTracerProvider(trace.WithSyncer(exp))
	tracer := provider.Tracer("github.com/MrAlias/collex")
	_, s := tracer.Start(ctx, "ExampleTracesExporter")
	s.End()
	if err := provider.Shutdown(ctx); err != nil {
		log.Fatal(err)
	}

	// Output: {"level":"info","msg":"TracesExporter","#spans":1}
}
