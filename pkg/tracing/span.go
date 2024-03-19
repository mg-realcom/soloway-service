package tracing

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

func CreateSpan(ctx context.Context, tracerName, spanName string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	newCtx, span := otel.Tracer(tracerName).Start(ctx, spanName, opts...)
	return newCtx, span
}

func EndSpanError(span trace.Span, err error, description string, withEnd bool) {
	span.RecordError(err)
	span.SetStatus(codes.Error, description)
	if withEnd {
		span.End()
	}
}

func EndSpanOk(span trace.Span, description string, withEnd bool) {
	span.SetStatus(codes.Ok, description)
	if withEnd {
		span.End()
	}
}

func SpanFromContext(ctx context.Context) trace.Span {
	return trace.SpanFromContext(ctx)
}

func SetSpanAttribute(span trace.Span, key string, value any) {
	attr := attribute.KeyValue{Key: attribute.Key(key)}
	switch val := value.(type) {
	case string:
		attr.Value = attribute.StringValue(val)
	case []string:
		attr.Value = attribute.StringSliceValue(val)
	case int8:
		attr.Value = attribute.Int64Value(int64(val))
	case int16:
		attr.Value = attribute.Int64Value(int64(val))
	case int32:
		attr.Value = attribute.Int64Value(int64(val))
	case int:
		attr.Value = attribute.IntValue(val)
	case int64:
		attr.Value = attribute.Int64Value(val)
	case float32:
		attr.Value = attribute.Float64Value(float64(val))
	case float64:
		attr.Value = attribute.Float64Value(val)
	case bool:
		attr.Value = attribute.BoolValue(val)
	case []bool:
		attr.Value = attribute.BoolSliceValue(val)
	case byte:
		attr.Value = attribute.Int64Value(int64(val))
	case []byte:
		attr.Value = attribute.StringValue(string(val))
	case []int8:
		int64Slice := make([]int64, len(val), cap(val))
		for i, b := range val {
			int64Slice[i] = int64(b)
		}
		attr.Value = attribute.Int64SliceValue(int64Slice)
	case []int16:
		int64Slice := make([]int64, len(val), cap(val))
		for i, b := range val {
			int64Slice[i] = int64(b)
		}
		attr.Value = attribute.Int64SliceValue(int64Slice)
	case []int32:
		int64Slice := make([]int64, len(val), cap(val))
		for i, b := range val {
			int64Slice[i] = int64(b)
		}
		attr.Value = attribute.Int64SliceValue(int64Slice)
	case []int64:
		attr.Value = attribute.Int64SliceValue(val)
	case []int:
		attr.Value = attribute.IntSliceValue(val)
	case []float32:
		float64Slice := make([]float64, len(val), cap(val))
		for i, b := range val {
			float64Slice[i] = float64(b)
		}
		attr.Value = attribute.Float64SliceValue(float64Slice)
	case []float64:
		attr.Value = attribute.Float64SliceValue(val)
	default:
		jsonByte, _ := json.Marshal(value)
		attr.Value = attribute.StringValue(string(jsonByte))
	}
	span.SetAttributes(attr)
}

func SetSpanAttributesByStructFields(span trace.Span, attributes any, preffix string) {
	jsonMap := make(map[string]any)
	fields, _ := json.Marshal(attributes)
	_ = json.Unmarshal(fields, &jsonMap)
	writeFields(span, jsonMap, preffix)
}

func writeFields(span trace.Span, source any, prefix string) {
	sourceMap, okSourceMap := source.(map[string]any)
	if okSourceMap {
		for key, val := range sourceMap {
			_, okMap := val.(map[string]any)
			_, okArray := val.([]any)
			keyPrefix := prefix

			if len(sourceMap) != 1 || !okMap {
				if prefix != "" {
					keyPrefix = fmt.Sprintf("%s.%s", prefix, key)
				} else {
					keyPrefix = key
				}
			}

			if okMap || okArray {
				writeFields(span, val, keyPrefix)
			} else {
				SetSpanAttribute(span, keyPrefix, fmt.Sprintf("%v", val))
			}
		}
	}
	sourceArray, okSourceArray := source.([]any)
	if okSourceArray {
		for index, val := range sourceArray {
			key := strconv.Itoa(index)
			_, okMap := val.(map[string]any)
			_, okArray := val.([]any)
			keyPrefix := prefix

			if len(sourceArray) != 1 || !okMap {
				if prefix != "" {
					keyPrefix = fmt.Sprintf("%s.%s", prefix, key)
				} else {
					keyPrefix = key
				}
			}

			if okMap || okArray {
				writeFields(span, val, keyPrefix)
			} else {
				SetSpanAttribute(span, keyPrefix, fmt.Sprintf("%v", val))
			}
		}
	}
}
