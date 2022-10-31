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

package transmute

import (
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/ptrace"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/sdk/instrumentation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	api "go.opentelemetry.io/otel/trace"
)

// Spans converts s to pdata Traces.
func Spans(s []trace.ReadOnlySpan) ptrace.Traces {
	t := ptrace.NewTraces()
	rMap := mapSpans(s)

	rs := t.ResourceSpans()
	rs.EnsureCapacity(len(rMap))
	for res, sMap := range rMap {
		r := rs.AppendEmpty()
		r.SetSchemaUrl(res.SchemaURL())
		setAttrMapIter(r.Resource().Attributes(), res.Iter())
		setScopeSpans(r.ScopeSpans(), sMap)
	}
	return t
}

type scopeMap map[instrumentation.Scope][]trace.ReadOnlySpan

type resMap map[resource.Resource]scopeMap

func mapSpans(spans []trace.ReadOnlySpan) resMap {
	if len(spans) == 0 {
		return nil
	}

	rMap := make(resMap)
	for _, s := range spans {
		sMap := rMap[*s.Resource()]
		if sMap == nil {
			sMap = make(scopeMap)
		}
		roSpans := sMap[s.InstrumentationScope()]
		roSpans = append(roSpans, s)
		sMap[s.InstrumentationScope()] = roSpans
		rMap[*s.Resource()] = sMap
	}
	return rMap
}

func setAttrMapIter(p pcommon.Map, o attribute.Iterator) {
	p.EnsureCapacity(o.Len())
	for o.Next() {
		a := o.Attribute()
		setAttribute(p, a)
	}
}

func setAttrMapSlice(p pcommon.Map, o []attribute.KeyValue) {
	p.EnsureCapacity(len(o))
	for _, a := range o {
		setAttribute(p, a)
	}
}

func setAttribute(p pcommon.Map, a attribute.KeyValue) {
	switch a.Value.Type() {
	case attribute.BOOL:
		p.PutBool(string(a.Key), a.Value.AsBool())
	case attribute.INT64:
		p.PutInt(string(a.Key), a.Value.AsInt64())
	case attribute.FLOAT64:
		p.PutDouble(string(a.Key), a.Value.AsFloat64())
	case attribute.STRING:
		p.PutStr(string(a.Key), a.Value.AsString())
	case attribute.BOOLSLICE:
		s := p.PutEmptySlice(string(a.Key))
		vSlice := a.Value.AsBoolSlice()
		s.EnsureCapacity(len(vSlice))
		for _, v := range vSlice {
			s.AppendEmpty().SetBool(v)
		}
	case attribute.INT64SLICE:
		s := p.PutEmptySlice(string(a.Key))
		vSlice := a.Value.AsInt64Slice()
		s.EnsureCapacity(len(vSlice))
		for _, v := range vSlice {
			s.AppendEmpty().SetInt(v)
		}
	case attribute.FLOAT64SLICE:
		s := p.PutEmptySlice(string(a.Key))
		vSlice := a.Value.AsFloat64Slice()
		s.EnsureCapacity(len(vSlice))
		for _, v := range vSlice {
			s.AppendEmpty().SetDouble(v)
		}
	case attribute.STRINGSLICE:
		s := p.PutEmptySlice(string(a.Key))
		vSlice := a.Value.AsStringSlice()
		s.EnsureCapacity(len(vSlice))
		for _, v := range vSlice {
			s.AppendEmpty().SetStr(v)
		}
	default:
		// drop unknown.
	}
}

func setScopeSpans(p ptrace.ScopeSpansSlice, o scopeMap) {
	p.EnsureCapacity(len(o))
	for scope, spans := range o {
		scopeSpans := p.AppendEmpty()
		setScope(scopeSpans.Scope(), scope)
		setSpans(scopeSpans.Spans(), spans)
	}
}

func setScope(p pcommon.InstrumentationScope, o instrumentation.Scope) {
	p.SetName(o.Name)
	p.SetVersion(o.Version)
	// TODO: support scope attributes when added to o.
}

func setSpans(p ptrace.SpanSlice, o []trace.ReadOnlySpan) {
	p.EnsureCapacity(len(o))
	for _, s := range o {
		setSpan(p.AppendEmpty(), s)
	}
}

func setSpan(p ptrace.Span, o trace.ReadOnlySpan) {
	p.SetName(o.Name())
	p.SetTraceID(pcommon.TraceID(o.SpanContext().TraceID()))
	p.SetSpanID(pcommon.SpanID(o.SpanContext().SpanID()))
	p.TraceState().FromRaw(o.SpanContext().TraceState().String())
	p.SetParentSpanID(pcommon.SpanID(o.Parent().SpanID()))
	p.SetKind(spanKind(o.SpanKind()))
	p.SetStartTimestamp(pcommon.NewTimestampFromTime(o.StartTime()))
	p.SetEndTimestamp(pcommon.NewTimestampFromTime(o.EndTime()))
	setAttrMapSlice(p.Attributes(), o.Attributes())
	setLinks(p.Links(), o.Links())
	setEvents(p.Events(), o.Events())
	setStatus(p.Status(), o.Status())
	p.SetDroppedAttributesCount(uint32(o.DroppedAttributes()))
	p.SetDroppedLinksCount(uint32(o.DroppedLinks()))
	p.SetDroppedEventsCount(uint32(o.DroppedEvents()))
}

func spanKind(o api.SpanKind) ptrace.SpanKind {
	switch o {
	case api.SpanKindInternal:
		return ptrace.SpanKindInternal
	case api.SpanKindServer:
		return ptrace.SpanKindServer
	case api.SpanKindClient:
		return ptrace.SpanKindClient
	case api.SpanKindProducer:
		return ptrace.SpanKindProducer
	case api.SpanKindConsumer:
		return ptrace.SpanKindConsumer
	}
	return ptrace.SpanKindUnspecified
}

func setLinks(p ptrace.SpanLinkSlice, o []trace.Link) {
	p.EnsureCapacity(len(o))
	for _, ol := range o {
		pl := p.AppendEmpty()
		pl.SetTraceID(pcommon.TraceID(ol.SpanContext.TraceID()))
		pl.SetSpanID(pcommon.SpanID(ol.SpanContext.SpanID()))
		pl.TraceState().FromRaw(ol.SpanContext.TraceState().String())
		setAttrMapSlice(pl.Attributes(), ol.Attributes)
		pl.SetDroppedAttributesCount(uint32(ol.DroppedAttributeCount))
	}
}

func setEvents(p ptrace.SpanEventSlice, o []trace.Event) {
	p.EnsureCapacity(len(o))
	for _, oe := range o {
		pe := p.AppendEmpty()
		pe.SetName(oe.Name)
		pe.SetTimestamp(pcommon.NewTimestampFromTime(oe.Time))
		setAttrMapSlice(pe.Attributes(), oe.Attributes)
		pe.SetDroppedAttributesCount(uint32(oe.DroppedAttributeCount))
	}
}

func setStatus(p ptrace.Status, o trace.Status) {
	p.SetMessage(o.Description)
	switch o.Code {
	case codes.Ok:
		p.SetCode(ptrace.StatusCodeOk)
	case codes.Error:
		p.SetCode(ptrace.StatusCodeError)
	default:
		p.SetCode(ptrace.StatusCodeUnset)
	}
}
