#include <stdint.h>
#include <stddef.h>
#include <stdbool.h>

#define MAX_SPANS 100
#define MAX_FILE_SIZE 1024 * 1024 * 50 // 50MB

struct AnyValue;
struct ArrayValue;
struct KeyValueList;
struct KeyValue;
struct InstrumentationScope;

typedef struct {
    int32_t attributes_offset;
} Resource;

typedef struct {
    int64_t name_offset;
    int32_t total_recorded_attributes;
    int32_t attributes_offset; // points to an array of KeyValue
    int32_t status_offset;
    int32_t parent_span_id_offset;
    int32_t kind;
    int32_t start_timestamp_offset;
    int32_t end_timestamp_offset;
    Resource resource;
    int32_t instrumentation_scope_offset;
    int32_t span_id_offset;
    int32_t trace_id_offset;
    int32_t tracestate_offset;
    int32_t trace_flags;
} Span;

typedef struct {
    int32_t code;
    int32_t description_offset;
} Status;


typedef struct AnyValue {
    enum {
        ANYVALUE_BOOL,
        ANYVALUE_INT,
        ANYVALUE_STRING,
        ANYVALUE_DOUBLE,
        ANYVALUE_ARRAY,
        ANYVALUE_KVLIST,
        ANYVALUE_BYTES,
        ANYVALUE_NONE
    } value_type;
    union {
        size_t string_value_offset;
        // char *string_value;
        bool bool_value;
        int64_t int_value;
        double double_value;
        struct ArrayValue *array_value;
        struct KeyValueList *kvlist_value;
        struct {
            uint8_t *data;
            size_t len;
        } bytes_value;
    } value;
} AnyValue;

// Define the ArrayValue struct
typedef struct ArrayValue {
    AnyValue *values;
    size_t values_count;
} ArrayValue;

// Define the KeyValue struct
typedef struct KeyValue {
    // char *key;
    long key_offset;
    size_t value_offset;
    // AnyValue value;
} KeyValue;

// Define the KeyValueList struct
typedef struct KeyValueList {
    KeyValue *values;
    size_t values_count;
} KeyValueList;