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

package collex

import (
	"context"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config/configtelemetry"
	"go.opentelemetry.io/collector/exporter"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/noop"
	"go.opentelemetry.io/otel/sdk/trace"
	"go.uber.org/zap"
)

// Factory wraps an OpenTelemetry collector ExporterFactory and initializes new
// OpenTelemetry Go exporters from it.
type Factory struct {
	createCfg   exporter.Settings
	collFactory exporter.Factory
}

// NewFactory returns a new configured *Factory. If set is nil, a default
// Settings will be used. These settings use a production ready Zap logger and
// a global OpenTelemetry Go TracerProvider.
func NewFactory(f exporter.Factory, set *exporter.Settings) (*Factory, error) {
	if set == nil {
		logger, err := zap.NewProduction()
		if err != nil {
			return nil, err
		}

		set = &exporter.Settings{
			TelemetrySettings: component.TelemetrySettings{
				Logger:         logger,
				TracerProvider: otel.GetTracerProvider(),
				LeveledMeterProvider: func(configtelemetry.Level) metric.MeterProvider {
					return noop.NewMeterProvider()
				},
			},
			BuildInfo: component.BuildInfo{
				Command:     "collex",
				Description: "OpenTelemetry Collector to OpenTelemetry Go translator",
				Version:     "latest",
			},
		}
	}
	return &Factory{*set, f}, nil
}

// SpanExporter returns an OpenTelemetry Go SpanExporter that can be registered
// with a TracerProvider. If cfg is nil the factory default configuration for
// the ExporterFactory is used.
func (f *Factory) SpanExporter(ctx context.Context, cfg component.Config) (trace.SpanExporter, error) {
	if cfg == nil {
		cfg = f.collFactory.CreateDefaultConfig()
	}
	collExp, err := f.collFactory.CreateTracesExporter(ctx, f.createCfg, cfg)
	if err != nil {
		return nil, err
	}
	return &spanExporter{cexp: collExp}, nil
}
